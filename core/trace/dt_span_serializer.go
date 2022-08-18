package trace

import (
	"errors"
	"fmt"

	attribute "go.opentelemetry.io/otel/attribute"
	codes "go.opentelemetry.io/otel/codes"
	resource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	trace "go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"

	"core/internal/version"

	protoCollectorCommon "dynatrace.com/odin/odin-proto/gen/go/collector/common/v1"
	protoCollector "dynatrace.com/odin/odin-proto/gen/go/collector/traces/v1"
	protoCommon "dynatrace.com/odin/odin-proto/gen/go/common/v1"
	protoResource "dynatrace.com/odin/odin-proto/gen/go/resource/v1"
	protoTrace "dynatrace.com/odin/odin-proto/gen/go/trace/v1"
)

func serializeSpans(
	spans dtSpanSet,
	tenantUUID string,
	agentId int64) ([]byte, error) {

	agSpanEnvelopes := make([]*protoCollector.ActiveGateSpanEnvelope, 0, len(spans))

	var resource *resource.Resource

	for span := range spans {
		fw4Tag := span.metadata.fw4Tag
		customTag, err := getProtoCustomTag(fw4Tag.CustomBlob)
		if err != nil {
			return nil, err
		}

		var spanMsg *protoTrace.Span
		spanMsg, resource, err = createProtoSpan(span, customTag)
		if err != nil {
			return nil, err
		}

		serializedClusterSpanEnvelope, err := createSerializedClusterSpanEnvelope(spanMsg, customTag, int32(fw4Tag.PathInfo))
		if err != nil {
			return nil, err
		}

		agSpanEnvelope := createAgSpanEnvelope(serializedClusterSpanEnvelope, int64(fw4Tag.ServerID), spanMsg.TraceId)
		agSpanEnvelopes = append(agSpanEnvelopes, agSpanEnvelope)
	}

	exportMetaInfo, err := proto.Marshal(&protoCollectorCommon.ExportMetaInfo{
		TimeSyncMode: protoCollectorCommon.ExportMetaInfo_NTPSync,
	})
	if err != nil {
		return nil, err
	}

	serializedResource, err := getSerializedResourceForSpanExport(resource)
	if err != nil {
		return nil, err
	}

	spanExport := &protoCollector.SpanExport{
		TenantUUID:     tenantUUID,
		AgentId:        agentId,
		ExportMetaInfo: exportMetaInfo,
		Resource:       serializedResource,
		Spans:          agSpanEnvelopes,
	}
	return proto.Marshal(spanExport)
}

func createProtoSpan(dtSpan *dtSpan, incomingCustomTag *protoTrace.CustomTag) (*protoTrace.Span, *resource.Resource, error) {
	spanMetadata := dtSpan.metadata
	span, ok := dtSpan.Span.(sdktrace.ReadOnlySpan)
	if !ok {
		return nil, nil, errors.New("failed to cast span to ReadOnlySpan")
	}

	spanContext := span.SpanContext()
	traceId := spanContext.TraceID()
	spanId := spanContext.SpanID()
	sendReason, err := getProtoSendReason(spanMetadata.sendState)
	if err != nil {
		return nil, nil, err
	}
	spanMsg := &protoTrace.Span{
		TraceId:          traceId[:],
		SpanId:           spanId[:],
		SendReason:       sendReason,
		UpdateSequenceNo: spanMetadata.seqNumber,
	}

	if sendReason := spanMetadata.sendState; sendReason == sendStateSpanEnded || sendReason == sendStateInitialSend {
		spanMsg.TenantParentSpanId = spanMetadata.tenantParentSpanId[:]
		parentSpanId := span.Parent().SpanID()
		spanMsg.ParentSpanId = parentSpanId[:]
		spanMsg.Name = span.Name()
		spanMsg.Kind = getProtoSpanKind(span.SpanKind())
		spanMsg.StartTimeUnixnano = uint64(span.StartTime().UnixNano())
		spanMsg.LastPropagateTimeUnixnano = uint64(spanMetadata.lastPropagationTime.UnixNano())

		if sendReason == sendStateSpanEnded {
			spanMsg.EndTimeUnixnano = uint64(span.EndTime().UnixNano())
		}

		spanMsg.Attributes = append(spanMsg.Attributes, getInstrumentationLibAttrs(span)...)
		protoAttributes, err := getProtoAttributes(span.Attributes())
		if err != nil {
			return nil, nil, err
		}
		spanMsg.Attributes = append(spanMsg.Attributes, protoAttributes...)
		
		protoEvents, err := getProtoEvents(span.Events())
		if err != nil {
			return nil, nil, err
		}
		spanMsg.Events = protoEvents

		protoLinks, err := getProtoLinks(span.Links())
		if err != nil {
			return nil, nil, err
		}
		spanMsg.Links = protoLinks

		status, err := getProtoStatus(span.Status())
		if err != nil {
			return nil, nil, err
		}
		spanMsg.Status = status
		
		if isRootSpan := !parentSpanId.IsValid(); isRootSpan {
			spanMsg.CustomTag = incomingCustomTag
		}
	}
	return spanMsg, span.Resource(), nil
}

func createSerializedClusterSpanEnvelope(spanMsg *protoTrace.Span, customTag *protoTrace.CustomTag, pathInfo int32) ([]byte, error) {
	spanContainer := protoCollector.SpanContainer{
		Spans: []*protoTrace.Span{spanMsg},
	}
	serializedSpanContainer, err := proto.Marshal(&spanContainer)
	if err != nil {
		return nil, err
	}

	var customTags []*protoTrace.CustomTag = nil
	if customTag != nil {
		customTags = append(customTags, customTag)
	}
	clusterSpanEnvelope := protoCollector.ClusterSpanEnvelope{
		TraceId:       spanMsg.TraceId,
		PathInfo:      pathInfo,
		CustomTags:    customTags,
		SpanContainer: serializedSpanContainer,
	}
	serializedClusterSpanEnvelope, err := proto.Marshal(&clusterSpanEnvelope)
	if err != nil {
		return nil, err
	}
	return serializedClusterSpanEnvelope, nil
}

func createAgSpanEnvelope(clusterSpanEnvelope []byte, serverId int64, traceId []byte) *protoCollector.ActiveGateSpanEnvelope {
	// Some code duplication is necessary here due to DestinationKey
	// having an internal interface type isActiveGateSpanEnvelope_DestinationKey
	if serverId != 0 {
		return &protoCollector.ActiveGateSpanEnvelope{
			ClusterSpanEnvelope: clusterSpanEnvelope,
			DestinationKey: &protoCollector.ActiveGateSpanEnvelope_ServerId{
				ServerId: serverId,
			},
		}
	} else {
		return &protoCollector.ActiveGateSpanEnvelope{
			ClusterSpanEnvelope: clusterSpanEnvelope,
			DestinationKey: &protoCollector.ActiveGateSpanEnvelope_TraceId{
				TraceId: traceId,
			},
		}
	}
}

func getProtoSendReason(sendState sendState) (protoTrace.Span_SendReason, error) {
	switch sendState {
	case sendStateNew, sendStateInitialSend:
		return protoTrace.Span_NewOrChanged, nil
	case sendStateDrop:
		return protoTrace.Span_Dropped, nil
	case sendStateAlive:
		return protoTrace.Span_KeepAlive, nil
	case sendStateSpanEnded:
		return protoTrace.Span_Ended, nil
	default:
		return -1, fmt.Errorf("unknown send state: %d", sendState)
	}
}

func getProtoSpanKind(spanKind trace.SpanKind) protoTrace.Span_SpanKind {
	switch spanKind {
	case trace.SpanKindUnspecified, trace.SpanKindInternal:
		return protoTrace.Span_INTERNAL
	case trace.SpanKindServer:
		return protoTrace.Span_SERVER
	case trace.SpanKindClient:
		return protoTrace.Span_CLIENT
	case trace.SpanKindProducer:
		return protoTrace.Span_PRODUCER
	case trace.SpanKindConsumer:
		return protoTrace.Span_CONSUMER
	default:
		return protoTrace.Span_SPAN_KIND_UNSPECIFIED
	}
}

func getInstrumentationLibAttrs(span sdktrace.ReadOnlySpan) []*protoCommon.AttributeKeyValue {
	instrumentationLib := span.InstrumentationLibrary()
	instrumentationLibNameAttr := &protoCommon.AttributeKeyValue{
		Type:        protoCommon.AttributeKeyValue_STRING,
		Key:         "otel.library.name", // TODO replace with SemConv constant
		StringValue: instrumentationLib.Name,
	}
	instrumentationLibVersionAttr := &protoCommon.AttributeKeyValue{
		Type:        protoCommon.AttributeKeyValue_STRING,
		Key:         "otel.library.version", // TODO replace with SemConv constant
		StringValue: instrumentationLib.Version,
	}
	return []*protoCommon.AttributeKeyValue{instrumentationLibNameAttr, instrumentationLibVersionAttr}
}

func getProtoAttributes(attributes []attribute.KeyValue) ([]*protoCommon.AttributeKeyValue, error) {
	protoAttrs := make([]*protoCommon.AttributeKeyValue, 0, len(attributes))
	for _, attr := range attributes {
		protoAttr, err := createProtoAttribute(attr)
		if err != nil {
			return nil, err
		}
		protoAttrs = append(protoAttrs, protoAttr)
	}
	return protoAttrs, nil
}

func createProtoAttribute(attr attribute.KeyValue) (*protoCommon.AttributeKeyValue, error) {
	attrKeyVal := protoCommon.AttributeKeyValue{
		Key: string(attr.Key),
	}
	switch attr.Value.Type() {
	case attribute.BOOL:
		attrKeyVal.Type = protoCommon.AttributeKeyValue_BOOL
		attrKeyVal.BoolValue = attr.Value.AsBool()
	case attribute.INT64:
		attrKeyVal.Type = protoCommon.AttributeKeyValue_INT
		attrKeyVal.IntValue = attr.Value.AsInt64()
	case attribute.FLOAT64:
		attrKeyVal.Type = protoCommon.AttributeKeyValue_DOUBLE
		attrKeyVal.DoubleValue = attr.Value.AsFloat64()
	case attribute.STRING:
		attrKeyVal.Type = protoCommon.AttributeKeyValue_STRING
		attrKeyVal.StringValue = attr.Value.AsString()
	case attribute.BOOLSLICE:
		attrKeyVal.Type = protoCommon.AttributeKeyValue_BOOL_ARRAY
		attrKeyVal.BoolValues = attr.Value.AsBoolSlice()
	case attribute.INT64SLICE:
		attrKeyVal.Type = protoCommon.AttributeKeyValue_INT_ARRAY
		attrKeyVal.IntValues = attr.Value.AsInt64Slice()
	case attribute.FLOAT64SLICE:
		attrKeyVal.Type = protoCommon.AttributeKeyValue_DOUBLE_ARRAY
		attrKeyVal.DoubleValues = attr.Value.AsFloat64Slice()
	case attribute.STRINGSLICE:
		attrKeyVal.Type = protoCommon.AttributeKeyValue_STRING_ARRAY
		attrKeyVal.StringValues = attr.Value.AsStringSlice()
	default:
		return nil, fmt.Errorf("unknown attribute type: %s", attr.Value.Type())
	}
	return &attrKeyVal, nil
}

func getProtoEvents(events []sdktrace.Event) ([]*protoTrace.Span_Event, error) {
	protoEvents := make([]*protoTrace.Span_Event, 0, len(events))
	for _, event := range events {
		protoAttributes, err := getProtoAttributes(event.Attributes)
		if err != nil {
			return nil, err
		}
		protoEvent := &protoTrace.Span_Event{
			Name:                   event.Name,
			TimeUnixnano:           uint64(event.Time.UnixNano()),
			Attributes:             protoAttributes,
			DroppedAttributesCount: uint32(event.DroppedAttributeCount),
		}
		protoEvents = append(protoEvents, protoEvent)
	}
	return protoEvents, nil
}

func getProtoLinks(links []sdktrace.Link) ([]*protoTrace.Span_Link, error) {
	protoLinks := make([]*protoTrace.Span_Link, 0, len(links))
	for _, link := range links {
		traceId := link.SpanContext.TraceID()
		spanId := link.SpanContext.SpanID()
		protoAttributes, err := getProtoAttributes(link.Attributes)
		if err != nil {
			return nil, err
		}
		protoLink := &protoTrace.Span_Link{
			TraceId:    traceId[:],
			SpanId:     spanId[:],
			Attributes: protoAttributes,
		}
		protoLinks = append(protoLinks, protoLink)
	}
	return protoLinks, nil
}

func getProtoStatus(status sdktrace.Status) (*protoTrace.Status, error) {
	statusCode, err := getProtoStatusCode(status.Code)
	if err != nil {
		return nil, err
	}
	return &protoTrace.Status{
		Code:    statusCode,
		Message: status.Description,
	}, nil
}

func getProtoStatusCode(code codes.Code) (protoTrace.Status_StatusCode, error) {
	switch code {
	case codes.Ok, codes.Unset:
		return protoTrace.Status_Ok, nil
	case codes.Error:
		return protoTrace.Status_UnknownError, nil
	default:
		return -1, fmt.Errorf("invalid status code: %d", code)
	}
}

func getProtoCustomTag(customBlob string) (*protoTrace.CustomTag, error) {
	if len(customBlob) > 0 {
		firstByte := customBlob[0]
		customTag := protoTrace.CustomTag{
			Type:      protoTrace.CustomTag_Type(firstByte),
			Direction: protoTrace.CustomTag_Incoming,
			TagValue:  []byte(customBlob[1:]),
		}
		return &customTag, nil
	}
	return nil, nil
}

func getSerializedResourceForSpanExport(spanResource *resource.Resource) ([]byte, error) {
	exporterResource := getExporterResource()
	mergedResource := mergeResources(spanResource, exporterResource)
	protoAttributes, err := getProtoAttributes(mergedResource.Attributes())
	if err != nil {
		return nil, err
	}
	resource := protoResource.Resource{
		Attributes: protoAttributes,
	}
	return proto.Marshal(&resource)
}

func getExporterResource() *resource.Resource {
	return resource.NewSchemaless(
		// TODO get the attribute keys from the generated semantic conventions
		attribute.Key("telemetry.exporter.name").String("odin"),
		attribute.Key("telemetry.exporter.version").String(version.FullVersion),
	)
}

// Merges the resources, taking the attribute values from resourceB if duplicate keys exist.
func mergeResources(resourceA, resourceB *resource.Resource) *resource.Resource {
	return resource.NewSchemaless(append(resourceA.Attributes(), resourceB.Attributes()...)...)
}
