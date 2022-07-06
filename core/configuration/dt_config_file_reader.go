package configuration

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
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
	readConfigFromFile() (fileConfig, error)
}

type jsonConfigFileReader struct {
}

// ReadConfigFromFile looks for a config file "dtconfig.json" in the current directory and attempts to parse it.
// Returns an error if the file can't be read or the parsing fails.
func (j *jsonConfigFileReader) readConfigFromFile() (fileConfig, error) {
	return j.readConfigFromFileByPath("./dtconfig.json")
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
