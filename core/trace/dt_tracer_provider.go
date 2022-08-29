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
	"errors"
	"sync"
	"time"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"github.com/dynatrace-oss/opentelemetry-exporter-go/core/configuration"
	"github.com/dynatrace-oss/opentelemetry-exporter-go/core/internal/logger"
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

func NewTracerProvider(opts ...sdktrace.TracerProviderOption) (*DtTracerProvider, error) {
	config, err := configuration.GlobalConfigurationProvider.GetConfiguration()
	if err != nil {
		return nil, err
	}

	tp := &DtTracerProvider{
		TracerProvider: sdktrace.NewTracerProvider(opts...),
		mu:             sync.Mutex{},
		wrappedTracers: make(map[trace.Tracer]*dtTracer),
		processor:      newDtSpanProcessor(config),
		logger:         logger.NewComponentLogger("TracerProvider"),
		config:         config,
	}

	tp.logger.Debug("TracerProvider created")
	return tp, nil
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
