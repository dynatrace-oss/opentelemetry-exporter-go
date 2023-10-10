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
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/protobuf/proto"

	"github.com/dynatrace-oss/opentelemetry-exporter-go/core/configuration"
	"github.com/dynatrace-oss/opentelemetry-exporter-go/core/internal/logger"
	protoCollectorCommon "github.com/dynatrace-oss/opentelemetry-exporter-go/core/internal/odin-proto/collector/common/v1"
	protoCollectorTraces "github.com/dynatrace-oss/opentelemetry-exporter-go/core/internal/odin-proto/collector/traces/v1"
	protoCommon "github.com/dynatrace-oss/opentelemetry-exporter-go/core/internal/odin-proto/common/v1"
	protoResource "github.com/dynatrace-oss/opentelemetry-exporter-go/core/internal/odin-proto/resource/v1"
	protoTrace "github.com/dynatrace-oss/opentelemetry-exporter-go/core/internal/odin-proto/trace/v1"
	"github.com/dynatrace-oss/opentelemetry-exporter-go/core/internal/semconv"
	"github.com/dynatrace-oss/opentelemetry-exporter-go/core/internal/version"
)

const (
	cMsgSizeMax  = 64 * 1024 * 1024 // 64 MB
	cMsgSizeWarn = 1 * 1024 * 1024  // 1 MB
)

type exportData []byte

type dtSpanSerializer struct {
	logger            *logger.ComponentLogger
	tenantUUID        string
	agentId           int64
	qualifiedTenantId configuration.QualifiedTenantId
}

func newSpanSerializer(
	tenantUUID string,
	agentId int64,
	qualifiedTenantId configuration.QualifiedTenantId) *dtSpanSerializer {
	return &dtSpanSerializer{
		logger:            logger.NewComponentLogger("SpanSerializer"),
		tenantUUID:        tenantUUID,
		agentId:           agentId,
		qualifiedTenantId: qualifiedTenantId,
	}
}

// serializeSpans serializes the spans into one or multiple SpanExport messages.
// Uses a "Next Fit" bin-packing algorithm.
// The spans are serialized in order and SpanExport messages are sent to the exportChannel.
// If an error occurs, the error is sent to the error channel and the function returns.
func (s *dtSpanSerializer) serializeSpans(spans dtSpanSet, exportChannel chan exportData, errorChannel chan error) {
	exportMetaInfo, err := proto.Marshal(&protoCollectorCommon.ExportMetaInfo{
		TimeSyncMode: protoCollectorCommon.ExportMetaInfo_NTPSync,
	})
	if err != nil {
		errorChannel <- err
		return
	}

	firstSpanResource, err := getFirstResource(spans)
	if err != nil {
		errorChannel <- err
		return
	}
	serializedResource, err := getSerializedResourceForSpanExport(firstSpanResource)
	if err != nil {
		errorChannel <- err
		return
	}

	spanExport := &protoCollectorTraces.SpanExport{
		TenantUUID:     s.tenantUUID,
		AgentId:        s.agentId,
		ExportMetaInfo: exportMetaInfo,
		Resource:       serializedResource,
	}

	spanlessMsgSize := proto.Size(spanExport)

	s.logger.Debugf("spanless message size: %v", spanlessMsgSize)

	if spanlessMsgSize > cMsgSizeMax {
		err = fmt.Errorf("resource too big (%v), cannot export any spans", spanlessMsgSize)
		errorChannel <- err
		return
	}

	sizeSoFar := spanlessMsgSize

	agSpanEnvelopes := make([]*protoCollectorTraces.ActiveGateSpanEnvelope, 0, len(spans))

	export := func(exp *protoCollectorTraces.SpanExport) error {
		serializedExport, err := proto.Marshal(exp)
		if err != nil {
			return err
		}
		exportChannel <- serializedExport
		return nil
	}

	for span := range spans {
		fw4Tag := span.metadata.fw4Tag
		customTag := getProtoCustomTag(fw4Tag.CustomBlob)

		spanMsg, err := createProtoSpan(span, customTag, s.qualifiedTenantId)
		if err != nil {
			errorChannel <- err
			return
		}

		serializedClusterSpanEnvelope, err := createSerializedClusterSpanEnvelope(spanMsg, customTag, int32(fw4Tag.PathInfo))
		if err != nil {
			errorChannel <- err
			return
		}

		agSpanEnvelope := createAgSpanEnvelope(serializedClusterSpanEnvelope, int64(fw4Tag.ServerID), spanMsg.TraceId)

		// Estimate 1 byte for the tag size and up to 4 byte for the varint
		// encoding the size of the cluster envelope.
		estimatedEnvelopeSize := proto.Size(agSpanEnvelope) + 1 + 4

		if sizeSoFar+estimatedEnvelopeSize > cMsgSizeWarn {
			if minSize := spanlessMsgSize + estimatedEnvelopeSize; minSize > cMsgSizeMax {
				// DROP: The size of this span + export msg is too big to ever fit, so we drop this span altogether
				// and try the next span
				s.logger.Warnf("span too big (%v), dropping", minSize)
				continue
			}

			// BUFFER: The size exceeds the desired size AND the export already contains a span,
			// so we buffer the current span into the next envelope
			if len(agSpanEnvelopes) > 0 {
				s.logger.Debugf("size (%v) exceeds desired size, creating new span export", sizeSoFar+estimatedEnvelopeSize)

				// export the previous spanExport
				if err := export(spanExport); err != nil {
					errorChannel <- err
					return
				}

				// Create a new SpanExport in which to fit the overhanging span
				spanExport = &protoCollectorTraces.SpanExport{
					TenantUUID:     s.tenantUUID,
					AgentId:        s.agentId,
					ExportMetaInfo: exportMetaInfo,
					Resource:       serializedResource,
				}
				agSpanEnvelopes = make([]*protoCollectorTraces.ActiveGateSpanEnvelope, 0)
				sizeSoFar = spanlessMsgSize
			}
		}

		// ADD: Add the span to the SpanExport
		sizeSoFar += estimatedEnvelopeSize
		agSpanEnvelopes = append(agSpanEnvelopes, agSpanEnvelope)
		spanExport.Spans = agSpanEnvelopes
	}

	if err := export(spanExport); err != nil {
		errorChannel <- err
		return
	}
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

func createProtoSpan(dtSpan *dtSpan, incomingCustomTag *protoTrace.CustomTag, qualifiedTenantId configuration.QualifiedTenantId) (*protoTrace.Span, error) {
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

			if parentSpanCtx.IsRemote() {
				encodedLinkId := spanMetadata.fw4Tag.EncodedLinkID()
				spanMsg.ParentFwtagEncodedLinkId = &encodedLinkId
			}

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
		protoAttributes, err := getProtoSpanAttributes(span.Attributes(), spanMetadata.propagatedResourceAttributes)
		if err != nil {
			return nil, err
		}
		spanMsg.Attributes = append(spanMsg.Attributes, protoAttributes...)

		protoEvents, err := getProtoEvents(span.Events())
		if err != nil {
			return nil, err
		}
		spanMsg.Events = protoEvents

		protoLinks, err := getProtoLinks(span.Links(), qualifiedTenantId)
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
	// TODO we want to do this later??
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
