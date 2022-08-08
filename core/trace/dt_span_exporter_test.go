package trace

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"core/configuration"
)

func TestDtSpanExporterVerifyNewRequest(t *testing.T) {
	exporter := newDtSpanExporter(testConfig).(*dtSpanExporterImpl)
	req, err := exporter.newRequest(context.Background(), bytes.NewBuffer([]byte{1, 2, 3, 4, 5}))

	require.NoError(t, err)
	require.Equal(t, req.Method, "POST")
	require.Equal(t, req.URL.String(), "https://example.com/odin/v1/spans")
	require.Equal(t, req.Header.Get("Content-Type"), "application/x-dt-span-export")
	require.Equal(t, req.Header.Get("Authorization"), "Dynatrace testAuthToken")
	require.Equal(t, req.Header.Get("User-Agent"), "odin-go/0.0.1.20220701-000000 0x4d65822107fcfd52 testTenant")
	require.Equal(t, req.Header.Get("Accept"), "*/*; q=0")
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
		require.Equal(t, req.Header.Get("User-Agent"), "odin-go/0.0.1.20220701-000000 0x000000000000000a testDtTenant")
		require.Equal(t, req.Header.Get("Accept"), "*/*; q=0")
		require.EqualValues(t, req.ContentLength, 3)

		body, err := ioutil.ReadAll(req.Body)
		require.NoError(t, err)
		require.EqualValues(t, body, []byte{10, 20, 30})

		rw.Write([]byte(`Ok`))
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
	req, _ := exporter.newRequest(context.Background(), bytes.NewBuffer([]byte{10, 20, 30}))
	resp, err := exporter.performHttpRequest(req, exportTypePeriodic)
	require.Equal(t, resp.StatusCode, http.StatusOK)
	require.NoError(t, err)

	defer testServer.Close()
}

func TestDtSpanExporterPerformHTTPRequestWithReachedFlushOperationTimeout(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// sleep to reach flush operation timeout
		time.Sleep(time.Millisecond * time.Duration(configuration.DefaultFlushExportConnTimeoutMs+configuration.DefaultFlushExportDataTimeoutMs+100))
		rw.Write([]byte(`Ok`))
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
	req, _ := exporter.newRequest(context.Background(), bytes.NewBuffer([]byte{10, 20, 30}))
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

// TODO: add unit tests for exporter.export function when Span Enricher and Span Serializer will be implemented
