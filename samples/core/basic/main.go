package main

import (
	"bufio"
	"context"
	"log"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	dtTrace "core/trace"
)

func main() {
	// Setup Dynatrace TracerProvider as a global TracerProvider
	tp := dtTrace.NewTracerProvider()
	otel.SetTracerProvider(tp)

	tracer := otel.Tracer("Dynatrace Tracer")

	ctx, spanA := tracer.Start(context.Background(), "Span A")
	endTimestamp := trace.WithTimestamp(time.Now().Add(750 * time.Millisecond))
	spanA.End(endTimestamp)

	ctx, spanB := tracer.Start(ctx, "Span B")
	endTimestamp = trace.WithTimestamp(time.Now().Add(1250 * time.Millisecond))
	spanB.End(endTimestamp)

	err := tp.ForceFlush(ctx)
	if err != nil {
		log.Printf("Can not perform flush operation: %s", err)
	}

	ctx, spanC := tracer.Start(ctx, "Span C")
	endTimestamp = trace.WithTimestamp(time.Now().Add(300 * time.Millisecond))
	spanC.End(endTimestamp)

	err = tp.Shutdown(ctx)
	if err != nil {
		log.Printf("Can not perform shutdown operation: %s", err)
	}

	// Wait for user input before finish
	_, _ = bufio.NewReader(os.Stdin).ReadBytes('\n')
}
