package trace

// import (
// 	"context"
// 	"core/fw4"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// 	"go.opentelemetry.io/otel"
// 	sdktrace "go.opentelemetry.io/otel/sdk/trace"
// 	trace "go.opentelemetry.io/otel/trace"
// )

// func createTracer() trace.Tracer {
// 	otel.SetTracerProvider(sdktrace.NewTracerProvider())
// 	return otel.Tracer("Test tracer")
// }

// func TestGetFw4TagFromMetadata(t *testing.T) {
// 	tag := fw4.EmptyTag()
// 	metadataWithTag := dtSpanMetadata{
// 		fw4Tag: &tag,
// 	}
// 	assert.Equal(t, &tag, getFw4Tag(&metadataWithTag))

// 	metadataWithoutTag := dtSpanMetadata{}
// 	assert.Nil(t, getFw4Tag(&metadataWithoutTag))
// }

// func TestCreateFw4Tag(t *testing.T) {
// 	clusterId := int32(1)
// 	tenantId := int32(2)
// 	tag := createFw4Tag(clusterId, tenantId)
// 	assert.Equal(t, int32(clusterId), tag.ClusterID)
// 	assert.Equal(t, int32(tenantId), tag.TenantID)
// }

// func TestCreateSpanMetadata(t *testing.T) {
// 	spanEnricher := dtSpanEnricher{}

// 	tracer := createTracer()
// 	_, span := tracer.Start(context.Background(), "test span name")
// 	rwSpan := span.(sdktrace.ReadWriteSpan)

// 	metadataMap := newDtSpanMetadataMap(99)
// 	metadata := spanEnricher.CreateSpanMetaData(rwSpan, defaultTransmitOptions, 1, 2, &metadataMap)
// 	assert.NotNil(t, metadata)
// 	assert.NotNil(t, metadata.fw4Tag)
// 	assert.Equal(t, int32(1), metadata.fw4Tag.ClusterID)
// 	assert.Equal(t, int32(2), metadata.fw4Tag.TenantID)
// }

// func TestCreateSpanMetadata_WithParentSpan(t *testing.T) {
// 	spanEnricher := dtSpanEnricher{}
// 	metadataMap := newDtSpanMetadataMap(99)

// 	tracer := createTracer()
// 	ctx, parentSpan := tracer.Start(context.Background(), "parent span")
// 	parentMetadata := newDtSpanMetadata(defaultTransmitOptions, parentSpan.(sdktrace.ReadOnlySpan))
// 	metadataMap.add(parentSpan.SpanContext(), parentMetadata)

// 	_, childSpan := tracer.Start(ctx, "child span")

// 	metadata := spanEnricher.CreateSpanMetaData(childSpan.(sdktrace.ReadWriteSpan), defaultTransmitOptions, 1, 2, &metadataMap)

// 	assert.Equal(t, parentSpan.SpanContext().SpanID(), metadata.tenantParentSpanId)
// 	assert.Equal(t, metadata.fw4Tag.ClusterID, int32(1))
// 	assert.Equal(t, metadata.fw4Tag.TenantID, int32(2))
// 	assert.True(t, metadata.lastPropagationTime.IsZero())
// 	assert.True(t, !parentMetadata.lastPropagationTime.IsZero())
// }

// func TestMarkSpanPropagatedNow_SetIfMetadataExistsInMap(t *testing.T) {
// 	spanContext := trace.SpanContext{}
// 	metadata := dtSpanMetadata{}
// 	metadataMap := newDtSpanMetadataMap(99)
// 	metadataMap.add(spanContext, &metadata)

// 	markSpanPropagatedNow(spanContext, &metadataMap)
// 	assert.True(t, !metadata.lastPropagationTime.IsZero())
// }

// func TestMarkSpanPropagatedNow_NotSetIfMetadaNotInMap(t *testing.T) {
// 	spanContext := trace.SpanContext{}
// 	metadata := dtSpanMetadata{}
// 	metadataMap := newDtSpanMetadataMap(99)
// 	// Intentionally don't add the metadata to the map

// 	markSpanPropagatedNow(spanContext, &metadataMap)
// 	assert.True(t, metadata.lastPropagationTime.IsZero())
// }
