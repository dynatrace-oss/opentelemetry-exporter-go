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

package fw4

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
)

func TestCreateFw4Tag(t *testing.T) {
	clusterId := int32(1)
	tenantId := int32(2)
	traceId, _ := trace.TraceIDFromHex("4bf92f3577b34da6a3ce929d0e0e4736")
	spanId, _ := trace.SpanIDFromHex("00f067aa0ba902b7")
	spanContext := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: traceId,
		SpanID:  spanId,
	})
	tag := NewFw4Tag(clusterId, tenantId, spanContext)
	assert.Equal(t, int32(clusterId), tag.ClusterID)
	assert.Equal(t, int32(tenantId), tag.TenantID)
	assert.Equal(t, spanContext.TraceID(), tag.TraceID)
	assert.Equal(t, spanContext.SpanID(), tag.SpanID)
}

func TestGetFw4TagFromContext_IsNil(t *testing.T) {
	ctx := context.Background()
	tag := Fw4TagFromContext(ctx)
	assert.Nil(t, tag)
}

func TestGetFw4TagFromContext_IsNilWhenNilIsAssigned(t *testing.T) {
	ctx := ContextWithFw4Tag(context.Background(), nil)
	tag := Fw4TagFromContext(ctx)
	assert.Nil(t, tag)
}

func TestGetFw4TagFromContext_IsNotNil(t *testing.T) {
	tag := EmptyTag()
	ctx := ContextWithFw4Tag(context.Background(), &tag)

	tagFromContext := Fw4TagFromContext(ctx)
	assert.NotNil(t, tagFromContext)
	assert.Equal(t, &tag, tagFromContext)
}

func TestFw4Propagate(t *testing.T) {
	tag, err := ParseXDynatrace("FW4;129;1;526;0;0;17;12345;ce1f;2h01;6h11223344556677889900112233445566;7h663055fc5bca216f")
	require.NoError(t, err)

	traceId, err := trace.TraceIDFromHex("89efed6de8a82184abe1f086e818dc61")
	require.NoError(t, err)

	spanId, err := trace.SpanIDFromHex("1122334455667788")
	require.NoError(t, err)

	config := trace.SpanContextConfig{
		TraceID: traceId,
		SpanID:  spanId,
	}

	require.Equal(t, tag.TraceID.String(), "11223344556677889900112233445566")
	require.Equal(t, tag.SpanID.String(), "663055fc5bca216f")

	propTag := tag.Propagate(trace.NewSpanContext(config))

	require.Equal(t, propTag.TraceID, traceId)
	require.Equal(t, propTag.SpanID, spanId)
	require.Equal(t, propTag.TraceStateKey(), "11-81@dt")
	require.Equal(t, propTag.ToTracestateEntryValue(), "fw4;1;0;0;0;1;0;3039;56d8;2h02;6h89efed6de8a82184abe1f086e818dc61;7h1122334455667788")
	require.Equal(t, propTag.ToTracestateEntryValueWithoutTraceId(), "fw4;1;0;0;0;1;0;3039;a591;2h02;7h1122334455667788")
	require.NoError(t, err)
}

func TestFw4SpanContext(t *testing.T) {
	tag, err := ParseXDynatrace("FW4;657;5;15;33;67;113948091;0;e03f;2h01;6h11223344556677889900112233445566;7h8877665544332211")
	require.NoError(t, err)

	ctx := tag.SpanContext()

	require.True(t, ctx.IsSampled())
	require.Equal(t, ctx.TraceID().String(), "11223344556677889900112233445566")
	require.Equal(t, ctx.SpanID().String(), "8877665544332211")
	require.Equal(t, ctx.TraceState().String(), "6cab5bb-291@dt=fw4;5;f;21;43;0;0;0;7db5;2h01;7h8877665544332211")
}

func TestFw4UpdateTracestate(t *testing.T) {
	traceId, err := trace.TraceIDFromHex("11223344556677889900112233445566")
	require.NoError(t, err)

	spanId, err := trace.SpanIDFromHex("8877665544332211")
	require.NoError(t, err)

	ts := trace.TraceState{}
	ts, err = ts.Insert("custom", "00f067aa0ba902b7")
	require.NoError(t, err)

	config := trace.SpanContextConfig{
		TraceID:    traceId,
		SpanID:     spanId,
		TraceState: ts,
	}

	tag, err := ParseXDynatrace("FW4;657;5;15;33;67;113948091;0;c546;2h01;6h732fed6de8a82184abe1f086e818dc61;7h1122334455667788")
	require.NoError(t, err)

	spanCtx := trace.NewSpanContext(config)
	updatedSpanCtx := UpdateTracestate(spanCtx, tag)

	require.Equal(t, spanCtx.TraceState().String(), "custom=00f067aa0ba902b7")
	require.Equal(t, updatedSpanCtx.TraceState().String(), "6cab5bb-291@dt=fw4;5;f;21;43;0;0;0;8d40;2h01;7h1122334455667788,custom=00f067aa0ba902b7")
}

func TestFw4UpdateTraceFlags(t *testing.T) {
	traceId, err := trace.TraceIDFromHex("11223344556677889900112233445566")
	require.NoError(t, err)

	spanId, err := trace.SpanIDFromHex("8877665544332211")
	require.NoError(t, err)

	ts := trace.TraceState{}
	ts, err = ts.Insert("custom", "00f067aa0ba902b7")
	require.NoError(t, err)

	config := trace.SpanContextConfig{
		TraceID:    traceId,
		SpanID:     spanId,
		TraceState: ts,
	}

	tag, err := ParseXDynatrace("FW4;657;5;15;33;67;113948091;0;c546;2h01;6h732fed6de8a82184abe1f086e818dc61;7h1122334455667788")
	require.NoError(t, err)

	spanCtx := trace.NewSpanContext(config)
	updatedSpanCtx := UpdateTraceFlags(spanCtx, tag)

	require.False(t, spanCtx.IsSampled())
	require.True(t, updatedSpanCtx.IsSampled())
}
