package trace

import (
	"context"
	// "core/configuration"
	"core/fw4"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	trace "go.opentelemetry.io/otel/trace"
)

func createTracer() trace.Tracer {
	otel.SetTracerProvider(NewTracerProvider())
	return otel.Tracer("Test tracer")
}

func TestGetFw4TagFromContext_IsNil(t *testing.T) {
	ctx :=  context.Background()
	tag := getFw4TagFromContext(ctx)
	assert.Nil(t, tag)
}

func TestGetFw4TagFromContext_IsNotNil(t *testing.T) {
	tag := fw4.EmptyTag()
	ctx := context.WithValue(context.Background(), fw4TagKey, &tag)

	tagFromContext := getFw4TagFromContext(ctx)
	assert.NotNil(t, tagFromContext)
	assert.Equal(t, &tag, tagFromContext)
}

func TestCreateFw4Tag(t *testing.T) {
	clusterId := int32(1)
	tenantId := int32(2)
	traceId, _ := trace.TraceIDFromHex("4bf92f3577b34da6a3ce929d0e0e4736")
	spanId, _ := trace.SpanIDFromHex("00f067aa0ba902b7")
	spanContext := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: traceId,
		SpanID: spanId,
	})
	tag := createFw4Tag(clusterId, tenantId, spanContext)
	assert.Equal(t, int32(clusterId), tag.ClusterID)
	assert.Equal(t, int32(tenantId), tag.TenantID)
	assert.Equal(t, spanContext.TraceID(), tag.TraceID)
	assert.Equal(t, spanContext.SpanID(), tag.SpanID)
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
	assert.Equal(t, childMetadata.fw4Tag, parentMetadata.fw4Tag, "Pointer to FW4Tag of child should be equal to parent")
	assert.True(t, childMetadata.lastPropagationTime.IsZero())
	assert.True(t, !parentMetadata.lastPropagationTime.IsZero())
}
