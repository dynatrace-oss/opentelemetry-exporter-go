// Copyright 2022 Dynatrace LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package trace

import (
	"errors"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/protobuf/proto"

	protoCollectorCommon "github.com/dynatrace-oss/opentelemetry-exporter-go/core/internal/odin-proto/collector/common/v1"
	protoCollectorTraces "github.com/dynatrace-oss/opentelemetry-exporter-go/core/internal/odin-proto/collector/traces/v1"
	protoCommon "github.com/dynatrace-oss/opentelemetry-exporter-go/core/internal/odin-proto/common/v1"
	protoResource "github.com/dynatrace-oss/opentelemetry-exporter-go/core/internal/odin-proto/resource/v1"
	protoTrace "github.com/dynatrace-oss/opentelemetry-exporter-go/core/internal/odin-proto/trace/v1"
	"github.com/dynatrace-oss/opentelemetry-exporter-go/core/internal/semconv"
	"github.com/dynatrace-oss/opentelemetry-exporter-go/core/internal/version"
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

	firstSpanResource, err := getFirstResource(spans)
	if err != nil {
		return nil, err
	}
	serializedResource, err := getSerializedResourceForSpanExport(firstSpanResource)
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
		Key:         semconv.OtelLibraryName,
		StringValue: instrumentationLib.Name,
	}
	instrumentationLibVersionAttr := &protoCommon.AttributeKeyValue{
		Type:        protoCommon.AttributeKeyValue_STRING,
		Key:         semconv.OtelLibraryVersion,
		StringValue: instrumentationLib.Version,
	}
	return []*protoCommon.AttributeKeyValue{instrumentationLibNameAttr, instrumentationLibVersionAttr}
}

func getResourceForSpanExport(spanResource *resource.Resource) (*resource.Resource, error) {
	mergedResource, err := resource.Merge(spanResource, getExporterResource())
	if err != nil {
		return nil, err
	}

	// Ensure the exported resource contains the following attributes from resource.Default():
	// - telemetry.sdk.language
	// - telemetry.sdk.name
	// - telemetry.sdk.version
	mergedResourceWithDefaults, err := resource.Merge(resource.Default(), mergedResource)
	if err != nil {
		return nil, err
	}
	return mergedResourceWithDefaults, nil
}

func getSerializedResourceForSpanExport(spanResource *resource.Resource) ([]byte, error) {
	resourceForExport, err := getResourceForSpanExport(spanResource)
	if err != nil {
		return nil, err
	}
	protoAttributes, err := getProtoAttributes(resourceForExport.Attributes())
	if err != nil {
		return nil, err
	}
	res := protoResource.Resource{
		Attributes: protoAttributes,
	}
	return proto.Marshal(&res)
}

func getExporterResource() *resource.Resource {
	return resource.NewSchemaless(
		attribute.Key(semconv.TelemetryExporterName).String(semconv.TelemetryExporterNameOdin),
		attribute.Key(semconv.TelemetryExporterVersion).String(version.FullVersion),
	)
}
