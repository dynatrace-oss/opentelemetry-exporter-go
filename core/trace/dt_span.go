package trace

import (
	"context"
	"errors"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type dtSpan struct {
	trace.Span
	tracer   *dtTracer
	metadata *dtSpanMetadata
}

func (s *dtSpan) End(options ...trace.SpanEndOption) {
	if !s.IsRecording() {
		return
	}

	s.Span.End(options...)
	s.tracer.provider.processor.onEnd(s)
}

func (s *dtSpan) TracerProvider() trace.TracerProvider {
	return s.tracer.provider
}

func (s *dtSpan) readOnlySpan() (sdktrace.ReadOnlySpan, error) {
	if readOnlySpan, ok := s.Span.(sdktrace.ReadOnlySpan); ok {
		return readOnlySpan, nil
	}
	return nil, errors.New("span is not a ReadOnlySpan")
}

// dtSpanFromContext return Dynatrace span instance from given context, nil if Dynatrace span is not found.
func dtSpanFromContext(ctx context.Context) *dtSpan {
	if s := trace.SpanFromContext(ctx); s != nil {
		if span, ok := s.(*dtSpan); ok {
			return span
		}
	}

	return nil
}

// prepareSend evaluates whether a span should be sent to Dynatrace Cluster and updates the metadata accordingly
func (s *dtSpan) prepareSend(sendTime int64) prepareResult {
	// No need to handle error, prepareSend will only ever be called on spans in the watchlist
	// which conform to the ReadWriteSpan interface
	readOnlySpan, _ := s.readOnlySpan()
	sdkSpanEnded := readOnlySpan.EndTime().IsZero()
	shouldSend := s.metadata.evaluateSendState(sendTime, !sdkSpanEnded)
	if shouldSend == prepareResultSend {
		s.metadata.lastSentMs = sendTime
		s.metadata.seqNumber++
	}

	return shouldSend
}
