package trace

import (
	"errors"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/protobuf/proto"

	"core/internal/version"

	protoCollectorCommon "dynatrace.com/odin/odin-proto/gen/go/collector/common/v1"
	protoCollectorTraces "dynatrace.com/odin/odin-proto/gen/go/collector/traces/v1"
	protoCommon "dynatrace.com/odin/odin-proto/gen/go/common/v1"
	protoResource "dynatrace.com/odin/odin-proto/gen/go/resource/v1"
	protoTrace "dynatrace.com/odin/odin-proto/gen/go/trace/v1"
)

func serializeSpans(
	spans dtSpanSet,
	tenantUUID string,
	agentId int64) ([]byte, error) {

	agSpanEnvelopes := make([]*protoCollectorTraces.ActiveGateSpanEnvelope, 0, len(spans))

	for span := range spans {
		fw4Tag := span.metadata.fw4Tag
		customTag := getProtoCustomTag(fw4Tag.CustomBlob)

		spanMsg, err := createProtoSpan(span, customTag)
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

	resource, err := getFirstResource(spans)
	if err != nil {
		return nil, err
	}
	serializedResource, err := getSerializedResourceForSpanExport(resource)
	if err != nil {
		return nil, err
	}

	spanExport := &protoCollectorTraces.SpanExport{
		TenantUUID:     tenantUUID,
		AgentId:        agentId,
		ExportMetaInfo: exportMetaInfo,
		Resource:       serializedResource,
		Spans:          agSpanEnvelopes,
	}
	return proto.Marshal(spanExport)
}

func getFirstResource(spans dtSpanSet) (*resource.Resource, error) {
	for span := range spans {
		readOnlySpan, err := span.readOnlySpan()
		if err != nil {
			return nil, err
		}
		return readOnlySpan.Resource(), nil
	}
	return nil, errors.New("span set is empty, can't retrieve resource")
}

func createProtoSpan(dtSpan *dtSpan, incomingCustomTag *protoTrace.CustomTag) (*protoTrace.Span, error) {
	if dtSpan == nil {
		return nil, errors.New("cannot create proto span from nil dtSpan")
	}

	span, err := dtSpan.readOnlySpan()
	if err != nil {
		return nil, err
	}

	spanMetadata := dtSpan.metadata
	if spanMetadata == nil {
		return nil, errors.New("cannot create proto span when dtSpan metadata is nil")
	}

	spanContext := span.SpanContext()
	traceId := spanContext.TraceID()
	spanId := spanContext.SpanID()
	sendReason, err := getProtoSendReason(spanMetadata.sendState)
	if err != nil {
		return nil, err
	}
	spanMsg := &protoTrace.Span{
		TraceId:          traceId[:],
		SpanId:           spanId[:],
		SendReason:       sendReason,
		UpdateSequenceNo: spanMetadata.seqNumber,
	}

	if sendReason := spanMetadata.sendState; sendReason == sendStateSpanEnded || sendReason == sendStateInitialSend {
		if spanMetadata.tenantParentSpanId.IsValid() {
			spanMsg.TenantParentSpanId = spanMetadata.tenantParentSpanId[:]
		}

		if parentSpanCtx := span.Parent(); parentSpanCtx.IsValid() {
			// This is not a root span and has a parent
			parentSpanId := parentSpanCtx.SpanID()
			spanMsg.ParentSpanId = parentSpanId[:]
		} else {
			// This is a root span
			spanMsg.CustomTag = incomingCustomTag
		}

		spanMsg.Name = span.Name()
		spanMsg.Kind = getProtoSpanKind(span.SpanKind())
		spanMsg.StartTimeUnixnano = uint64(span.StartTime().UnixNano())
		spanMsg.LastPropagateTimeUnixnano = uint64(spanMetadata.lastPropagationTime.UnixNano())

		if sendReason == sendStateSpanEnded {
			spanMsg.EndTimeUnixnano = uint64(span.EndTime().UnixNano())
		}

		spanMsg.Attributes = append(spanMsg.Attributes, createInstrumentationLibAttrs(span)...)
		protoAttributes, err := getProtoAttributes(span.Attributes())
		if err != nil {
			return nil, err
		}
		spanMsg.Attributes = append(spanMsg.Attributes, protoAttributes...)

		protoEvents, err := getProtoEvents(span.Events())
		if err != nil {
			return nil, err
		}
		spanMsg.Events = protoEvents

		protoLinks, err := getProtoLinks(span.Links())
		if err != nil {
			return nil, err
		}
		spanMsg.Links = protoLinks

		status, err := getProtoStatus(span.Status())
		if err != nil {
			return nil, err
		}
		spanMsg.Status = status
	}
	return spanMsg, nil
}

func createSerializedClusterSpanEnvelope(spanMsg *protoTrace.Span, customTag *protoTrace.CustomTag, pathInfo int32) ([]byte, error) {
	spanContainer := protoCollectorTraces.SpanContainer{
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
	clusterSpanEnvelope := protoCollectorTraces.ClusterSpanEnvelope{
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

func createAgSpanEnvelope(clusterSpanEnvelope []byte, serverId int64, traceId []byte) *protoCollectorTraces.ActiveGateSpanEnvelope {
	envelope := &protoCollectorTraces.ActiveGateSpanEnvelope{
		ClusterSpanEnvelope: clusterSpanEnvelope,
	}

	if serverId != 0 {
		envelope.DestinationKey = &protoCollectorTraces.ActiveGateSpanEnvelope_ServerId{ServerId: serverId}
	} else {
		envelope.DestinationKey = &protoCollectorTraces.ActiveGateSpanEnvelope_TraceId{TraceId: traceId}
	}

	return envelope
}

func createInstrumentationLibAttrs(span sdktrace.ReadOnlySpan) []*protoCommon.AttributeKeyValue {
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
