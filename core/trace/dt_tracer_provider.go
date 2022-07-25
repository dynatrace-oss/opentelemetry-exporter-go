package trace

import (
	"context"
	"errors"
	"sync"
	"time"

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
}

func NewTracerProvider(opts ...sdktrace.TracerProviderOption) *DtTracerProvider {
	tp := &DtTracerProvider{
		TracerProvider: sdktrace.NewTracerProvider(opts...),
		mu:             sync.Mutex{},
		wrappedTracers: make(map[trace.Tracer]*dtTracer),
		processor:      newDtSpanProcessor(),
		logger:         logger.NewComponentLogger("TracerProvider"),
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

	return measureExecutionTime(ctx, p.processor.forceFlush, "ForceFlush", p.logger)
}

// Shutdown stops exporting goroutine and exports all remaining spans to Dynatrace Cluster.
// It executes only once, subsequent call does nothing.
func (p *DtTracerProvider) Shutdown(ctx context.Context) error {
	if p.processor == nil {
		return errInvalidSpanProcessor
	}

	return measureExecutionTime(ctx, p.processor.shutdown, "Shutdown", p.logger)
}

// measureExecutionTime measure execution time of a given function
// and log a warning message if it takes more than a third of the operation timeout
func measureExecutionTime(ctx context.Context, f func(context.Context) error, opName string, logger *logger.ComponentLogger) error {
	timeout := time.Millisecond * time.Duration(configuration.DefaultFlushOrShutdownTimeoutMs)
	deadline, ok := ctx.Deadline()
	if ok {
		deadlineTimeout := time.Until(deadline)
		if timeout > deadlineTimeout {
			timeout = deadlineTimeout
		}
	}

	start := time.Now()
	err := f(ctx)
	timeTaken := time.Since(start)

	if timeTaken > (timeout / 3) {
		logger.Warnf("%s execution took %s which more than a third of the operation timeout %s", opName, timeTaken, timeout)
	}

	return err
}
