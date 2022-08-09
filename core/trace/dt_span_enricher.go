package trace

import (
	"context"

	"go.opentelemetry.io/otel/trace"

	"core/internal/fw4"
)

func createSpanMetadata(
	parentCtx context.Context,
	span trace.Span,
	clusterId,
	tenantId int32,
	spanProcessingIntervalMs int64,
) *dtSpanMetadata {
	markParentSpanPropagatedNow(parentCtx)

	metadata := newDtSpanMetadata(spanProcessingIntervalMs)
	tenantParentSpanId, fw4Tag := extractTenantParentSpanIdAndTagFromParentSpanContext(parentCtx)
	metadata.tenantParentSpanId = tenantParentSpanId

	// No FW4Tag was found for the parent span, so create one.
	if fw4Tag == nil {
		fw4Tag = fw4.NewFw4Tag(clusterId, tenantId, span.SpanContext())
	}

	if fw4Tag.ServerID == 0 {
		serverId := getServerIdFromContext(parentCtx)
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
