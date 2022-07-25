package trace

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"strconv"
	"time"

	"core/configuration"
	"core/internal/logger"
	"core/internal/version"
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

func newDtSpanExporter() dtSpanExporter {
	log := logger.NewComponentLogger("SpanExporter")

	config, err := configuration.GlobalConfigurationProvider.GetConfiguration()
	if err != nil {
		log.Errorf("Can not create Span exporter due to Configuration error: %s", err)
		return nil
	}

	d := &net.Dialer{}
	exporter := &dtSpanExporterImpl{
		logger: log,
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
	// TODO: Serialize spans that have to be exported.
	// As an optimization, stop spans serialization if cMaxSizeSend size is reached
	serializedSpans := []byte{}
	e.logger.Debugf("Serialization process took %s", time.Since(start))

	serializedSpansLen := len(serializedSpans)
	if serializedSpansLen > cMaxSizeSend {
		errMsg := fmt.Sprintf("skip exporting, serialized spans reached %d bytes. Maximum allowed size is %d bytes",
			serializedSpansLen, cMaxSizeSend)
		return errors.New(errMsg)
	} else if serializedSpansLen > cMaxSizeWarning {
		e.logger.Warnf("Size of serialized spans reached %d bytes", serializedSpansLen)
	}

	body := bytes.NewBuffer(serializedSpans)
	req, err := e.newRequest(ctx, body)
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

func (e *dtSpanExporterImpl) newRequest(ctx context.Context, body io.Reader) (*http.Request, error) {
	config, err := configuration.GlobalConfigurationProvider.GetConfiguration()
	if err != nil {
		e.logger.Errorf("Can not prepare a new request due to Configuration error: " + err.Error())
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", config.BaseUrl+cSpansPath, body)
	if err != nil {
		e.logger.Errorf("Can not create HTTP request: %s", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-dt-span-export")
	req.Header.Set("Authorization", "Dynatrace "+config.AuthToken)
	req.Header.Set("User-Agent", fmt.Sprintf("odin-go/%s %#016x %s",
		version.FullVersion, config.AgentId, config.Tenant))
	req.Header.Set("Accept", "*/*; q=0")

	return req, nil
}

func (e *dtSpanExporterImpl) performHttpRequest(req *http.Request, t exportType) (resp *http.Response, err error) {
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

	e.updateTimeouts(t)

	start := time.Now()
	resp, err = e.client.Do(req)
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

// updateTimeouts updates connection and data timeouts for HTTP client
func (e *dtSpanExporterImpl) updateTimeouts(t exportType) {
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
