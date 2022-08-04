package trace

import (
	"time"

	"core/configuration"
	"core/fw4"

	"go.opentelemetry.io/otel/trace"
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

	fw4Tag              *fw4.Fw4Tag
	lastPropagationTime time.Time
	tenantParentSpanId  trace.SpanID
	serverId            int64
}

type transmitOptions struct {
	updateIntervalMs    int64
	keepAliveIntervalMs int64
	openSpanTimeoutMs   int64
}

func newDtSpanMetadata(spanProcessingIntervalMs int64) *dtSpanMetadata {
	return &dtSpanMetadata{
		sendState:   sendStateNew,
		firstSeenMs: time.Now().UnixNano() / int64(time.Millisecond),
		lastSentMs:  0,
		seqNumber:   -1,
		options: &transmitOptions{
			updateIntervalMs:    spanProcessingIntervalMs,
			keepAliveIntervalMs: configuration.DefaultKeepAliveIntervalMs,
			openSpanTimeoutMs:   configuration.DefaultOpenSpanTimeoutMs,
		},
	}
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
