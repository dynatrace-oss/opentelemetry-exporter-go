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
	"context"

	"go.opentelemetry.io/otel/trace"

	"core/internal/fw4"
	"core/trace/internal/util"
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
		fw4Tag.ServerID = util.GetServerIdFromContext(parentCtx)
	}

	metadata.setFw4Tag(fw4Tag)
	return metadata
}

func tenantParentSpanIdFromContext(ctx context.Context) trace.SpanID {
	parentSpanContext := trace.SpanFromContext(ctx).SpanContext()
	if parentSpanContext.IsRemote() {
		if fw4Tag := fw4.Fw4TagFromContext(ctx); fw4Tag != nil {
			return fw4Tag.SpanID
		}
	} else {
		return parentSpanContext.SpanID()
	}

	return trace.SpanID{}
}

func fw4TagFromContextOrMetadata(ctx context.Context) *fw4.Fw4Tag {
	parentSpan := trace.SpanFromContext(ctx)
	if parentSpan.SpanContext().IsRemote() {
		// For remote parent spans, the FW4 tag is stored in the context, and no metadata will exist.
		return fw4.Fw4TagFromContext(ctx)
	} else if parentSpanMetaData := dtSpanMetadataFromSpan(parentSpan); parentSpanMetaData != nil {
		return parentSpanMetaData.getFw4Tag()
	}
	return nil
}
