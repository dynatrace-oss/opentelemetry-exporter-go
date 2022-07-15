package export

import (
	"log"
	"sync"
	"time"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"core/fw4"
)

// TODO: move to ODIN config package and replace default values with proper values from config package
const (
	defaultMaxSpansWatchlistSize = 2048
)

type sendState int

const (
	sendStateNew sendState = iota
	sendStateSkip
	sendStateDrop
	sendStateInitialSend
	sendStateAlive
	sendStateSpanEnded
)

type prepareResult int

const (
	prepareResultDrop prepareResult = iota
	prepareResultSend
	prepareResultSkip
)

type dtSpanMetadata struct {
	sendState   sendState
	firstSeenMs int64
	lastSentMs  int64
	seqNumber   int32
	options     *transmitOptions
	span        sdktrace.ReadOnlySpan

	fw4Tag              *fw4.Fw4Tag
	lastPropagationTime time.Time
	tenantParentSpanId  trace.SpanID
	serverId            int64
	xDtc                string
}

type transmitOptions struct {
	updateIntervalMs    int64
	keepAliveIntervalMs int64
	openSpanTimeoutMs   int64
}

func newDtSpanMetadata(transmitOptions *transmitOptions, span sdktrace.ReadOnlySpan) *dtSpanMetadata {
	return &dtSpanMetadata{
		sendState:   sendStateNew,
		firstSeenMs: time.Now().UnixNano() / int64(time.Millisecond),
		lastSentMs:  0,
		seqNumber:   -1,
		options: transmitOptions,
		span:    span,
	}
}

// prepareSend evaluates whether a span should be sent to Dynatrace Cluster and updates the last sent timestamp,
// sequence number and send state accordingly
func (p *dtSpanMetadata) prepareSend(sendTime int64) prepareResult {
	shouldSend := p.evaluateSendState(sendTime, !p.span.EndTime().IsZero())
	if shouldSend == prepareResultSend {
		p.lastSentMs = sendTime
		p.seqNumber++
	}

	return shouldSend
}

func (p *dtSpanMetadata) evaluateSendState(sendTime int64, isFinished bool) prepareResult {
	// always send finished spans
	if isFinished {
		p.sendState = sendStateSpanEnded
		return prepareResultSend
	}

	// drop outdated, non-finished spans
	spanAgeMs := sendTime - p.firstSeenMs
	if spanAgeMs >= p.options.openSpanTimeoutMs {
		p.sendState = sendStateDrop
		return prepareResultDrop
	}

	// new, non-finished spans should be sent only if they are old enough
	if p.sendState == sendStateNew {
		if spanAgeMs >= p.options.updateIntervalMs {
			p.sendState = sendStateInitialSend
			return prepareResultSend
		}

		return prepareResultSkip
	}

	// open spans should be sent as keep alive
	if sendTime-p.lastSentMs >= p.options.keepAliveIntervalMs {
		p.sendState = sendStateAlive
		return prepareResultSend
	}

	p.sendState = sendStateSkip
	return prepareResultSkip
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

func (p *dtSpanMetadataMap) get(spanContext trace.SpanContext) *dtSpanMetadata {
	p.lock.Lock()
	defer p.lock.Unlock()

	return p.spans[getSpanKey(spanContext)]
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

// getSpansToExport evaluates available spans and returns those that have to be sent
func (p *dtSpanMetadataMap) getSpansToExport() map[spanKey]*dtSpanMetadata {
	p.lock.Lock()
	// make a copy of the spans map to avoid keeping the lock for a longer period of time
	spansToExport := make(map[spanKey]*dtSpanMetadata, len(p.spans))
	for k, v := range p.spans {
		spansToExport[k] = v
	}
	p.lock.Unlock()

	now := time.Now().UnixNano() / int64(time.Millisecond)
	for key, metadata := range spansToExport {
		prepareResult := metadata.prepareSend(now)

		if prepareResult == prepareResultSkip {
			delete(spansToExport, key)
		}

		if prepareResult == prepareResultDrop || (prepareResult == prepareResultSend && metadata.sendState == sendStateSpanEnded) {
			// remove dropped or ended spans from the span processor map, so they will not be watched anymore
			p.remove(metadata.span.SpanContext())
		}
	}

	return spansToExport
}
