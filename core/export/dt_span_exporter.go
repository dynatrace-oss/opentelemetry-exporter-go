package export

import (
	"context"
	"log"
	"time"
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

type dtSpanExporterImpl struct{}

func newDtSpanExporter() dtSpanExporter {
	return &dtSpanExporterImpl{}
}

func (e *dtSpanExporterImpl) export(_ context.Context, spans map[spanKey]*dtSpanMetadata) error {
	log.Printf("DtSpanExporter: Nums spans to export: %d", len(spans))
	for _, v := range spans {
		log.Printf("DtSpanExporter: Exporting span: %s", v.span.Name())
	}

	// emulate export operation
	time.Sleep(500 * time.Millisecond)
	// TODO
	return nil
}
