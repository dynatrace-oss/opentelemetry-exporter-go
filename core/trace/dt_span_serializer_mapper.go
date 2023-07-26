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
	"fmt"

	"github.com/dynatrace-oss/opentelemetry-exporter-go/core/configuration"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"github.com/dynatrace-oss/opentelemetry-exporter-go/core/internal/fw4"
	protoCommon "github.com/dynatrace-oss/opentelemetry-exporter-go/core/internal/odin-proto/common/v1"
	protoTrace "github.com/dynatrace-oss/opentelemetry-exporter-go/core/internal/odin-proto/trace/v1"
)

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

func getProtoCustomTag(customBlob string) *protoTrace.CustomTag {
	if len(customBlob) > 0 {
		// The first char in the blob is the type (see OAAD tagging) and can be directly assigned as enum value.
		firstByte := customBlob[0]
		customTag := protoTrace.CustomTag{
			Type:      protoTrace.CustomTag_Type(firstByte),
			Direction: protoTrace.CustomTag_Incoming,
			TagValue:  []byte(customBlob[1:]),
		}
		return &customTag
	}
	return nil
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

func getProtoSpanAttributes(attributes []attribute.KeyValue, propagatedAttributes propagatedResourceAttributes) ([]*protoCommon.AttributeKeyValue, error) {
	if len(propagatedAttributes) == 0 {
		return getProtoAttributes(attributes)
	}

	protoAttrs := make([]*protoCommon.AttributeKeyValue, 0, len(attributes)+len(propagatedAttributes))
	for _, attr := range propagatedAttributes {
		protoAttr, err := createProtoAttribute(attr)
		if err != nil {
			return nil, err
		}
		protoAttrs = append(protoAttrs, protoAttr)
	}

	for _, attr := range attributes {
		if propagatedAttr, ok := (propagatedAttributes)[attr.Key]; ok {
			if areEqual, err := attributeValueEquals(propagatedAttr.Value, attr.Value); areEqual {
				continue
			} else if err != nil {
				return nil, err
			}

			attr = attribute.KeyValue{
				// Since we support at most 1 overwritten attribute per key here we can hardcode the prefix to index 1
				Key:   attribute.Key("overwritten1." + attr.Key),
				Value: attr.Value,
			}
		}

		protoAttr, err := createProtoAttribute(attr)
		if err != nil {
			return nil, err
		}
		protoAttrs = append(protoAttrs, protoAttr)
	}

	return protoAttrs, nil
}

func attributeValueEquals(first attribute.Value, second attribute.Value) (bool, error) {
	attrType := first.Type()
	if attrType != second.Type() {
		return false, nil
	}

	switch attrType {
	case attribute.BOOL:
		return first.AsBool() == second.AsBool(), nil
	case attribute.INT64:
		return first.AsInt64() == second.AsInt64(), nil
	case attribute.FLOAT64:
		return first.AsFloat64() == second.AsFloat64(), nil
	case attribute.STRING:
		return first.AsString() == second.AsString(), nil
	case attribute.BOOLSLICE:
		return boolSliceEquals(first.AsBoolSlice(), second.AsBoolSlice()), nil
	case attribute.INT64SLICE:
		return int64SliceEquals(first.AsInt64Slice(), second.AsInt64Slice()), nil
	case attribute.FLOAT64SLICE:
		return float64SliceEquals(first.AsFloat64Slice(), second.AsFloat64Slice()), nil
	case attribute.STRINGSLICE:
		return stringSliceEquals(first.AsStringSlice(), second.AsStringSlice()), nil
	default:
		return false, fmt.Errorf("unknown attribute type: %s", attrType)
	}
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

func getProtoLinks(links []sdktrace.Link, qualifiedTenantId configuration.QualifiedTenantId) ([]*protoTrace.Span_Link, error) {
	protoLinks := make([]*protoTrace.Span_Link, 0, len(links))
	for _, link := range links {
		spanContext := link.SpanContext
		traceId := spanContext.TraceID()
		spanId := spanContext.SpanID()
		protoAttributes, err := getProtoAttributes(link.Attributes)
		if err != nil {
			return nil, err
		}

		protoLink := &protoTrace.Span_Link{
			TraceId:                traceId[:],
			SpanId:                 spanId[:],
			Attributes:             protoAttributes,
			DroppedAttributesCount: uint32(link.DroppedAttributeCount),
		}

		if spanContext.IsRemote() {
			fw4Tag, err := fw4.GetMatchingFw4FromTracestate(spanContext.TraceState(), qualifiedTenantId)
			if err != nil {
				return nil, err
			}

			encodedLinkID := fw4Tag.EncodedLinkID()
			protoLink.FwtagEncodedLinkId = &encodedLinkID
		}

		protoLinks = append(protoLinks, protoLink)
	}
	return protoLinks, nil
}
