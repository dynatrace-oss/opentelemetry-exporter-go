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
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"github.com/dynatrace-oss/opentelemetry-exporter-go/core/configuration"
	"github.com/dynatrace-oss/opentelemetry-exporter-go/core/internal/version"
)

func TestDtSpanExporterVerifyNewRequest(t *testing.T) {
	config := &configuration.DtConfiguration{
		ClusterId:                -1234,
		Tenant:                   "testTenant",
		AgentId:                  10,
		BaseUrl:                  "https://example.com",
		AuthToken:                "testDtToken",
		SpanProcessingIntervalMs: configuration.DefaultSpanProcessingIntervalMs,
		LoggingDestination:       configuration.LoggingDestination_Stdout,
		LoggingFlags:             "SpanExporter=true,SpanProcessor=true,TracerProvider=true",
		RumClientIpHeaders:       nil,
		DebugAddStackOnStart:     false,
	}

	exporter := newDtSpanExporter(config).(*dtSpanExporterImpl)
	req, err := exporter.newRequest(context.Background(), bytes.NewReader([]byte{1, 2, 3, 4, 5}))

	require.NoError(t, err)
	require.Equal(t, req.Method, "POST")
	require.Equal(t, req.URL.String(), "https://example.com/odin/v1/spans")
	require.Equal(t, req.Header.Get("Content-Type"), "application/x-dt-span-export")
	require.Equal(t, req.Header.Get("Authorization"), "Dynatrace testDtToken")
	require.Equal(t, req.Header.Get("User-Agent"), fmt.Sprintf("odin-go/%s 0x000000000000000a testTenant", version.FullVersion))
	require.Equal(t, req.Header.Get("Accept"), "*/*; q=0")
	require.Equal(t, req.Header.Get("Idempotency-Key"), "")
	require.EqualValues(t, req.ContentLength, 5)

	body, err := ioutil.ReadAll(req.Body)
	require.NoError(t, err)
	require.EqualValues(t, body, []byte{1, 2, 3, 4, 5})
}

func TestDtSpanExporterPerformHTTPRequest(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		require.Equal(t, req.URL.String(), "/odin/v1/spans")
		require.Equal(t, req.Method, "POST")
		require.Equal(t, req.Header.Get("Content-Type"), "application/x-dt-span-export")
		require.Equal(t, req.Header.Get("Authorization"), "Dynatrace testDtToken")
		require.Equal(t, req.Header.Get("User-Agent"), fmt.Sprintf("odin-go/%s 0x000000000000000a testDtTenant", version.FullVersion))
		require.Equal(t, req.Header.Get("Accept"), "*/*; q=0")
		require.Equal(t, req.Header.Get("Idempotency-Key"), "")
		require.EqualValues(t, req.ContentLength, 3)

		body, err := ioutil.ReadAll(req.Body)
		require.NoError(t, err)
		require.EqualValues(t, body, []byte{10, 20, 30})

		rw.Write([]byte(`Ok`)) //nolint:errcheck
	}))

	config := &configuration.DtConfiguration{
		ClusterId:                -1234,
		Tenant:                   "testDtTenant",
		AgentId:                  10,
		BaseUrl:                  testServer.URL,
		AuthToken:                "testDtToken",
		SpanProcessingIntervalMs: configuration.DefaultSpanProcessingIntervalMs,
		LoggingDestination:       configuration.LoggingDestination_Stdout,
		LoggingFlags:             "SpanExporter=true,SpanProcessor=true,TracerProvider=true",
		RumClientIpHeaders:       nil,
		DebugAddStackOnStart:     false,
	}

	exporter := newDtSpanExporter(config).(*dtSpanExporterImpl)
	req, _ := exporter.newRequest(context.Background(), bytes.NewReader([]byte{10, 20, 30}))
	resp, err := exporter.performHttpRequest(req, exportTypePeriodic)
	require.Equal(t, resp.StatusCode, http.StatusOK)
	require.NoError(t, err)

	defer testServer.Close()
}

func TestDtSpanExporterPerformHTTPRequestWithReachedFlushOperationTimeout(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// sleep to reach flush operation timeout
		time.Sleep(time.Millisecond * time.Duration(configuration.DefaultFlushExportConnTimeoutMs+configuration.DefaultFlushExportDataTimeoutMs+100))
		rw.Write([]byte(`Ok`)) //nolint:errcheck
	}))

	config := &configuration.DtConfiguration{
		ClusterId:                -1234,
		Tenant:                   "testDtTenant",
		AgentId:                  10,
		BaseUrl:                  testServer.URL,
		AuthToken:                "testDtToken",
		SpanProcessingIntervalMs: configuration.DefaultSpanProcessingIntervalMs,
		LoggingDestination:       configuration.LoggingDestination_Stdout,
		LoggingFlags:             "SpanExporter=true,SpanProcessor=true,TracerProvider=true",
		RumClientIpHeaders:       nil,
		DebugAddStackOnStart:     false,
	}

	exporter := newDtSpanExporter(config).(*dtSpanExporterImpl)
	req, _ := exporter.newRequest(context.Background(), bytes.NewReader([]byte{10, 20, 30}))
	resp, err := exporter.performHttpRequest(req, exportTypeForceFlush)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "context deadline exceeded (Client.Timeout exceeded while awaiting headers")

	defer testServer.Close()
}

func TestDtSpanExporterUpdateHttpClientTimeouts(t *testing.T) {
	exporter := newDtSpanExporter(testConfig).(*dtSpanExporterImpl)

	exporter.setTimeouts(exportTypeForceFlush)
	require.Equal(t, exporter.dialer.Timeout, time.Millisecond*time.Duration(configuration.DefaultFlushExportConnTimeoutMs))
	require.Equal(t, exporter.client.Timeout, time.Millisecond*time.Duration(configuration.DefaultFlushExportConnTimeoutMs+configuration.DefaultFlushExportDataTimeoutMs))

	exporter.setTimeouts(exportTypePeriodic)
	require.Equal(t, exporter.dialer.Timeout, time.Millisecond*time.Duration(configuration.DefaultRegularExportConnTimeoutMs))
	require.Equal(t, exporter.client.Timeout, time.Millisecond*time.Duration(configuration.DefaultRegularExportConnTimeoutMs+configuration.DefaultRegularExportDataTimeoutMs))
}

func TestSpanExportWithoutErrors(t *testing.T) {
	numRequests := 0
	testServer, config := createTestServerAndConfig(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		numRequests++
		rw.Write([]byte(`Ok`)) //nolint:errcheck
	}))
	defer testServer.Close()

	exporter := newDtSpanExporter(config).(*dtSpanExporterImpl)
	tracer := createTracer()

	_, span1 := tracer.Start(context.Background(), "span1")
	_, span2 := tracer.Start(context.Background(), "span2")
	_, span3 := tracer.Start(context.Background(), "span3")

	spans := makeSpanSet(span1, span2, span3)
	err := exporter.export(context.Background(), exportTypeForceFlush, spans)

	require.NoError(t, err)
	require.Equal(t, 1, numRequests, "spans are small enough so only one export request is expected")
}

func TestSpanExportWithoutErrors_MultipleExports(t *testing.T) {
	largeString := strings.Repeat("r", 1024*1024) // 1 MB

	numRequests := 0
	testServer, config := createTestServerAndConfig(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		numRequests++
		rw.Write([]byte(`Ok`)) //nolint:errcheck
	}))
	defer testServer.Close()

	exporter := newDtSpanExporter(config).(*dtSpanExporterImpl)
	tracer := createTracer()

	_, span1 := tracer.Start(context.Background(), "span1",
		trace.WithAttributes(attribute.String("large-attr-key", largeString)))
	_, span2 := tracer.Start(context.Background(), "span2")
	_, span3 := tracer.Start(context.Background(), "span3")

	(span1.(*dtSpan)).metadata.sendState = sendStateSpanEnded

	spans := makeSpanSet(span1, span2, span3)
	err := exporter.export(context.Background(), exportTypeForceFlush, spans)

	require.NoError(t, err)
	require.Equal(t, 2, numRequests, "since the first span exceeds the warning size, 2 exports must be done")
}

func TestSpanExportWithoutErrors_DroppedSpans(t *testing.T) {
	largeString := strings.Repeat("r", 64*1024*1024) // 64 MB

	numRequests := 0
	testServer, config := createTestServerAndConfig(func(rw http.ResponseWriter, req *http.Request) {
		numRequests++
		rw.Write([]byte(`Ok`)) //nolint:errcheck
	})
	defer testServer.Close()

	exporter := newDtSpanExporter(config).(*dtSpanExporterImpl)
	tracer := createTracer()

	_, span1 := tracer.Start(context.Background(), "span1",
		trace.WithAttributes(attribute.String("large-attr-key", largeString)))
	_, span2 := tracer.Start(context.Background(), "span2")
	_, span3 := tracer.Start(context.Background(), "span3")

	(span1.(*dtSpan)).metadata.sendState = sendStateSpanEnded

	spans := makeSpanSet(span1, span2, span3)
	err := exporter.export(context.Background(), exportTypeForceFlush, spans)

	require.NoError(t, err)
	require.Equal(t, 1, numRequests, "the first span exceeds the maximum size, it must be dropped -> 1 export")
}

func TestSpanExportWithError_ResourceTooBig(t *testing.T) {
	numRequests := 0
	testServer, config := createTestServerAndConfig(func(rw http.ResponseWriter, req *http.Request) {
		numRequests++
		rw.Write([]byte(`Ok`)) //nolint:errcheck
	})
	defer testServer.Close()

	exporter := newDtSpanExporter(config).(*dtSpanExporterImpl)
	largeString := strings.Repeat("r", 64*1024*1024) // 64 MB
	tracer := createTracer(sdktrace.WithResource(resource.NewSchemaless(attribute.String("large-string", largeString))))

	_, span1 := tracer.Start(context.Background(), "span1")
	_, span2 := tracer.Start(context.Background(), "span2")
	_, span3 := tracer.Start(context.Background(), "span3")

	spans := makeSpanSet(span1, span2, span3)
	err := exporter.export(context.Background(), exportTypeForceFlush, spans)

	require.ErrorContains(t, err, "resource too big")
	require.Equal(t, 0, numRequests, "the resource is too big -> 0 exports")
}

func TestSpanExportWithoutError_LargeResource(t *testing.T) {
	numRequests := 0
	testServer, config := createTestServerAndConfig(func(rw http.ResponseWriter, req *http.Request) {
		numRequests++
		rw.Write([]byte(`Ok`)) //nolint:errcheck
	})
	defer testServer.Close()

	exporter := newDtSpanExporter(config).(*dtSpanExporterImpl)
	largeString := strings.Repeat("r", 1024*1024) // 1 MB
	tracer := createTracer(sdktrace.WithResource(resource.NewSchemaless(attribute.String("large-string", largeString))))

	_, span1 := tracer.Start(context.Background(), "span1")
	_, span2 := tracer.Start(context.Background(), "span2")
	_, span3 := tracer.Start(context.Background(), "span3")

	spans := makeSpanSet(span1, span2, span3)
	err := exporter.export(context.Background(), exportTypeForceFlush, spans)

	require.NoError(t, err)
	require.Equal(t, 3, numRequests, "large resource attached to each SpanExport -> split into 3 requests")
}

func makeSpanSet(spans ...trace.Span) dtSpanSet {
	spanSet := make(dtSpanSet)
	for _, span := range spans {
		spanSet[span.(*dtSpan)] = struct{}{}
	}
	return spanSet
}

func createTestServerAndConfig(handler http.HandlerFunc) (*httptest.Server, *configuration.DtConfiguration) {
	testServer := httptest.NewServer(handler)
	config := &configuration.DtConfiguration{
		ClusterId:                -1234,
		Tenant:                   "testDtTenant",
		AgentId:                  10,
		BaseUrl:                  testServer.URL,
		AuthToken:                "testDtToken",
		SpanProcessingIntervalMs: configuration.DefaultSpanProcessingIntervalMs,
		LoggingDestination:       configuration.LoggingDestination_Stdout,
		LoggingFlags:             "SpanExporter=true,SpanProcessor=true,TracerProvider=true",
		RumClientIpHeaders:       nil,
		DebugAddStackOnStart:     false,
	}
	return testServer, config
}
