# -
# Variables
# -

VERSION := $(shell git rev-parse HEAD)
GIT_SUMMARY := $(shell git describe --tags --dirty --always)
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)

CIRCLE_SHA1 ?= $(VERSION)
CIRCLE_BRANCH ?= $(GIT_BRANCH)

LDFLAGS := -ldflags "-X main.version=$(VERSION)"

ifeq ($(CIRCLE_BRANCH), master)
    DOCKER_TAG := latest
else
    DOCKER_TAG := $(CIRCLE_BRANCH)
endif

ifeq ($(COMMAND), request-certs)
    BUILD_PATH := ./cmd/request-certs
    DOCKER_IMAGE_NAME := registry.uw.systems/cockroach-cfssl-certs
else
    BUILD_PATH := ./cmd/health-checker
    DOCKER_IMAGE_NAME := registry.uw.systems/cockroach-health-checker
endif

check-command:
	@ if [ "${COMMAND}" = "" ]; then \
		echo "Environment variable COMMAND not set"; \
		exit 1; \
	fi
	@ if [ "${COMMAND}" != "request-certs" ] && [ "${COMMAND}" != "health-checker" ]; then \
		echo "${COMMAND} is not a valid value for COMMAND, allowed values are 'request-certs' and 'health-checker'"; \
		exit 1; \
	fi

info: check-command
	@echo COMMAND: $(COMMAND)
	@echo VERSION: $(VERSION)
	@echo GIT_SUMMARY: $(GIT_SUMMARY)
	@echo GIT_BRANCH: $(GIT_BRANCH)
	@echo BUILD_PATH: $(BUILD_PATH)
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

install-all-packages:
	go get -t -d -v ./...

install-packages: check-command
	go get -t -d -v $(BUILD_PATH)/...

lint:
	golangci-lint run

test: lint

fast: check-command
	go build $(LDFLAGS) $(BUILD_PATH)

static: check-command
	CGO_ENABLED=0 GOOS=linux go build -a $(LDFLAGS) $(BUILD_PATH)

# -
# Docker Tasks
# -

docker-build: check-command
	docker build \
		--build-arg GIT_SUMMARY=$(GIT_SUMMARY) \
		--build-arg GIT_BRANCH=$(GIT_BRANCH) \
		-t $(DOCKER_IMAGE_NAME):$(CIRCLE_SHA1) \
		. -f $(BUILD_PATH)/Dockerfile

docker-tag: check-command
	docker tag $(DOCKER_IMAGE_NAME):$(CIRCLE_SHA1) $(DOCKER_IMAGE_NAME):$(DOCKER_TAG)

docker-push: check-command
	docker push $(DOCKER_IMAGE_NAME)

docker-run: check-command
	docker run $(DOCKER_IMAGE_NAME):$(CIRCLE_SHA1)
