ifeq ($(VERBOSE),)
	VERBOSE = FALSE
endif
# Hide or not the calls depending of VERBOSE
ifeq ($(VERBOSE),TRUE)
	HIDE =
else
	HIDE = @
endif
ifeq ($(DOCKER_REPO),)
	DOCKER_REPO = docker-local
endif

DOCKER_IMAGE ?= artifactory-kfs.habana-labs.com/$(DOCKER_REPO)/$(RELEASE_VERSION)/habanalabs/metric-exporter
DOCKER_TAG ?= $(RELEASE_VERSION)-$(RELEASE_BUILD_ID)
DOCKER_BASE_IMAGE ?= artifactory-kfs.habana-labs.com/docker-local/$(RELEASE_VERSION)/ubuntu22.04/habanalabs/external/base-installer:$(RELEASE_VERSION)-$(RELEASE_BUILD_ID)
SRC_DIR := $(CURDIR)/../..

.PHONY: help validate init build test build/bin

help: ## This help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help

clean: ## clean the build env
	$(HIDE)docker rm -f metric-exporter-test

validate: ## make sure we have the right variables
ifndef RELEASE_VERSION
	$(error RELEASE_VERSION is not set)
endif
ifndef RELEASE_BUILD_ID
	$(error RELEASE_BUILD_ID is not set)
endif
ifndef SRC_DIR
	$(error SRC_DIR is not set)
endif

init: ## init the environemnt - pull image and create folders
	@echo pull base image $(DOCKER_BASE_IMAGE)
	$(HIDE)docker pull -q $(DOCKER_BASE_IMAGE)
	@echo remove old metric-exporter image
	$(HIDE)docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) || true

GCFLAGS ?= all=-spectre=all -N -l
ASMFLAGS ?= all=-spectre=all
LDFLAGS ?= all=-s -w
TARGETARCH ?= amd64
TARGETOS ?= linux

build/bin:
	CGO_ENABLED=0 GOOS=$(TARGETOS) GOARCH=$(TARGETARCH) go build -trimpath -mod=readonly -gcflags="$(GCFLAGS)" -asmflags="$(ASMFLAGS)" -ldflags="$(LDFLAGS)" -a -o bin/metric-exporter $(SRC_DIR)

build: validate init ## build the metric exporter image
	@echo building binaries
	@echo "#!/usr/bin/env bash" > $(SRC_DIR)/scripts/setup-env.sh
	@echo "set -eu" >> $(SRC_DIR)/scripts/setup-env.sh
	@echo "DOCKER_IMG_NAME=$(DOCKER_IMAGE)" >> $(SRC_DIR)/scripts/setup-env.sh
	@echo "DOCKER_IMG_TAG=$(DOCKER_TAG)" >> $(SRC_DIR)/scripts/setup-env.sh
	@echo "BASE_INSTALLER_IMAGE=$(DOCKER_BASE_IMAGE)" >> $(SRC_DIR)/scripts/setup-env.sh

	# point the ds to the appropriate version
	cd $(SRC_DIR) && sed -i 's|vault.habana.ai/gaudi-metric-exporter/metric-exporter:latest|vault.habana.ai/gaudi-metric-exporter/metric-exporter:$(DOCKER_TAG)|g' deploy/manifests/metric-exporter-daemonset.yaml
	cd $(SRC_DIR)/build/docker && ./docker-build.sh local

test: validate ## test the metric exporter image
	@echo start the metric-exporter container in detached mode
	docker run -itd --privileged --network=host --name metric-exporter-test -v /dev:/dev $(DOCKER_IMAGE):$(DOCKER_TAG); \
	curl localhost:41611/metrics | tee metrics.log; \
	docker rm -f metric-exporter-test
	@echo validate the metrics
	grep 'habanalabs_memory_total_bytes{' metrics.log
