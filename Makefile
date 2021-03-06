# -
# Variables
# -

-include .env
VERSION := $(shell git rev-parse HEAD)
GIT_SUMMARY := $(shell git describe --tags --dirty --always)
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)

BINARY_NAME := cockroach-certs
DOCKER_IMAGE_NAME := registry.uw.systems/cockroach-cfssl-certs

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
	@echo BINARY_NAME: $(BINARY_NAME)
	@echo DOCKER_TAG: $(DOCKER_TAG)
	@echo DOCKER_IMAGE_NAME: $(DOCKER_IMAGE_NAME)
	@echo LDFLAGS: $(LDFLAGS)

# -
# Local Setup Tasks
# -

install-tools:
	go get -u github.com/golangci/golangci-lint/cmd/golangci-lint

# -
# Application Tasks
# -

tidy:
	go mod tidy

install:
	go get -d -v ./...

lint:
	golangci-lint run

test: lint

fast:
	go build $(LDFLAGS) -o=$(BINARY_NAME) .

static:
	CGO_ENABLED=0 GOOS=linux go build -a $(LDFLAGS) -o=$(BINARY_NAME) .


# -
# Docker Tasks
# -

docker-build:
	docker build \
		--build-arg GIT_SUMMARY=$(GIT_SUMMARY) \
		--build-arg GIT_BRANCH=$(GIT_BRANCH) \
		-t $(DOCKER_IMAGE_NAME):$(CIRCLE_SHA1) \
		.

docker-tag:
	docker tag $(DOCKER_IMAGE_NAME):$(CIRCLE_SHA1) $(DOCKER_IMAGE_NAME):$(DOCKER_TAG)

docker-push:
	docker push $(DOCKER_IMAGE_NAME)

docker-run:
	docker run --env-file=.env $(DOCKER_IMAGE_NAME):$(CIRCLE_SHA1)
