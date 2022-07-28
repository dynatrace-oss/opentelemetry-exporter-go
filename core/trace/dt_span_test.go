package trace

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestDtSpanEndsSdkSpan(t *testing.T) {
	tp, _ := newDtTracerProviderWithTestExporter()
	tr := tp.Tracer("Dynatrace tracer")

	_, span := tr.Start(context.Background(), "Test span")
	dynatraceSpan := span.(*dtSpan)
	sdkSpan := dynatraceSpan.Span.(sdktrace.ReadOnlySpan)

	require.True(t, sdkSpan.EndTime().IsZero())
	span.End()
	require.False(t, sdkSpan.EndTime().IsZero())
}

func TestDtSpanGetTracerProvider(t *testing.T) {
	tp, _ := newDtTracerProviderWithTestExporter()
	tr := tp.Tracer("Dynatrace tracer")
	_, s := tr.Start(context.Background(), "Test span")

	require.Equal(t, tp, s.TracerProvider())
}
