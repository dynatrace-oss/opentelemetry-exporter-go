package trace

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"go.opentelemetry.io/otel/trace"

	"core/internal/fw4"
)

type fw4TagKeyType int

const fw4TagKey fw4TagKeyType = iota

func createSpanMetadata(
	ctx context.Context,
	span trace.Span,
	clusterId,
	tenantId int32,
	spanProcessingIntervalMs int64,
) *dtSpanMetadata {
	markParentSpanPropagatedNow(ctx)

	metadata := newDtSpanMetadata(spanProcessingIntervalMs)
	tenantParentSpanId, fw4Tag := extractTenantParentSpanIdAndTagFromParentSpanContext(ctx)
	metadata.tenantParentSpanId = tenantParentSpanId

	// No FW4Tag was found for the parent span, so create one.
	if fw4Tag == nil {
		fw4Tag = createFw4Tag(clusterId, tenantId, span.SpanContext())
	}

	if fw4Tag.ServerID == 0 {
		serverId := getServerIdFromContext(ctx)
		fw4Tag.ServerID = serverId
	}

	metadata.fw4Tag = fw4Tag
	return metadata
}

func extractTenantParentSpanIdAndTagFromParentSpanContext(ctx context.Context) (trace.SpanID, *fw4.Fw4Tag) {
	parentSpan := trace.SpanFromContext(ctx)
	parentSpanContext := parentSpan.SpanContext()

	if parentSpanContext.IsRemote() {
		// For remote parent spans, the FW4 tag is stored in the context, and no metadata will exist.
		if fw4Tag := fw4.Fw4TagFromContext(ctx); fw4Tag != nil {
			return fw4Tag.SpanID, fw4Tag
		}
	} else {
		if parentSpanMetaData := getParentSpanMetadata(parentSpan); parentSpanMetaData != nil {
			return parentSpanContext.SpanID(), parentSpanMetaData.fw4Tag
		}
		return parentSpanContext.SpanID(), nil
	}
	return trace.SpanID{}, nil
}

func getParentSpanMetadata(parentSpan trace.Span) *dtSpanMetadata {
	if parentDtSpan, ok := parentSpan.(*dtSpan); ok {
		return parentDtSpan.metadata
	}
	return nil
}

func getParentSpanMetadataFromContext(ctx context.Context) *dtSpanMetadata {
	parentSpan := trace.SpanFromContext(ctx)
	return getParentSpanMetadata(parentSpan)
}

func markParentSpanPropagatedNow(ctx context.Context) {
	if parentMetadata := getParentSpanMetadataFromContext(ctx); parentMetadata != nil {
		parentMetadata.markPropagatedNow()
	}
}

type pathInfoGenerator struct {
	randomNumberGenerator *rand.Rand
	mutex                 sync.Mutex
}

func (p *pathInfoGenerator) generatePathInfo() uint32 {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	// PathInfo must be an unsigned 32 bit integer
	// whose lowest 8 bits are a pseudo-random number in the range [0, 255]
	return uint32(p.randomNumberGenerator.Intn(256))
}

var pathInfoGeneratorInstance = pathInfoGenerator{
	randomNumberGenerator: rand.New(rand.NewSource(time.Now().UnixNano())),
	mutex:                 sync.Mutex{},
}

func createFw4Tag(clusterId, tenantId int32, spanContext trace.SpanContext) *fw4.Fw4Tag {
	tag := fw4.EmptyTag()
	tag.ClusterID = clusterId
	tag.TenantID = tenantId
	tag.PathInfo = pathInfoGeneratorInstance.generatePathInfo()
	tag.TraceID = spanContext.TraceID()
	tag.SpanID = spanContext.SpanID()
	return &tag
}
