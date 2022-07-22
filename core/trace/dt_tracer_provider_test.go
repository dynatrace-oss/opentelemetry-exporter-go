package trace

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
)

func TestTracerProviderCreatesDtTracer(t *testing.T) {
	tp := NewTracerProvider()
	otel.SetTracerProvider(tp)

	tr := otel.Tracer("Dynatrace Tracer")
	require.IsType(t, &dtTracer{}, tr)
}

func TestTracerProviderReturnsTheSameInstanceOfDtTracer(t *testing.T) {
	tp := NewTracerProvider()
	otel.SetTracerProvider(tp)

	tracerName := "Dynatrace Tracer"
	tr := otel.Tracer(tracerName)
	tr2 := otel.Tracer(tracerName)
	require.Equal(t, tr, tr2)
}

func TestTracerProviderShutdown(t *testing.T) {
	tp := NewTracerProvider()
	require.Zero(t, tp.processor.exportingStopped)

	otel.SetTracerProvider(tp)
	err := tp.Shutdown(context.Background())

	require.NoError(t, err, "shutdown has failed")
	require.NotZero(t, tp.processor.exportingStopped, "Exporting goroutine has not been stopped")
}

func TestTracerProviderForceFlush(t *testing.T) {
	tp := NewTracerProvider()
	require.Zero(t, tp.processor.exportingStopped)
	require.Zero(t, tp.processor.spanWatchlist.len())

	otel.SetTracerProvider(tp)
	tr := otel.Tracer("Dynatrace Tracer")

	generateSpans(tr, spanGeneratorOption{
		numSpans:   20,
		endedSpans: true,
	})
	
	require.EqualValues(t, tp.processor.spanWatchlist.len(), 20)
	tp.ForceFlush(context.Background())
	require.EqualValues(t, tp.processor.spanWatchlist.len(), 0)
}
