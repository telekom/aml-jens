
BUILD_DIR ?="../_build"
ASSETS_DIR = "../../assets"
BIN_DIR = "../../bin"
OUT_DIR = "../../out"
all: build
build: | pre-build copy-binaries
	@echo [2] create .deb package
	@sed --in-place -e "s%TEMPLATE_VERSION%${VERSION}%g" ${BUILD_DIR}/DEBIAN/control
	@dpkg-deb --build ${BUILD_DIR} > /dev/null
	@mv ${BUILD_DIR}.deb $(OUT_DIR)/jens-cli_$(VERSION)_x86_64.deb
pre-build:
	@echo [0] copy relevant package information
	@mkdir -p ${BUILD_DIR}/etc/jens-cli
	@cp ${ASSETS_DIR}/default_bundle/* ${BUILD_DIR}/etc/jens-cli/
	@mkdir -p ${BUILD_DIR}/etc/bash-completion/completions
	@cp ${ASSETS_DIR}/bash-completion/* ${BUILD_DIR}/etc/bash-completion/completions/
	@mkdir -p ${BUILD_DIR}/usr/share/man/man1
	@cp ${ASSETS_DIR}/man/* ${BUILD_DIR}/usr/share/man/man1/
	@mkdir -p ${BUILD_DIR}/doc/jens-cli
	@cp ${ASSETS_DIR}/doc/copyright ${BUILD_DIR}/doc/jens-cli
	@cp -r ./DEBIAN ${BUILD_DIR}
copy-binaries: 
	@echo [1] copy built binaries
	@mkdir -p ${BUILD_DIR}/usr/bin
	@echo adding drplay, drshow and drbenchmark
	@cp ${BIN_DIR}/drplay ${BUILD_DIR}/usr/bin/drplay
	@cp ${BIN_DIR}/drshow ${BUILD_DIR}/usr/bin/drshow
	@cp ${BIN_DIR}/drbenchmark ${BUILD_DIR}/usr/bin/drbenchmark

	
clean:
	@echo Deleting old build
	@rm -rf $(BUILD_DIR)