package trace

import (
	"context"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"core/configuration"
)

type testExporter struct {
	exportingFinished   chan bool
	iterationIntervalMs int
	numIterations       int
}

func (e *testExporter) export(ctx context.Context, _ dtSpanSet) (err error) {
	for i := e.numIterations; i > 0; i-- {
		// simulate exporting operation
		time.Sleep(time.Millisecond * time.Duration(e.iterationIntervalMs))

		if err = ctx.Err(); err != nil {
			break
		}
	}

	e.exportingFinished <- true
	return
}

type spanGeneratorOption struct {
	numSpans   int
	endedSpans bool
}

func generateSpans(tr trace.Tracer, option spanGeneratorOption) {
	wg := &sync.WaitGroup{}
	for i := 0; i < option.numSpans; i++ {
		wg.Add(1)
		go func(idx int) {
			_, span := tr.Start(context.Background(), fmt.Sprintf("Span #%d", idx))
			endTimestamp := trace.WithTimestamp(time.Now().Add(750 * time.Millisecond))
			if option.endedSpans {
				span.End(endTimestamp)
			}
			wg.Done()
		}(i)
	}

	wg.Wait()
}

func createSpanProcessor() *dtSpanProcessor {
	return newDtSpanProcessor(&configuration.DtConfiguration{})
}

func createTracerProvider() *DtTracerProvider {
	return NewTracerProvider(&configuration.DtConfiguration{})
}

func TestDtSpanProcessorStartSpans(t *testing.T) {
	tp := createTracerProvider()
	require.Zero(t, tp.processor.exportingStopped)
	require.Zero(t, tp.processor.spanWatchlist.len())

	otel.SetTracerProvider(tp)
	tr := otel.Tracer("Dynatrace Tracer")

	generateSpans(tr, spanGeneratorOption{
		numSpans:   20,
		endedSpans: true,
	})
	require.EqualValues(t, tp.processor.spanWatchlist.len(), 20)

	generateSpans(tr, spanGeneratorOption{
		numSpans:   5,
		endedSpans: true,
	})
	require.EqualValues(t, tp.processor.spanWatchlist.len(), 25)
}

func TestDtSpanProcessorShutdown(t *testing.T) {
	p := createSpanProcessor()
	require.Zero(t, p.exportingStopped)
	require.Zero(t, p.spanWatchlist.len())

	err := p.shutdown(context.Background())
	require.NoError(t, err, "shutdown has failed")
	require.NotZero(t, p.exportingStopped, "Exporting goroutine has not been stopped")
}

func TestDtSpanProcessorGenerateSpansAndShutdown(t *testing.T) {
	tp := createTracerProvider()
	require.Zero(t, tp.processor.exportingStopped)
	require.Zero(t, tp.processor.spanWatchlist.len())

	otel.SetTracerProvider(tp)
	tr := otel.Tracer("Dynatrace Tracer")

	generateSpans(tr, spanGeneratorOption{
		numSpans:   20,
		endedSpans: true,
	})
	require.EqualValues(t, tp.processor.spanWatchlist.len(), 20)

	err := tp.Shutdown(context.Background())
	require.NoError(t, err, "shutdown has failed")
	require.NotZero(t, tp.processor.exportingStopped, "Exporting goroutine has not been stopped")
	require.Zero(t, tp.processor.spanWatchlist.len(), "Spans have not been flushed on shutdown")
}

func TestDtSpanProcessorShutdownCancelContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	p := createSpanProcessor()
	require.Zero(t, p.exportingStopped)

	err := p.shutdown(ctx)
	require.ErrorIs(t, err, context.Canceled)
}

func TestDtSpanProcessorShutdownTimeoutReached(t *testing.T) {
	exporter := &testExporter{
		exportingFinished:   make(chan bool, 1),
		iterationIntervalMs: 500,
		numIterations:       20,
	}

	p := createSpanProcessor()
	p.exporter = exporter
	require.Zero(t, p.exportingStopped)

	err := p.shutdown(context.Background())
	require.ErrorIs(t, err, context.Canceled)

	// check whether the exporting operation was aborted due to the reached flush timeout
	select {
	case <-exporter.exportingFinished:
		log.Println("Long running export operation has been aborted")
	// the exporter must finish exporting within the next interval due to canceled context caused by the flush timeout
	case <-time.After(time.Millisecond * time.Duration(exporter.iterationIntervalMs*2)):
		require.Fail(t, "Exporter operation has not been aborted after the flush timeout reached")
	}
}

func TestDtSpanProcessorForceFlushNonEndedSpan(t *testing.T) {
	tp := createTracerProvider()
	require.Zero(t, tp.processor.exportingStopped)
	require.Zero(t, tp.processor.spanWatchlist.len())

	otel.SetTracerProvider(tp)
	tr := otel.Tracer("Dynatrace Tracer")

	generateSpans(tr, spanGeneratorOption{
		numSpans:   20,
		endedSpans: true,
	})
	generateSpans(tr, spanGeneratorOption{
		numSpans:   5,
		endedSpans: false,
	})
	require.EqualValues(t, tp.processor.spanWatchlist.len(), 25)

	// Non-ended spans still must be watched
	tp.ForceFlush(context.Background())
	require.EqualValues(t, tp.processor.spanWatchlist.len(), 5)
}

func TestDtSpanProcessorForceFlushCancelContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	exporter := &testExporter{
		exportingFinished:   make(chan bool, 1),
		iterationIntervalMs: 500,
		numIterations:       20,
	}

	p := createSpanProcessor()
	p.exporter = exporter
	require.Zero(t, p.exportingStopped)

	err := p.forceFlush(ctx)
	require.ErrorIs(t, err, context.Canceled)

	// check whether the exporting operation was aborted due to canceled force flush context
	select {
	case <-exporter.exportingFinished:
		log.Println("Long running export operation has been aborted")
	// the exporter must finish exporting operation within the next interval due to canceled flush context
	case <-time.After(time.Millisecond * time.Duration(exporter.iterationIntervalMs*2)):
		require.Fail(t, "The export operation was not aborted prior to reaching the flush timeout")
	}
}

func TestDtSpanProcessorForceFlushTimeoutReached(t *testing.T) {
	exporter := &testExporter{
		exportingFinished:   make(chan bool, 1),
		iterationIntervalMs: 500,
		numIterations:       20,
	}

	p := createSpanProcessor()
	p.exporter = exporter
	require.Zero(t, p.exportingStopped)

	err := p.forceFlush(context.Background())
	require.ErrorIs(t, err, context.Canceled)

	// check whether the exporting operation was aborted due to the reached flush timeout
	select {
	case <-exporter.exportingFinished:
		log.Println("Long running export operation has been aborted")
	// the exporter must finish exporting within the next interval due to canceled context caused by the flush timeout
	case <-time.After(time.Millisecond * time.Duration(exporter.iterationIntervalMs*2)):
		require.Fail(t, "The export operation was not aborted prior to reaching the flush timeout")
	}
}

func TestDtSpanProcessorWaitForScheduledFlushOperation(t *testing.T) {
	exporter := &testExporter{
		exportingFinished:   make(chan bool, 5),
		iterationIntervalMs: 500,
		numIterations:       4,
	}

	tp := createTracerProvider()
	tp.processor.exporter = exporter
	require.Zero(t, tp.processor.exportingStopped)
	require.Zero(t, tp.processor.spanWatchlist.len())

	otel.SetTracerProvider(tp)
	tr := otel.Tracer("Dynatrace Tracer")

	generateSpans(tr, spanGeneratorOption{
		numSpans:   5,
		endedSpans: true,
	})

	var wg sync.WaitGroup

	go func() {
		wg.Add(1)
		defer wg.Done()

		tp.ForceFlush(context.Background())
	}()

	// 1st flush operation will be started ASAP, thus the flush channel must be empty
	// all spans are ended, so after evaluation the spans watchlist size must be 0
	require.Eventually(t, func() bool {
		return len(tp.processor.flushRequestCh) == 0 && tp.processor.spanWatchlist.len() == 0
	}, 3*time.Second, 100*time.Millisecond)

	go func() {
		wg.Add(1)
		defer wg.Done()

		generateSpans(tr, spanGeneratorOption{
			numSpans:   10,
			endedSpans: true,
		})
		tp.ForceFlush(context.Background())
	}()

	require.Eventually(t, func() bool {
		// the 1st flush is running, so the 2nd one must be scheduled, as a result, the flush channel must not be empty
		// 10 spans were started after the 1st flush call, thus span watchlist size must be equal to 10
		return len(tp.processor.flushRequestCh) == 1 && tp.processor.spanWatchlist.len() == 10
	}, 3*time.Second, 100*time.Millisecond)

	go func() {
		wg.Add(1)
		defer wg.Done()

		generateSpans(tr, spanGeneratorOption{
			numSpans:   5,
			endedSpans: true,
		})

		tp.ForceFlush(context.Background())

		// the 3rd flush call have to wait until the 2nd pending flush will be finished
		// and export spans started upon 3rd flush call
		require.Equal(t, tp.processor.spanWatchlist.len(), 0)
	}()

	wg.Wait()
}
