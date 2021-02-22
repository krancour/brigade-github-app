SHELL ?= /bin/bash

.DEFAULT_GOAL := build

################################################################################
# Version details                                                              #
################################################################################

# This will reliably return the short SHA1 of HEAD or, if the working directory
# is dirty, will return that + "-dirty"
GIT_VERSION = $(shell git describe --always --abbrev=7 --dirty --match=NeVeRmAtCh)

################################################################################
# Containerized development environment-- or lack thereof                      #
################################################################################

ifneq ($(SKIP_DOCKER),true)
	PROJECT_ROOT := $(dir $(realpath $(firstword $(MAKEFILE_LIST))))
	GO_DEV_IMAGE := brigadecore/go-tools:v0.1.0

	GO_DOCKER_CMD := docker run \
		-it \
		--rm \
		-e SKIP_DOCKER=true \
		-e GOCACHE=/workspaces/brigade-github-app/.gocache \
		-v $(PROJECT_ROOT):/workspaces/brigade-github-app \
		-w /workspaces/brigade-github-app \
		$(GO_DEV_IMAGE)

	KANIKO_IMAGE := brigadecore/kaniko:v0.2.0

	KANIKO_DOCKER_CMD := docker run \
		-it \
		--rm \
		-e SKIP_DOCKER=true \
		-e DOCKER_PASSWORD=$${DOCKER_PASSWORD} \
		-v $(PROJECT_ROOT):/workspaces/brigade-github-app \
		-w /workspaces/brigade-github-app \
		$(KANIKO_IMAGE)

	HELM_IMAGE := brigadecore/helm-tools:v0.1.0

	HELM_DOCKER_CMD := docker run \
	  -it \
		--rm \
		-e SKIP_DOCKER=true \
		-e HELM_PASSWORD=$${HELM_PASSWORD} \
		-v $(PROJECT_ROOT):/workspaces/brigade-github-app \
		-w /workspaces/brigade-github-app \
		$(HELM_IMAGE)
endif

################################################################################
# Docker images and charts we build and publish                                #
################################################################################

ifdef DOCKER_REGISTRY
	DOCKER_REGISTRY := $(DOCKER_REGISTRY)/
endif

ifdef DOCKER_ORG
	DOCKER_ORG := $(DOCKER_ORG)/
endif

DOCKER_IMAGE_PREFIX := $(DOCKER_REGISTRY)$(DOCKER_ORG)

ifdef HELM_REGISTRY
	HELM_REGISTRY := $(HELM_REGISTRY)/
endif

ifdef HELM_ORG
	HELM_ORG := $(HELM_ORG)/
endif

HELM_CHART_PREFIX := $(HELM_REGISTRY)$(HELM_ORG)

ifdef VERSION
	MUTABLE_DOCKER_TAG := latest
else
	VERSION            := $(GIT_VERSION)
	MUTABLE_DOCKER_TAG := edge
endif

IMMUTABLE_DOCKER_TAG := $(VERSION)

################################################################################
# Tests                                                                        #
################################################################################

.PHONY: lint
lint:
	$(GO_DOCKER_CMD) sh -c ' \
		cd v2 && \
		golangci-lint run --config ../golangci.yaml \
	'

.PHONY: test-unit
test-unit:
	$(GO_DOCKER_CMD) sh -c ' \
	cd v2 && \
	go test \
		-v \
		-timeout=60s \
		-race \
		-coverprofile=coverage.txt \
		-covermode=atomic \
		./... \
	'

################################################################################
# Build / Publish                                                              #
################################################################################

.PHONY: build
build: build-brigadier build-images build-cli

.PHONY: build-images
build-images: build-brigade-github-app

.PHONY: build-%
build-%:
	$(KANIKO_DOCKER_CMD) kaniko \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(GIT_VERSION) \
		--dockerfile /workspaces/brigade-github-app/v2/$*/Dockerfile \
		--context dir:///workspaces/brigade-github-app/ \
		--no-push

################################################################################
# Publish                                                                      #
################################################################################

.PHONY: publish
publish: push-images publish-chart

.PHONY: push-images
push-images: push-brigade-github-app

.PHONY: push-%
push-%:
	$(KANIKO_DOCKER_CMD) sh -c ' \
		docker login $(DOCKER_REGISTRY) -u $(DOCKER_USERNAME) -p $${DOCKER_PASSWORD} && \
		kaniko \
			--build-arg VERSION="$(VERSION)" \
			--build-arg COMMIT="$(GIT_VERSION)" \
			--dockerfile /workspaces/brigade-github-app/v2/$*/Dockerfile \
			--context dir:///workspaces/brigade-github-app/ \
			--destination $(DOCKER_IMAGE_PREFIX)$*:$(IMMUTABLE_DOCKER_TAG) \
			--destination $(DOCKER_IMAGE_PREFIX)$*:$(MUTABLE_DOCKER_TAG) \
	'

.PHONY: publish-chart
publish-chart:
	$(HELM_DOCKER_CMD) sh	-c ' \
		helm registry login $(HELM_REGISTRY) -u $(HELM_USERNAME) -p $${HELM_PASSWORD} && \
		cd charts/brigade-github-app && \
		helm dep up && \
		sed -i "s/^version:.*/version: $(VERSION)/" Chart.yaml && \
		sed -i "s/^appVersion:.*/appVersion: $(VERSION)/" Chart.yaml && \
		helm chart save . $(HELM_CHART_PREFIX)brigade-github-app:$(VERSION) && \
		helm chart push $(HELM_CHART_PREFIX)brigade-github-app:$(VERSION) \
	'

################################################################################
# Targets to facilitate hacking on this gateway.                               #
################################################################################

.PHONY: hack-build-%
hack-build-%:
	docker build \
		-f v2/$*/Dockerfile \
		-t $(DOCKER_IMAGE_PREFIX)$*:$(IMMUTABLE_DOCKER_TAG) \
		--build-arg VERSION='$(VERSION)' \
		--build-arg COMMIT='$(GIT_VERSION)' \
		.

.PHONY: hack-push-images
hack-push-images: hack-push-gateway

.PHONY: hack-push-%
hack-push-%: hack-build-%
	docker push $(DOCKER_IMAGE_PREFIX)$*:$(IMMUTABLE_DOCKER_TAG)

.PHONY: hack
hack: hack-push-images
	kubectl get namespace brigade-github-app || kubectl create namespace brigade-github-app
	helm dep up charts/brigade-github-app && \
	helm upgrade brigade-github-app charts/brigade-github-app \
		--install \
		--namespace brigade-github-app
