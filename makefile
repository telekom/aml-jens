.PHONY: all test clean clean-go-testdata

BUILD_DIR = "build"
DEB_DIR = "build/deb"
DEB_PACKAGE = "jens-cli"
VERSION ?="99.99"

default: clean-local-logs test go-build-dev

deb-package: clean go-build-release
	@echo building package
	./$(DEB_DIR)/create-deb-files.sh $(VERSION)

clean: clean-deb-packages clean-local-logs clean-go-testdata
	@echo Removing executables
	@rm -f "$(BUILD_DIR)/drplay" "$(BUILD_DIR)/drshow"
	@rm -f "$(BUILD_DIR)/drbenchmark" "$(BUILD_DIR)/drbcreate"
	

go-build-dev: 
	@echo Building executables
	@go build -ldflags "-s" -o $(BUILD_DIR) ./cmd/*
	@chmod +x ./cmd/dr*
	@echo Done

go-build-release:
	@echo Building executables
	@go build -o $(BUILD_DIR) ./cmd/*
	@chmod +x ./cmd/dr*

test: | go-test clean-go-testdata

go-test: go-pre-test
	@echo Testing
	@go test ./...
	

go-pre-test:
	@cp -rf testdata drbenchmark
	@cp -rf testdata cmd/drplay/
	@cp -rf testdata drcommon/persistence/jsonp
	@cp -rf testdata drcommon/persistence/datatypes

clean-go-testdata:
	@echo Removing Testdata
	@rm -rf drbenchmark/testdata
	@rm -rf testdata cmd/drplay/testdata
	@rm -rf drcommon/persistence/jsonp/testdata
	@rm -rf drcommon/persistence/datatypes/testdata


clean-deb-packages:
	@echo Removing debian-packages
	@rm -rf $(DEB_DIR)/$(DEB_PACKAGE)-[0-9]*

clean-local-logs:
	@rm -rf dr.log
	@rm -f /etc/jens-cli/dr.log
