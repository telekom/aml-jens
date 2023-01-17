TEST_DATA_DIR="./test/testdata"
VERSION ?= "00.00-99"
all: help

build-binaries: # Build executables under cmd/*
build-binaries: clean-binaries
	@echo Building executables
	@go build -v -o ./bin/ ./cmd/*
	@chmod +x ./bin/*
clean-binaries:
	@rm -rf ./bin/*
clean-packages:
	@rm -rf ./out/*
test: # Run go tests
test: pre-test
	@go test ./...
pre-test:
	@cp -rf ${TEST_DATA_DIR} cmd/drbenchmark/internal
	@cp -rf ${TEST_DATA_DIR} internal/persistence/datatypes

package: # Create a (.deb) package containing all executabels
package: test build-binaries
	@mkdir -p out 
	@cd deployments/debian && VERSION=$(VERSION) $(MAKE) build
	@echo Removing artifacts
	@cd deployments/debian && VERSION=$(VERSION) $(MAKE) clean

help: # Display this help
	@echo 'Help:'
	@cat makefile | grep ': \#' | grep --invert-match column | sed "s/^/  /" | column -s"\#" -t