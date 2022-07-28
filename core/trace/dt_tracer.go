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
		Span:     sdkSpan,
		tracer:   tr,
		metadata: newDtSpanMetadata(int64(tr.config.SpanProcessingIntervalMs)),
	}

	if sdkSpan.IsRecording() {
		tr.provider.processor.onStart(ctx, span)
	}

	return trace.ContextWithSpan(sdkCtx, span), span
}
