package export

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestDtSpanMetadataSkipNewNonEndedSpan(t *testing.T) {
	otel.SetTracerProvider(sdktrace.NewTracerProvider())
	tr := otel.Tracer("SDK Tracer")

	_, nonEndedSpan := tr.Start(context.Background(), "Span A")
	spanMetadata := newDtSpanMetadata(nonEndedSpan.(sdktrace.ReadOnlySpan))

	require.Equal(t, spanMetadata.prepareSend(time.Now().UnixNano()/int64(time.Millisecond)), prepareResultSkip)
	require.EqualValues(t, spanMetadata.lastSentMs, 0)
	require.EqualValues(t, spanMetadata.seqNumber, -1)
	require.Equal(t, spanMetadata.sendState, sendStateNew)
}

func TestDtSpanMetadataSendNonEndedSpanOlderThanUpdateInterval(t *testing.T) {
	otel.SetTracerProvider(sdktrace.NewTracerProvider())
	tr := otel.Tracer("SDK Tracer")

	_, nonEndedSpan := tr.Start(context.Background(), "Span A")
	spanMetadata := newDtSpanMetadata(nonEndedSpan.(sdktrace.ReadOnlySpan))

	// new non-ended span must be sent if it is older than update interval
	sendTime := (time.Now().UnixNano() / int64(time.Millisecond)) + spanMetadata.options.updateIntervalMs
	require.Equal(t, spanMetadata.prepareSend(sendTime), prepareResultSend)
	require.Equal(t, spanMetadata.lastSentMs, sendTime)
	require.Zero(t, spanMetadata.seqNumber)
	require.Equal(t, spanMetadata.sendState, sendStateInitialSend)
}

func TestDtSpanMetadataSendEndedSpan(t *testing.T) {
	otel.SetTracerProvider(sdktrace.NewTracerProvider())
	tr := otel.Tracer("SDK Tracer")

	_, nonEndedSpan := tr.Start(context.Background(), "Span A")
	nonEndedSpan.End()

	spanMetadata := newDtSpanMetadata(nonEndedSpan.(sdktrace.ReadOnlySpan))

	sendTime := time.Now().UnixNano() / int64(time.Millisecond)
	require.Equal(t, spanMetadata.prepareSend(sendTime), prepareResultSend)
	require.Equal(t, spanMetadata.lastSentMs, sendTime)
	require.Zero(t, spanMetadata.seqNumber)
	require.Equal(t, spanMetadata.sendState, sendStateSpanEnded)
}

func TestDtSpanMetadataDropNonEndedOutdatedSpan(t *testing.T) {
	otel.SetTracerProvider(sdktrace.NewTracerProvider())
	tr := otel.Tracer("SDK Tracer")

	_, nonEndedSpan := tr.Start(context.Background(), "Span A")
	spanMetadata := newDtSpanMetadata(nonEndedSpan.(sdktrace.ReadOnlySpan))

	// non-ended span older than openSpanTimeout interval must be dropped
	sendTime := (time.Now().UnixNano() / int64(time.Millisecond)) + spanMetadata.options.openSpanTimeoutMs
	require.Equal(t, spanMetadata.prepareSend(sendTime), prepareResultDrop)
	require.EqualValues(t, spanMetadata.lastSentMs, 0)
	require.EqualValues(t, spanMetadata.seqNumber, -1)
	require.Equal(t, spanMetadata.sendState, sendStateDrop)
}

func TestDtSpanMetadataSendSpanAfterKeepAliveInterval(t *testing.T) {
	otel.SetTracerProvider(sdktrace.NewTracerProvider())
	tr := otel.Tracer("SDK Tracer")

	_, nonEndedSpan := tr.Start(context.Background(), "Span A")
	spanMetadata := newDtSpanMetadata(nonEndedSpan.(sdktrace.ReadOnlySpan))

	// new non-ended span must be sent if it is older than update interval
	sendTime := (time.Now().UnixNano() / int64(time.Millisecond)) + spanMetadata.options.updateIntervalMs
	require.Equal(t, spanMetadata.prepareSend(sendTime), prepareResultSend)
	require.Equal(t, spanMetadata.lastSentMs, sendTime)
	require.Zero(t, spanMetadata.seqNumber)
	require.Equal(t, spanMetadata.sendState, sendStateInitialSend)

	// open span must be sent as keep alive after keep alive interval
	sendTime = sendTime + spanMetadata.options.keepAliveIntervalMs
	require.Equal(t, spanMetadata.prepareSend(sendTime), prepareResultSend)
	require.Equal(t, spanMetadata.lastSentMs, sendTime)
	require.EqualValues(t, spanMetadata.seqNumber, 1)
	require.Equal(t, spanMetadata.sendState, sendStateAlive)
}
