package export

import (
	"context"
	"errors"
	"log"
	"sync"
	"sync/atomic"
	"time"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// TODO: move to ODIN config package and replace default values with proper values from config package
const (
	defaultUpdateIntervalMs    = 3000
	defaultForceFlushTimeoutMs = 5000
)

var flushTimeoutReached = errors.New("flush operation timeout is reached")

type flushContext struct {
	ctx             context.Context
	flushOpFinished chan error
}

type DtSpanProcessor struct {
	exporter            dtSpanExporter
	metadata            dtSpanMetadataMap
	stopExportingCh     chan struct{}
	stopExportingWait   sync.WaitGroup
	exportingStopped    int32
	shutdownOnce        sync.Once
	flushRequestCh      chan *flushContext
	periodicSendOpTimer *time.Timer
}

// NewDtSpanProcessor creates a Dynatrace span processor that will send spans to Dynatrace Cluster.
func NewDtSpanProcessor() *DtSpanProcessor {
	p := &DtSpanProcessor{
		exporter:            newDtSpanExporter(),
		metadata:            newDtSpanMetadataMap(defaultMaxSpansWatchlistSize),
		stopExportingCh:     make(chan struct{}),
		exportingStopped:    0,
		flushRequestCh:      make(chan *flushContext, 1),
		periodicSendOpTimer: time.NewTimer(time.Millisecond * time.Duration(defaultUpdateIntervalMs)),
	}

	p.stopExportingWait.Add(1)
	go func() {
		defer p.stopExportingWait.Done()
		p.startSpanExportingLoop()
	}()

	return p
}

// OnStart adds a newly created span with a corresponding metadata struct in a processor map for later processing.
func (p *DtSpanProcessor) OnStart(_ context.Context, s sdktrace.ReadWriteSpan) {
	if p.isExportingStopped() {
		return
	}

	log.Printf("DtSpanProcessor: Start span %s", s.Name())

	metadata := newDtSpanMetadata(s)
	p.metadata.add(s.SpanContext(), metadata)
}

// OnEnd adds ended span in a processor map for later processing if it wasn't added upon span start call due to a
// reaching max spans limit of the processor map.
func (p *DtSpanProcessor) OnEnd(s sdktrace.ReadOnlySpan) {
	if p.isExportingStopped() {
		return
	}

	log.Printf("DtSpanProcessor: End span %s", s.Name())
	if !p.metadata.exist(s.SpanContext()) {
		// most likely span metadata map was full on span start, so try to re-add span
		metadata := newDtSpanMetadata(s)
		p.metadata.add(s.SpanContext(), metadata)
	}
}

// Shutdown stops exporting goroutine and exports all remaining spans to Dynatrace Cluster.
// It executes only once, subsequent call does nothing.
func (p *DtSpanProcessor) Shutdown(ctx context.Context) error {
	var err error
	p.shutdownOnce.Do(func() {
		log.Println("DtSpanProcessor: Shutting down")

		ctx, cancelFlush := context.WithCancel(ctx)
		defer cancelFlush()

		waitShutdown := make(chan struct{})
		go func() {
			close(p.stopExportingCh)
			p.stopExportingWait.Wait()
			err = p.exportSpans(ctx, false, exportTypeForceFlush)
			close(waitShutdown)
		}()

		// wait until shutdown is finished or the context is cancelled
		select {
		case <-waitShutdown:
			log.Println("DtSpanProcessor: Shutdown operation has been finished")
		case <-time.After(time.Millisecond * defaultForceFlushTimeoutMs):
			log.Println("DtSpanProcessor: Flush operation timeout is reached")
			cancelFlush()
			err = flushTimeoutReached
		case <-ctx.Done():
			err = ctx.Err()
			log.Printf("DtSpanProcessor: Context of shutdown operation has been cancelled: %s", err)
		}
	})

	return err
}

// ForceFlush waits for any in-progress send operation to be finished and starts a new send operation to serve this
// flush call.
func (p *DtSpanProcessor) ForceFlush(ctx context.Context) error {
	if p.isExportingStopped() {
		log.Println("DtSpanProcessor: the processor is already shut down, ignore flush call")
		return nil
	}

	log.Println("DtSpanProcessor: Force flush")
	var err error

	ctx, cancelFlush := context.WithCancel(ctx)
	defer cancelFlush()
	flushCtx := &flushContext{
		flushOpFinished: make(chan error),
		ctx:             ctx,
	}

	select {
	case p.flushRequestCh <- flushCtx:
		log.Println("DtSpanProcessor: Flush operation is scheduled")
	default:
		log.Println("DtSpanProcessor: There is pending flush operation, ignore flush call")
		return nil
	}

	select {
	case <-time.After(time.Millisecond * defaultForceFlushTimeoutMs):
		log.Println("DtSpanProcessor: Flush operation timeout is reached")
		err = flushTimeoutReached

		// the flush operation SHOULD abort any in-progress send operation,
		// thus cancel flush context to inform exporting goroutine
		cancelFlush()
	case err = <-flushCtx.flushOpFinished:
		log.Println("DtSpanProcessor: Flush operation has been finished")
	case <-ctx.Done():
		err = ctx.Err()
		log.Printf("DtSpanProcessor: Context of flush operation has been cancelled: %s", err)
	}

	return err
}

// startSpanExportingLoop starts exporting loop to process flush and periodic send operations
func (p *DtSpanProcessor) startSpanExportingLoop() {
	defer p.periodicSendOpTimer.Stop()
	for {
		select {
		case <-p.stopExportingCh:
			atomic.StoreInt32(&p.exportingStopped, 1)
			log.Println("DtSpanProcessor: The shutdown has been requested, stop exporting loop")
			return
		case <-p.periodicSendOpTimer.C:
			log.Println("DtSpanProcessor: Execute periodic send operation...")
			err := p.exportSpans(context.Background(), true, exportTypePeriodic)
			if err != nil {
				log.Printf("DtSpanProcessor: periodic send operation has failed: %s", err)
			}
		case flushCtx := <-p.flushRequestCh:
			log.Println("DtSpanProcessor: Execute flush operation...")
			// stop the periodic send operation, the timer will be reset once exporting will be completed
			if !p.periodicSendOpTimer.Stop() {
				<-p.periodicSendOpTimer.C
			}

			err := p.exportSpans(flushCtx.ctx, true, exportTypeForceFlush)
			if err != nil {
				log.Printf("DtSpanProcessor: flush send operation has failed: %s", err)
			}

			flushCtx.flushOpFinished <- err
		}
	}
}

func (p *DtSpanProcessor) isExportingStopped() bool {
	return atomic.LoadInt32(&p.exportingStopped) == 1
}

// exportSpans collect spans that are ready to be exported and pass them to Dynatrace spans exporter
func (p *DtSpanProcessor) exportSpans(ctx context.Context, resetPeriodicSendOpTimer bool, t exportType) error {
	log.Printf("DtSpanProcessor: Spans export call: Export type: %d", t)
	if resetPeriodicSendOpTimer {
		// reset periodic send operation timer at the end of the send operation since
		// the time the operation takes must increase the update interval
		defer p.periodicSendOpTimer.Reset(time.Millisecond * time.Duration(defaultUpdateIntervalMs))
	}

	// TODO: depends on span metadata, collect spans that have to be exported and pass them over to DT span exporter
	err := p.exporter.export(ctx, nil)
	return err
}
