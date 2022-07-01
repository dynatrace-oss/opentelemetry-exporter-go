package configuration

import (
	"encoding/json"
	"os"
	"path/filepath"
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
	OpenTelemetry struct {
		DisabledSensors       []string
		OverrideMaxApiVersion string
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

// Looks for a config file "dtconfig.json" in the executable's directory and attempts to parse it.
// Returns an error if the file can't be read or the parsing fails.
func (j *jsonConfigFileReader) ReadConfigFromFile() (fileConfig, error) {
	cwd, _ := os.Getwd()

	path := filepath.Join(cwd, "dtconfig.json")
	fileData, err := os.ReadFile(path)
	if err != nil {
		return fileConfig{}, err
	}

	var config fileConfig
	err = json.Unmarshal(fileData, &config)
	return config, err
}
