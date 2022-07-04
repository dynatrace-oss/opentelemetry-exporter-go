package configuration

import (
	"fmt"
	"math/rand"
	"net/url"

	"github.com/dynatrace/opentelemetry-exporter-go/core/configuration/util"
)

type DtConfiguration struct {
	ClusterId                int32
	Tenant                   string
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

// Usage: create an instance of ConfigurationProvider and call GetConfiguration() to get the configuration.
// You may pass around the ConfigurationProvider or the returned DtConfiguration to other parts of the application.
type ConfigurationProvider struct {
	configuration *DtConfiguration
}

// Returns configuration from environment variables or from file.
// Will return a cached configuration when called multiple times.
func (cp *ConfigurationProvider) GetConfiguration() (*DtConfiguration, error) {
	if cp.configuration == nil {
		config, err := detectConfiguration(&jsonConfigFileReader{})
		if err != nil {
			return nil, err
		}
		cp.configuration = config
	}
	return cp.configuration, nil
}

// Consolidates configuration provided by environment variables and by file into a single struct.
// Configuration provided by environment variables overrides configuration provided by file.
func detectConfiguration(configFileReader configFileReader) (*DtConfiguration, error) {
	fileConfig, _ := configFileReader.ReadConfigFromFile()

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

	setDefaultConfigValues(config)
	validateConfiguration(config)
	return config, nil
}

func setDefaultConfigValues(config *DtConfiguration) {
	if config.LoggingDestination == "" {
		config.LoggingDestination = LoggingDestination_Off
	}

	if config.RumClientIpHeaders == nil {
		config.RumClientIpHeaders = []string{"forwarded", "x-forwarded-for"}
	}
}

func validateConfiguration(config *DtConfiguration) {
	if config.Tenant == "" {
		panic("Tenant must be specified in configuration.")
	}

	if config.ClusterId == 0 {
		panic("ClusterId must be specified in configuration.")
	}

	if config.BaseUrl == "" {
		panic("BaseUrl must be specified in configuration.")
	} else {
		_, err := url.ParseRequestURI(config.BaseUrl)
		if err != nil {
			panic("BaseUrl does does not have valid format.")
		}
	}

	if config.AuthToken == "" {
		panic("AuthToken must be specified in configuration.")
	}

	switch config.LoggingDestination {
	case LoggingDestination_Off, LoggingDestination_Stdout, LoggingDestination_Stderr:
		// valid, do nothing
	default:
		panic(fmt.Sprintf("LoggingDestionation must be one of: %s, %s, %s",
			LoggingDestination_Off, LoggingDestination_Stdout, LoggingDestination_Stderr))
	}
}

func generateAgentId() int64 {
	return rand.Int63()
}
