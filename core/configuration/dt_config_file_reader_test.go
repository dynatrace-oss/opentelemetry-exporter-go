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

package configuration

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJsonConfigFileReader_NoErrorForValidFile(t *testing.T) {
	reader := jsonConfigFileReader{}
	_, err := reader.readConfigFromFileByPath("./testdata/dtconfig_test_valid.json")
	assert.NoError(t, err)
}

func TestJsonConfigFileReader_CustomPath(t *testing.T) {
	os.Setenv("DT_CONFIG_FILE_PATH", "./testdata/subfolder/dtconfig_test_valid.json")
	defer os.Unsetenv("DT_CONFIG_FILE_PATH")

	reader := jsonConfigFileReader{}
	cfg, err := reader.readConfigFromFile()
	assert.NoError(t, err)
	assert.Equal(t, "subfolder_config", cfg.Tenant)
}

func TestJsonConfigFileReader_ErrorForInvalidFile(t *testing.T) {
	reader := jsonConfigFileReader{}
	_, err := reader.readConfigFromFileByPath("./testdata/dtconfig_test_invalid.json")
	assert.Error(t, err)
}

func TestJsonConfigFileReader_ErrorForMissingFile(t *testing.T) {
	reader := jsonConfigFileReader{}
	_, err := reader.readConfigFromFileByPath("./testdata/dtconfig_test_missing.json")
	assert.Error(t, err)
}

func TestJsonConfigFileReader_ErrorForEmptyFile(t *testing.T) {
	reader := jsonConfigFileReader{}
	_, err := reader.readConfigFromFileByPath("./testdata/dtconfig_test_empty.json")
	assert.Error(t, err)
}

func TestJsonConfigFileReader_ValidFileIsCorrectlyDeserialized(t *testing.T) {
	reader := jsonConfigFileReader{}
	config, err := reader.readConfigFromFileByPath("./testdata/dtconfig_test_valid.json")

	assert.NoError(t, err)

	assert.Equal(t, config.AgentActive, true)
	assert.Equal(t, config.ClusterID, 12345)
	assert.Equal(t, config.Tenant, "schnitzel")
	assert.Equal(t, config.Connection.BaseUrl, "https://ag.xyz.com")
	assert.Equal(t, config.Connection.AuthToken, "dt0a01.schnitzel.xsdffdedr")
	assert.Equal(t, config.RUM.ClientIpHeaders, []string{"x-forwarded-for"})
	assert.Equal(t, config.Testability.SpanProcessingIntervalMs, 3000)
	assert.Equal(t, config.Testability.KeepAliveIntervalMs, 30000)
	assert.Equal(t, config.Testability.MetricCollectionIntervalMs, 10000)
	assert.Equal(t, config.Testability.MetricCollectionsPerExport, 6)
	assert.Equal(t, config.Logging.Destination, LoggingDestination_Stderr)
	assert.Equal(t, config.Logging.Go.Flags, "Exporter=true,Propagator=false")
	assert.Equal(t, config.Debug.AddStackOnStart, true)
}
