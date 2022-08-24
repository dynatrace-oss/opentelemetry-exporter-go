package trace

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
)

func TestDtSpanIsOfDtSpanType(t *testing.T) {
	tp, _ := newDtTracerProviderWithTestExporter()
	tr := tp.Tracer("Dynatrace tracer")

	_, s := tr.Start(context.Background(), "Test span")
	span, ok := s.(*dtSpan)

	require.NotNil(t, span)
	require.True(t, ok)
}

func TestDtSpanContainsSdkSpan(t *testing.T) {
	tp, _ := newDtTracerProviderWithTestExporter()
	tr := tp.Tracer("Dynatrace tracer")

	_, s := tr.Start(context.Background(), "Test span")
	span := s.(*dtSpan)

	sdkSpan, err := span.readOnlySpan()
	require.NotNil(t, sdkSpan)
	require.NoError(t, err)
}

func TestDtSpanContainsValidTracer(t *testing.T) {
	tp, _ := newDtTracerProviderWithTestExporter()
	tr := tp.Tracer("Dynatrace tracer")

	_, s := tr.Start(context.Background(), "Test span")
	span := s.(*dtSpan)

	require.Equal(t, tr, span.tracer)
}

func TestReturnedContextContainsDtSpan(t *testing.T) {
	tp, _ := newDtTracerProviderWithTestExporter()
	tr := tp.Tracer("Dynatrace tracer")

	ctx, s := tr.Start(context.Background(), "Test span")

	spanFromCtx := trace.SpanFromContext(ctx)
	require.Equal(t, spanFromCtx, s)
}
