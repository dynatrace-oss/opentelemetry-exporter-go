package main

import (
	"bufio"
	"context"
	"log"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"core/configuration"
	"core/export"
)

func main() {
	config, err := configuration.GlobalConfigurationProvider.GetConfiguration()
	if err != nil {
		log.Printf("Cannot get configuration: %s", err)
		return
	}

	processor := export.NewDtSpanProcessor(config)
	opt := sdktrace.WithSpanProcessor(processor)
	otel.SetTracerProvider(sdktrace.NewTracerProvider(opt))

	tracer := otel.Tracer("BasicOtelUseCase")

	ctx, spanA := tracer.Start(context.Background(), "Span A")
	endTimestamp := trace.WithTimestamp(time.Now().Add(750 * time.Millisecond))
	spanA.End(endTimestamp)

	ctx, spanB := tracer.Start(ctx, "Span B")
	endTimestamp = trace.WithTimestamp(time.Now().Add(1250 * time.Millisecond))
	spanB.End(endTimestamp)

	err = processor.ForceFlush(ctx)
	if err != nil {
		log.Printf("Can not perform flush operation: %s", err)
	}

	ctx, spanC := tracer.Start(ctx, "Span C")
	endTimestamp = trace.WithTimestamp(time.Now().Add(300 * time.Millisecond))
	spanC.End(endTimestamp)

	err = processor.Shutdown(ctx)
	if err != nil {
		log.Printf("Can not perform shutdown operation: %s", err)
	}

	// wait for user input before finish
	_, _ = bufio.NewReader(os.Stdin).ReadBytes('\n')
}
