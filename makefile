# -
# Variables
# -

-include .env
VERSION := $(shell git rev-parse HEAD)
GIT_SUMMARY := $(shell git describe --tags --dirty --always)
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)

SERVICE := $(shell basename $(shell pwd))
UW_DOCKER_REGISTRY := registry.uw.systems
DOCKER_IMAGE_NAME := $(UW_DOCKER_REGISTRY)/$(SERVICE)

LDFLAGS := -ldflags "-X main.version=$(VERSION)"

CIRCLE_SHA1 ?= $(VERSION)
CIRCLE_BRANCH ?= $(GIT_BRANCH)

ifeq ($(CIRCLE_BRANCH), master)
    DOCKER_TAG := latest
else
    DOCKER_TAG := $(CIRCLE_BRANCH)
endif

info:
	@echo VERSION: $(VERSION)
	@echo GIT_SUMMARY: $(GIT_SUMMARY)
	@echo GIT_BRANCH: $(GIT_BRANCH)
	@echo UW_DOCKER_REGISTRY: $(UW_DOCKER_REGISTRY)
	@echo NAMESPACE: $(NAMESPACE)
	@echo SERVICE: $(SERVICE)
	@echo DOCKER_TAG: $(DOCKER_TAG)
	@echo DOCKER_IMAGE_NAME: $(DOCKER_IMAGE_NAME)
	@echo LDFLAGS: $(LDFLAGS)

# -
# Local Setup Tasks
# -

install:
	go get -u github.com/golangci/golangci-lint/cmd/golangci-lint

# -
# Application Tasks
# -

install-packages:
	go get ./...

lint:
	golangci-lint run

test: lint

fast:
	go build $(LDFLAGS)

static:
	CGO_ENABLED=0 GOOS=linux go build -a $(LDFLAGS) .


# -
# Docker Tasks
# -

docker-build:
	docker build \
		--build-arg GIT_SUMMARY=$(GIT_SUMMARY) \
		--build-arg GIT_BRANCH=$(GIT_BRANCH) \
		--build-arg SERVICE=$(SERVICE) \
		-t $(DOCKER_IMAGE_NAME):$(CIRCLE_SHA1) \
		.

docker-tag:
	docker tag $(DOCKER_IMAGE_NAME):$(CIRCLE_SHA1) $(DOCKER_IMAGE_NAME):$(DOCKER_TAG)

docker-push:
	docker push $(DOCKER_IMAGE_NAME)

docker-run:
	docker run --env-file=.env $(DOCKER_IMAGE_NAME):$(CIRCLE_SHA1)
