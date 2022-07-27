package trace

import (
	"context"
	"time"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"core/internal/logger"
)

type exportType int

const (
	exportTypePeriodic exportType = iota
	exportTypeForceFlush
)

type dtSpanExporter interface {
	// TODO: discuss how to pass spans to export
	export(ctx context.Context, spans dtSpanSet) error
}

type dtSpanExporterImpl struct {
	logger *logger.ComponentLogger
}

func newDtSpanExporter() dtSpanExporter {
	return &dtSpanExporterImpl{
		logger: logger.NewComponentLogger("SpanExporter"),
	}
}

func (e *dtSpanExporterImpl) export(_ context.Context, spans dtSpanSet) error {
	e.logger.Debugf("Number of spans to export: %d", len(spans))
	for s := range spans {
		e.logger.Debugf("Exporting span: %s", s.Span.(sdktrace.ReadOnlySpan).Name())
	}

	// emulate export operation
	time.Sleep(500 * time.Millisecond)
	return nil
}
