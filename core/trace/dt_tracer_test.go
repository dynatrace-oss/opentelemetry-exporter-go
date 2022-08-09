package trace

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
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

	sdkSpan, ok := span.Span.(sdktrace.ReadOnlySpan)
	require.NotNil(t, sdkSpan)
	require.True(t, ok)
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

func TestLocalParentWithServerId(t *testing.T) {
	c := propagation.HeaderCarrier{}
	c.Set(traceparentHeader, "00-11223344556677889900112233445566-8877665544332211-01")
	c.Set(tracestateHeader, "34d2ae74-7b@dt=fw4;fffffff8;0;0;0;0;0;0;7db5;2h01;7h8877665544332211")

	p := NewTextMapPropagator()
	require.NotNil(t, p)
	remoteCtx := p.Extract(context.Background(), c)

	tp, _ := newDtTracerProviderWithTestExporter()
	tr := tp.Tracer("Dynatrace tracer")
	ctx, localParent := tr.Start(remoteCtx, "local root")
	_, span := tr.Start(ctx, "child")

	metadata := span.(*dtSpan).metadata
	require.NotNil(t, metadata)

	tag := metadata.getFw4Tag()
	require.NotNil(t, tag)
	require.EqualValues(t, tag.ServerID, -8)
	require.Equal(t, metadata.tenantParentSpanId, localParent.SpanContext().SpanID())
}

func TestLocalParentWithNoServerId(t *testing.T) {
	tp, _ := newDtTracerProviderWithTestExporter()
	tr := tp.Tracer("Dynatrace tracer")
	ctx, localParent := tr.Start(context.Background(), "local root")
	_, span := tr.Start(ctx, "child")

	metadata := span.(*dtSpan).metadata
	require.NotNil(t, metadata)

	tag := metadata.getFw4Tag()
	require.NotNil(t, tag)
	require.EqualValues(t, tag.ServerID, 0)
	require.True(t, tag.PathInfo < 256)
	require.Equal(t, metadata.tenantParentSpanId, localParent.SpanContext().SpanID())
}

func TestRemoteParentWithTracestateAndWrongTenantId(t *testing.T) {
	c := propagation.HeaderCarrier{}
	c.Set(traceparentHeader, "00-11223344556677889900112233445566-aaffffeebbaabbee-01")
	c.Set(tracestateHeader, "6cab5bb-7b@dt=fw4;fffffff8;0;0;0;0;0;0;7db5;2h01;7h8877665544332211")

	p := NewTextMapPropagator()
	require.NotNil(t, p)

	remoteCtx := p.Extract(context.Background(), c)

	tp, _ := newDtTracerProviderWithTestExporter()
	tr := tp.Tracer("Dynatrace tracer")
	_, span := tr.Start(remoteCtx, "local root")

	metadata := span.(*dtSpan).metadata
	require.NotNil(t, metadata)
	require.Equal(t, span.SpanContext().TraceID().String(), "11223344556677889900112233445566")
	require.Zero(t, metadata.tenantParentSpanId)

	tag := metadata.getFw4Tag()
	require.NotNil(t, tag)
	require.EqualValues(t, tag.ServerID, 0)
	require.True(t, tag.PathInfo < 256)
}
