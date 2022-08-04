package trace

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"

	"core/configuration"
)

var testConfig *configuration.DtConfiguration

func TestMain(m *testing.M) {
	defer os.Clearenv()

	os.Setenv("DT_CLUSTER_ID", "123")
	os.Setenv("DT_TENANT", "testTenant")
	os.Setenv("DT_CONNECTION_BASE_URL", "https://example.com")
	os.Setenv("DT_CONNECTION_AUTH_TOKEN", "testAuthToken")
	os.Setenv("DT_LOGGING_DESTINATION", "stdout")
	os.Setenv("DT_LOGGING_GO_FLAGS", "SpanExporter=true,SpanProcessor=true,TracerProvider=true")

	var err error
	testConfig, err = configuration.GlobalConfigurationProvider.GetConfiguration()
	if err != nil {
		fmt.Println("Can not get test configurations: " + err.Error())
	}

	m.Run()
}

type testExporterOptions struct {
	// duration of a single iteration
	iterationIntervalMs int
	// number of iterations per export call
	numIterations int
}

// testExporter test exporter that executes dummy iterations based on given options
type testExporter struct {
	option testExporterOptions
}

func (e *testExporter) export(ctx context.Context, _ exportType, _ dtSpanSet) (err error) {
	for i := e.option.numIterations; i > 0; i-- {
		// simulate exporting operation
		time.Sleep(time.Millisecond * time.Duration(e.option.iterationIntervalMs))

		if err = ctx.Err(); err != nil {
			break
		}
	}

	return
}

func newTestExporter(o testExporterOptions) *testExporter {
	return &testExporter{
		option: o,
	}
}

// newDtTracerProviderWithTestExporter create Dynatrace Tracer Provider with testExporter
func newDtTracerProviderWithTestExporter() (*DtTracerProvider, *testExporter) {
	defaultTextExporterOptions := testExporterOptions{
		iterationIntervalMs: 500,
		numIterations:       1,
	}

	if tp := NewTracerProvider(); tp != nil {
		exporter := newTestExporter(defaultTextExporterOptions)
		tp.processor.exporter = exporter

		return tp, exporter
	}

	return nil, nil
}

type spanGeneratorOptions struct {
	numSpans   int
	endedSpans bool
}

func generateSpans(tr trace.Tracer, option spanGeneratorOptions) {
	wg := &sync.WaitGroup{}
	for i := 0; i < option.numSpans; i++ {
		wg.Add(1)
		go func(idx int) {
			_, span := tr.Start(context.Background(), fmt.Sprintf("Span #%d", idx))
			endTimestamp := trace.WithTimestamp(time.Now().Add(750 * time.Millisecond))
			if option.endedSpans {
				span.End(endTimestamp)
			}
			wg.Done()
		}(i)
	}

	wg.Wait()
}

func TestDtSpanProcessorStartSpans(t *testing.T) {
	tp, _ := newDtTracerProviderWithTestExporter()

	require.Zero(t, tp.processor.exportingStopped)
	require.Zero(t, tp.processor.spanWatchlist.len())

	tr := tp.Tracer("Dynatrace Tracer")

	generateSpans(tr, spanGeneratorOptions{
		numSpans:   20,
		endedSpans: true,
	})
	require.EqualValues(t, tp.processor.spanWatchlist.len(), 20)

	generateSpans(tr, spanGeneratorOptions{
		numSpans:   5,
		endedSpans: true,
	})
	require.EqualValues(t, tp.processor.spanWatchlist.len(), 25)
}

func TestDtSpanProcessorShutdown(t *testing.T) {
	p := newDtSpanProcessor(testConfig)
	require.Zero(t, p.exportingStopped)
	require.Zero(t, p.spanWatchlist.len())

	err := p.shutdown(context.Background())
	require.NoError(t, err, "shutdown has failed")
	require.NotZero(t, p.exportingStopped, "Exporting goroutine has not been stopped")
}

func TestDtSpanProcessorGenerateSpansAndShutdown(t *testing.T) {
	tp, _ := newDtTracerProviderWithTestExporter()

	require.Zero(t, tp.processor.exportingStopped)
	require.Zero(t, tp.processor.spanWatchlist.len())

	tr := tp.Tracer("Dynatrace Tracer")

	generateSpans(tr, spanGeneratorOptions{
		numSpans:   20,
		endedSpans: true,
	})
	require.EqualValues(t, tp.processor.spanWatchlist.len(), 20)

	err := tp.Shutdown(context.Background())
	require.NoError(t, err, "shutdown has failed")
	require.NotZero(t, tp.processor.exportingStopped, "Exporting goroutine has not been stopped")
	require.Zero(t, tp.processor.spanWatchlist.len(), "Spans have not been flushed on shutdown")
}

func TestDtSpanProcessorShutdownCancelContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	p := newDtSpanProcessor(testConfig)
	require.Zero(t, p.exportingStopped)

	err := p.shutdown(ctx)
	require.ErrorIs(t, err, context.Canceled)
}

func TestDtSpanProcessorShutdownTimeoutReached(t *testing.T) {
	p := newDtSpanProcessor(testConfig)
	p.exporter = newTestExporter(testExporterOptions{
		iterationIntervalMs: 500,
		numIterations:       20,
	})
	require.Zero(t, p.exportingStopped)

	err := p.shutdown(context.Background())
	require.ErrorIs(t, err, context.Canceled)
}

func TestDtSpanProcessorForceFlushNonEndedSpan(t *testing.T) {
	tp, _ := newDtTracerProviderWithTestExporter()

	require.Zero(t, tp.processor.exportingStopped)
	require.Zero(t, tp.processor.spanWatchlist.len())

	tr := tp.Tracer("Dynatrace Tracer")

	generateSpans(tr, spanGeneratorOptions{
		numSpans:   20,
		endedSpans: true,
	})
	generateSpans(tr, spanGeneratorOptions{
		numSpans:   5,
		endedSpans: false,
	})
	require.EqualValues(t, tp.processor.spanWatchlist.len(), 25)

	// Non-ended spans still must be watched
	tp.ForceFlush(context.Background())
	require.EqualValues(t, tp.processor.spanWatchlist.len(), 5)
}

func TestDtSpanProcessorForceFlushCancelContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	p := newDtSpanProcessor(testConfig)
	p.exporter = newTestExporter(testExporterOptions{
		iterationIntervalMs: 500,
		numIterations:       20,
	})

	require.Zero(t, p.exportingStopped)

	err := p.forceFlush(ctx)
	require.ErrorIs(t, err, context.Canceled)
}

func TestDtSpanProcessorForceFlushTimeoutReached(t *testing.T) {
	p := newDtSpanProcessor(testConfig)
	p.exporter = newTestExporter(testExporterOptions{
		iterationIntervalMs: 500,
		numIterations:       20,
	})

	require.Zero(t, p.exportingStopped)

	err := p.forceFlush(context.Background())
	require.ErrorIs(t, err, context.Canceled)
}

func TestDtSpanProcessorWaitForScheduledFlushOperation(t *testing.T) {
	tp, exporter := newDtTracerProviderWithTestExporter()
	exporter.option.iterationIntervalMs = 500
	exporter.option.numIterations = 4

	require.Zero(t, tp.processor.exportingStopped)
	require.Zero(t, tp.processor.spanWatchlist.len())

	tr := tp.Tracer("Dynatrace Tracer")

	generateSpans(tr, spanGeneratorOptions{
		numSpans:   5,
		endedSpans: true,
	})

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		tp.ForceFlush(context.Background())
	}()

	// 1st flush operation will be started ASAP, thus the flush channel must be empty
	// all spans are ended, so after evaluation the spans watchlist size must be 0
	require.Eventually(t, func() bool {
		return len(tp.processor.flushRequestCh) == 0 && tp.processor.spanWatchlist.len() == 0
	}, 3*time.Second, 100*time.Millisecond)

	wg.Add(1)
	go func() {
		defer wg.Done()

		generateSpans(tr, spanGeneratorOptions{
			numSpans:   10,
			endedSpans: true,
		})
		tp.ForceFlush(context.Background())
	}()

	require.Eventually(t, func() bool {
		// the 1st flush is running, so the 2nd one must be scheduled, as a result, the flush channel must not be empty
		// 10 spans were started after the 1st flush call, thus span watchlist size must be equal to 10
		return len(tp.processor.flushRequestCh) == 1 && tp.processor.spanWatchlist.len() == 10
	}, 3*time.Second, 100*time.Millisecond)

	wg.Add(1)
	go func() {
		defer wg.Done()

		generateSpans(tr, spanGeneratorOptions{
			numSpans:   5,
			endedSpans: true,
		})

		tp.ForceFlush(context.Background())

		// the 3rd flush call have to wait until the 2nd pending flush will be finished
		// and export spans started upon 3rd flush call
		require.Equal(t, tp.processor.spanWatchlist.len(), 0)
	}()

	wg.Wait()
}
