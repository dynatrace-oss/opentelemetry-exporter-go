package export

import (
	"context"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func generateSpans(tr trace.Tracer, numSpans int) {
	wg := &sync.WaitGroup{}
	for i := 0; i < numSpans; i++ {
		wg.Add(1)
		go func(idx int) {
			_, span := tr.Start(context.Background(), fmt.Sprintf("Span #%d", idx))
			endTimestamp := trace.WithTimestamp(time.Now().Add(750 * time.Millisecond))
			span.End(endTimestamp)
			wg.Done()
		}(i)
	}

	wg.Wait()
}

func TestDtSpanProcessorStartSpans(t *testing.T) {
	p := NewDtSpanProcessor()
	assert.Zero(t, p.exportingStopped)
	assert.Zero(t, len(p.metadata.spans))

	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(p))
	otel.SetTracerProvider(tp)
	tr := otel.Tracer("SDK Tracer")

	generateSpans(tr, 20)
	assert.EqualValues(t, len(p.metadata.spans), 20)

	generateSpans(tr, 5)
	assert.EqualValues(t, len(p.metadata.spans), 25)
}

func TestDtSpanProcessorShutdown(t *testing.T) {
	p := NewDtSpanProcessor()
	assert.Zero(t, p.exportingStopped)

	err := p.Shutdown(context.Background())
	require.NoError(t, err, "Shutdown has failed")
	require.NotZero(t, p.exportingStopped, "Exporting goroutine has not been stopped")
}

func TestDtSpanProcessorPostShutdown(t *testing.T) {
	p := NewDtSpanProcessor()
	assert.Zero(t, p.exportingStopped)
	assert.Zero(t, len(p.metadata.spans))

	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(p))
	otel.SetTracerProvider(tp)
	tr := otel.Tracer("SDK Tracer")

	generateSpans(tr, 20)
	assert.EqualValues(t, len(p.metadata.spans), 20)

	err := p.Shutdown(context.Background())
	require.NoError(t, err, "Shutdown has failed")
	require.NotZero(t, p.exportingStopped, "Exporting goroutine has not been stopped")

	// span started after shutdown must not be processed by the processor
	generateSpans(tr, 5)
	assert.EqualValues(t, len(p.metadata.spans), 20)
}

func TestDtSpanProcessorShutdownCancelContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	p := NewDtSpanProcessor()
	assert.Zero(t, p.exportingStopped)

	err := p.Shutdown(ctx)
	require.ErrorIs(t, err, context.Canceled)
}

type testStuckExporter struct {
	waitExporter        chan struct{}
	iterationIntervalMs int
}

func (e *testStuckExporter) export(ctx context.Context, _ map[spanKey]*dtSpanMetadata) (err error) {
	for {
		// simulate exporting operation
		time.Sleep(time.Millisecond * time.Duration(e.iterationIntervalMs))

		if err = ctx.Err(); err != nil {
			break
		}
	}

	close(e.waitExporter)
	return
}

func TestDtSpanProcessorForceFlushCancelContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	stuckExporter := &testStuckExporter{
		waitExporter:        make(chan struct{}),
		iterationIntervalMs: 500,
	}

	p := NewDtSpanProcessor()
	p.exporter = stuckExporter
	assert.Zero(t, p.exportingStopped)

	err := p.ForceFlush(ctx)
	require.ErrorIs(t, err, context.Canceled)

	// check whether the exporter was aborted due to canceled force flush context
	select {
	case <-stuckExporter.waitExporter:
		log.Println("Exporter operation has been aborted")
	// the stuck exporter must finish exporting operation within the next interval
	// due to canceled flush context
	case <-time.After(time.Millisecond * time.Duration(stuckExporter.iterationIntervalMs*2)):
		assert.Fail(t, "Exporter operation has not been aborted after flush context was canceled")
	}
}

func TestDtSpanProcessorForceFlushTimeoutReached(t *testing.T) {
	stuckExporter := &testStuckExporter{
		waitExporter:        make(chan struct{}),
		iterationIntervalMs: 500,
	}

	p := NewDtSpanProcessor()
	p.exporter = stuckExporter
	assert.Zero(t, p.exportingStopped)

	err := p.ForceFlush(context.Background())
	require.ErrorIs(t, err, flushTimeoutReached)

	// check whether the exporting operation was aborted due to the reached flush timeout
	select {
	case <-stuckExporter.waitExporter:
		log.Println("Exporter operation has been aborted")
	// the stuck exporter must finish exporting operation within the next interval
	// due to canceled context caused by the flush timeout
	case <-time.After(time.Millisecond * time.Duration(stuckExporter.iterationIntervalMs*2)):
		assert.Fail(t, "Exporter operation has not been aborted after the flush timeout reached")
	}
}
