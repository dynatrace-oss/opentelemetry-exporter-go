package trace

import (
	"context"
	"testing"
	"time"

	protoCollectorTraces "dynatrace.com/odin/odin-proto/gen/go/collector/traces/v1"
	protoTrace "dynatrace.com/odin/odin-proto/gen/go/trace/v1"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func TestCreateProtoSpan_NilDtSpan(t *testing.T) {
	protoSpan, err := createProtoSpan(nil, nil)
	require.Nil(t, protoSpan)
	require.Error(t, err)
}

func TestCreateProtoSpan_NonReadOnlySpan(t *testing.T) {
	tp := trace.NewNoopTracerProvider()
	tracer := tp.Tracer("test")
	// span will be of type noopSpan
	_, span := tracer.Start(context.Background(), "test span")
	dtSpan := &dtSpan{
		Span:     span,
		metadata: newDtSpanMetadata(123),
	}
	protoSpan, err := createProtoSpan(dtSpan, nil)
	require.Nil(t, protoSpan)
	require.Error(t, err)
}

func TestCreateProtoSpan_NoMetadata(t *testing.T) {
	tp := sdktrace.NewTracerProvider()
	tracer := tp.Tracer("test")
	_, span := tracer.Start(context.Background(), "test span")
	dtSpan := &dtSpan{
		Span:     span,
		metadata: nil,
	}
	protoSpan, err := createProtoSpan(dtSpan, nil)
	require.Nil(t, protoSpan)
	require.Error(t, err)
}

func TestCreateProtoSpan(t *testing.T) {
	tp, _ := newDtTracerProviderWithTestExporter()
	tracer := tp.Tracer("test")
	_, span := tracer.Start(context.Background(), "test span")
	span.End()
	err := tp.ForceFlush(context.Background())
	require.NoError(t, err)

	dtSpan := span.(*dtSpan)

	customTag := &protoTrace.CustomTag{
		Type:      protoTrace.CustomTag_Generic,
		Direction: protoTrace.CustomTag_Incoming,
	}
	protoSpan, err := createProtoSpan(dtSpan, customTag)
	require.NoError(t, err)
	require.NotNil(t, protoSpan)
	require.NotNil(t, protoSpan.GetTraceId())
	require.NotNil(t, protoSpan.GetSpanId())
	require.Nil(t, protoSpan.GetParentSpanId())
	require.Equal(t, protoSpan.GetName(), "test span")
	require.Same(t, protoSpan.GetCustomTag(), customTag)
}

func TestGetFirstResource(t *testing.T) {
	tracer := createTracer().(*dtTracer)
	_, span1 := tracer.Start(context.Background(), "span1")
	_, span2 := tracer.Start(context.Background(), "span2")
	_, span3 := tracer.Start(context.Background(), "span3")

	spans := make(dtSpanSet)
	spans[span1.(*dtSpan)] = struct{}{}
	spans[span2.(*dtSpan)] = struct{}{}
	spans[span3.(*dtSpan)] = struct{}{}
	resource, err := getFirstResource(spans)
	require.NoError(t, err)
	require.NotNil(t, resource)
	require.Same(t, resource, span1.(*dtSpan).Span.(sdktrace.ReadOnlySpan).Resource())
}

func TestGetFirstResource_FailsIfSetEmpty(t *testing.T) {
	resource, err := getFirstResource(make(dtSpanSet))
	require.Error(t, err)
	require.Nil(t, resource)
}

func TestMergeResources_WithDuplicateKey(t *testing.T) {
	resource1 := resource.NewSchemaless(
		attribute.String("A", "res1_A"),
		attribute.String("B", "res1_B"),
	)

	resource2 := resource.NewSchemaless(
		attribute.String("A", "res2_A"),
		attribute.String("C", "res2_C"),
	)

	merged := mergeResources(resource1, resource2)

	require.ElementsMatch(t, []attribute.KeyValue{
		attribute.String("A", "res2_A"),
		attribute.String("B", "res1_B"),
		attribute.String("C", "res2_C"),
	}, merged.Attributes(), "Duplicate key should be overwritten by second arg of mergeResources")
}

func TestMergeResources_WithoutAttributes(t *testing.T) {
	resource1 := resource.NewSchemaless()
	resource2 := resource.NewSchemaless()
	merged := mergeResources(resource1, resource2)

	require.Empty(t, merged.Attributes(), "Merged resource attributes should be returned if both resources have no attributes")
}

func TestCreateAgSpanEnvelope_WithServerId(t *testing.T) {
	clusterSpanEnvelope := []byte{1, 2, 3}
	agSpanEnvelope := createAgSpanEnvelope(clusterSpanEnvelope, 99, nil)

	require.Equal(t, clusterSpanEnvelope, agSpanEnvelope.GetClusterSpanEnvelope())
	require.EqualValues(t, 99, agSpanEnvelope.DestinationKey.(*protoCollectorTraces.ActiveGateSpanEnvelope_ServerId).ServerId,
		"ServerId should be set in the envelope when provided to createAgSpanEnvelope")

	agSpanEnvelope = createAgSpanEnvelope(clusterSpanEnvelope, 99, []byte{4, 5, 6})
	require.Equal(t, clusterSpanEnvelope, agSpanEnvelope.GetClusterSpanEnvelope())
	require.EqualValues(t, 99, agSpanEnvelope.DestinationKey.(*protoCollectorTraces.ActiveGateSpanEnvelope_ServerId).ServerId,
		"ServerId should be set in the envelope when provided to createAgSpanEnvelope, even if traceId was provided too")
}

func TestCreateAgSpanEnvelope_WithTraceId(t *testing.T) {
	clusterSpanEnvelope := []byte{1, 2, 3}
	agSpanEnvelope := createAgSpanEnvelope(clusterSpanEnvelope, 0, []byte{4, 5, 6})

	require.Equal(t, clusterSpanEnvelope, agSpanEnvelope.GetClusterSpanEnvelope())
	require.Equal(t, []byte{4, 5, 6}, agSpanEnvelope.DestinationKey.(*protoCollectorTraces.ActiveGateSpanEnvelope_TraceId).TraceId,
		"TraceId should be set in the envelope when provided to createAgSpanEnvelope and serverId is 0")
}

func TestGetProtoSendReason(t *testing.T) {
	sendStateToSendReasonMap := map[sendState]protoTrace.Span_SendReason{
		sendStateNew:         protoTrace.Span_NewOrChanged,
		sendStateInitialSend: protoTrace.Span_NewOrChanged,
		sendStateDrop:        protoTrace.Span_Dropped,
		sendStateAlive:       protoTrace.Span_KeepAlive,
		sendStateSpanEnded:   protoTrace.Span_Ended,
	}

	for state, expectedReason := range sendStateToSendReasonMap {
		reason, err := getProtoSendReason(state)
		require.NoError(t, err)
		require.Equal(t, expectedReason, reason)
	}
}

func TestGetProtoSendReason_InvalidState(t *testing.T) {
	_, err := getProtoSendReason(sendStateNew - 1)
	require.Error(t, err)

	_, err = getProtoSendReason(sendStateSpanEnded + 1)
	require.Error(t, err)
}

func TestGetProtoSpanKind(t *testing.T) {
	spanKindToProtoSpanKindMap := map[trace.SpanKind]protoTrace.Span_SpanKind{
		trace.SpanKindUnspecified:     protoTrace.Span_INTERNAL,
		trace.SpanKindInternal:        protoTrace.Span_INTERNAL,
		trace.SpanKindServer:          protoTrace.Span_SERVER,
		trace.SpanKindClient:          protoTrace.Span_CLIENT,
		trace.SpanKindProducer:        protoTrace.Span_PRODUCER,
		trace.SpanKindConsumer:        protoTrace.Span_CONSUMER,
		trace.SpanKindUnspecified - 1: protoTrace.Span_SPAN_KIND_UNSPECIFIED,
		trace.SpanKindConsumer + 1:    protoTrace.Span_SPAN_KIND_UNSPECIFIED,
	}

	for spanKind, expectedProtoSpanKind := range spanKindToProtoSpanKindMap {
		protoSpanKind := getProtoSpanKind(spanKind)
		require.Equal(t, expectedProtoSpanKind, protoSpanKind)
	}
}

func TestGetProtoEvents(t *testing.T) {
	events := []sdktrace.Event{
		{
			Name:                  "event1",
			Attributes:            []attribute.KeyValue{attribute.String("key1", "value1")},
			Time:                  time.Now(),
			DroppedAttributeCount: 0,
		},
		{
			Name:                  "event2",
			Attributes:            []attribute.KeyValue{attribute.String("key2", "value2")},
			Time:                  time.Now(),
			DroppedAttributeCount: 1,
		},
	}

	protoEvents, err := getProtoEvents(events)
	require.NoError(t, err)
	require.Len(t, protoEvents, len(events))

	for i, protoEvent := range protoEvents {
		require.Equal(t, protoEvent.GetName(), events[i].Name)
		require.EqualValues(t, protoEvent.GetTimeUnixnano(), events[i].Time.UnixNano())
		require.Len(t, protoEvent.GetAttributes(), len(events[i].Attributes))
		require.EqualValues(t, protoEvent.GetAttributes()[0].GetKey(), events[i].Attributes[0].Key)
		require.Equal(t, protoEvent.GetAttributes()[0].GetStringValue(), events[i].Attributes[0].Value.AsString())
		require.EqualValues(t, protoEvent.GetDroppedAttributesCount(), events[i].DroppedAttributeCount)
	}
}

func TestGetProtoLinks(t *testing.T) {
	links := []sdktrace.Link{
		{
			SpanContext: trace.NewSpanContext(trace.SpanContextConfig{
				TraceID:    trace.TraceID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
				SpanID:     trace.SpanID{1, 2, 3, 4, 5, 6, 7, 8},
				TraceFlags: 0,
			}),
			Attributes:            []attribute.KeyValue{attribute.String("key1", "value1")},
			DroppedAttributeCount: 0,
		},
		{
			SpanContext: trace.NewSpanContext(trace.SpanContextConfig{
				TraceID:    trace.TraceID{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
				SpanID:     trace.SpanID{8, 7, 6, 5, 4, 3, 2, 1},
				TraceFlags: 0,
			}),
			Attributes:            []attribute.KeyValue{attribute.String("key2", "value2")},
			DroppedAttributeCount: 1,
		},
	}

	protoLinks, err := getProtoLinks(links)
	require.NoError(t, err)
	require.Len(t, protoLinks, len(links))

	for i, protoLink := range protoLinks {
		linkTraceId := links[i].SpanContext.TraceID()
		require.Equal(t, protoLink.GetTraceId(), linkTraceId[:])
		linkSpanId := links[i].SpanContext.SpanID()
		require.EqualValues(t, protoLink.GetSpanId(), linkSpanId[:])
		require.Len(t, protoLink.GetAttributes(), len(links[i].Attributes))
		require.EqualValues(t, protoLink.GetAttributes()[0].GetKey(), links[i].Attributes[0].Key)
		require.Equal(t, protoLink.GetAttributes()[0].GetStringValue(), links[i].Attributes[0].Value.AsString())
		require.EqualValues(t, protoLink.GetDroppedAttributesCount(), links[i].DroppedAttributeCount)
	}
}
