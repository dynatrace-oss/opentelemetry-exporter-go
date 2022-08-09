package trace

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"core/configuration"
	"core/internal/fw4"
)

const (
	invalidTraceparent = "00-00000000000000000000000000000000-0000000000000000-00"
)

func TestPropagatorExtractInvalidSpanContext(t *testing.T) {
	p := NewTextMapPropagator()
	require.NotNil(t, p)

	c := propagation.HeaderCarrier{}
	c.Set(traceparentHeader, invalidTraceparent)
	ctx := p.Extract(context.Background(), c)

	require.Equal(t, ctx, context.Background())
}

func TestPropagatorExtractEmptyTaggingHeaders(t *testing.T) {
	p := NewTextMapPropagator()
	require.NotNil(t, p)

	c := propagation.HeaderCarrier{}
	ctx := p.Extract(context.Background(), c)

	require.Equal(t, ctx, context.Background())
}

func TestPropagatorExtractInvalidXDynatrace(t *testing.T) {
	p := NewTextMapPropagator()
	require.NotNil(t, p)

	c := propagation.HeaderCarrier{}
	c.Set(xDtHeader, "FW4;123;5;15;33;67;886222452;0")

	ctx := p.Extract(context.Background(), c)

	require.Equal(t, ctx, context.Background())
}

func TestPropagatorExtractOnlyXDynatrace(t *testing.T) {
	p := NewTextMapPropagator()
	require.NotNil(t, p)

	c := propagation.HeaderCarrier{}
	c.Set(xDtHeader, "FW4;123;5;15;33;-2147483581;886222452;0;e03f;2h01;6h11223344556677889900112233445566;7h8877665544332211")

	ctx := p.Extract(context.Background(), c)
	span := trace.SpanFromContext(ctx)
	require.NotNil(t, span)

	// check span context
	spanCtx := span.SpanContext()
	require.Equal(t, spanCtx.TraceID().String(), "11223344556677889900112233445566")
	require.Equal(t, spanCtx.SpanID().String(), "8877665544332211")
	require.True(t, spanCtx.IsRemote())
	require.Zero(t, spanCtx.TraceFlags())
	require.Equal(t, spanCtx.TraceState().String(), "34d2ae74-7b@dt=fw4;5;f;21;43;1;0;0;7db5;2h01;7h8877665544332211")

	// check FW4 tag
	tag := fw4.Fw4TagFromContext(ctx)
	require.NotNil(t, tag)
	require.EqualValues(t, tag.AgentID, 15)
	require.EqualValues(t, tag.TagID, 33)
	require.EqualValues(t, tag.LinkID(), 67)
	require.EqualValues(t, tag.ServerID, 5)
	require.EqualValues(t, tag.ClusterID, 123)
	require.EqualValues(t, tag.TenantID, 886222452)
	require.True(t, tag.IsIgnored())
	require.Equal(t, tag.TraceID.String(), "11223344556677889900112233445566")
	require.Equal(t, tag.SpanID.String(), "8877665544332211")
}

func TestPropagatorExtractXDynatraceWithInvalidTraceIdSpanId(t *testing.T) {
	p := NewTextMapPropagator()
	require.NotNil(t, p)

	c := propagation.HeaderCarrier{}
	c.Set(xDtHeader, "FW4;123;5;15;33;-2147483581;886222452;0;e03f;2h01;6h11223344556677889900112233445566;7h8877665544332211")
	c.Set(traceparentHeader, invalidTraceparent)

	ctx := p.Extract(context.Background(), c)
	span := trace.SpanFromContext(ctx)
	require.NotNil(t, span)

	// check span context
	spanCtx := span.SpanContext()
	require.Equal(t, spanCtx.TraceID().String(), "11223344556677889900112233445566")
	require.Equal(t, spanCtx.SpanID().String(), "8877665544332211")
	require.True(t, spanCtx.IsRemote())
	require.Zero(t, spanCtx.TraceFlags())
	require.Equal(t, spanCtx.TraceState().String(), "34d2ae74-7b@dt=fw4;5;f;21;43;1;0;0;7db5;2h01;7h8877665544332211")

	// check FW4 tag
	tag := fw4.Fw4TagFromContext(ctx)
	require.NotNil(t, tag)
	require.EqualValues(t, tag.AgentID, 15)
	require.EqualValues(t, tag.TagID, 33)
	require.EqualValues(t, tag.LinkID(), 67)
	require.EqualValues(t, tag.ServerID, 5)
	require.EqualValues(t, tag.ClusterID, 123)
	require.EqualValues(t, tag.TenantID, 886222452)
	require.True(t, tag.IsIgnored())
	require.Equal(t, tag.TraceID.String(), "11223344556677889900112233445566")
	require.Equal(t, tag.SpanID.String(), "8877665544332211")
}

func TestPropagatorExtractXDynatracePreferredMismatchTraceId(t *testing.T) {
	p := NewTextMapPropagator()
	require.NotNil(t, p)

	c := propagation.HeaderCarrier{}
	c.Set(xDtHeader, "FW4;123;5;15;33;67;886222452;0;e03f;2h01;6h11223344556677889900112233445566;7h8877665544332211")
	c.Set(traceparentHeader, "00-22334455667788990011223344556677-9988776655443322-00")
	c.Set(tracestateHeader, "34d2ae74-7b@dt=fw4;fffffffd;0;0;0;0;0;0;8355;2h01;7haaffffeebbaabbee,custom=00f067aa0ba902b7")

	ctx := p.Extract(context.Background(), c)

	span := trace.SpanFromContext(ctx)
	require.NotNil(t, span)

	// check span context
	spanCtx := span.SpanContext()
	require.Equal(t, spanCtx.TraceID().String(), "11223344556677889900112233445566")
	require.Equal(t, spanCtx.SpanID().String(), "8877665544332211")
	require.True(t, spanCtx.IsRemote())
	require.Equal(t, spanCtx.TraceFlags(), trace.FlagsSampled)
	// 'custom' tracestate entry is removed due to traceId mismatch
	require.Equal(t, spanCtx.TraceState().String(), "34d2ae74-7b@dt=fw4;5;f;21;43;0;0;0;7db5;2h01;7h8877665544332211")

	// check FW4 tag
	tag := fw4.Fw4TagFromContext(ctx)
	require.NotNil(t, tag)
	require.EqualValues(t, tag.AgentID, 15)
	require.EqualValues(t, tag.TagID, 33)
	require.EqualValues(t, tag.LinkID(), 67)
	require.EqualValues(t, tag.ServerID, 5)
	require.EqualValues(t, tag.ClusterID, 123)
	require.EqualValues(t, tag.TenantID, 886222452)
	require.False(t, tag.IsIgnored())
	require.Equal(t, tag.TraceID.String(), "11223344556677889900112233445566")
	require.Equal(t, tag.SpanID.String(), "8877665544332211")
}

func TestPropagatorExtractXDynatracePreferredMatchingTraceId(t *testing.T) {
	p := NewTextMapPropagator()
	require.NotNil(t, p)

	c := propagation.HeaderCarrier{}
	c.Set(xDtHeader, "FW4;123;5;15;33;67;886222452;0;e03f;2h01;6h11223344556677889900112233445566;7h8877665544332211")
	c.Set(traceparentHeader, "00-11223344556677889900112233445566-aaffffeebbaabbee-00")
	c.Set(tracestateHeader, "34d2ae74-7b@dt=fw4;fffffffd;0;0;0;0;0;0;8355;2h01;7haaffffeebbaabbee,custom=00f067aa0ba902b7")

	ctx := p.Extract(context.Background(), c)

	span := trace.SpanFromContext(ctx)
	require.NotNil(t, span)

	// check span context
	spanCtx := span.SpanContext()
	require.Equal(t, spanCtx.TraceID().String(), "11223344556677889900112233445566")
	require.Equal(t, spanCtx.SpanID().String(), "aaffffeebbaabbee")
	require.True(t, spanCtx.IsRemote())
	require.Equal(t, spanCtx.TraceFlags(), trace.FlagsSampled)
	// 'custom' tracestate entry remains since traceId is matching
	require.Equal(t, spanCtx.TraceState().Get("34d2ae74-7b@dt"), "fw4;5;f;21;43;0;0;0;7db5;2h01;7h8877665544332211")
	require.Equal(t, spanCtx.TraceState().Get("custom"), "00f067aa0ba902b7")

	// check FW4 tag
	tag := fw4.Fw4TagFromContext(ctx)
	require.NotNil(t, tag)
	require.EqualValues(t, tag.AgentID, 15)
	require.EqualValues(t, tag.TagID, 33)
	require.EqualValues(t, tag.LinkID(), 67)
	require.EqualValues(t, tag.ServerID, 5)
	require.EqualValues(t, tag.ClusterID, 123)
	require.EqualValues(t, tag.TenantID, 886222452)
	require.Equal(t, tag.TraceID.String(), "11223344556677889900112233445566")
	require.Equal(t, tag.SpanID.String(), "8877665544332211")
}

func TestPropagatorExtractXDynatraceTraceparentWithoutTracestate(t *testing.T) {
	p := NewTextMapPropagator()
	require.NotNil(t, p)

	c := propagation.HeaderCarrier{}
	c.Set(xDtHeader, "FW4;123;5;15;33;67;886222452;0;e03f;2h01;6h11223344556677889900112233445566;7h8877665544332211")
	c.Set(traceparentHeader, "00-11223344556677889900112233445566-aaffffeebbaabbee-00")

	ctx := p.Extract(context.Background(), c)

	span := trace.SpanFromContext(ctx)
	require.NotNil(t, span)

	// check span context
	spanCtx := span.SpanContext()
	require.Equal(t, spanCtx.TraceID().String(), "11223344556677889900112233445566")
	require.Equal(t, spanCtx.SpanID().String(), "aaffffeebbaabbee")
	require.True(t, spanCtx.IsRemote())
	require.Equal(t, spanCtx.TraceFlags(), trace.FlagsSampled)
	require.Equal(t, spanCtx.TraceState().Get("34d2ae74-7b@dt"), "fw4;5;f;21;43;0;0;0;7db5;2h01;7h8877665544332211")

	// check FW4 tag
	tag := fw4.Fw4TagFromContext(ctx)
	require.NotNil(t, tag)
	require.EqualValues(t, tag.AgentID, 15)
	require.EqualValues(t, tag.TagID, 33)
	require.EqualValues(t, tag.LinkID(), 67)
	require.EqualValues(t, tag.ServerID, 5)
	require.EqualValues(t, tag.ClusterID, 123)
	require.EqualValues(t, tag.TenantID, 886222452)
	require.Equal(t, tag.TraceID.String(), "11223344556677889900112233445566")
	require.Equal(t, tag.SpanID.String(), "8877665544332211")
}

func TestPropagatorExtractXDynatraceForeignWithTraceparent(t *testing.T) {
	p := NewTextMapPropagator()
	require.NotNil(t, p)

	c := propagation.HeaderCarrier{}
	c.Set(xDtHeader, "FW4;656;5;15;33;67;113948091;0;e03f;2h01;6h11223344556677889900112233445566;7h8877665544332211")
	c.Set(traceparentHeader, "00-22334455667788990011223344556677-8877665544332211-01")
	c.Set(tracestateHeader, "34d2ae74-7b@dt=fw4;0;0;0;0;1;0;0;8355;2h01;7haaffffeebbaabbee")
	// X-Dynatrace entry is foreign since tracestate and X-Dynatrace clusterId does not match

	ctx := p.Extract(context.Background(), c)

	span := trace.SpanFromContext(ctx)
	require.NotNil(t, span)

	// check span context
	spanCtx := span.SpanContext()
	require.Equal(t, spanCtx.TraceID().String(), "22334455667788990011223344556677")
	require.Equal(t, spanCtx.SpanID().String(), "8877665544332211")
	require.True(t, spanCtx.IsRemote())
	require.False(t, spanCtx.TraceFlags().IsSampled())
	// 'custom' tracestate entry remains since traceId is matching
	require.Equal(t, spanCtx.TraceState().Get("34d2ae74-7b@dt"), "fw4;0;0;0;0;1;0;0;8355;2h01;7haaffffeebbaabbee")

	// check FW4 tag
	tag := fw4.Fw4TagFromContext(ctx)
	require.NotNil(t, tag)
	require.EqualValues(t, tag.ServerID, 0)
	require.EqualValues(t, tag.ClusterID, 123)
	require.EqualValues(t, tag.TenantID, 886222452)
	require.Equal(t, tag.SpanID.String(), "aaffffeebbaabbee")
}

func TestPropagatorExtractOnlyXDynatraceWithMissingTraceIdSpanId(t *testing.T) {
	p := NewTextMapPropagator()
	require.NotNil(t, p)

	c := propagation.HeaderCarrier{}
	c.Set(xDtHeader, "FW4;123;5;15;33;67;886222452;0")

	ctx := p.Extract(context.Background(), c)
	require.Equal(t, ctx, context.Background())
}

func TestPropagatorExtractParseTracestateZeroServerIdMismatchSpanId(t *testing.T) {
	p := NewTextMapPropagator()
	require.NotNil(t, p)

	c := propagation.HeaderCarrier{}
	c.Set(traceparentHeader, "00-11223344556677889900112233445566-8877665544332211-01")
	c.Set(tracestateHeader, "34d2ae74-7b@dt=fw4;0;0;0;0;1;0;0;8355;2h01;7haaffffeebbaabbee")

	ctx := p.Extract(context.Background(), c)

	span := trace.SpanFromContext(ctx)
	require.NotNil(t, span)

	// check span context
	spanCtx := span.SpanContext()
	require.Equal(t, spanCtx.TraceID().String(), "11223344556677889900112233445566")
	require.Equal(t, spanCtx.SpanID().String(), "8877665544332211")
	require.True(t, spanCtx.IsRemote())
	require.False(t, spanCtx.TraceFlags().IsSampled())
	// 'custom' tracestate entry remains since traceId is matching
	require.Equal(t, spanCtx.TraceState().Get("34d2ae74-7b@dt"), "fw4;0;0;0;0;1;0;0;8355;2h01;7haaffffeebbaabbee")

	// check FW4 tag
	tag := fw4.Fw4TagFromContext(ctx)
	require.NotNil(t, tag)
	require.EqualValues(t, tag.ServerID, 0)
	require.EqualValues(t, tag.ClusterID, 123)
	require.EqualValues(t, tag.TenantID, 886222452)
	require.Equal(t, tag.SpanID.String(), "aaffffeebbaabbee")
}

func TestPropagatorExtractParseTracestateMismatchTraceId(t *testing.T) {
	p := NewTextMapPropagator()
	require.NotNil(t, p)

	c := propagation.HeaderCarrier{}
	c.Set(traceparentHeader, "00-11223344556677889900112233445566-8877665544332211-01")
	c.Set(tracestateHeader, "34d2ae74-7b@dt=fw4;0;0;0;0;1;0;0;dd95;2h01;6haa2233445566778899001122334455bb;7h8877665544332211")

	type ctxKey struct{}
	parentCtx := context.WithValue(context.Background(), ctxKey{}, "testValue")
	ctx := p.Extract(parentCtx, c)
	require.Equal(t, ctx, parentCtx)
}

func TestPropagatorExtractParseTracestateNegativeServerMatchingSpanId(t *testing.T) {
	p := NewTextMapPropagator()
	require.NotNil(t, p)

	c := propagation.HeaderCarrier{}
	c.Set(traceparentHeader, "00-11223344556677889900112233445566-8877665544332211-01")
	c.Set(tracestateHeader, "34d2ae74-7b@dt=fw4;fffffffd;0;0;0;0;0;0;7db5;2h01;7h8877665544332211")

	ctx := p.Extract(context.Background(), c)

	span := trace.SpanFromContext(ctx)
	require.NotNil(t, span)

	// check span context
	spanCtx := span.SpanContext()
	require.Equal(t, spanCtx.TraceID().String(), "11223344556677889900112233445566")
	require.Equal(t, spanCtx.SpanID().String(), "8877665544332211")
	require.True(t, spanCtx.IsRemote())
	require.True(t, spanCtx.TraceFlags().IsSampled())
	// 'custom' tracestate entry remains since traceId is matching
	require.Equal(t, spanCtx.TraceState().Get("34d2ae74-7b@dt"), "fw4;fffffffd;0;0;0;0;0;0;7db5;2h01;7h8877665544332211")

	// check FW4 tag
	tag := fw4.Fw4TagFromContext(ctx)
	require.NotNil(t, tag)
	require.EqualValues(t, tag.ServerID, -3)
	require.EqualValues(t, tag.ClusterID, 123)
	require.EqualValues(t, tag.TenantID, 886222452)
	require.Equal(t, tag.SpanID.String(), "8877665544332211")
}

func TestPropagatorExtractParseForeignTracestate(t *testing.T) {
	p := NewTextMapPropagator()
	require.NotNil(t, p)

	c := propagation.HeaderCarrier{}
	c.Set(traceparentHeader, "00-11223344556677889900112233445566-8877665544332211-01")
	c.Set(tracestateHeader, "custom=00f067aa0ba902b7,6cab5bb-290@dt=fw4;0;0;0;0;0;0;0;8355;2h01;7haaffffeebbaabbee")

	ctx := p.Extract(context.Background(), c)

	span := trace.SpanFromContext(ctx)
	require.NotNil(t, span)

	// check span context
	spanCtx := span.SpanContext()
	require.Equal(t, spanCtx.TraceID().String(), "11223344556677889900112233445566")
	require.Equal(t, spanCtx.SpanID().String(), "8877665544332211")
	require.True(t, spanCtx.IsRemote())
	require.True(t, spanCtx.TraceFlags().IsSampled())
	// 'custom' tracestate entry remains since traceId is matching
	require.Equal(t, spanCtx.TraceState().Get("6cab5bb-290@dt"), "fw4;0;0;0;0;0;0;0;8355;2h01;7haaffffeebbaabbee")
	require.Equal(t, spanCtx.TraceState().Get("custom"), "00f067aa0ba902b7")

	// check FW4 tag
	tag := fw4.Fw4TagFromContext(ctx)
	require.Nil(t, tag)
}

func TestPropagatorExtractOnlyTraceparent(t *testing.T) {
	p := NewTextMapPropagator()
	require.NotNil(t, p)

	c := propagation.HeaderCarrier{}
	c.Set(traceparentHeader, "00-11223344556677889900112233445566-8877665544332211-00")

	ctx := p.Extract(context.Background(), c)

	span := trace.SpanFromContext(ctx)
	require.NotNil(t, span)

	// check span context
	spanCtx := span.SpanContext()
	require.Equal(t, spanCtx.TraceID().String(), "11223344556677889900112233445566")
	require.Equal(t, spanCtx.SpanID().String(), "8877665544332211")
	require.True(t, spanCtx.IsRemote())
	require.False(t, spanCtx.TraceFlags().IsSampled())
	// 'custom' tracestate entry remains since traceId is matching
	require.Empty(t, spanCtx.TraceState())

	// check FW4 tag
	tag := fw4.Fw4TagFromContext(ctx)
	require.Nil(t, tag)
}

func TestPropagatorInjectInvalidSpanContext(t *testing.T) {
	p := NewTextMapPropagator()
	require.NotNil(t, p)

	config := trace.SpanContextConfig{
		TraceID:    trace.TraceID{},
		SpanID:     trace.SpanID{},
		TraceFlags: trace.FlagsSampled,
	}

	ctx, span := newTestDtSpan(config, p.config)
	require.False(t, span.metadata.fw4Tag.HasTagDepth())

	c := propagation.HeaderCarrier{}
	p.Inject(ctx, c)
	require.Empty(t, c)
}

func TestPropagatorInjectMinimalSpanContext(t *testing.T) {
	p := NewTextMapPropagator()
	require.NotNil(t, p)

	traceId, err := trace.TraceIDFromHex("11223344556677889900112233445566")
	require.NoError(t, err)

	spanId, err := trace.SpanIDFromHex("8877665544332211")
	require.NoError(t, err)

	config := trace.SpanContextConfig{
		TraceID:    traceId,
		SpanID:     spanId,
		TraceFlags: trace.FlagsSampled,
		TraceState: trace.TraceState{},
	}

	ctx, span := newTestDtSpan(config, p.config)
	require.False(t, span.metadata.fw4Tag.HasTagDepth())

	c := propagation.HeaderCarrier{}
	p.Inject(ctx, c)

	require.False(t, span.metadata.fw4Tag.HasTagDepth())
	require.Equal(t, c.Get(traceparentHeader), "00-11223344556677889900112233445566-8877665544332211-01")
	require.Equal(t, c.Get(tracestateHeader), "34d2ae74-7b@dt=fw4;0;0;0;0;0;0;0;7db5;2h01;7h8877665544332211")
	require.Equal(t, c.Get(xDtHeader), "FW4;123;0;0;0;0;886222452;0;e03f;2h01;6h11223344556677889900112233445566;7h8877665544332211")
}

func TestPropagatorInjectValuesOfFieldsCall(t *testing.T) {
	p := NewTextMapPropagator()
	require.NotNil(t, p)

	traceId, err := trace.TraceIDFromHex("11223344556677889900112233445566")
	require.NoError(t, err)

	spanId, err := trace.SpanIDFromHex("8877665544332211")
	require.NoError(t, err)

	config := trace.SpanContextConfig{
		TraceID:    traceId,
		SpanID:     spanId,
		TraceFlags: trace.FlagsSampled,
		TraceState: trace.TraceState{},
	}

	ctx, span := newTestDtSpan(config, p.config)
	require.False(t, span.metadata.fw4Tag.HasTagDepth())

	c := propagation.HeaderCarrier{}
	p.Inject(ctx, c)

	fields := p.Fields()
	require.Equal(t, fields, []string{traceparentHeader, tracestateHeader, xDtHeader})

	require.False(t, span.metadata.fw4Tag.HasTagDepth())
	require.Equal(t, c.Get(traceparentHeader), "00-11223344556677889900112233445566-8877665544332211-01")
	require.Equal(t, c.Get(tracestateHeader), "34d2ae74-7b@dt=fw4;0;0;0;0;0;0;0;7db5;2h01;7h8877665544332211")
	require.Equal(t, c.Get(xDtHeader), "FW4;123;0;0;0;0;886222452;0;e03f;2h01;6h11223344556677889900112233445566;7h8877665544332211")
}

func TestPropagatorInjectExtendTracestate(t *testing.T) {
	p := NewTextMapPropagator()
	require.NotNil(t, p)

	traceId, err := trace.TraceIDFromHex("11223344556677889900112233445566")
	require.NoError(t, err)

	spanId, err := trace.SpanIDFromHex("8877665544332211")
	require.NoError(t, err)

	ts := trace.TraceState{}
	ts, err = ts.Insert("congo", "t61rcWkgMzE")
	require.NoError(t, err)

	ts, err = ts.Insert("rojo", "00f067aa0ba902b7")
	require.NoError(t, err)

	config := trace.SpanContextConfig{
		TraceID:    traceId,
		SpanID:     spanId,
		TraceFlags: trace.FlagsSampled,
		TraceState: ts,
	}

	ctx, span := newTestDtSpan(config, p.config)
	require.False(t, span.metadata.fw4Tag.HasTagDepth())
	span.metadata.fw4Tag.PathInfo = 0x53

	c := propagation.HeaderCarrier{}
	p.Inject(ctx, c)

	require.False(t, span.metadata.fw4Tag.HasTagDepth())
	require.Equal(t, c.Get(traceparentHeader), "00-11223344556677889900112233445566-8877665544332211-01")
	require.Equal(t, c.Get(tracestateHeader), "34d2ae74-7b@dt=fw4;0;0;0;0;0;0;53;7db5;2h01;7h8877665544332211,rojo=00f067aa0ba902b7,congo=t61rcWkgMzE")
	require.Equal(t, c.Get(xDtHeader), "FW4;123;0;0;0;0;886222452;83;e03f;2h01;6h11223344556677889900112233445566;7h8877665544332211")
}

func TestPropagatorInjectUpdateTracestate(t *testing.T) {
	p := NewTextMapPropagator()
	require.NotNil(t, p)

	traceId, err := trace.TraceIDFromHex("11223344556677889900112233445566")
	require.NoError(t, err)

	spanId, err := trace.SpanIDFromHex("8877665544332211")
	require.NoError(t, err)

	ts := trace.TraceState{}
	ts, err = ts.Insert("rojo", "00f067aa0ba902b7")
	require.NoError(t, err)

	// tracestate value will be updated within inject call
	tracestateValue := "fw4;1;2;3;4;0;0;6;8355;2h01;7haaffffeebbaabbee"
	ts, err = ts.Insert("34d2ae74-7b@dt", tracestateValue)
	require.NoError(t, err)

	config := trace.SpanContextConfig{
		TraceID:    traceId,
		SpanID:     spanId,
		TraceFlags: trace.FlagsSampled,
		TraceState: ts,
	}

	ctx, span := newTestDtSpan(config, p.config)
	tag, err := fw4.ParseTracestateEntryValue(tracestateValue)
	require.NoError(t, err)

	tag.TenantID = p.config.TenantId()
	tag.ClusterID = p.config.ClusterId
	span.metadata.fw4Tag = &tag

	c := propagation.HeaderCarrier{}
	p.Inject(ctx, c)

	require.True(t, span.metadata.fw4Tag.HasTagDepth())
	require.Equal(t, c.Get(traceparentHeader), "00-11223344556677889900112233445566-8877665544332211-01")
	require.Equal(t, c.Get(tracestateHeader), "34d2ae74-7b@dt=fw4;1;0;0;0;0;0;6;5561;2h02;7h8877665544332211,rojo=00f067aa0ba902b7")
	require.Equal(t, c.Get(xDtHeader), "FW4;123;1;0;0;0;886222452;6;34ad;2h02;6h11223344556677889900112233445566;7h8877665544332211")
}

func TestPropagatorInjectSetLastPropagationTime(t *testing.T) {
	p := NewTextMapPropagator()
	require.NotNil(t, p)

	traceId, err := trace.TraceIDFromHex("11223344556677889900112233445566")
	require.NoError(t, err)

	spanId, err := trace.SpanIDFromHex("8877665544332211")
	require.NoError(t, err)

	config := trace.SpanContextConfig{
		TraceID:    traceId,
		SpanID:     spanId,
		TraceFlags: trace.FlagsSampled,
	}

	ctx, span := newTestDtSpan(config, p.config)
	require.False(t, span.metadata.fw4Tag.HasTagDepth())
	require.Zero(t, span.metadata.lastPropagationTime)

	p.Inject(ctx, propagation.HeaderCarrier{})

	require.NotZero(t, span.metadata.lastPropagationTime)
	require.False(t, span.metadata.fw4Tag.HasTagDepth())
}

func TestPropagatorIgnoreNonSampledFlagFromSpanContext(t *testing.T) {
	p := NewTextMapPropagator()
	require.NotNil(t, p)

	c := propagation.HeaderCarrier{}
	c.Set(traceparentHeader, "00-11223344556677889900112233445566-8877665544332211-00")
	c.Set(tracestateHeader, "34d2ae74-7b@dt=fw4;8;0;0;0;0;0;0;7db5;2h01;7h8877665544332211")

	ctx := p.Extract(context.Background(), c)

	span := trace.SpanFromContext(ctx)
	require.NotNil(t, span)

	// traceparent non-sampling decision must be ignored, thus span must be sampled based on FW4 tag
	require.True(t, span.SpanContext().TraceFlags().IsSampled())

	tag := fw4.Fw4TagFromContext(ctx)
	require.NotNil(t, tag)
	require.False(t, tag.IsIgnored())
}

func TestPropagatorIgnoreSampledFlagFromSpanContext(t *testing.T) {
	p := NewTextMapPropagator()
	require.NotNil(t, p)

	c := propagation.HeaderCarrier{}
	c.Set(traceparentHeader, "00-11223344556677889900112233445566-8877665544332211-01")
	c.Set(tracestateHeader, "34d2ae74-7b@dt=fw4;8;0;0;0;1;0;0;7db5;2h01;7h8877665544332211")

	ctx := p.Extract(context.Background(), c)

	span := trace.SpanFromContext(ctx)
	require.NotNil(t, span)

	// traceparent sampling decision must be ignored, thus span must be non-sampled based on FW4 tag
	require.False(t, span.SpanContext().TraceFlags().IsSampled())

	tag := fw4.Fw4TagFromContext(ctx)
	require.NotNil(t, tag)
	require.True(t, tag.IsIgnored())
}

func newTestDtSpan(spanConfig trace.SpanContextConfig, dtConfig *configuration.DtConfiguration) (context.Context, *dtSpan) {
	spanCtx := trace.NewSpanContext(spanConfig)
	ctx := trace.ContextWithSpanContext(context.Background(), spanCtx)

	spanMetadata := newDtSpanMetadata(0)
	spanMetadata.fw4Tag = fw4.NewFw4Tag(dtConfig.ClusterId, dtConfig.TenantId(), spanCtx)
	spanMetadata.fw4Tag.PathInfo = 0

	span := &dtSpan{
		Span:     trace.SpanFromContext(ctx),
		metadata: spanMetadata,
	}

	ctx = trace.ContextWithSpan(ctx, span)
	return ctx, span
}
