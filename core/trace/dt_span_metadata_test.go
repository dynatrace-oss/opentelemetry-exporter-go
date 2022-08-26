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
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestDtSpanMetadataSkipNewNonEndedSpan(t *testing.T) {
	tp, _ := newDtTracerProviderWithTestExporter()
	tr := tp.Tracer("Dynatrace Tracer")

	_, span := tr.Start(context.Background(), "Span A")
	s := span.(*dtSpan)

	sendTime := time.Now().UnixNano() / int64(time.Millisecond)
	require.Equal(t, s.prepareSend(sendTime), prepareResultSkip)
	require.EqualValues(t, s.metadata.lastSentMs, 0)
	require.EqualValues(t, s.metadata.seqNumber, -1)
	require.Equal(t, s.metadata.sendState, sendStateNew)
}

func TestDtSpanMetadataSendNonEndedSpanOlderThanUpdateInterval(t *testing.T) {
	tp, _ := newDtTracerProviderWithTestExporter()
	tr := tp.Tracer("Dynatrace Tracer")

	_, span := tr.Start(context.Background(), "Span A")
	s := span.(*dtSpan)

	// new non-ended span must be sent if it is older than update interval
	sendTime := (time.Now().UnixNano() / int64(time.Millisecond)) + s.metadata.options.updateIntervalMs
	require.Equal(t, s.prepareSend(sendTime), prepareResultSend)
	require.Equal(t, s.metadata.lastSentMs, sendTime)
	require.Zero(t, s.metadata.seqNumber)
	require.Equal(t, s.metadata.sendState, sendStateInitialSend)
}

func TestDtSpanMetadataSendEndedSpan(t *testing.T) {
	tp, _ := newDtTracerProviderWithTestExporter()
	tr := tp.Tracer("Dynatrace Tracer")

	_, span := tr.Start(context.Background(), "Span A")
	span.End()
	s := span.(*dtSpan)

	sendTime := time.Now().UnixNano() / int64(time.Millisecond)
	require.Equal(t, s.prepareSend(sendTime), prepareResultSend)
	require.Equal(t, s.metadata.lastSentMs, sendTime)
	require.Zero(t, s.metadata.seqNumber)
	require.Equal(t, s.metadata.sendState, sendStateSpanEnded)
}

func TestDtSpanMetadataDropNonEndedOutdatedSpan(t *testing.T) {
	tp, _ := newDtTracerProviderWithTestExporter()
	tr := tp.Tracer("Dynatrace Tracer")

	_, span := tr.Start(context.Background(), "Span A")
	s := span.(*dtSpan)

	// non-ended span older than openSpanTimeout interval must be dropped
	sendTime := (time.Now().UnixNano() / int64(time.Millisecond)) + s.metadata.options.openSpanTimeoutMs
	require.Equal(t, s.prepareSend(sendTime), prepareResultDrop)
	require.EqualValues(t, s.metadata.lastSentMs, 0)
	require.EqualValues(t, s.metadata.seqNumber, -1)
	require.Equal(t, s.metadata.sendState, sendStateDrop)
}

func TestDtSpanMetadataSendSpanAfterKeepAliveInterval(t *testing.T) {
	tp, _ := newDtTracerProviderWithTestExporter()
	tr := tp.Tracer("Dynatrace Tracer")

	_, span := tr.Start(context.Background(), "Span A")
	s := span.(*dtSpan)

	// new non-ended span must be sent if it is older than update interval
	sendTime := (time.Now().UnixNano() / int64(time.Millisecond)) + s.metadata.options.updateIntervalMs
	require.Equal(t, s.prepareSend(sendTime), prepareResultSend)
	require.Equal(t, s.metadata.lastSentMs, sendTime)
	require.Zero(t, s.metadata.seqNumber)
	require.Equal(t, s.metadata.sendState, sendStateInitialSend)

	// open span must be sent as keep alive after keep alive interval
	sendTime = sendTime + s.metadata.options.keepAliveIntervalMs
	require.Equal(t, s.prepareSend(sendTime), prepareResultSend)
	require.Equal(t, s.metadata.lastSentMs, sendTime)
	require.EqualValues(t, s.metadata.seqNumber, 1)
	require.Equal(t, s.metadata.sendState, sendStateAlive)
}

func TestDtSpanContainsMetadata(t *testing.T) {
	tp, _ := newDtTracerProviderWithTestExporter()
	tr := tp.Tracer("Dynatrace tracer")
	_, span := tr.Start(context.Background(), "Test span")
	s := span.(*dtSpan)

	require.NotNil(t, s.metadata)
	require.Equal(t, s.metadata.sendState, sendStateNew)
	require.EqualValues(t, s.metadata.lastSentMs, 0)
}

func TestNotSampledDtSpanContainMetadata(t *testing.T) {
	sampler := sdktrace.WithSampler(sdktrace.NeverSample())
	tp := NewTracerProvider(sampler)
	tp.processor.exporter = newTestExporter(testExporterOptions{
		iterationIntervalMs: 500,
		numIterations:       1,
	})

	tr := tp.Tracer("Dynatrace tracer")
	_, span := tr.Start(context.Background(), "Test span")
	s := span.(*dtSpan)

	require.NotNil(t, s.metadata)
}
