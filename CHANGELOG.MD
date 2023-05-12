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
