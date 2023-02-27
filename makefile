TEST_DATA_DIR="./test/testdata"
VERSION ?= "-dev-00.00-99"
LD_FLAGS = "-X 'github.com/telekom/aml-jens/internal/assets.VERSION=$(VERSION)' -X 'github.com/telekom/aml-jens/internal/assets.BUILD_TIME=$(shell date --iso-8601=seconds)'"
.PHONY: clean-binaries clean-packages test package help build-binaries coverage
all: help

build-binaries: # Build executables under cmd/*
build-binaries: clean-binaries
	@echo Building executables
	@go build -v -ldflags=${LD_FLAGS} -o ./bin/ ./cmd/*
	@chmod +x ./bin/*
clean-binaries:
	@rm -rf ./bin/*
clean-packages:
	@rm -rf ./out/*
clean: # Cleans binaries and deployments
clean: clean-binaries clean-packages
	@echo "Done cleaning"

test: # Run go tests
test: 
	@go test ./...
package: # Create a (.deb) package containing all executabels
package: test build-binaries
	@mkdir -p out 
	@cd deployments/debian && VERSION=$(VERSION) $(MAKE) build
	@echo Removing artifacts
	@cd deployments/debian && VERSION=$(VERSION) $(MAKE) clean
coverage: # Create a html coverage report 
	@go test -coverprofile out/coverage.out ./...
	@go tool cover -html=out/coverage.out -o out/coverage.html
	@echo "View Coverage in ./out/coverage.html"
help: # Display this help
	@echo 'Help:'
	@cat makefile | grep ': \#' | grep --invert-match column | sed "s/^/  /" | column -s"\#" -t