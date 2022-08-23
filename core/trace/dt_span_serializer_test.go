package trace

import (
	"context"
	"fmt"
	"testing"
	"time"

	protoCollectorTraces "dynatrace.com/odin/odin-proto/gen/go/collector/traces/v1"
	protoCommon "dynatrace.com/odin/odin-proto/gen/go/common/v1"
	protoTrace "dynatrace.com/odin/odin-proto/gen/go/trace/v1"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func TestCreateProtoSpan_NilDtSpan(t *testing.T) {
	protoSpan, resource, err := createProtoSpan(nil, nil)
	require.Nil(t, protoSpan)
	require.Nil(t, resource)
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
	protoSpan, resource, err := createProtoSpan(dtSpan, nil)
	require.Nil(t, protoSpan)
	require.Nil(t, resource)
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
	protoSpan, resource, err := createProtoSpan(dtSpan, nil)
	require.Nil(t, protoSpan)
	require.Nil(t, resource)
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
		Type: protoTrace.CustomTag_Generic,
		Direction: protoTrace.CustomTag_Incoming,
	}
	protoSpan, resource, err := createProtoSpan(dtSpan, customTag)
	require.NoError(t, err)
	require.NotNil(t, protoSpan)
	require.NotNil(t, resource)
	require.NotNil(t, protoSpan.GetTraceId())
	require.NotNil(t, protoSpan.GetSpanId())
	require.Nil(t, protoSpan.GetParentSpanId())
	require.Equal(t, protoSpan.GetName(), "test span")
	require.Same(t, protoSpan.GetCustomTag(), customTag)
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

func TestGetProtoCustomTag(t *testing.T) {
	require.NotNil(t, getProtoCustomTag("customTag"))
	require.Nil(t, getProtoCustomTag(""))
}

func TestCreateProtoAttribute(t *testing.T) {
	testCases := []struct {
		attribute                       attribute.KeyValue
		valueGetter                     func(*protoCommon.AttributeKeyValue) interface{}
		expectedProtoAttributeValueType protoCommon.AttributeKeyValue_ValueType
	}{
		{
			attribute: attribute.String("string_attr", "value"),
			valueGetter: func(kv *protoCommon.AttributeKeyValue) interface{} {
				return kv.GetStringValue()
			},
			expectedProtoAttributeValueType: protoCommon.AttributeKeyValue_STRING,
		},
		{
			attribute: attribute.Int("int_attr", 123),
			valueGetter: func(kv *protoCommon.AttributeKeyValue) interface{} {
				return kv.GetIntValue()
			},
			expectedProtoAttributeValueType: protoCommon.AttributeKeyValue_INT,
		},
		{
			attribute: attribute.Float64("double_attr", 123.45),
			valueGetter: func(kv *protoCommon.AttributeKeyValue) interface{} {
				return kv.GetDoubleValue()
			},
			expectedProtoAttributeValueType: protoCommon.AttributeKeyValue_DOUBLE,
		},
		{
			attribute: attribute.Bool("bool_attr", true),
			valueGetter: func(kv *protoCommon.AttributeKeyValue) interface{} {
				return kv.GetBoolValue()
			},
			expectedProtoAttributeValueType: protoCommon.AttributeKeyValue_BOOL,
		},
		{
			attribute: attribute.StringSlice("string_array_attr", []string{"foo", "bar"}),
			valueGetter: func(kv *protoCommon.AttributeKeyValue) interface{} {
				return kv.GetStringValues()
			},
			expectedProtoAttributeValueType: protoCommon.AttributeKeyValue_STRING_ARRAY,
		},
		{
			attribute: attribute.Int64Slice("int_array_attr", []int64{1, 2, 3}),
			valueGetter: func(kv *protoCommon.AttributeKeyValue) interface{} {
				return kv.GetIntValues()
			},
			expectedProtoAttributeValueType: protoCommon.AttributeKeyValue_INT_ARRAY,
		},
		{
			attribute: attribute.Float64Slice("double_array_attr", []float64{1.1, 2.2, 3.3}),
			valueGetter: func(kv *protoCommon.AttributeKeyValue) interface{} {
				return kv.GetDoubleValues()
			},
			expectedProtoAttributeValueType: protoCommon.AttributeKeyValue_DOUBLE_ARRAY,
		},
		{
			attribute: attribute.BoolSlice("bool_array_attr", []bool{true, false, true}),
			valueGetter: func(kv *protoCommon.AttributeKeyValue) interface{} {
				return kv.GetBoolValues()
			},
			expectedProtoAttributeValueType: protoCommon.AttributeKeyValue_BOOL_ARRAY,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Test proto attribute creation for attribute %s", tc.attribute.Key), func(t *testing.T) {
			protoAttribute, err := createProtoAttribute(tc.attribute)
			require.NoError(t, err)
			require.Equal(t, string(tc.attribute.Key), protoAttribute.GetKey())
			require.Equal(t, tc.expectedProtoAttributeValueType, protoAttribute.GetType())
			require.Equal(t, tc.valueGetter(protoAttribute), tc.attribute.Value.AsInterface())
		})
	}
}

func TestGetProtoAttributes(t *testing.T) {
	attributes := []attribute.KeyValue{
		attribute.String("string_attr", "value"),
		attribute.Int("int_attr", 123),
		attribute.Float64("double_attr", 123.45),
		attribute.Bool("bool_attr", true),
		attribute.StringSlice("string_array_attr", []string{"foo", "bar"}),
		attribute.Int64Slice("int_array_attr", []int64{1, 2, 3}),
		attribute.Float64Slice("double_array_attr", []float64{1.1, 2.2, 3.3}),
		attribute.BoolSlice("bool_array_attr", []bool{true, false, true}),
	}

	protoAttributes, err := getProtoAttributes(attributes)
	require.NoError(t, err)
	require.Len(t, protoAttributes, len(attributes))
}

func TestGetProtoStatus(t *testing.T) {
	status := sdktrace.Status{
		Code:        codes.Ok,
		Description: "description",
	}
	protoStatus, err := getProtoStatus(status)
	require.NoError(t, err)
	require.Equal(t, protoStatus.GetCode(), protoTrace.Status_Ok)
	require.Equal(t, protoStatus.GetMessage(), status.Description)
}

func TestGetProtoStatusCode(t *testing.T) {
	codeToProtoStatusCodeMap := map[codes.Code]protoTrace.Status_StatusCode{
		codes.Ok:    protoTrace.Status_Ok,
		codes.Unset: protoTrace.Status_Ok,
		codes.Error: protoTrace.Status_UnknownError,
	}

	for code, expectedProtoStatusCode := range codeToProtoStatusCodeMap {
		protoStatusCode, err := getProtoStatusCode(code)
		require.NoError(t, err)
		require.Equal(t, expectedProtoStatusCode, protoStatusCode)
	}
}

func TestGetProtoStatusCode_InvalidCode(t *testing.T) {
	_, err := getProtoStatusCode(codes.Ok + 1)
	require.Error(t, err)
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
