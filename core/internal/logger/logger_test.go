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

package logger

import (
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"core/configuration"
)

const testMsg string = "test log message"

func TestDontLogByDefault(t *testing.T) {
	require.Nil(t, internalDtLogger.logger)
	require.False(t, internalDtLogger.enabled())
}

func TestLogDestStdOut(t *testing.T) {
	stdout, fakeStdout := replaceStdout(t)
	defer func() { os.Stdout = stdout }()

	internalDtLogger.configure(configuration.LoggingDestination_Stdout, nil)
	require.True(t, internalDtLogger.enabled())

	internalDtLogger.log(logKindErr, "TestLogDestStdOut", testMsg)
	output := read(t, fakeStdout)
	require.Contains(t, output, testMsg)
}

func TestLogDestStdErr(t *testing.T) {
	stderr, fakeStderr := replaceStderr(t)
	defer func() { os.Stderr = stderr }()

	internalDtLogger.configure(configuration.LoggingDestination_Stderr, nil)
	require.True(t, internalDtLogger.enabled())

	internalDtLogger.log(logKindErr, "TestLogDestStdErr", testMsg)
	output := read(t, fakeStderr)
	require.Contains(t, output, testMsg)
}

func TestLogDestOff(t *testing.T) {
	internalDtLogger.configure(configuration.LoggingDestination_Off, nil)
	require.Nil(t, internalDtLogger.logger)
	require.False(t, internalDtLogger.enabled())
}

func TestParsingLoggerDebugFlags(t *testing.T) {
	flags := parseLogFlags("FlagA=true,FlagB=false,FlagC=true")
	require.Equal(t, len(flags), 3)
	require.True(t, flags["FlagA"])
	require.False(t, flags["FlagB"])
	require.True(t, flags["FlagC"])
	require.False(t, flags["Unknown"])
}

func TestComponentLoggerLogLevels(t *testing.T) {
	stdout, fakeStdout := replaceStdout(t)
	defer func() { os.Stdout = stdout }()

	componentName := "TestLogLevels"
	flags := make(debugLogFlags)
	flags[componentName] = true

	internalDtLogger.configure(configuration.LoggingDestination_Stdout, flags)
	require.True(t, internalDtLogger.enabled())

	logger := NewComponentLogger(componentName)
	logger.Debug(testMsg)
	logger.Info(testMsg)
	logger.Warn(testMsg)
	logger.Error(testMsg)

	output := read(t, fakeStdout)
	require.Contains(t, output, fmt.Sprintf("%-7s [%s] %s", logKindDebug, componentName, testMsg))
	require.Contains(t, output, fmt.Sprintf("%-7s [%s] %s", logKindInfo, componentName, testMsg))
	require.Contains(t, output, fmt.Sprintf("%-7s [%s] %s", logKindWarn, componentName, testMsg))
	require.Contains(t, output, fmt.Sprintf("%-7s [%s] %s", logKindErr, componentName, testMsg))
}

func TestComponentLoggerRespectDebugFlags(t *testing.T) {
	stdout, fakeStdout := replaceStdout(t)
	defer func() { os.Stdout = stdout }()

	// enabled debug flags only for ComponentA
	flags := parseLogFlags("ComponentA=true")
	internalDtLogger.configure(configuration.LoggingDestination_Stdout, flags)
	require.True(t, internalDtLogger.enabled())

	loggerComponentA := NewComponentLogger("ComponentA")
	require.True(t, loggerComponentA.Enabled())
	require.True(t, loggerComponentA.DebugEnabled())

	loggerComponentB := NewComponentLogger("ComponentB")
	require.True(t, loggerComponentB.Enabled())
	require.False(t, loggerComponentB.DebugEnabled())

	loggerComponentA.Debug(testMsg)
	loggerComponentA.Info(testMsg)

	loggerComponentB.Debug(testMsg)
	loggerComponentB.Info(testMsg)

	output := read(t, fakeStdout)
	require.Contains(t, output, fmt.Sprintf("%-7s [%s] %s", logKindDebug, "ComponentA", testMsg))
	require.Contains(t, output, fmt.Sprintf("%-7s [%s] %s", logKindInfo, "ComponentB", testMsg))

	// debug message from ComponentB must not be logged since there is no debug flag for it
	require.NotContains(t, output, fmt.Sprintf("%-7s [%s] %s", logKindDebug, "ComponentB", testMsg))
	require.Contains(t, output, fmt.Sprintf("%-7s [%s] %s", logKindInfo, "ComponentB", testMsg))
}

func TestComponentLoggerCheckLogOutputFormat(t *testing.T) {
	stderr, fakeStderr := replaceStderr(t)
	defer func() { os.Stderr = stderr }()

	componentName := "TestLogLevels"
	flags := make(debugLogFlags)
	flags[componentName] = true

	internalDtLogger.configure(configuration.LoggingDestination_Stderr, flags)
	require.True(t, internalDtLogger.enabled())

	logger := NewComponentLogger(componentName)
	logger.Info(testMsg)

	output := read(t, fakeStderr)

	logMsgRexExp, err := regexp.Compile(`^\[Dynatrace] \d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d{3} UTC \[(\d+)-00000000] (info   |warning|error  |debug  ) \[\w+] .*`)
	require.NoError(t, err)
	require.Regexp(t, logMsgRexExp, output)
}

func read(t *testing.T, file *os.File) string {
	const testStdOutBuffSize int = 1024
	const defaultReadContentTimeoutMs = 10000

	buf := make([]byte, testStdOutBuffSize)
	waitReadContent := make(chan struct{})
	go func() {
		n, err := file.Read(buf)
		require.NoError(t, err)
		buf = buf[:n]
		close(waitReadContent)
	}()

	select {
	case <-time.After(time.Millisecond * defaultReadContentTimeoutMs):
		file.Close()
		require.Fail(t, "read content timeout is reached")
	case <-waitReadContent:
	}

	return string(buf)
}

func replaceStdout(t *testing.T) (*os.File, *os.File) {
	r, w, err := os.Pipe()
	require.NoError(t, err)

	origin := os.Stdout
	os.Stdout = w

	return origin, r
}

func replaceStderr(t *testing.T) (*os.File, *os.File) {
	r, w, err := os.Pipe()
	require.NoError(t, err)

	origin := os.Stderr
	os.Stderr = w

	return origin, r
}
