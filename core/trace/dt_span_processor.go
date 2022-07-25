package trace

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"core/configuration"
	"core/internal/logger"
)

var errInvalidSpanExporter = errors.New("span exporter is invalid")

type flushContext struct {
	ctx                  context.Context
	flushRequestFinished chan struct{}
	err                  error
}

type dtSpanProcessor struct {
	exporter                dtSpanExporter
	spanWatchlist           dtSpanWatchlist
	stopExportingCh         chan struct{}
	stopExportingWait       sync.WaitGroup
	exportingStopped        int32
	shutdownOnce            sync.Once
	flushRequestCh          chan *flushContext
	flushRequestLock        sync.Mutex
	lastFlushRequestContext *flushContext
	periodicSendOpTimer     *time.Timer
	logger                  *logger.ComponentLogger
}

// newDtSpanProcessor creates a Dynatrace span processor that will send spans to Dynatrace Cluster.
func newDtSpanProcessor() *dtSpanProcessor {
	p := &dtSpanProcessor{
		exporter:            newDtSpanExporter(),
		spanWatchlist:       newDtSpanWatchlist(configuration.DefaultMaxSpansWatchlistSize),
		stopExportingCh:     make(chan struct{}, 1),
		exportingStopped:    0,
		flushRequestCh:      make(chan *flushContext, 1),
		periodicSendOpTimer: time.NewTimer(time.Millisecond * time.Duration(configuration.DefaultUpdateIntervalMs)),
		logger:              logger.NewComponentLogger("SpanProcessor"),
	}

	p.stopExportingWait.Add(1)
	go func() {
		defer p.stopExportingWait.Done()
		p.runSpanExportingLoop()
	}()

	return p
}

// onStart adds a newly created span with a corresponding metadata struct to the span watchlist for later processing.
func (p *dtSpanProcessor) onStart(ctx context.Context, s *dtSpan) {
	if p.isExportingStopped() {
		return
	}

	// only recording span has to be sent to Dynatrace Cluster
	span, ok := s.Span.(sdktrace.ReadWriteSpan)
	if !ok {
		return
	}

	p.logger.Debugf("Start span %s", span.Name())

	// TODO: Example how to handle parent span context, will be used by Span Enricher
	parentSpan := trace.SpanFromContext(ctx)
	if s, ok := parentSpan.(*dtSpan); ok {
		// Parent span context holds Dynatrace Span started locally by Dynatrace TracerProvider, thus metadata is available
		p.logger.Debugf("Parent span last sent ms %d", s.metadata.lastSentMs)
	} else if parentSpan.SpanContext().IsValid() && parentSpan.SpanContext().IsRemote() {
		// Parent span context holds remote span.
		// TraceContext Propagator creates remote span by calling trace.ContextWithRemoteSpanContext(ctx, sc) method.
		// The remote span is nonRecordingSpan with SpanContext extracted via TraceContext TextMapPropagator.
		//
		// Important outcome:
		// Remote span isn't create via a configured Dynatrace TracerProvider, thus it can not be handled by us as an internal span.
		//
		// Discuss possible solutions:
		// Dynatrace propagator on a client side will have to inject metadata(FW4) into tracestate,
		// thus it can be extracted and processed on server side.
		json, err := parentSpan.SpanContext().MarshalJSON()
		if err != nil {
			p.logger.Infof("Can not serialize Remote SpanContext: %s", err)
		} else {
			p.logger.Debugf("SpanContext of Remote Span: %s", string(json))
		}
	} else {
		// Parent span is neither Dynatrace nor Remote span, thus can't be handled by Dynatrace span processor
		p.logger.Debugf("Span processor can not handle Parent span of %T type", parentSpan)
	}

	if !p.spanWatchlist.add(s) {
		p.logger.Infof("Span watchlist map is full, can not add metadata for started span: %s", span.Name())
	}
}

// onEnd adds ended span in a processor map for later processing if it wasn't added upon span start call due to a
// reaching max spans limit of the processor map.
func (p *dtSpanProcessor) onEnd(s *dtSpan) {
	if p.isExportingStopped() {
		return
	}

	// only recording span has to be sent to Dynatrace Cluster
	span, ok := s.Span.(sdktrace.ReadWriteSpan)
	if !ok {
		return
	}

	p.logger.Debugf("End span %s", span.Name())
	if !p.spanWatchlist.contains(s) {
		// most likely the span watchlist map was full on span start, so try to re-add span
		if !p.spanWatchlist.add(s) {
			p.logger.Infof("Span watchlist map is full, can not add metadata for ended span: %s", span.Name())
		}
	}
}

// shutdown stops exporting goroutine and exports all remaining spans to Dynatrace Cluster.
// It executes only once, subsequent call does nothing.
func (p *dtSpanProcessor) shutdown(ctx context.Context) error {
	var err error
	p.shutdownOnce.Do(func() {
		p.logger.Debugf("Shutting down is called")

		ctx, cancelFlush := context.WithCancel(ctx)
		defer cancelFlush()

		waitShutdown := make(chan struct{})
		go func() {
			p.requestStopSpanExportingLoop()
			p.stopExportingWait.Wait()

			err = p.sendSpansToExport(ctx, false, exportTypeForceFlush)
			if err != nil {
				p.logger.Warnf("Shutdown operation has failed: %s", err)
			}

			p.lastFlushRequestContext = nil
			close(waitShutdown)
		}()

		// wait until shutdown is finished or the context is cancelled
		select {
		case <-waitShutdown:
			p.logger.Debug("Shutdown operation is finished")
		case <-time.After(time.Millisecond * configuration.DefaultFlushOrShutdownTimeoutMs):
			p.logger.Warn("Shutdown operation timeout is reached")
			cancelFlush()
			err = ctx.Err()
		case <-ctx.Done():
			err = ctx.Err()
			p.logger.Warnf("Context of shutdown operation has been cancelled: %s", err)
		}
	})

	return err
}

// forceFlush waits for any in-progress send operation to be finished and starts a new send operation to serve this
// flush call.
func (p *dtSpanProcessor) forceFlush(ctx context.Context) error {
	if p.isExportingStopped() {
		p.logger.Debug("Ignore Flush operation, exporting is stopped")
		return nil
	}

	p.logger.Debug("Force flush is called")
	ctx, cancelFlush := context.WithCancel(ctx)
	defer cancelFlush()
	flushCtx := &flushContext{
		flushRequestFinished: make(chan struct{}),
		ctx:                  ctx,
		err:                  nil,
	}

	p.flushRequestLock.Lock()
	select {
	case p.flushRequestCh <- flushCtx:
		p.lastFlushRequestContext = flushCtx
		p.flushRequestLock.Unlock()
		p.logger.Debug("Flush operation is scheduled")
	default:
		lastFlushCtx := p.lastFlushRequestContext
		p.flushRequestLock.Unlock()

		p.logger.Debug("Wait until the scheduled flush operation is finished")
		<-lastFlushCtx.ctx.Done()
		return lastFlushCtx.err
	}

	select {
	case <-time.After(time.Millisecond * configuration.DefaultFlushOrShutdownTimeoutMs):
		// the flush operation SHOULD abort any in-progress send operation,
		// thus cancel flush context to inform exporting goroutine
		cancelFlush()
		flushCtx.err = ctx.Err()
		p.logger.Warn("Flush operation timeout is reached")
	case <-flushCtx.flushRequestFinished:
		p.logger.Debug("Flush operation is finished")
	case <-ctx.Done():
		flushCtx.err = ctx.Err()
		p.logger.Warnf("Context of flush operation has been cancelled: %s", ctx.Err())
	}

	return flushCtx.err
}

// runSpanExportingLoop starts exporting loop to process flush and periodic send operations
func (p *dtSpanProcessor) runSpanExportingLoop() {
	defer p.periodicSendOpTimer.Stop()
	for {
		select {
		case <-p.stopExportingCh:
			atomic.StoreInt32(&p.exportingStopped, 1)
			p.logger.Debug("Finish exporting loop")
			return
		case <-p.periodicSendOpTimer.C:
			p.logger.Debug("Execute periodic send operation...")
			err := p.sendSpansToExport(context.Background(), true, exportTypePeriodic)
			if err != nil {
				p.logger.Warnf("Periodic send operation has failed: %s", err)
			}
		case flushCtx := <-p.flushRequestCh:
			p.logger.Debug("Execute flush operation...")
			// stop the periodic send operation, the timer will be reset once exporting will be completed
			if !p.periodicSendOpTimer.Stop() {
				<-p.periodicSendOpTimer.C
			}

			flushCtx.err = p.sendSpansToExport(flushCtx.ctx, true, exportTypeForceFlush)
			if flushCtx.err != nil {
				p.logger.Warnf("Flush send operation has failed: %s", flushCtx.err)
			}

			close(flushCtx.flushRequestFinished)
		}
	}
}

func (p *dtSpanProcessor) isExportingStopped() bool {
	return atomic.LoadInt32(&p.exportingStopped) == 1
}

// requestStopSpanExportingLoop send request to stop span exporting loop
func (p *dtSpanProcessor) requestStopSpanExportingLoop() {
	select {
	case p.stopExportingCh <- struct{}{}:
		p.logger.Debug("Stop exporting is requested")
	default:
		p.logger.Debug("Stop exporting has been already requested")
	}
}

// sendSpansToExport collect spans that are ready to be exported and pass them to Dynatrace Span Exporter
func (p *dtSpanProcessor) sendSpansToExport(ctx context.Context, resetPeriodicSendOpTimer bool, t exportType) error {
	p.logger.Debugf("Preparing spans to export. Export type: %d", t)

	if resetPeriodicSendOpTimer {
		// reset periodic send operation timer at the end of the send operation since
		// the time the operation takes must increase the update interval
		defer p.periodicSendOpTimer.Reset(time.Millisecond * time.Duration(configuration.DefaultUpdateIntervalMs))
	}

	if p.exporter == nil {
		p.logger.Debugf("Can not perform exporting operation, Span Exporter is not initialized")
		return errInvalidSpanExporter
	}

	start := time.Now()
	err := p.exporter.export(ctx, t, p.spanWatchlist.getSpansToExport())
	p.logger.Debugf("Export operation took %s", time.Since(start))

	if err == errNotAuthorizedRequest {
		// not authorized request can be fixed only if the auth token is changed, thus stop exporting loop
		p.logger.Debugf("Stop exporting loop because span export request is not authorized")
		p.requestStopSpanExportingLoop()
	}

	return err
}
