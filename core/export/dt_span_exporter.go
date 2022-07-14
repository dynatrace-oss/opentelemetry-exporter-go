package export

import (
	"context"
	"time"

	"core/internal/logger"
)

type exportType int

const (
	exportTypePeriodic exportType = iota
	exportTypeForceFlush
)

type dtSpanExporter interface {
	// TODO: discuss how to pass spans to export
	export(ctx context.Context, spans map[spanKey]*dtSpanMetadata) error
}

type dtSpanExporterImpl struct {
	logger *logger.ComponentLogger
}

func newDtSpanExporter() dtSpanExporter {
	return &dtSpanExporterImpl{
		logger: logger.NewComponentLogger("SpanExporter"),
	}
}

func (e *dtSpanExporterImpl) export(_ context.Context, spans map[spanKey]*dtSpanMetadata) error {
	e.logger.Debugf("Number of spans to export: %d", len(spans))
	for _, v := range spans {
		e.logger.Debugf("Exporting span: %s", v.span.Name())
	}

	// emulate export operation
	time.Sleep(500 * time.Millisecond)
	// TODO
	return nil
}
