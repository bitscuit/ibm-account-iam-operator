VERSION ?= 1.0.0

BUNDLE_IMG ?= ibm-account-iam-operator-bundle:v$(VERSION)

IMG ?= ibm-account-iam-operator:v$(VERSION)

YQ_VERSION ?= v4.44.1

# change the image to dev when applying deployment manifests
deploy: IMG = quay.io/bedrockinstallerfid/ibm-account-iam-operator:dev

# change the image tag from VERSION to tag when building dev image
docker-build-dev: IMG = quay.io/bedrockinstallerfid/ibm-account-iam-operator:dev

.PHONY: docker-build-dev
docker-build-dev: build
	$(CONTAINER_TOOL) build -t ${IMG} .
	$(MAKE) docker-push IMG=${IMG}

clean-before-commit:
	cd config/manager && $(KUSTOMIZE) edit set image controller=controller:latest
	cp ./config/manager/manager.yaml ./config/manager/tmp.yaml
	sed -e 's/Always/IfNotPresent/g' ./config/manager/tmp.yaml > ./config/manager/manager.yaml
	rm ./config/manager/tmp.yaml

.PHONY: yq
YQ ?= $(LOCALBIN)/yq
yq: ## Download operator-sdk locally if necessary.
ifeq (,$(wildcard $(YQ)))
ifeq (, $(shell which yq 2>/dev/null))
	@{ \
	set -e ;\
	mkdir -p $(dir $(YQ)) ;\
	OS=$(shell go env GOOS) && ARCH=$(shell go env GOARCH) && \
	curl -sSLo $(YQ) https://github.com/mikefarah/yq/releases/download/$(YQ_VERSION)/yq_$${OS}_$${ARCH} ;\
	chmod +x $(YQ) ;\
	}
else
YQ = $(shell which yq)
endif
endif

