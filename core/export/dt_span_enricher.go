package export

import (
	"log"
	"math/rand"
	"time"

	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"core/fw4"
)

type dtSpanEnricher struct{}

func (se *dtSpanEnricher) CreateSpanMetaData(
	span sdktrace.ReadWriteSpan,
	transmitOptions *transmitOptions,
	clusterId,
	tenantId int32,
	metadataMap *dtSpanMetadataMap,
) *dtSpanMetadata {
	metadata := newDtSpanMetadata(transmitOptions, span)

	parentSpanContext := span.Parent()
	shouldCreateFw4Tag, tenantParentSpanId := extractMetaDataFromParentSpanContext(parentSpanContext, metadataMap)

	metadata.tenantParentSpanId = tenantParentSpanId

	if shouldCreateFw4Tag {
		metadata.fw4Tag = createFw4Tag(clusterId, tenantId)
	}

	// Set serverId if not provided and not yet set. Then, set the XDtc attribute if provided.
	fw4Tag := getFw4Tag(metadata)
	if fw4Tag != nil && fw4Tag.ServerID == 0 {
		if serverId := metadata.serverId; serverId != -1 {
			fw4Tag.ServerID = int32(serverId)
			xDtc := metadata.xDtc
			// TODO "dt.rum.dtc" string literal should come from SemConv constants instead
			span.SetAttributes(attribute.String("dt.rum.dtc", xDtc))
		}
	}

	return metadata
}

func extractMetaDataFromParentSpanContext(
	parentSpanContext trace.SpanContext,
	metadataMap *dtSpanMetadataMap,
) (shouldCreateFw4Tag bool, tenantParentSpanId trace.SpanID) {
	shouldCreateFw4Tag = true
	parentSpanMetaData := metadataMap.get(parentSpanContext)

	if parentSpanContext.IsRemote() {
		if fw4Tag := getFw4Tag(parentSpanMetaData); fw4Tag != nil {
			shouldCreateFw4Tag = false
			tenantParentSpanId = getTenantParentSpanIdFromFw4Tag(fw4Tag)
		}
	} else {
		tenantParentSpanId = parentSpanContext.SpanID()
		markSpanPropagatedNow(parentSpanContext, metadataMap)

		if getFw4Tag(parentSpanMetaData) != nil {
			shouldCreateFw4Tag = false
		} else {
			log.Printf("Expected FW4 tag to exist for span (traceId: %s, spanId: %s) but none found.\n",
				parentSpanContext.TraceID(), parentSpanContext.SpanID())
		}
	}

	return shouldCreateFw4Tag, tenantParentSpanId
}

func getFw4Tag(metaData *dtSpanMetadata) *fw4.Fw4Tag {
	if metaData != nil {
		return metaData.fw4Tag
	}
	return nil
}

func getTenantParentSpanIdFromFw4Tag(fw4Tag *fw4.Fw4Tag) trace.SpanID {
	return fw4Tag.SpanID
}

func createFw4Tag(clusterId, tenantId int32) *fw4.Fw4Tag {
	tag := fw4.EmptyTag()
	tag.ClusterID = clusterId
	tag.TenantID = tenantId
	// Set lowest 8 bits of PathInfo to a pseudo-random number in the range [0, 255]
	tag.PathInfo = uint32(rand.Intn(256))
	return &tag
}

func markSpanPropagatedNow(spanContext trace.SpanContext, metadataMap *dtSpanMetadataMap) {
	if metaData := metadataMap.get(spanContext); metaData != nil {
		metaData.lastPropagationTime = time.Now()
	}
}
