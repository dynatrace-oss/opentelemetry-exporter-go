package trace

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTracerProviderCreatesDtTracer(t *testing.T) {
	tp, _ := newDtTracerProviderWithTestExporter()
	tr := tp.Tracer("Dynatrace Tracer")
	require.IsType(t, &dtTracer{}, tr)
}

func TestTracerProviderReturnsTheSameInstanceOfDtTracer(t *testing.T) {
	tp, _ := newDtTracerProviderWithTestExporter()

	tracerName := "Dynatrace Tracer"
	tr := tp.Tracer(tracerName)
	tr2 := tp.Tracer(tracerName)

	require.Equal(t, tr, tr2)
}

func TestTracerProviderShutdown(t *testing.T) {
	tp, _ := newDtTracerProviderWithTestExporter()
	require.Zero(t, tp.processor.exportingStopped)

	err := tp.Shutdown(context.Background())

	require.NoError(t, err, "shutdown has failed")
	require.NotZero(t, tp.processor.exportingStopped, "Exporting goroutine has not been stopped")
}

func TestTracerProviderForceFlush(t *testing.T) {
	tp, _ := newDtTracerProviderWithTestExporter()

	require.Zero(t, tp.processor.exportingStopped)
	require.Zero(t, tp.processor.spanWatchlist.len())

	tr := tp.Tracer("Dynatrace Tracer")

	generateSpans(tr, spanGeneratorOptions{
		numSpans:   20,
		endedSpans: true,
	})

	require.EqualValues(t, tp.processor.spanWatchlist.len(), 20)
	tp.ForceFlush(context.Background())
	require.EqualValues(t, tp.processor.spanWatchlist.len(), 0)
}
