package trace

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"core/configuration"
	"core/internal/logger"
)

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

	clusterId int32
	tenantId  int32
}

// newDtSpanProcessor creates a Dynatrace span processor that will send spans to Dynatrace Cluster.
func newDtSpanProcessor(config *configuration.DtConfiguration) *dtSpanProcessor {
	p := &dtSpanProcessor{
		exporter:            newDtSpanExporter(),
		spanWatchlist:       newDtSpanWatchlist(configuration.DefaultMaxSpansWatchlistSize),
		stopExportingCh:     make(chan struct{}),
		exportingStopped:    0,
		flushRequestCh:      make(chan *flushContext, 1),
		periodicSendOpTimer: time.NewTimer(time.Millisecond * time.Duration(configuration.DefaultUpdateIntervalMs)),
		logger:              logger.NewComponentLogger("SpanProcessor"),
		clusterId:           config.ClusterId,
		tenantId:            config.TenantId,
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

	s.metadata = createSpanMetadata(ctx, span, p.clusterId, p.tenantId)

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
			close(p.stopExportingCh)
			p.stopExportingWait.Wait()
			err = p.exportSpans(ctx, false, exportTypeForceFlush)
			p.lastFlushRequestContext = nil
			close(waitShutdown)
		}()

		// wait until shutdown is finished or the context is cancelled
		select {
		case <-waitShutdown:
			p.logger.Debug("Shutdown operation has been finished")
		case <-time.After(time.Millisecond * configuration.DefaultForceFlushTimeoutMs):
			p.logger.Warn("Flush operation timeout is reached")
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
		p.logger.Debug("The processor is already shut down, ignore flush call")
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
	case <-time.After(time.Millisecond * configuration.DefaultForceFlushTimeoutMs):
		// the flush operation SHOULD abort any in-progress send operation,
		// thus cancel flush context to inform exporting goroutine
		cancelFlush()
		flushCtx.err = ctx.Err()
		p.logger.Warnf("Flush operation timeout is reached")
	case <-flushCtx.flushRequestFinished:
		p.logger.Debug("Flush operation has been finished")
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
			p.logger.Debug("The shutdown has been requested, stop exporting loop")
			return
		case <-p.periodicSendOpTimer.C:
			p.logger.Debug("Execute periodic send operation...")
			err := p.exportSpans(context.Background(), true, exportTypePeriodic)
			if err != nil {
				p.logger.Warnf("Periodic send operation has failed: %s", err)
			}
		case flushCtx := <-p.flushRequestCh:
			p.logger.Debug("Execute flush operation...")
			// stop the periodic send operation, the timer will be reset once exporting will be completed
			if !p.periodicSendOpTimer.Stop() {
				<-p.periodicSendOpTimer.C
			}

			flushCtx.err = p.exportSpans(flushCtx.ctx, true, exportTypeForceFlush)
			if flushCtx.err != nil {
				p.logger.Warnf("flush send operation has failed: %s", flushCtx.err)
			}

			close(flushCtx.flushRequestFinished)
		}
	}
}

func (p *dtSpanProcessor) isExportingStopped() bool {
	return atomic.LoadInt32(&p.exportingStopped) == 1
}

// exportSpans collect spans that are ready to be exported and pass them to Dynatrace spans exporter
func (p *dtSpanProcessor) exportSpans(ctx context.Context, resetPeriodicSendOpTimer bool, t exportType) error {
	p.logger.Debugf("Spans export is called. export type: %d", t)
	if resetPeriodicSendOpTimer {
		// reset periodic send operation timer at the end of the send operation since
		// the time the operation takes must increase the update interval
		defer p.periodicSendOpTimer.Reset(time.Millisecond * time.Duration(configuration.DefaultUpdateIntervalMs))
	}

	spans := p.spanWatchlist.getSpansToExport()
	return p.exporter.export(ctx, spans)
}
