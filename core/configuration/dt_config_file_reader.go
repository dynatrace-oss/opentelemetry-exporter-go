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
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
)

type fileConfig struct {
	AgentActive bool
	ClusterID   int
	Tenant      string
	Connection  struct {
		AuthToken string
		BaseUrl   string
	}
	RUM struct {
		ClientIpHeaders []string
	}
	Testability struct {
		SpanProcessingIntervalMs   int
		KeepAliveIntervalMs        int
		MetricCollectionIntervalMs int
		MetricCollectionsPerExport int
	}
	Logging struct {
		Destination LoggingDestination
		Go          struct {
			Flags string
		}
	}
	Debug struct {
		AddStackOnStart bool
	}
}

type configFileReader interface {
	readConfigFromFile() (fileConfig, error)
}

type jsonConfigFileReader struct {
}

// ReadConfigFromFile looks for a config file "dtconfig.json" in the current directory and attempts to parse it.
// Returns an error if the file can't be read or the parsing fails.
func (j *jsonConfigFileReader) readConfigFromFile() (fileConfig, error) {
	filePath := configFilePath()
	return j.readConfigFromFileByPath(filePath)
}

func configFilePath() string {
	if configFilePathFromEnv := os.Getenv("DT_CONFIG_FILE_PATH"); configFilePathFromEnv != "" {
		return configFilePathFromEnv
	}

	// When running in a Google Cloud Functions Go runtime, we need to find the config file at a different path.
	// For reference on the K_SERVICE environment variable, see:
	// https://cloud.google.com/functions/docs/configuring/env-var#runtime_environment_variables_set_automatically
	_, inGcf := os.LookupEnv("K_SERVICE")
	if inGcf {
		// In the GCF Go runtime, the root directory of your function source code is
		// beneath the current working directory at ./serverless_function_source_code
		// See https://cloud.google.com/functions/docs/concepts/execution-environment#memory-file-system
		return "./serverless_function_source_code/dtconfig.json"
	}

	return "./dtconfig.json"
}

func (j *jsonConfigFileReader) readConfigFromFileByPath(filePath string) (fileConfig, error) {
	fileData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fileConfig{}, err
	}

	var config fileConfig
	decoder := json.NewDecoder(bytes.NewReader(fileData))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&config)
	return config, err
}
