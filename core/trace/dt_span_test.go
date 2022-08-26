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

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"

	"core/internal/fw4"
)

func TestDtSpanEndsSdkSpan(t *testing.T) {
	tp, _ := newDtTracerProviderWithTestExporter()
	tr := tp.Tracer("Dynatrace tracer")

	_, span := tr.Start(context.Background(), "Test span")
	dynatraceSpan := span.(*dtSpan)
	sdkSpan, err := dynatraceSpan.readOnlySpan()

	require.NoError(t, err)
	require.True(t, sdkSpan.EndTime().IsZero())
	span.End()
	require.False(t, sdkSpan.EndTime().IsZero())
}

func TestDtSpanGetTracerProvider(t *testing.T) {
	tp, _ := newDtTracerProviderWithTestExporter()
	tr := tp.Tracer("Dynatrace tracer")
	_, s := tr.Start(context.Background(), "Test span")

	require.Equal(t, tp, s.TracerProvider())
}

func TestDtSpanGetSpanContext(t *testing.T) {
	traceId, err := trace.TraceIDFromHex("11223344556677889900112233445566")
	require.NoError(t, err)

	spanId, err := trace.SpanIDFromHex("8877665544332211")
	require.NoError(t, err)

	var ts trace.TraceState
	ts, err = ts.Insert("custom", "00f067aa0ba902b7")
	require.NoError(t, err)

	config := trace.SpanContextConfig{
		TraceID:    traceId,
		SpanID:     spanId,
		TraceFlags: trace.FlagsSampled,
		TraceState: ts,
	}

	spanCtx := trace.NewSpanContext(config)
	ctx := trace.ContextWithSpanContext(context.Background(), spanCtx)

	spanMetadata := newDtSpanMetadata(0)
	spanMetadata.fw4Tag = fw4.NewFw4Tag(testConfig.ClusterId, testConfig.TenantId(), spanCtx)
	spanMetadata.fw4Tag.PathInfo = 0

	span := &dtSpan{
		Span:     trace.SpanFromContext(ctx),
		metadata: spanMetadata,
	}

	spanCtx = span.SpanContext()
	require.Equal(t, spanCtx.TraceID(), traceId)
	require.Equal(t, spanCtx.SpanID(), spanId)

	// serialized FW4 tag must be added to tracestate
	require.Equal(t, spanCtx.TraceState().String(), "34d2ae74-7b@dt=fw4;0;0;0;0;0;0;0;7db5;2h01;7h8877665544332211,custom=00f067aa0ba902b7")
}
