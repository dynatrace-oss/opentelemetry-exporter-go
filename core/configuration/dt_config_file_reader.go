package configuration

import (
	"encoding/json"
	"os"
)

type fileConfig struct {
	AgentActive bool
	ClusterID   int
	Tenant      string
	Connection  struct {
		AuthToken string
		BaseUrl   string
		Proxy     string
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
	ReadConfigFromFile() (fileConfig, error)
}

type jsonConfigFileReader struct {
}

// Looks for a config file "dtconfig.json" in the current directory and attempts to parse it.
// Returns an error if the file can't be read or the parsing fails.
func (j *jsonConfigFileReader) ReadConfigFromFile() (fileConfig, error) {
	fileData, err := os.ReadFile("./dtconfig.json")
	if err != nil {
		return fileConfig{}, err
	}

	var config fileConfig
	err = json.Unmarshal(fileData, &config)
	return config, err
}
