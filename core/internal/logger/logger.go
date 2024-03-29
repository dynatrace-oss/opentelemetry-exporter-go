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
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dynatrace-oss/opentelemetry-exporter-go/core/configuration"
	"github.com/dynatrace-oss/opentelemetry-exporter-go/core/internal/version"
)

type debugLogFlags map[string]bool
type logKind string

const (
	logKindInfo  logKind = "info"
	logKindWarn  logKind = "warning"
	logKindErr   logKind = "error"
	logKindDebug logKind = "debug"
)

var internalDtLogger dtLogger

type dtLogger struct {
	logger        *log.Logger
	configureOnce sync.Once
	flags         debugLogFlags
}

func (p *dtLogger) enabled() bool {
	return p.logger != nil
}

func (p *dtLogger) debugFlagEnabled(flag string) bool {
	if p.flags == nil {
		return false
	}

	return p.flags[flag]
}

func (p *dtLogger) configure(dest configuration.LoggingDestination, flags debugLogFlags) {
	switch dest {
	case configuration.LoggingDestination_Off:
		p.logger = nil
	case configuration.LoggingDestination_Stdout:
		p.logger = log.New(os.Stdout, "[Dynatrace] ", 0)
	case configuration.LoggingDestination_Stderr:
		p.logger = log.New(os.Stderr, "[Dynatrace] ", 0)
	}

	p.flags = flags
}

func (p *dtLogger) log(kind logKind, component string, msg string) {
	if !p.enabled() {
		return
	}

	utcTime := time.Now().UTC().Format("2006-01-02 15:04:05.000")

	// TODO: find a replacement for thread ID
	logMsg := fmt.Sprintf("%s UTC [%d-00000000] %-7s [%s] %s", utcTime, os.Getpid(), kind, component, msg)
	p.logger.Println(logMsg)
}

type ComponentLogger struct {
	componentName    string
	debugFlagEnabled bool
}

func init() {
	// We do not use the internalDtLogger for this since it might never be correctly configured.
	// We ensure this line is always printed to stderr if a development version of the library is used.
	if version.Version == "0.0.0" {
		fmt.Fprintf(os.Stderr,
			"[Dynatrace] This is a development version (%s) and not intended for use in production environments.\n",
			version.FullVersion)
	}
}

func NewComponentLogger(componentName string) *ComponentLogger {
	internalDtLogger.configureOnce.Do(func() {
		config, err := configuration.GlobalConfigurationProvider.GetConfiguration()
		if err != nil {
			fmt.Println("Dynatrace Logger cannot be instantiated due to a configuration error: " + err.Error())
			return
		}

		flags := parseLogFlags(config.LoggingFlags)
		internalDtLogger.configure(config.LoggingDestination, flags)

		logStartupBanner(config)
	})

	return newComponentLogger(componentName)
}

func newComponentLogger(componentName string) *ComponentLogger {
	logger := &ComponentLogger{componentName: componentName}
	logger.debugFlagEnabled = internalDtLogger.debugFlagEnabled(componentName)
	return logger
}

func (p *ComponentLogger) Enabled() bool {
	return internalDtLogger.enabled()
}

func (p *ComponentLogger) DebugEnabled() bool {
	return p.Enabled() && p.debugFlagEnabled
}

func (p *ComponentLogger) Debug(msg string) {
	if !p.DebugEnabled() {
		return
	}

	internalDtLogger.log(logKindDebug, p.componentName, msg)
}

func (p *ComponentLogger) Debugf(format string, v ...interface{}) {
	if !p.DebugEnabled() {
		return
	}

	internalDtLogger.log(logKindDebug, p.componentName, fmt.Sprintf(format, v...))
}

func (p *ComponentLogger) Info(msg string) {
	internalDtLogger.log(logKindInfo, p.componentName, msg)
}

func (p *ComponentLogger) Infof(format string, v ...interface{}) {
	if !p.Enabled() {
		return
	}

	p.Info(fmt.Sprintf(format, v...))
}

func (p *ComponentLogger) Warn(msg string) {
	internalDtLogger.log(logKindWarn, p.componentName, msg)
}

func (p *ComponentLogger) Warnf(format string, v ...interface{}) {
	if !p.Enabled() {
		return
	}

	p.Warn(fmt.Sprintf(format, v...))
}

func (p *ComponentLogger) Error(msg string) {
	internalDtLogger.log(logKindErr, p.componentName, msg)
}

func (p *ComponentLogger) Errorf(format string, v ...interface{}) {
	if !p.Enabled() {
		return
	}

	p.Error(fmt.Sprintf(format, v...))
}

// parseLogFlags parses debug logging flags stored in the following format "SpanExporter=true,SpanProcessor=false"
func parseLogFlags(str string) debugLogFlags {
	if str == "" {
		return nil
	}

	values := strings.Split(str, ",")
	flags := make(debugLogFlags, len(values))

	for _, keyValue := range values {
		flag := strings.SplitN(keyValue, "=", 2)
		if len(flag) != 2 {
			fmt.Println("Can not split Logger flag on key-value pair: " + keyValue)
			continue
		}

		value, err := strconv.ParseBool(flag[1])
		if err != nil {
			fmt.Printf("Can not parse bool value of Logger flag: %s, err: %s \n", flag, err)
			continue
		}

		flags[flag[0]] = value
	}

	return flags
}
