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
	"fmt"

	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/dynatrace-oss/opentelemetry-exporter-go/core/configuration"
	"github.com/dynatrace-oss/opentelemetry-exporter-go/core/internal/fw4"
	"github.com/dynatrace-oss/opentelemetry-exporter-go/core/internal/logger"
)

const (
	traceparentHeader = "traceparent"
	tracestateHeader  = "tracestate"
	xDtHeader         = "x-dynatrace"
)

type DtTextMapPropagator struct {
	sdkPropagator propagation.TraceContext
	logger        *logger.ComponentLogger
	config        *configuration.DtConfiguration
}

func NewTextMapPropagator() *DtTextMapPropagator {
	config, err := configuration.GlobalConfigurationProvider.GetConfiguration()
	if err != nil {
		fmt.Println("Dynatrace TextMapPropagator cannot be instantiated due to a configuration error: " + err.Error())
		return nil
	}

	p := &DtTextMapPropagator{
		sdkPropagator: propagation.TraceContext{},
		logger:        logger.NewComponentLogger("TextMapPropagator"),
		config:        config,
	}

	p.logger.Debug("TextMapPropagator created")
	return p
}

func (p *DtTextMapPropagator) Inject(ctx context.Context, carrier propagation.TextMapCarrier) {
	span := dtSpanFromContext(ctx)
	if span == nil {
		p.logger.Debug("Attempted to inject non DT Span")
		p.sdkPropagator.Inject(ctx, carrier)
		return
	}

	spanCtx := span.SpanContext()
	if !spanCtx.IsValid() {
		p.logger.Info("Attempted to inject span with an invalid span context")
		return
	}

	span.metadata.markPropagatedNow()

	var tag *fw4.Fw4Tag
	if parentTag := span.metadata.getFw4Tag(); parentTag == nil {
		// usually we should have the tag in metadata created by Span Enricher
		p.logger.Warnf("There is no FW4 tag for Span - traceId: %s, spanId: %s", spanCtx.TraceID(), spanCtx.SpanID())
		tag = fw4.NewFw4Tag(p.config.ClusterId, p.config.TenantId(), spanCtx)
		span.metadata.setFw4Tag(tag)

		tag = tag.Propagate(spanCtx)
	} else {
		tag = parentTag.Propagate(spanCtx)
	}

	// set x-dynatrace header
	xDt := tag.ToXDynatrace()
	carrier.Set(xDtHeader, xDt)
	p.sdkPropagator.Inject(ctx, carrier)

	if p.logger.DebugEnabled() {
		spanCtx = span.SpanContext()
		p.logger.Debugf("Inject %s: %s - traceId: %s, spanId: %s, traceState: %s  ", xDtHeader, xDt,
			spanCtx.TraceID(), spanCtx.SpanID(), spanCtx.TraceState())
	}
}

func (p *DtTextMapPropagator) Extract(parentCtx context.Context, carrier propagation.TextMapCarrier) context.Context {
	remoteContext := p.sdkPropagator.Extract(parentCtx, carrier)
	remoteSpanCtx := trace.SpanContextFromContext(remoteContext)
	if p.logger.DebugEnabled() {
		if json, err := remoteSpanCtx.MarshalJSON(); err == nil {
			p.logger.Debugf("Remote span context: %s", string(json))
		} else {
			p.logger.Debugf("Can not parse remote span context: %s", err)
		}
	}

	// look for matching FW4 tag in x-dynatrace header
	if xDt := carrier.Get(xDtHeader); len(xDt) > 0 {
		p.logger.Debugf("FW4 tag from x-dynatrace header: %s", xDt)

		if tag, err := fw4.GetMatchingFw4FromXDynatrace(xDt, p.config.TenantId(), p.config.ClusterId); err == nil {
			if remoteSpanCtx.IsValid() && remoteSpanCtx.TraceID() == tag.TraceID {
				return contextWithFw4TagAndUpdatedSpanContext(parentCtx, remoteSpanCtx, tag)
			} else {
				if remoteSpanCtx.IsValid() {
					p.logger.Debugf("Mismatch between x-dynatrace and traceparent traceId, "+
						"discard W3C data, span context: %s, FW4 tag: %s",
						remoteSpanCtx.TraceID(), tag.TraceID)
				}

				// traceId in traceparent and FW4 mismatches, so discard W3C data and use FW4 tag as a new root
				dtSpanCtx := trace.ContextWithRemoteSpanContext(parentCtx, tag.SpanContext())
				return fw4.ContextWithFw4Tag(dtSpanCtx, &tag)
			}
		} else {
			p.logger.Infof("Can not extract FW4 tag from x-dynatrace: %s", err)
		}
	}

	// look for matching FW4 tag in tracestate field
	if tag, err := fw4.GetMatchingFw4FromTracestate(remoteSpanCtx.TraceState(), p.config.TenantId(), p.config.ClusterId); err == nil {
		p.logger.Debugf("FW4 tag from tracestate field: %s", remoteSpanCtx.TraceState())

		if tag.TraceID.IsValid() && remoteSpanCtx.TraceID() != tag.TraceID {
			p.logger.Debugf("Mismatch between @dt and traceparent traceId, span context: %s, FW4 tag: %s",
				remoteSpanCtx.TraceID(), tag.TraceID)
			return parentCtx
		} else {
			return contextWithFw4TagAndUpdatedSpanContext(parentCtx, remoteSpanCtx, tag)
		}
	} else {
		p.logger.Infof("Can not extract FW4 tag from tracestate: %s", err)
	}

	// FW4 tag is found neither in x-dynatrace nor in tracestate, so return remote context without FW4 tag
	return remoteContext
}

func (p *DtTextMapPropagator) Fields() []string {
	return []string{traceparentHeader, tracestateHeader, xDtHeader}
}

func contextWithFw4TagAndUpdatedSpanContext(ctx context.Context, spanCtx trace.SpanContext, tag fw4.Fw4Tag) context.Context {
	spanCtx = fw4.UpdateTracestate(spanCtx, tag)
	spanCtx = fw4.UpdateTraceFlags(spanCtx, tag)
	ctx = trace.ContextWithRemoteSpanContext(ctx, spanCtx)

	return fw4.ContextWithFw4Tag(ctx, &tag)
}
