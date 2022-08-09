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
	if parentMetadata := dtSpanMetadataFromContext(parentCtx); parentMetadata != nil {
		parentMetadata.markPropagatedNow()
	}

	metadata := newDtSpanMetadata(spanProcessingIntervalMs)
	metadata.tenantParentSpanId = tenantParentSpanIdFromContext(parentCtx)

	fw4Tag := fw4TagFromContextOrMetadata(parentCtx)

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

func tenantParentSpanIdFromContext(ctx context.Context) trace.SpanID {
	parentSpanContext := trace.SpanFromContext(ctx).SpanContext()
	if parentSpanContext.IsRemote() {
		if fw4Tag := fw4.Fw4TagFromContext(ctx); fw4Tag != nil {
			return fw4Tag.SpanID
		}
	}
	return parentSpanContext.SpanID()
}

func fw4TagFromContextOrMetadata(ctx context.Context) *fw4.Fw4Tag {
	parentSpan := trace.SpanFromContext(ctx)
	parentSpanContext := parentSpan.SpanContext()
	if parentSpanContext.IsRemote() {
		// For remote parent spans, the FW4 tag is stored in the context, and no metadata will exist.
		return fw4.Fw4TagFromContext(ctx)
	} else {
		if parentSpanMetaData := dtSpanMetadataFromSpan(parentSpan); parentSpanMetaData != nil {
			return parentSpanMetaData.fw4Tag
		}
	}
	return nil
}
