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

package trace

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTracerProviderCreatesDtTracer(t *testing.T) {
	tp, _ := newDtTracerProviderWithTestExporter()
	tr := tp.Tracer("Dynatrace Tracer")
	require.IsType(t, &dtTracer{}, tr)
}

func TestTracerProviderReturnsTheSameInstanceOfDtTracer(t *testing.T) {
	tp, _ := newDtTracerProviderWithTestExporter()

	tracerName := "Dynatrace Tracer"
	tr := tp.Tracer(tracerName)
	tr2 := tp.Tracer(tracerName)

	require.Equal(t, tr, tr2)
}

func TestTracerProviderShutdown(t *testing.T) {
	tp, _ := newDtTracerProviderWithTestExporter()
	require.Zero(t, tp.processor.exportingStopped)

	err := tp.Shutdown(context.Background())

	require.NoError(t, err, "shutdown has failed")
	require.NotZero(t, tp.processor.exportingStopped, "Exporting goroutine has not been stopped")
}

func TestTracerProviderForceFlush(t *testing.T) {
	tp, _ := newDtTracerProviderWithTestExporter()

	require.Zero(t, tp.processor.exportingStopped)
	require.Zero(t, tp.processor.spanWatchlist.len())

	tr := tp.Tracer("Dynatrace Tracer")

	generateSpans(tr, spanGeneratorOptions{
		numSpans:   20,
		endedSpans: true,
	})

	require.EqualValues(t, tp.processor.spanWatchlist.len(), 20)
	tp.ForceFlush(context.Background())
	require.EqualValues(t, tp.processor.spanWatchlist.len(), 0)
}
