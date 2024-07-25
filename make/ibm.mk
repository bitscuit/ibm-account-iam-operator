# Override default variable values from top level Makefile #

VERSION ?= 0.0.1
CHANNELS ?= v0.1

# IBM specific # 

ifeq ($(BUILD_LOCALLY),0)
	IMG_REGISTRY ?= docker-na-public.artifactory.swg-devops.com/hyc-cloud-private-integration-docker-local/ibmcom
	export CONFIG_DOCKER_TARGET = config-docker
else
	IMG_REGISTRY ?= docker-na-public.artifactory.swg-devops.com/hyc-cloud-private-scratch-docker-local/ibmcom
endif

IMG ?= $(IMAGE_TAG_BASE):latest

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

ICR_REGISTRY ?= icr.io/cpopen
ICR_IMAGE_TAG_BASE ?= $(ICR_REGISTRY)/ibm-user-management-operator

DEV_VERSION ?= dev # Could be other string or version number
DEV_REGISTRY ?= quay.io/bedrockinstallerfid
DEV_IMAGE_TAG_BASE ?= $(DEV_REGISTRY)/ibm-user-management-operator

ifneq ($(shell echo "$(DEV_VERSION)" | grep -E '^[^0-9]'),)
	TAG := $(DEV_VERSION)
else
	TAG := v$(DEV_VERSION)
endif

DEV_IMG ?= $(DEV_IMAGE_TAG_BASE):$(TAG)
DEV_BUNDLE_IMG ?= $(DEV_IMAGE_TAG_BASE)-bundle:$(TAG)
DEV_CATALOG_IMG ?= $(DEV_IMAGE_TAG_BASE)-catalog:$(TAG)

# Configure the varaiable for the dev build
.PHONY: configure-dev
configure-dev:
	$(eval VERSION := $(DEV_VERSION))
	$(eval IMG := $(DEV_IMG))
	$(eval BUNDLE_IMG := $(DEV_BUNDLE_IMG))
	$(eval CATALOG_IMG := $(DEV_CATALOG_IMG))
	$(MAKE) bundle
	
##@ Development Build
.PHONY: docker-build-dev
docker-build-dev: configure-dev docker-build 

.PHONY: docker-build-push-dev
docker-build-push-dev: configure-dev docker-build-dev docker-push

.PHONY: bundle-build-dev
bundle-build-dev: configure-dev bundle-build

.PHONY: bundle-build-push-dev
bundle-build-push-dev: configure-dev bundle-build-dev bundle-push

.PHONY: catalog-build-dev
catalog-build-dev: configure-dev catalog-build

.PHONY: catalog-build-push-dev
catalog-build-push-dev: configure-dev catalog-build-dev catalog-push

##@ Production Build
.PHONY: docker-build-prod
docker-build-prod: docker-build

.PHONY: docker-build-push-prod
docker-build-push-prod: docker-build-prod docker-push
	$(CONTAINER_TOOL) tag $(IMG) $(IMAGE_TAG_BASE):v$(VERSION)
	$(MAKE) docker-push IMG=$(IMAGE_TAG_BASE):v$(VERSION)

clean-before-commit:
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(ICR_IMAGE_TAG_BASE):latest
	cp ./bundle/manifests/ibm-user-management-operator.clusterserviceversion.yaml ./bundle/manifests/tmp.yaml
	sed -e 's|image: .*|image: $(ICR_IMAGE_TAG_BASE):latest|g' \
		-e 's|Always|IfNotPresent|g' \
		./bundle/manifests/tmp.yaml > ./bundle/manifests/ibm-user-management-operator.clusterserviceversion.yaml
	rm ./bundle/manifests/tmp.yaml

get-cluster-credentials:
	mkdir -p ~/.kube; cp -v /etc/kubeconfig/config ~/.kube; kubectl config use-context default; kubectl get nodes; echo going forward retiring google cloud

config-docker: get-cluster-credentials
	@scripts/makefile-config/artifactory_config_docker.sh

# Test
.PHONY: check
check: ## @code Run the code check
	@echo "Running check for the code."

# Override default variable values or prerequisites from top level Makefile #

bundle: IMG = icr.io/cpopen/ibm-user-management-operator:latest

docker-build: $(CONFIG_DOCKER_TARGET)

# Change the image to dev when applying deployment manifests
deploy: configure-dev

docker-buildx: require-docker-buildx
