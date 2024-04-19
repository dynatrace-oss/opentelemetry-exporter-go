## 1.289.1 (2024-04-18)

### Fixes

* Update vulnerable dependencies (as a precaution):
    * golang.org/x/sys to version 0.1.0.
    * gopkg.in/yaml.v3 to version 3.0.0.
    * google.golang.org/protobuf to version 1.33.0.

### Features

* Update go version to 1.17

## 1.279.1 (2023-12-21)

### Features

* Update default location for GCF config file to `./serverless_function_source_code/dtconfig.json`
* Allow specifying config path via DT_CONFIG_FILE_PATH env variable
* Improve reliability when sending large traces
* Update semantic constants
* Add pseudo resource attribute propagation
* Export encoded link ID

## 1.267.1 (2023-05-11)

### Fixes

* Ensure that remote span contexts are always sampled, unless an FW4Tag exists which explicitly has the ignored flag.
* Agent ID generation now uses a properly seeded random number generator.

## 1.259.1 (2023-02-15)

### Features

* Mark span export requests as idempotent so that they are automatically retried if they fail due to a connection reset error.


## 1.251.1 (2022-10-11)

### Features

* First release. Please refer to the [README](README.md) for more information.
