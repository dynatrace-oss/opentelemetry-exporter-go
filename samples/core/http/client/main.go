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

package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
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

func main() {
	// Configure zipkin exporter
	const zipkinCollectorURL string = "http://localhost:9411/api/v2/spans"
	opts := configureZipkinExporter(zipkinCollectorURL)

	// Setup Dynatrace TracerProvider as a global TracerProvider
	tp := dtTrace.NewTracerProvider(opts...)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(dtTrace.NewTextMapPropagator())

	// Create HTTP client wrapped with OpenTelemetry
	client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
	req, err := http.NewRequest("GET", "http://localhost:8080", nil)
	if err != nil {
		fmt.Printf("Can not create HTTP request: %s", err)
		return
	}

	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("HTTP request has failed: %s", err)
		return
	}

	defer func() { _ = res.Body.Close() }()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("Can not read HTTP body: %s", err)
		return
	}

	fmt.Printf("HTTP response: %s", body)

	// Wait for user input before finish
	_, _ = bufio.NewReader(os.Stdin).ReadBytes('\n')
}
