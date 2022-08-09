package trace

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	trace "go.opentelemetry.io/otel/trace"

	"core/internal/fw4"
)

func createTracer() trace.Tracer {
	tp, _ := newDtTracerProviderWithTestExporter()
	otel.SetTracerProvider(tp)
	return otel.Tracer("Test tracer")
}

func TestSpanHasMetadata(t *testing.T) {
	tracer := createTracer()
	_, span := tracer.Start(context.Background(), "test span name")

	metadata := span.(*dtSpan).metadata
	assert.NotNil(t, metadata)
	assert.NotNil(t, metadata.fw4Tag)
	assert.Equal(t, tracer.(*dtTracer).config.ClusterId, metadata.fw4Tag.ClusterID)
	assert.Equal(t, tracer.(*dtTracer).config.TenantId(), metadata.fw4Tag.TenantID)
}

func TestCreateSpanMetadata_WithParentSpan(t *testing.T) {
	tracer := createTracer()
	ctx, parentSpan := tracer.Start(context.Background(), "parent span")
	parentDtSpan := parentSpan.(*dtSpan)

	_, childSpan := tracer.Start(ctx, "child span")
	childDtSpan := childSpan.(*dtSpan)

	parentMetadata := parentDtSpan.metadata
	childMetadata := childDtSpan.metadata

	assert.Equal(t, parentSpan.SpanContext().SpanID(), childMetadata.tenantParentSpanId)
	assert.Equal(t, childMetadata.fw4Tag.ClusterID, tracer.(*dtTracer).config.ClusterId)
	assert.Equal(t, childMetadata.fw4Tag.TenantID, tracer.(*dtTracer).config.TenantId())
	assert.Same(t, childMetadata.fw4Tag, parentMetadata.fw4Tag, "Pointer to FW4Tag of child should be equal to parent")
	assert.True(t, childMetadata.lastPropagationTime.IsZero())
	assert.False(t, parentMetadata.lastPropagationTime.IsZero())
}

func TestCreateSpanMetadata_WithRemoteParentSpan(t *testing.T) {
	// Create remote span context
	spanId, _ := trace.SpanIDFromHex("00f067aa0ba902b7")
	spanCtx := trace.NewSpanContext(trace.SpanContextConfig{
		SpanID: spanId,
	})
	ctx := trace.ContextWithRemoteSpanContext(context.Background(), spanCtx)

	// Add Fw4Tag to remote span context
	tag := fw4.NewFw4Tag(1, 2, spanCtx)
	ctx = fw4.ContextWithFw4Tag(ctx, tag)

	tracer := createTracer()
	_, childSpan := tracer.Start(ctx, "child span with remote span context")
	childDtSpan := childSpan.(*dtSpan)
	childMetadata := childDtSpan.metadata

	assert.Equal(t, spanCtx.SpanID(), childMetadata.tenantParentSpanId)
	assert.Equal(t, childMetadata.fw4Tag.ClusterID, int32(1))
	assert.Equal(t, childMetadata.fw4Tag.TenantID, int32(2))
	assert.Same(t, childMetadata.fw4Tag, tag, "Pointer to FW4Tag of child should be equal to tag in parent context")
	assert.True(t, childMetadata.lastPropagationTime.IsZero())
}
