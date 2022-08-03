package trace

import (
	"context"
	"math/rand"
	"time"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"core/fw4"
)

type fw4TagKeyType int
const fw4TagKey fw4TagKeyType = iota

func createSpanMetadata(
	ctx context.Context,
	span sdktrace.ReadOnlySpan,
	clusterId,
	tenantId int32,
) *dtSpanMetadata {
	metadata := newDtSpanMetadata()

	tenantParentSpanId, fw4Tag := extractTenantParentSpanIdFromParentSpanContext(span, ctx)
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

func extractTenantParentSpanIdFromParentSpanContext(span sdktrace.ReadOnlySpan, ctx context.Context) (trace.SpanID, *fw4.Fw4Tag) {
	parentSpanContext := span.Parent()
	parentSpanMetaData := getParentSpanMetadata(ctx)

	if parentSpanContext.IsRemote() {
		// For remote parent spans, the FW4 tag is stored in the context, and no metadata will exist.
		if fw4Tag := getFw4TagFromContext(ctx); fw4Tag != nil {
			return fw4Tag.SpanID, fw4Tag
		}
	} else {
		if parentSpanMetaData != nil {
			parentSpanMetaData.lastPropagationTime = time.Now()
			return parentSpanContext.SpanID(), parentSpanMetaData.fw4Tag
		}
		return parentSpanContext.SpanID(), nil
	}
	return trace.SpanID{}, nil
}

func getParentSpanMetadata(ctx context.Context) *dtSpanMetadata {
	parentSpan := trace.SpanFromContext(ctx)
	if parentDtSpan, ok := parentSpan.(*dtSpan); ok {
		return parentDtSpan.metadata
	}
	return nil
}

func getFw4TagFromContext(ctx context.Context) *fw4.Fw4Tag {
	if fw4Tag, ok := ctx.Value(fw4TagKey).(*fw4.Fw4Tag); ok {
		return fw4Tag
	}
	return nil
}

var random = rand.New(rand.NewSource(time.Now().UnixNano()))

func createFw4Tag(clusterId, tenantId int32, spanContext trace.SpanContext) *fw4.Fw4Tag {
	tag := fw4.EmptyTag()
	tag.ClusterID = clusterId
	tag.TenantID = tenantId
	// Set lowest 8 bits of PathInfo to a pseudo-random number in the range [0, 255]
	tag.PathInfo = uint32(random.Intn(256))
	tag.TraceID = spanContext.TraceID()
	tag.SpanID = spanContext.SpanID()
	return &tag
}