package fw4

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"
)

func TestCreateFw4Tag(t *testing.T) {
	clusterId := int32(1)
	tenantId := int32(2)
	traceId, _ := trace.TraceIDFromHex("4bf92f3577b34da6a3ce929d0e0e4736")
	spanId, _ := trace.SpanIDFromHex("00f067aa0ba902b7")
	spanContext := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: traceId,
		SpanID: spanId,
	})
	tag := NewFw4Tag(clusterId, tenantId, spanContext)
	assert.Equal(t, int32(clusterId), tag.ClusterID)
	assert.Equal(t, int32(tenantId), tag.TenantID)
	assert.Equal(t, spanContext.TraceID(), tag.TraceID)
	assert.Equal(t, spanContext.SpanID(), tag.SpanID)
}

func TestGetFw4TagFromContext_IsNil(t *testing.T) {
	ctx := context.Background()
	tag := Fw4TagFromContext(ctx)
	assert.Nil(t, tag)
}

func TestGetFw4TagFromContext_IsNilWhenNilIsAssigned(t *testing.T) {
	ctx := ContextWithFw4Tag(context.Background(), nil)
	tag := Fw4TagFromContext(ctx)
	assert.Nil(t, tag)
}

func TestGetFw4TagFromContext_IsNotNil(t *testing.T) {
	tag := EmptyTag()
	ctx := ContextWithFw4Tag(context.Background(), &tag)

	tagFromContext := Fw4TagFromContext(ctx)
	assert.NotNil(t, tagFromContext)
	assert.Equal(t, &tag, tagFromContext)
}
