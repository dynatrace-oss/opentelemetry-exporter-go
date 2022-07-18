package export

import (
	"context"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"core/configuration"
)

type testExporter struct {
	exportingFinished   chan bool // TODO
	iterationIntervalMs int
	numIterations       int
}

func (e *testExporter) export(ctx context.Context, _ map[spanKey]*dtSpanMetadata) (err error) {
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

func createSpanProcessor() *DtSpanProcessor {
	return NewDtSpanProcessor(&configuration.DtConfiguration{})
}

func TestDtSpanProcessorStartSpans(t *testing.T) {
	p := createSpanProcessor()
	require.Zero(t, p.exportingStopped)
	require.Zero(t, len(p.spanWatchlist.spans))

	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(p))
	otel.SetTracerProvider(tp)
	tr := otel.Tracer("SDK Tracer")

	generateSpans(tr, spanGeneratorOption{
		numSpans:   20,
		endedSpans: true,
	})
	require.EqualValues(t, len(p.spanWatchlist.spans), 20)

	generateSpans(tr, spanGeneratorOption{
		numSpans:   5,
		endedSpans: true,
	})
	require.EqualValues(t, len(p.spanWatchlist.spans), 25)
}

func TestDtSpanProcessorShutdown(t *testing.T) {
	p := createSpanProcessor()
	require.Zero(t, p.exportingStopped)

	err := p.Shutdown(context.Background())
	require.NoError(t, err, "Shutdown has failed")
	require.NotZero(t, p.exportingStopped, "Exporting goroutine has not been stopped")
}

func TestDtSpanProcessorPostShutdown(t *testing.T) {
	p := createSpanProcessor()
	require.Zero(t, p.exportingStopped)
	require.Zero(t, len(p.spanWatchlist.spans))

	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(p))
	otel.SetTracerProvider(tp)
	tr := otel.Tracer("SDK Tracer")

	generateSpans(tr, spanGeneratorOption{
		numSpans:   20,
		endedSpans: true,
	})
	require.EqualValues(t, len(p.spanWatchlist.spans), 20)

	err := p.Shutdown(context.Background())
	require.NoError(t, err, "Shutdown has failed")
	require.NotZero(t, p.exportingStopped, "Exporting goroutine has not been stopped")
	require.Zero(t, len(p.spanWatchlist.spans), "Spans have not been flushed on shutdown")
}

func TestDtSpanProcessorShutdownCancelContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	p := createSpanProcessor()
	require.Zero(t, p.exportingStopped)

	err := p.Shutdown(ctx)
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

	err := p.Shutdown(context.Background())
	require.ErrorIs(t, err, context.Canceled)

	// check whether the exporting operation was aborted due to the reached flush timeout
	select {
	case <-exporter.exportingFinished:
		log.Println("Exporter operation has been aborted")
	// the exporter must finish exporting operation within the next interval
	// due to canceled context caused by the flush timeout
	case <-time.After(time.Millisecond * time.Duration(exporter.iterationIntervalMs*2)):
		require.Fail(t, "Exporter operation has not been aborted after the flush timeout reached")
	}
}

func TestDtSpanProcessorForceFlush(t *testing.T) {
	p := createSpanProcessor()
	require.Zero(t, p.exportingStopped)
	require.Zero(t, len(p.spanWatchlist.spans))

	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(p))
	otel.SetTracerProvider(tp)
	tr := otel.Tracer("SDK Tracer")

	generateSpans(tr, spanGeneratorOption{
		numSpans:   20,
		endedSpans: true,
	})
	require.EqualValues(t, len(p.spanWatchlist.spans), 20)

	p.ForceFlush(context.Background())
	require.EqualValues(t, len(p.spanWatchlist.spans), 0)

	generateSpans(tr, spanGeneratorOption{
		numSpans:   5,
		endedSpans: true,
	})
	require.EqualValues(t, len(p.spanWatchlist.spans), 5)
}

func TestDtSpanProcessorForceFlushNonEndedSpan(t *testing.T) {
	p := createSpanProcessor()
	require.Zero(t, p.exportingStopped)
	require.Zero(t, len(p.spanWatchlist.spans))

	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(p))
	otel.SetTracerProvider(tp)
	tr := otel.Tracer("SDK Tracer")

	generateSpans(tr, spanGeneratorOption{
		numSpans:   20,
		endedSpans: true,
	})
	generateSpans(tr, spanGeneratorOption{
		numSpans:   5,
		endedSpans: false,
	})
	require.EqualValues(t, len(p.spanWatchlist.spans), 25)

	// Non-ended spans still must be watched
	p.ForceFlush(context.Background())
	require.EqualValues(t, len(p.spanWatchlist.spans), 5)
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

	err := p.ForceFlush(ctx)
	require.ErrorIs(t, err, context.Canceled)

	// check whether the exporting operation was aborted due to canceled force flush context
	select {
	case <-exporter.exportingFinished:
		log.Println("Exporter operation has been aborted")
	// the exporter must finish exporting operation within the next interval
	// due to canceled flush context
	case <-time.After(time.Millisecond * time.Duration(exporter.iterationIntervalMs*2)):
		require.Fail(t, "Exporter operation has not been aborted after flush context was canceled")
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

	err := p.ForceFlush(context.Background())
	require.ErrorIs(t, err, context.Canceled)

	// check whether the exporting operation was aborted due to the reached flush timeout
	select {
	case <-exporter.exportingFinished:
		log.Println("Exporter operation has been aborted")
	// the exporter must finish exporting operation within the next interval
	// due to canceled context caused by the flush timeout
	case <-time.After(time.Millisecond * time.Duration(exporter.iterationIntervalMs*2)):
		require.Fail(t, "Exporter operation has not been aborted after the flush timeout reached")
	}
}

func TestDtSpanProcessorWaitForScheduledFlushOperation(t *testing.T) {
	exporter := &testExporter{
		exportingFinished:   make(chan bool, 5),
		iterationIntervalMs: 500,
		numIterations:       4,
	}

	p := createSpanProcessor()
	p.exporter = exporter
	require.Zero(t, p.exportingStopped)
	require.Zero(t, len(p.spanWatchlist.spans))

	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(p))
	otel.SetTracerProvider(tp)
	tr := otel.Tracer("SDK Tracer")

	generateSpans(tr, spanGeneratorOption{
		numSpans:   5,
		endedSpans: true,
	})

	var wg sync.WaitGroup

	go func() {
		wg.Add(1)
		defer wg.Done()

		p.ForceFlush(context.Background())
	}()

	// 1st flush operation will be started ASAP, thus the flush channel must be empty
	// all spans are ended, so after evaluation the spans watchlist size must be 0
	require.Eventually(t, func() bool {
		return len(p.flushRequestCh) == 0 && len(p.spanWatchlist.spans) == 0
	}, 3*time.Second, 100*time.Millisecond)

	go func() {
		wg.Add(1)
		defer wg.Done()

		generateSpans(tr, spanGeneratorOption{
			numSpans:   10,
			endedSpans: true,
		})
		p.ForceFlush(context.Background())
	}()

	require.Eventually(t, func() bool {
		// the 1st flush is running, so the 2nd one must be scheduled, as a result, the flush channel must not be empty
		// 10 spans were started after the 1st flush call, thus span watchlist size must be equal to 10
		return len(p.flushRequestCh) == 1 && len(p.spanWatchlist.spans) == 10
	}, 3*time.Second, 100*time.Millisecond)

	go func() {
		wg.Add(1)
		defer wg.Done()

		generateSpans(tr, spanGeneratorOption{
			numSpans:   5,
			endedSpans: true,
		})

		p.ForceFlush(context.Background())

		// the 3rd flush call have to wait until the 2nd pending flush will be finished
		// and export spans started upon 3rd flush call
		require.Equal(t, len(p.spanWatchlist.spans), 0)
	}()

	wg.Wait()
}
