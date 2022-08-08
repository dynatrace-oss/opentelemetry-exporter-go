package configuration

import (
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"strings"

	"core/configuration/internal/util"
)

const (
	DefaultFlushExportConnTimeoutMs   = 1000
	DefaultFlushExportDataTimeoutMs   = 5000
	DefaultRegularExportConnTimeoutMs = 10000
	DefaultRegularExportDataTimeoutMs = 60000
)

const (
	DefaultSpanProcessingIntervalMs = 3000
	DefaultKeepAliveIntervalMs      = 25000
	DefaultOpenSpanTimeoutMs        = 115 * 60 * 1000 // 1h 55 mins (in millis)
	DefaultFlushOrShutdownTimeoutMs = DefaultFlushExportConnTimeoutMs + DefaultFlushExportDataTimeoutMs
	DefaultMaxSpansWatchlistSize    = 2048
)

type DtConfiguration struct {
	ClusterId                int32
	Tenant                   string
	tenantId                 int32
	AgentId                  int64
	BaseUrl                  string
	AuthToken                string
	SpanProcessingIntervalMs int
	LoggingDestination       LoggingDestination
	LoggingFlags             string
	RumClientIpHeaders       []string
	DebugAddStackOnStart     bool
}

type LoggingDestination string

const (
	LoggingDestination_Off    LoggingDestination = "off"
	LoggingDestination_Stdout LoggingDestination = "stdout"
	LoggingDestination_Stderr LoggingDestination = "stderr"
)

func (config *DtConfiguration) TenantId() int32 {
	return config.tenantId
}

// ConfigurationProvider
// Usage: create an instance of ConfigurationProvider and call GetConfiguration() to get the configuration.
// You may pass around the ConfigurationProvider or the returned DtConfiguration to other parts of the application.
// You may also use the GlobalConfigurationProvider singleton instead of creating your own instance.
type ConfigurationProvider struct {
	configuration *DtConfiguration
}

var GlobalConfigurationProvider = &ConfigurationProvider{}

// GetConfiguration returns configuration from environment variables or from file.
// Will return a cached configuration when called multiple times.
func (cp *ConfigurationProvider) GetConfiguration() (*DtConfiguration, error) {
	if cp.configuration == nil {
		config, err := loadConfiguration(&jsonConfigFileReader{})
		if err != nil {
			return nil, err
		}
		cp.configuration = config
	}
	return cp.configuration, nil
}

// loadConfiguration consolidates configuration provided by environment variables and by file into a single struct.
// Configuration provided by environment variables overrides configuration provided by file.
func loadConfiguration(configFileReader configFileReader) (*DtConfiguration, error) {
	fileConfig, err := configFileReader.readConfigFromFile()
	if err != nil {
		fmt.Println("Could not read configuration file: " + err.Error())
	}

	config := &DtConfiguration{
		AgentId:                  generateAgentId(),
		ClusterId:                int32(util.GetIntFromEnvWithDefault("DT_CLUSTER_ID", fileConfig.ClusterID)),
		Tenant:                   util.GetStringFromEnvWithDefault("DT_TENANT", fileConfig.Tenant),
		BaseUrl:                  util.GetStringFromEnvWithDefault("DT_CONNECTION_BASE_URL", fileConfig.Connection.BaseUrl),
		AuthToken:                util.GetStringFromEnvWithDefault("DT_CONNECTION_AUTH_TOKEN", fileConfig.Connection.AuthToken),
		SpanProcessingIntervalMs: util.GetIntFromEnvWithDefault("DT_TESTABILITY_SPAN_PROCESSING_INTERVAL_MS", fileConfig.Testability.SpanProcessingIntervalMs),
		RumClientIpHeaders:       util.GetStringSliceFromEnvWithDefault("DT_RUM_CLIENT_IP_HEADERS", fileConfig.RUM.ClientIpHeaders),
		DebugAddStackOnStart:     util.GetBoolFromEnvWithDefault("DT_DEBUG_ADD_STACK_ON_START", fileConfig.Debug.AddStackOnStart),
		LoggingDestination:       LoggingDestination(util.GetStringFromEnvWithDefault("DT_LOGGING_DESTINATION", string(fileConfig.Logging.Destination))),
		LoggingFlags:             util.GetStringFromEnvWithDefault("DT_LOGGING_GO_FLAGS", fileConfig.Logging.Go.Flags),
	}

	// A potential trailing forward slash in BaseUrl value must be gracefully handled
	config.BaseUrl = strings.TrimSuffix(config.BaseUrl, "/")

	setDefaultConfigValues(config)
	if validationErr := validateConfiguration(config); validationErr != nil {
		return nil, validationErr
	}

	config.tenantId = util.CalculateTenantId(config.Tenant)

	return config, nil
}

func setDefaultConfigValues(config *DtConfiguration) {
	if config.LoggingDestination == "" {
		config.LoggingDestination = LoggingDestination_Off
	}

	if config.RumClientIpHeaders == nil {
		config.RumClientIpHeaders = []string{"forwarded", "x-forwarded-for"}
	}

	if config.SpanProcessingIntervalMs == 0 {
		config.SpanProcessingIntervalMs = DefaultSpanProcessingIntervalMs
	}
}

func validateConfiguration(config *DtConfiguration) error {
	if config.Tenant == "" {
		return errors.New("Tenant must be specified in configuration.")
	}

	if config.ClusterId == 0 {
		return errors.New("ClusterId must be specified in configuration.")
	}

	if config.BaseUrl == "" {
		return errors.New("BaseUrl must be specified in configuration.")
	} else {
		_, err := url.ParseRequestURI(config.BaseUrl)
		if err != nil {
			return errors.New("BaseUrl does does not have valid format.")
		}
	}

	if config.AuthToken == "" {
		return errors.New("AuthToken must be specified in configuration.")
	}

	switch config.LoggingDestination {
	case LoggingDestination_Off, LoggingDestination_Stdout, LoggingDestination_Stderr:
		// valid, do nothing
	default:
		return fmt.Errorf("LoggingDestionation must be one of: %s, %s, %s",
			LoggingDestination_Off, LoggingDestination_Stdout, LoggingDestination_Stderr)
	}

	return nil
}

func generateAgentId() int64 {
	return int64(rand.Uint64())
}
