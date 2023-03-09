# Build / Developer guide

## Building
###
Tests can be run using `make test`.
### For Development-Use
Using `make` ( == `make build-binaries` )

Enable debug messages by specifying the EnvironmentVariable `JENS_DEBUG` to `1`.
### For Release
Using `VERSION=00.01-01 make build-binaries`

## Packaging
Using `make package`.

The Version can be specified by setting VERSION: `VERSION=00.01-01 make package`. It defaults to "-dev-00.00-99". 

The package is generated into folder `out`.