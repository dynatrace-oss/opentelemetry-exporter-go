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

	"core/configuration"
)

type dtTracer struct {
	trace.Tracer
	provider *DtTracerProvider
	config   *configuration.DtConfiguration
}

func (tr *dtTracer) Start(ctx context.Context, name string, options ...trace.SpanStartOption) (context.Context, trace.Span) {
	parentCtx := ctx
	if parentSpan := dtSpanFromContext(ctx); parentSpan != nil {
		parentCtx = trace.ContextWithSpan(ctx, parentSpan.Span)
	}

	sdkCtx, sdkSpan := tr.Tracer.Start(parentCtx, name, options...)
	span := &dtSpan{
		Span:   sdkSpan,
		tracer: tr,
		metadata: createSpanMetadata(
			ctx,
			sdkSpan,
			tr.config.ClusterId,
			tr.config.TenantId(),
			int64(tr.config.SpanProcessingIntervalMs)),
	}

	if sdkSpan.IsRecording() {
		tr.provider.processor.onStart(ctx, span)
	}

	return trace.ContextWithSpan(sdkCtx, span), span
}
