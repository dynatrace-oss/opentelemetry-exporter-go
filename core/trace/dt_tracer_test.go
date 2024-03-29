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

	"github.com/dynatrace-oss/opentelemetry-exporter-go/core/internal/semconv"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func TestDtSpanIsOfDtSpanType(t *testing.T) {
	tp, _ := newDtTracerProviderWithTestExporter()
	tr := tp.Tracer("Dynatrace tracer")

	_, s := tr.Start(context.Background(), "Test span")
	span, ok := s.(*dtSpan)

	require.NotNil(t, span)
	require.True(t, ok)
}

func TestDtSpanContainsSdkSpan(t *testing.T) {
	tp, _ := newDtTracerProviderWithTestExporter()
	tr := tp.Tracer("Dynatrace tracer")

	_, s := tr.Start(context.Background(), "Test span")
	span := s.(*dtSpan)

	sdkSpan, err := span.readOnlySpan()
	require.NotNil(t, sdkSpan)
	require.NoError(t, err)
}

func TestDtSpanContainsValidTracer(t *testing.T) {
	tp, _ := newDtTracerProviderWithTestExporter()
	tr := tp.Tracer("Dynatrace tracer")

	_, s := tr.Start(context.Background(), "Test span")
	span := s.(*dtSpan)

	require.Equal(t, tr, span.tracer)
}

func TestReturnedContextContainsDtSpan(t *testing.T) {
	tp, _ := newDtTracerProviderWithTestExporter()
	tr := tp.Tracer("Dynatrace tracer")

	ctx, s := tr.Start(context.Background(), "Test span")

	spanFromCtx := trace.SpanFromContext(ctx)
	require.Equal(t, spanFromCtx, s)
}

func TestLocalParentWithServerId(t *testing.T) {
	c := propagation.HeaderCarrier{}
	c.Set(traceparentHeader, "00-11223344556677889900112233445566-8877665544332211-01")
	c.Set(tracestateHeader, "34d2ae74-7b@dt=fw4;fffffff8;0;0;0;0;0;0;7db5;2h01;7h8877665544332211")

	p, err := NewTextMapPropagator()
	require.NotNil(t, p)
	require.NoError(t, err)
	remoteCtx := p.Extract(context.Background(), c)

	tp, _ := newDtTracerProviderWithTestExporter()
	tr := tp.Tracer("Dynatrace tracer")
	ctx, localParent := tr.Start(remoteCtx, "local root")
	_, span := tr.Start(ctx, "child")

	metadata := span.(*dtSpan).metadata
	require.NotNil(t, metadata)

	tag := metadata.getFw4Tag()
	require.NotNil(t, tag)
	require.EqualValues(t, tag.ServerID, -8)
	require.Equal(t, metadata.tenantParentSpanId, localParent.SpanContext().SpanID())
}

func TestLocalParentWithNoServerId(t *testing.T) {
	tp, _ := newDtTracerProviderWithTestExporter()
	tr := tp.Tracer("Dynatrace tracer")
	ctx, localParent := tr.Start(context.Background(), "local root")
	_, span := tr.Start(ctx, "child")

	metadata := span.(*dtSpan).metadata
	require.NotNil(t, metadata)

	tag := metadata.getFw4Tag()
	require.NotNil(t, tag)
	require.EqualValues(t, tag.ServerID, 0)
	require.True(t, tag.PathInfo < 256)
	require.Equal(t, metadata.tenantParentSpanId, localParent.SpanContext().SpanID())
}

func TestRemoteParentWithTracestateAndWrongTenantId(t *testing.T) {
	c := propagation.HeaderCarrier{}
	c.Set(traceparentHeader, "00-11223344556677889900112233445566-aaffffeebbaabbee-01")
	c.Set(tracestateHeader, "6cab5bb-7b@dt=fw4;fffffff8;0;0;0;0;0;0;7db5;2h01;7h8877665544332211")

	p, err := NewTextMapPropagator()
	require.NotNil(t, p)
	require.NoError(t, err)

	remoteCtx := p.Extract(context.Background(), c)

	tp, _ := newDtTracerProviderWithTestExporter()
	tr := tp.Tracer("Dynatrace tracer")
	_, span := tr.Start(remoteCtx, "local root")

	metadata := span.(*dtSpan).metadata
	require.NotNil(t, metadata)
	require.Equal(t, span.SpanContext().TraceID().String(), "11223344556677889900112233445566")
	require.Zero(t, metadata.tenantParentSpanId)

	tag := metadata.getFw4Tag()
	require.NotNil(t, tag)
	require.EqualValues(t, tag.ServerID, 0)
	require.True(t, tag.PathInfo < 256)
}

var testPropagatingAttributes = []attribute.KeyValue{
	attribute.Key(semconv.CloudAccountId).String("my-account"),
	attribute.Key(semconv.CloudPlatform).String("some-platform"),
	attribute.Key(semconv.CloudProvider).String("the-cloud"),
	attribute.Key(semconv.CloudRegion).String("moonbase-1"),
	attribute.Key(semconv.CloudAvailabilityZone).String("sector-5"),
	attribute.Key(semconv.FaasId).String("123-456"),
	attribute.Key(semconv.FaasName).String("the-function"),
	attribute.Key(semconv.FaasVersion).Int64(777),
	attribute.Key(semconv.FaasInstance).String("xyz"),
	attribute.Key(semconv.GcpRegion).String("moonbase-1"),
	attribute.Key(semconv.GcpProjectId).String("my-project"),
	attribute.Key(semconv.GcpInstanceName).String("the-instance"),
	attribute.Key(semconv.GcpResourceType).String("resource-type"),
}

func TestLocalParent_PropagateResourceAttributes(t *testing.T) {
	attributes := []attribute.KeyValue{attribute.Key("parent.attr").String("hello parent")}
	attributes = append(attributes, testPropagatingAttributes...)

	tp, _ := newDtTracerProviderWithTestExporter()
	tr := tp.Tracer("Dynatrace tracer")
	ctx, localParent := tr.Start(context.Background(), "local root", trace.WithAttributes(attributes...))
	childCtx, child := tr.Start(ctx, "child", trace.WithAttributes(attribute.Key("child.attr").String("hello child")))
	_, sibling := tr.Start(ctx, "sibling", trace.WithAttributes(attribute.Key("sibling.attr").String("hello sibling")))
	_, grandChild := tr.Start(childCtx, "grandchild", trace.WithAttributes(attribute.Key("grandchild.attr").String("hello grandchild")))

	expectedPropagatedAttributes := attributesToMap(t, testPropagatingAttributes)

	parentMetadata := localParent.(*dtSpan).metadata
	require.Nil(t, parentMetadata.propagatedResourceAttributes)

	childMetadata := child.(*dtSpan).metadata
	require.NotNil(t, childMetadata.propagatedResourceAttributes)
	assertAttributeEquals(t, expectedPropagatedAttributes, childMetadata.propagatedResourceAttributes)

	siblingMetadata := sibling.(*dtSpan).metadata
	require.NotNil(t, siblingMetadata.propagatedResourceAttributes)
	assertAttributeEquals(t, expectedPropagatedAttributes, siblingMetadata.propagatedResourceAttributes)

	grandChildMetadata := grandChild.(*dtSpan).metadata
	require.Equal(t, childMetadata.propagatedResourceAttributes, grandChildMetadata.propagatedResourceAttributes)
	assertAttributeEquals(t, expectedPropagatedAttributes, grandChildMetadata.propagatedResourceAttributes)
}

func TestRemoteParent_PropagateResourceAttributes(t *testing.T) {
	c := propagation.HeaderCarrier{}
	c.Set(traceparentHeader, "00-11223344556677889900112233445566-aaffffeebbaabbee-01")
	c.Set(tracestateHeader, "6cab5bb-7b@dt=fw4;fffffff8;0;0;0;0;0;0;7db5;2h01;7h8877665544332211")

	p, err := NewTextMapPropagator()
	require.NotNil(t, p)
	require.NoError(t, err)

	remoteCtx := p.Extract(context.Background(), c)

	tp, _ := newDtTracerProviderWithTestExporter()
	tr := tp.Tracer("Dynatrace tracer")

	attributes := []attribute.KeyValue{attribute.Key("parent.attr").String("hello parent")}
	attributes = append(attributes, testPropagatingAttributes...)
	ctx, localRoot := tr.Start(remoteCtx, "local root", trace.WithAttributes(attributes...))
	childCtx, child := tr.Start(ctx, "child", trace.WithAttributes(attribute.Key("child.attr").String("hello child")))
	_, sibling := tr.Start(ctx, "sibling", trace.WithAttributes(attribute.Key("sibling.attr").String("hello sibling")))
	_, grandChild := tr.Start(childCtx, "grandchild", trace.WithAttributes(attribute.Key("grandchild.attr").String("hello grandchild")))

	rootMetadata := localRoot.(*dtSpan).metadata
	require.Nil(t, rootMetadata.propagatedResourceAttributes)

	childMetadata := child.(*dtSpan).metadata
	require.NotNil(t, childMetadata.propagatedResourceAttributes)

	siblingMetadata := sibling.(*dtSpan).metadata
	require.NotNil(t, siblingMetadata.propagatedResourceAttributes)

	grandChildMetadata := grandChild.(*dtSpan).metadata
	require.Equal(t, childMetadata.propagatedResourceAttributes, grandChildMetadata.propagatedResourceAttributes)
}
