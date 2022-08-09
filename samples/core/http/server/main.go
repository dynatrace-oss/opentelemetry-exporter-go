package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"

	dtTrace "core/trace"
)

func configureZipkinExporter(collectorURL string) []sdktrace.TracerProviderOption {
	var opts []sdktrace.TracerProviderOption
	if exp, err := zipkin.New(collectorURL, zipkin.WithLogger(log.New(os.Stdout, "Zipkin: ", log.Ldate|log.Ltime|log.Llongfile))); err == nil {
		opts = append(opts, sdktrace.WithBatcher(exp))
		opts = append(opts, sdktrace.WithResource(resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceNameKey.String("Zipkin service"))))
	} else {
		fmt.Printf("Can not configure Zipkin exporter: %s", err)
	}

	return opts
}

// http://localhost:<port>/
func helloHandler(w http.ResponseWriter, _ *http.Request) {
	_, _ = fmt.Fprintln(w, "Hello from Golang")
}

func main() {
	// Configure zipkin exporter
	const zipkinCollectorURL string = "http://localhost:9411/api/v2/spans"
	opts := configureZipkinExporter(zipkinCollectorURL)

	// Setup Dynatrace TracerProvider as a global TracerProvider
	tp := dtTrace.NewTracerProvider(opts...)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(dtTrace.NewTextMapPropagator())

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Printf("Can not start listener: %s", err)
		return
	}

	mux := http.NewServeMux()

	// Register HTTP handler wrapped with OpenTelemetry
	mux.Handle("/", otelhttp.NewHandler(http.HandlerFunc(helloHandler), "ServerHelloHandler"))

	fmt.Println("Starting HTTP server on port 8080")
	srv := &http.Server{Handler: mux}
	srv.Serve(listener) //nolint:errcheck
}
