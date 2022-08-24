package logger

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"core/configuration"
	"core/internal/version"
)

func logStartupBanner(config *configuration.DtConfiguration) {
	logger := newComponentLogger("Core")

	logger.Infof("OneAgent ODIN Go version .... %s, build date %s", version.FullVersion, version.BuildDate)

	if version.Version == "0.0.0" {
		logger.Infof("This is a development version (%s) and not intended for use in production environments.", version.FullVersion)
	}

	exePath, err := os.Executable()
	if err == nil {
		logger.Infof("Executable path ............. %s", exePath)
	} else {
		logger.Warnf("Can not get executable path: %s", err)
	}

	logger.Infof("Local timezone .............. %s", getUTCOffsetToLocalTimezone())
	logger.Infof("Cluster ID .................. %#x", uint32(config.ClusterId))
	logger.Infof("Tenant ...................... %s", config.Tenant)
	logger.Infof("Agent ID .................... %#x", uint64(config.AgentId))
	logger.Infof("Connection URL .............. %s", config.BaseUrl)
	logger.Infof("Span processing interval .... %d", config.SpanProcessingIntervalMs)
	logger.Infof("Logging destination ......... %s", config.LoggingDestination)
	logger.Infof("Logging flags ............... %s", config.LoggingFlags)
	logger.Infof("Rum ClientIp Headers ........ %s", config.RumClientIpHeaders)
	logger.Infof("Process ID .................. %d", os.Getpid())
	logger.Infof("Command line is ............. %s", os.Args)

	hostname, err := os.Hostname()
	if err == nil {
		logger.Infof("Agent host .................. %s", hostname)
	} else {
		logger.Warnf("Can not get agent hostname: %s", err)
	}

	logger.Infof("Go version .................. %s", runtime.Version())
	logger.Infof("Platform .................... %s %s", runtime.GOOS, runtime.GOARCH)

}

func getUTCOffsetToLocalTimezone() string {
	_, offset := time.Now().Zone()

	d := time.Second * time.Duration(offset)
	d = d.Round(time.Minute)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute

	return fmt.Sprintf("UTC+%02d%02d", h, m)
}
