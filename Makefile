#!/usr/bin/env bash

# Copyright (C) 2025 Intel Corporation

# This program is free software; you can redistribute it and/or modify it
# under the terms of the GNU General Public License version 2 or later, as published
# by the Free Software Foundation.

# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
# GNU General Public License for more details.

# You should have received a copy of the GNU General Public License
# along with this program; if not, see <http://www.gnu.org/licenses/>.

# SPDX-License-Identifier: GPL-2.0-or-later

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
SRC_DIR := $(CURDIR)
UPDATE_TYPE ?= "patch"

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

# upgrade
.PHONY: update
update:
	@if [ "$(UPDATE_TYPE)" = "patch" ]; then \
		GO_MINOR=$$(awk '/^go / {split($$2, v, "."); print v[1] "." v[2]; exit}' go.mod) && \
		go get go@$$GO_MINOR && \
		go get toolchain@go$$GO_MINOR; \
	else \
		go get go@latest && \
		go get toolchain@latest; \
	fi
	go get -u ./... && \
	GO_VERSION=$$(awk '/^go / {print $$2; exit}' go.mod) && \
		sed -i "s/FROM golang:.* AS golang/FROM golang:$$GO_VERSION AS golang/g" build/docker/Dockerfile
	cd rhlml && cargo update
	cd rhlml/hlml-sys && cargo update
	cd rhlml/hlml && cargo update

.PHONY: tidy
tidy: ## Run go mod tidy.
	go mod tidy

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

## Remember to 'export GOTOOLCHAIN=auto' before running this target to use the latest Go toolchain.
.PHONY: upgrade
upgrade: update tidy fmt vet
