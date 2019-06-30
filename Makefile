# -------------------------------------------------------------
# This makefile defines the following targets
#
#
#   -api[-clean]           - build[/clean] the native hfrdserver binary
#   -api-docker[-clean]    - build[/clean] the hfrd/server docker image
#   -api-clean-all         - clean both hfrdserver binary and docker image
#   -gosdk-docker          - build test modules gosdk

EXECUTABLES ?= go docker git curl
K := $(foreach exec,$(EXECUTABLES),\
	$(if $(shell which $(exec)),some string,$(error "No $(exec) in PATH: Check dependencies")))

BUILD_DIR ?= .build
PROJECT_NAME = hfrd
API_PKGNAME = $(PROJECT_NAME)/api
GOSDK_PKGNAME = $(PROJECT_NAME)/modules/gosdk
STATIC_PATH = api/static
VERSION = $(shell git rev-parse --short HEAD)
UNAME := $(shell uname -m)
ifeq ($(UNAME),s390x)
	API_IMAGE_NAME = hfrd/server:s390x-latest
	GOSDK_IMAGE_NAME = hfrd/gosdk:s390x-latest
	GO_ALPINE_IMAGE_NAME = s390x/golang:alpine
	ALPINE_IMAGE_NAME = s390x\/alpine
else
	API_IMAGE_NAME = hfrd/server:amd64-latest
	GOSDK_IMAGE_NAME = hfrd/gosdk:amd64-latest
	GO_ALPINE_IMAGE_NAME = golang:alpine
	ALPINE_IMAGE_NAME = alpine
endif


# defined in metadata/metadata.go
METADATA_VAR = Version=$(VERSION)
API_GO_LDFLAGS = $(patsubst %,-X $(API_PKGNAME)/metadata.%,$(METADATA_VAR))
GOSDK_GO_LDFLAGS = $(patsubst %,-X $(GOSDK_PKGNAME)/metadata.%,$(METADATA_VAR))

# hfrd api Golang src code files
API_SRC_FILES = $(shell git ls-files  | grep ^api/ | grep .go$ | \
 	grep -v _test.go$ | \
	grep -v ^.git | grep -v ^vendor )

# web UI related files
STATIC_FILES = $(shell git ls-files  | grep ^$(STATIC_PATH) )

# test modules gosdk
GOSDK_SRC_FILES = $(shell git ls-files  | grep ^modules/gosdk/ |grep -v ^test | \
 	grep -v _test.go$ | grep -v .md$ | \
	grep -v ^.git | grep -v .png$ | \
	grep -v ^LICENSE | grep -v ^vendor | grep -v /fixtures/ | grep -v /sample-config/ | \
	grep -v gosdk_example.sh )

DEP_URL = https://raw.githubusercontent.com/golang/dep/master/install.sh

api: $(BUILD_DIR)/bin/hfrdserver $(BUILD_DIR)/bin/static
api-docker: $(BUILD_DIR)/docker/api/hfrdserver $(BUILD_DIR)/docker/api/Dockerfile $(BUILD_DIR)/bin/static
	docker build -t $(API_IMAGE_NAME) $(BUILD_DIR)/docker/api
api-clean-all: api-clean api-docker-clean
api-clean:
	@if [ -f "${BUILD_DIR}/bin/hfrdserver" ]; \
	then \
		rm -f ${BUILD_DIR}/bin/hfrdserver; \
	fi
	@rm -rf $(BUILD_DIR)/bin/static
api-docker-clean:
	docker images -q ${API_IMAGE_NAME}| xargs -I '{}' docker rmi -f '{}'

gosdk-docker: $(BUILD_DIR)/docker/gosdk/gosdk $(BUILD_DIR)/docker/gosdk/Dockerfile
	docker build -t $(GOSDK_IMAGE_NAME) $(BUILD_DIR)/docker/gosdk


# Build api binary for docker
$(BUILD_DIR)/docker/api/hfrdserver: $(API_SRC_FILES)
	@mkdir -p $(@D)
	docker run -i --rm -v $(abspath .)/api:/go/src/$(API_PKGNAME) \
		-v $(abspath $(BUILD_DIR)/docker/api/):/go/bin/hfrd/ \
		${GO_ALPINE_IMAGE_NAME} \
		sh -c "set -e; \
			apk add --no-cache --update git curl; \
			git clone https://github.com/golang/dep.git /go/src/github.com/golang/dep; \
			export GOPATH=/go ; \
			cd /go/src/github.com/golang/dep/cmd/dep/ &&\
			go build && mv dep /go/bin/ ; \
			cd /go/src/$(API_PKGNAME) && \
			dep ensure -vendor-only -v; \
			go build -o /go/bin/hfrd/hfrdserver -ldflags \"$(API_GO_LDFLAGS)\" $(API_PKGNAME)"

# Build test modules gosdk cli binary for docker
$(BUILD_DIR)/docker/gosdk/gosdk: $(GOSDK_SRC_FILES)
	@mkdir -p $(@D)
	docker run -i --rm -v $(abspath .)/modules/gosdk:/go/src/$(GOSDK_PKGNAME) \
		-v $(abspath $(BUILD_DIR)/docker/gosdk/):/go/bin/hfrd/ \
		${GO_ALPINE_IMAGE_NAME} \
		sh -c "set -e; \
			apk add --no-cache --update git curl; \
			git clone https://github.com/golang/dep.git /go/src/github.com/golang/dep; \
			export GOPATH=/go ; \
			cd /go/src/github.com/golang/dep/cmd/dep/ &&\
			go build && mv dep /go/bin/ ; \
			cd /go/src/$(GOSDK_PKGNAME) && \
			dep ensure -vendor-only -v; \
			CGO_ENABLED=0 go build -o /go/bin/hfrd/gosdk -ldflags \"$(GOSDK_GO_LDFLAGS)\" $(GOSDK_PKGNAME); \
			rm -rf vendor"

$(BUILD_DIR)/docker/api/Dockerfile: api/Dockerfile.in $(STATIC_FILES)
	@mkdir -p $(@D)
	@cp api/Dockerfile.in $(@D)/Dockerfile
	@sed -i 's/alpine/${ALPINE_IMAGE_NAME}/g' $(@D)/Dockerfile
	@rm -rf $(@D)/static
	@cp -r $(STATIC_PATH) $(@D)/static

$(BUILD_DIR)/docker/gosdk/Dockerfile: modules/gosdk/Dockerfile.in
	@mkdir -p $(@D)
	@cp modules/gosdk/Dockerfile.in $(@D)/Dockerfile
	@sed -i 's/alpine/${ALPINE_IMAGE_NAME}/g' $(@D)/Dockerfile
	@cp -r modules/gosdk/config $(@D)

# Build api binary locally
$(BUILD_DIR)/bin/hfrdserver: $(BUILD_DIR)/tools/dep $(API_SRC_FILES)
	@mkdir -p $(@D)
	@cd api && ../$(BUILD_DIR)/tools/dep ensure --vendor-only -v
	@go build -o $(abspath $@) -ldflags "$(API_GO_LDFLAGS)" $(API_PKGNAME)
	@echo "Binary available as $@"

$(BUILD_DIR)/bin/static: $(STATIC_FILES)
	@mkdir -p $(@D)
	@rm -rf $(@D)/static
	@cp -r $(STATIC_PATH) $(@D)

$(BUILD_DIR)/tools/dep: Makefile
	@mkdir -p $(@D)
	@if [ ! -e $(BUILD_DIR)/tools/dep ]; \
	then \
		echo "Installing dep for Golang vendor management..."; \
		curl $(DEP_URL) | INSTALL_DIRECTORY=$(abspath $(@D)) sh; \
	else echo "dep binary is already installed, skipping"; \
	fi;
