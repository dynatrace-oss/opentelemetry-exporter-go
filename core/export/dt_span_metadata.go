package export

import (
	"log"
	"sync"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// TODO: move to ODIN config package and replace default values with proper values from config package
const (
	defaultMaxSpansWatchlistSize = 2048
)

type sendState int

const (
	SendStateNew sendState = iota
	SendStateSkip
	SendStateDrop
	SendStateUpdate
	SendStateAlive
	SendStateFinished
)

type dtSpanMetadata struct {
	sendState sendState
	span      sdktrace.ReadOnlySpan
}

func newDtSpanMetadata(span sdktrace.ReadOnlySpan) *dtSpanMetadata {
	return &dtSpanMetadata{
		sendState: SendStateNew,
		span:      span,
	}
}

// TODO: Investigate whether it makes sense to use another entity as a key
type spanKey struct {
	traceID string
	spanID  string
}

func getSpanKey(spanCtx trace.SpanContext) spanKey {
	return spanKey{
		traceID: spanCtx.TraceID().String(),
		spanID:  spanCtx.SpanID().String(),
	}
}

type dtSpanMetadataMap struct {
	spans    map[spanKey]*dtSpanMetadata
	maxSpans int
	lock     sync.Mutex
}

func newDtSpanMetadataMap(maxSpansWatchlistSize int) dtSpanMetadataMap {
	return dtSpanMetadataMap{
		spans:    map[spanKey]*dtSpanMetadata{},
		lock:     sync.Mutex{},
		maxSpans: maxSpansWatchlistSize,
	}
}

func (p *dtSpanMetadataMap) add(spanContext trace.SpanContext, metadata *dtSpanMetadata) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if len(p.spans) >= p.maxSpans {
		log.Println("DtSpanMetaDataMap: spans map is full, could not add new span: " + metadata.span.Name())
		return
	}

	p.spans[getSpanKey(spanContext)] = metadata
}

func (p *dtSpanMetadataMap) remove(spanContext trace.SpanContext) {
	p.lock.Lock()
	defer p.lock.Unlock()

	delete(p.spans, getSpanKey(spanContext))
}

func (p *dtSpanMetadataMap) exist(spanContext trace.SpanContext) bool {
	p.lock.Lock()
	defer p.lock.Unlock()

	_, found := p.spans[getSpanKey(spanContext)]
	return found
}
