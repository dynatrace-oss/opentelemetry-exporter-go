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
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"strconv"
	"time"

	"github.com/dynatrace-oss/opentelemetry-exporter-go/core/configuration"
	"github.com/dynatrace-oss/opentelemetry-exporter-go/core/internal/logger"
	"github.com/dynatrace-oss/opentelemetry-exporter-go/core/internal/version"
)

type exportType int

const (
	exportTypePeriodic exportType = iota
	exportTypeForceFlush
)

const cSpansPath = "/odin/v1/spans"

const cMaxSizeWarning = 1 * 1024 * 1024 // 1 MB
const cMaxSizeSend = 64 * 1024 * 1024   // 64 MB

var errNotAuthorizedRequest = errors.New("Span Exporter is not authorized to send spans")

type dtSpanExporter interface {
	export(ctx context.Context, t exportType, spans dtSpanSet) error
}

type dtSpanExporterImpl struct {
	logger   *logger.ComponentLogger
	config   *configuration.DtConfiguration
	dialer   *net.Dialer
	client   *http.Client
	disabled bool
}

func newDtSpanExporter(config *configuration.DtConfiguration) dtSpanExporter {
	d := &net.Dialer{}
	exporter := &dtSpanExporterImpl{
		logger: logger.NewComponentLogger("SpanExporter"),
		config: config,
		dialer: d,
		client: &http.Client{
			Transport: &http.Transport{
				DialContext: d.DialContext,
			},
		},
		disabled: false,
	}

	return exporter
}

func (e *dtSpanExporterImpl) export(ctx context.Context, t exportType, spans dtSpanSet) error {
	if e.disabled {
		e.logger.Debug("Skip exporting, Span Exporter is disabled")
		return nil
	}

	if len(spans) == 0 {
		e.logger.Debug("Skip exporting, no spans to export")
		return nil
	}

	e.logger.Debugf("Serialize %d spans to export", len(spans))

	start := time.Now()
	// TODO: In order to support large amounts of spans, implement a splitting algorithm
	// so that we can send spans in batches whose sizes do not exceed cMaxSizeSend.
	serializedSpans, err := serializeSpans(spans, e.config.Tenant, e.config.AgentId, e.config.TenantId(), e.config.ClusterId)
	if err != nil {
		return err
	}

	e.logger.Debugf("Serialization process took %s", time.Since(start))

	serializedSpansLen := len(serializedSpans)
	if serializedSpansLen > cMaxSizeSend {
		errMsg := fmt.Sprintf("skip exporting, serialized spans reached %d bytes. Maximum allowed size is %d bytes",
			serializedSpansLen, cMaxSizeSend)
		return errors.New(errMsg)
	} else if serializedSpansLen > cMaxSizeWarning {
		e.logger.Warnf("Size of serialized spans reached %d bytes", serializedSpansLen)
	}

	reqBody := bytes.NewReader(serializedSpans)
	req, err := e.newRequest(ctx, reqBody)
	if err != nil {
		return err
	}
	resp, err := e.performHttpRequest(req, t)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		// 401/403 is permanent, so avoid further exporting
		e.disabled = true
		return errNotAuthorizedRequest
	} else if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.New("unexpected response code: " + strconv.Itoa(resp.StatusCode))
	}

	return nil
}

func (e *dtSpanExporterImpl) newRequest(ctx context.Context, body *bytes.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", e.config.BaseUrl+cSpansPath, body)
	if err != nil {
		e.logger.Errorf("Can not create HTTP request: %s", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-dt-span-export")
	req.Header.Set("Authorization", "Dynatrace "+e.config.AuthToken)
	req.Header.Set("User-Agent", fmt.Sprintf("odin-go/%s %#016x %s",
		version.FullVersion, e.config.AgentId, e.config.Tenant))
	req.Header.Set("Accept", "*/*; q=0")
	// Setting just the header Idempotency-Key with an empty value ensures that the request is
	// treated as idempotent but the header is not sent over the wire. See net/http/transport.go
	// req.GetBody must also be set. It is set automatically by http.NewRequestWithContext since the body is of type *bytes.Reader.
	req.Header.Set("Idempotency-Key", "")

	return req, nil
}

func (e *dtSpanExporterImpl) performHttpRequest(req *http.Request, t exportType) (*http.Response, error) {
	if e.logger.DebugEnabled() {
		// Authorization token must not be logged
		reqCopy := req.Clone(context.Background())
		reqCopy.Header.Del("Authorization")

		dump, err := httputil.DumpRequest(reqCopy, false)
		if err != nil {
			e.logger.Warnf("Can not dump HTTP request: %s", err)
		} else {
			e.logger.Debugf("About to perform HTTP request %s", string(dump))
		}
	}

	e.setTimeouts(t)

	start := time.Now()
	resp, err := e.client.Do(req)
	e.logger.Debugf("HTTP request took %s", time.Since(start))

	if err != nil {
		e.logger.Errorf("Can not perform HTTP request: %s", err)
	}

	if e.logger.DebugEnabled() && resp != nil {
		dump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			e.logger.Warnf("Can not dump HTTP response: %s", err)
		} else {
			e.logger.Debugf("HTTP response %s", string(dump))
		}
	}

	return resp, err
}

// setTimeouts updates connection and data timeouts for HTTP client
func (e *dtSpanExporterImpl) setTimeouts(t exportType) {
	var conn, data int64
	if t == exportTypeForceFlush {
		conn = configuration.DefaultFlushExportConnTimeoutMs
		data = configuration.DefaultFlushExportDataTimeoutMs
	} else {
		if t != exportTypePeriodic {
			e.logger.Warnf("Unknown export type: %d", t)
		}

		conn = configuration.DefaultRegularExportConnTimeoutMs
		data = configuration.DefaultRegularExportDataTimeoutMs
	}

	e.dialer.Timeout = time.Millisecond * time.Duration(conn)
	e.client.Timeout = time.Millisecond * time.Duration(conn+data)
}
