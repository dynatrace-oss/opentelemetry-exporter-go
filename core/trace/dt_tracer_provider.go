package trace

import (
	"context"
	"errors"
	"sync"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"core/configuration"
	"core/internal/logger"
)

var errInvalidSpanProcessor = errors.New("span processor is invalid")

type DtTracerProvider struct {
	trace.TracerProvider
	mu             sync.Mutex
	wrappedTracers map[trace.Tracer]*dtTracer
	processor      *dtSpanProcessor
	logger         *logger.ComponentLogger
	config         *configuration.DtConfiguration
}

func NewTracerProvider(config *configuration.DtConfiguration, opts ...sdktrace.TracerProviderOption) *DtTracerProvider {
	tp := &DtTracerProvider{
		TracerProvider: sdktrace.NewTracerProvider(opts...),
		mu:             sync.Mutex{},
		wrappedTracers: make(map[trace.Tracer]*dtTracer),
		processor:      newDtSpanProcessor(config),
		logger:         logger.NewComponentLogger("TracerProvider"),
		config:         config,
	}

	tp.logger.Debug("TracerProvider created")
	return tp
}

func (p *DtTracerProvider) Tracer(name string, opts ...trace.TracerOption) trace.Tracer {
	p.mu.Lock()
	defer p.mu.Unlock()

	sdkTracer := p.TracerProvider.Tracer(name, opts...)

	// Dynatrace TracerProvider is just a wrapper over SDK TracerProvider, thus use SDK Tracer as a key
	tr, found := p.wrappedTracers[sdkTracer]
	if !found {
		tr = &dtTracer{
			Tracer:   sdkTracer,
			provider: p,
			config:   p.config,
		}
		p.wrappedTracers[sdkTracer] = tr
		p.logger.Debugf("Tracer '%s' created", name)
	}

	return tr
}

// ForceFlush exports spans that have not been exported yet to Dynatrace Cluster
func (p *DtTracerProvider) ForceFlush(ctx context.Context) error {
	if p.processor == nil {
		return errInvalidSpanProcessor
	}

	return p.processor.forceFlush(ctx)
}

// Shutdown stops exporting goroutine and exports all remaining spans to Dynatrace Cluster.
// It executes only once, subsequent call does nothing.
func (p *DtTracerProvider) Shutdown(ctx context.Context) error {
	if p.processor == nil {
		return errInvalidSpanProcessor
	}

	return p.processor.shutdown(ctx)
}
