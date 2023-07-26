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
	"sync"
	"time"

	"github.com/dynatrace-oss/opentelemetry-exporter-go/core/configuration"
	"github.com/dynatrace-oss/opentelemetry-exporter-go/core/internal/fw4"

	"go.opentelemetry.io/otel/attribute"
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
type propagatedResourceAttributes map[attribute.Key]attribute.KeyValue

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

	propagatedResourceAttributes propagatedResourceAttributes

	mutex sync.Mutex
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

func (p *dtSpanMetadata) markPropagatedNow() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.lastPropagationTime = time.Now()
}

func dtSpanMetadataFromSpan(parentSpan trace.Span) *dtSpanMetadata {
	if parentDtSpan, ok := parentSpan.(*dtSpan); ok {
		return parentDtSpan.metadata
	}
	return nil
}

func dtSpanMetadataFromContext(ctx context.Context) *dtSpanMetadata {
	parentSpan := trace.SpanFromContext(ctx)
	return dtSpanMetadataFromSpan(parentSpan)
}

func (p *dtSpanMetadata) getFw4Tag() *fw4.Fw4Tag {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.fw4Tag
}

func (p *dtSpanMetadata) setFw4Tag(tag *fw4.Fw4Tag) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.fw4Tag = tag
}
