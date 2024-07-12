## General

ROOT_DIR ?= $(abspath $(dir $(firstword $(MAKEFILE_LIST))))
# Local bin folder used 
LOCAL_BIN_DIR ?= $(ROOT_DIR)/bin
# Local scripts folder used
LOCAL_SCRIPTS_DIR ?= $(ROOT_DIR)/scripts


# Must be created if doesn't exist, as some targets place dependencies into it
.PHONY: require-local-bin-dir
require-local-bin-dir:
	mkdir -p $(LOCAL_BIN_DIR)
	

ARCH := $(shell uname -m)
LOCAL_ARCH := "amd64"
ifeq ($(ARCH),x86_64)
    LOCAL_ARCH="amd64"
else ifeq ($(ARCH),ppc64le)
    LOCAL_ARCH="ppc64le"
else ifeq ($(ARCH),s390x)
    LOCAL_ARCH="s390x"
else ifeq ($(ARCH),arm64)
    LOCAL_ARCH="arm64"
else
    $(error "This system's ARCH $(ARCH) isn't recognized/supported")
endif

OS := $(shell uname)
ifeq ($(OS),Linux)
    LOCAL_OS ?= linux
else ifeq ($(OS),Darwin)
    LOCAL_OS ?= darwin
else
    $(error "This system's OS $(OS) isn't recognized/supported")
endif

DEV_VERSION ?= dev # Could be other string or version number
DEV_REGISTRY ?= quay.io/bedrockinstallerfid


ifneq ($(shell echo "$(DEV_VERSION)" | grep -E '^[^0-9]'),)
	TAG := $(DEV_VERSION)
else
	TAG := v$(DEV_VERSION)
endif

DEV_IMG ?= $(DEV_REGISTRY)/ibm-account-iam-operator:$(TAG)

DEV_BUNDLE_IMG ?= $(DEV_REGISTRY)/ibm-account-iam-operator-bundle:$(TAG)

DEV_CATALOG_IMG ?= $(DEV_REGISTRY)/ibm-account-iam-operator-catalog:$(TAG)

bundle: IMG = $(DEV_IMG)

# Change the image to dev when applying deployment manifests
deploy: configure-dev

# Configure the varaiable for the dev build
.PHONY: configure-dev
configure-dev:
	$(eval VERSION := $(DEV_VERSION))
	$(eval IMG := $(DEV_IMG))
	$(eval BUNDLE_IMG := $(DEV_BUNDLE_IMG))
	$(eval CATALOG_IMG := $(DEV_CATALOG_IMG))
	
##@ Development Build
.PHONY: docker-build-dev
docker-build-dev: configure-dev docker-build 

.PHONY: docker-build-push-dev
docker-build-push-dev: docker-build-dev docker-push

.PHONY: bundle-build-dev
bundle-build-dev: configure-dev bundle-build

.PHONY: bundle-build-push-dev
bundle-build-push-dev: bundle-build-dev bundle-push

.PHONY: catalog-build-dev
catalog-build-dev: configure-dev catalog-build

.PHONY: catalog-build-push-dev
catalog-build-push-dev: catalog-build-dev catalog-push

clean-before-commit:
	cd config/manager && $(KUSTOMIZE) edit set image controller=controller:latest
	cp ./config/manager/manager.yaml ./config/manager/tmp.yaml
	sed -e 's/Always/IfNotPresent/g' ./config/manager/tmp.yaml > ./config/manager/manager.yaml
	rm ./config/manager/tmp.yaml


# Test
.PHONY: check
check: ## @code Run the code check
	@echo "Running check for the code."
	@echo "Runing require docker buildx as pre-check"
	$(MAKE) require-docker-buildx
