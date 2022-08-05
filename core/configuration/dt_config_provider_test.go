package configuration

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockConfigFileReader struct {
	fileConfig fileConfig
}

func (m *mockConfigFileReader) readConfigFromFile() (fileConfig, error) {
	return m.fileConfig, nil
}

func createMockConfigFileReader(fileConfig fileConfig) *mockConfigFileReader {
	return &mockConfigFileReader{
		fileConfig: fileConfig,
	}
}

func createMockConfigFileReaderWithRequiredFields() *mockConfigFileReader {
	fileConfig := fileConfig{
		Tenant:    "tenant",
		ClusterID: 123,
	}
	fileConfig.Connection.BaseUrl = "http://localhost:8080"
	fileConfig.Connection.AuthToken = "authToken"

	return createMockConfigFileReader(fileConfig)
}

func createMockConfigFileReaderWithCompleteConfig() *mockConfigFileReader {
	fileConfig := fileConfig{
		Tenant:    "tenant",
		ClusterID: 123,
	}
	fileConfig.Connection.BaseUrl = "http://localhost:8080"
	fileConfig.Connection.AuthToken = "authToken"

	fileConfig.RUM.ClientIpHeaders = []string{"ip_header_1", "ip_header_2", "ip_header_3"}

	fileConfig.Testability.SpanProcessingIntervalMs = 999
	fileConfig.Testability.KeepAliveIntervalMs = 1000
	fileConfig.Testability.MetricCollectionIntervalMs = 2000
	fileConfig.Testability.MetricCollectionsPerExport = 3

	fileConfig.Logging.Destination = LoggingDestination_Stderr
	fileConfig.Logging.Go.Flags = "f1=true,f2=false,f3=true"

	fileConfig.Debug.AddStackOnStart = true

	return createMockConfigFileReader(fileConfig)
}

func TestDefaultConfigValues(t *testing.T) {
	mockConfigFileReader := createMockConfigFileReaderWithRequiredFields()
	config, _ := loadConfiguration(mockConfigFileReader)

	// If these values are not explicitly defined in the config, they should
	// have these default values.
	assert.Equal(t, config.LoggingDestination, LoggingDestination_Off)
	assert.Equal(t, config.RumClientIpHeaders, []string{"forwarded", "x-forwarded-for"})
	assert.Equal(t, config.SpanProcessingIntervalMs, DefaultSpanProcessingIntervalMs)
}

func TestConfigurationViaEnvironment_EmptyConfigFile(t *testing.T) {
	defer os.Clearenv()
	os.Setenv("DT_CLUSTER_ID", "123")
	os.Setenv("DT_TENANT", "tenant")
	os.Setenv("DT_CONNECTION_BASE_URL", "http://1111:2222")
	os.Setenv("DT_CONNECTION_AUTH_TOKEN", "authToken")
	os.Setenv("DT_TESTABILITY_SPAN_PROCESSING_INTERVAL_MS", "999")
	os.Setenv("DT_RUM_CLIENT_IP_HEADERS", "header1:header2")
	os.Setenv("DT_DEBUG_ADD_STACK_ON_START", "true")
	os.Setenv("DT_LOGGING_DESTINATION", "stdout")
	os.Setenv("DT_LOGGING_GO_FLAGS", "flag1=true,flag2=false")

	mockConfigFileReader := createMockConfigFileReader(fileConfig{})
	config, _ := loadConfiguration(mockConfigFileReader)

	// Config values should be derived from environment variables when available,
	// even if the config file is empty or does not exist.
	assert.Equal(t, config.ClusterId, int32(123))
	assert.Equal(t, config.Tenant, "tenant")
	assert.Equal(t, config.TenantId(), int32(1238414539))
	assert.Equal(t, config.BaseUrl, "http://1111:2222")
	assert.Equal(t, config.AuthToken, "authToken")
	assert.Equal(t, config.SpanProcessingIntervalMs, 999)
	assert.Equal(t, config.RumClientIpHeaders, []string{"header1", "header2"})
	assert.Equal(t, config.DebugAddStackOnStart, true)
	assert.Equal(t, config.LoggingDestination, LoggingDestination_Stdout)
	assert.Equal(t, config.LoggingFlags, "flag1=true,flag2=false")
}

func TestConfigurationViaEnvironment_NonEmptyConfigFile(t *testing.T) {
	defer os.Clearenv()
	os.Setenv("DT_CLUSTER_ID", "123")
	os.Setenv("DT_TENANT", "tenant")
	os.Setenv("DT_CONNECTION_BASE_URL", "http://1111:2222")
	os.Setenv("DT_CONNECTION_AUTH_TOKEN", "authToken")
	os.Setenv("DT_TESTABILITY_SPAN_PROCESSING_INTERVAL_MS", "999")
	os.Setenv("DT_RUM_CLIENT_IP_HEADERS", "header1:header2")
	os.Setenv("DT_DEBUG_ADD_STACK_ON_START", "true")
	os.Setenv("DT_LOGGING_DESTINATION", "stdout")
	os.Setenv("DT_LOGGING_GO_FLAGS", "flag1=true,flag2=false")

	mockConfigFileReader := createMockConfigFileReaderWithCompleteConfig()
	config, _ := loadConfiguration(mockConfigFileReader)

	// Config values should be derived from environment variables when available,
	// even if the values are defined in the config file.
	assert.Equal(t, config.ClusterId, int32(123))
	assert.Equal(t, config.Tenant, "tenant")
	assert.Equal(t, config.TenantId(), int32(1238414539))
	assert.Equal(t, config.BaseUrl, "http://1111:2222")
	assert.Equal(t, config.AuthToken, "authToken")
	assert.Equal(t, config.SpanProcessingIntervalMs, 999)
	assert.Equal(t, config.RumClientIpHeaders, []string{"header1", "header2"})
	assert.Equal(t, config.DebugAddStackOnStart, true)
	assert.Equal(t, config.LoggingDestination, LoggingDestination_Stdout)
	assert.Equal(t, config.LoggingFlags, "flag1=true,flag2=false")
}
