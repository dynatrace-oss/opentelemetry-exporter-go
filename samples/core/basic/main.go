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
