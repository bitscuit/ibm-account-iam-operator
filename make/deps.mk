# Dependency binaries
DOCKER_BUILDX ?= $(LOCAL_BIN_DIR)/buildx
YQ ?= $(LOCAL_BIN_DIR)/yq

# Dependency versions
DOCKER_BUILDX_VERSION ?= v0.12.1
YQ_VERSION ?= v4.44.1

# Dependency check and install scripts, for ease of use
LOCAL_SCRIPTS_MAKEFILE_CHECK_DIR ?= $(LOCAL_SCRIPTS_DIR)/makefile-check
LOCAL_SCRIPTS_MAKEFILE_INSTALL_DIR ?= $(LOCAL_SCRIPTS_DIR)/makefile-install


## Docker Buildx

DOCKER_CLI_PLUGINS ?= ~/.docker/cli-plugins
# Dir must exist for plugin installation (dir will not be created if buildx command already works and passes the check)
.PHONY: require-cli-plugins-dir
require-cli-plugins-dir:
	mkdir -p $(DOCKER_CLI_PLUGINS)

.PHONY: require-docker-buildx
require-docker-buildx:
	@ $(MAKE) check-docker-buildx || $(MAKE) install-docker-buildx

.PHONY: check-docker-buildx
check-docker-buildx: require-local-bin-dir
	@ echo "Checking dependency: docker buildx"
	@ $(LOCAL_SCRIPTS_MAKEFILE_CHECK_DIR)/check-docker-buildx.sh $(DOCKER_BUILDX) $(DOCKER_BUILDX_VERSION) $(DOCKER_CLI_PLUGINS)
	@ echo "Dependency satisfied: docker buildx"

.PHONY: install-docker-buildx
install-docker-buildx: require-local-bin-dir require-cli-plugins-dir
	@ echo "Installing dependency: docker buildx"
	@ $(LOCAL_SCRIPTS_MAKEFILE_INSTALL_DIR)/install-docker-buildx.sh $(DOCKER_BUILDX) $(DOCKER_BUILDX_VERSION) $(DOCKER_CLI_PLUGINS) $(LOCAL_OS) $(LOCAL_ARCH)
	@ echo "Dependency installed: docker buildx"
	@ echo "Checking if installation successful: docker buildx"
	@ $(MAKE) check-docker-buildx

## YQ is a lightweight and portable command-line YAML processor
.PHONY: yq
yq:
ifeq (,$(wildcard $(YQ)))
	ifeq (, $(shell which yq 2>/dev/null))
		@{ \
		set -e ;\
		mkdir -p $(dir $(YQ)) ;\
		curl -sSLo $(YQ) https://github.com/mikefarah/yq/releases/download/$(YQ_VERSION)/yq_$(LOCAL_OS)_$(LOCAL_ARCH) ;\
		chmod +x $(YQ) ;\
		}
	else
		YQ = $(shell which yq)
	endif
endif