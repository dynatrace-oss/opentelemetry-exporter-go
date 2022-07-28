package trace

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"core/configuration"
)

func TestSpanWatchlistMaximumSizeIsReached(t *testing.T) {
	tp, _ := newDtTracerProviderWithTestExporter()
	tr := tp.Tracer("Dynatrace Tracer")

	// generate more spans than watchlist can hold
	numSpans := configuration.DefaultMaxSpansWatchlistSize + 10
	generateSpans(tr, spanGeneratorOptions{
		numSpans:   numSpans,
		endedSpans: true,
	})

	require.Equal(t, tp.processor.spanWatchlist.len(), configuration.DefaultMaxSpansWatchlistSize)
}

func TestSpanWatchlistAdd(t *testing.T) {
	tp, _ := newDtTracerProviderWithTestExporter()
	tr := tp.Tracer("Dynatrace Tracer")

	numSpans := 10
	generateSpans(tr, spanGeneratorOptions{
		numSpans:   numSpans,
		endedSpans: true,
	})

	require.Equal(t, tp.processor.spanWatchlist.len(), numSpans)
}

func TestSpanWatchlistExistRemove(t *testing.T) {
	tp, _ := newDtTracerProviderWithTestExporter()
	tr := tp.Tracer("Dynatrace Tracer")

	ctx, spanA := tr.Start(context.Background(), "Span A")
	defer func() { spanA.End() }()

	_, spanB := tr.Start(ctx, "Span B")
	defer func() { spanB.End() }()

	require.True(t, tp.processor.spanWatchlist.contains(spanA.(*dtSpan)))
	tp.processor.spanWatchlist.remove(spanA.(*dtSpan))
	require.False(t, tp.processor.spanWatchlist.contains(spanA.(*dtSpan)))
}

func TestSpanWatchlistSpansToExport(t *testing.T) {
	tp, _ := newDtTracerProviderWithTestExporter()
	tr := tp.Tracer("Dynatrace Tracer")

	numSpans := 10
	generateSpans(tr, spanGeneratorOptions{
		numSpans:   numSpans,
		endedSpans: true,
	})

	_, _ = tr.Start(context.Background(), "Non-ended span")

	// all finished spans have to be exported
	spans := tp.processor.spanWatchlist.getSpansToExport()
	require.Equal(t, len(spans), numSpans)
}
