# Go parameters
BINARY_NAME=proxy
DOCKER_IMAGE=ghcr.io/kdex-tech/proxy
DOCKER_TAG=$(shell git describe --tags --dirty --always)
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GO_VERSION=$(shell go mod edit -json | jq -r .Go)
LDFLAGS=-ldflags "-s -w"
MAIN_PATH=cmd/main.go
LICENSE_HOLDER=KDex Tech
PLATFORMS?=linux/arm64,linux/amd64,linux/s390x,linux/ppc64le

.PHONY: all build test clean run debug deps tidy docker-build docker-build-debug docker-buildx docker-run docker-push install-addlicense license check-license

all: deps test build

build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PATH)

test:
	$(GOTEST) -v ./...

clean:
	@rm -f $(BINARY_NAME)
	@rm -f coverage.out

run: build
	./$(BINARY_NAME)

deps:
	$(GOMOD) download

tidy: license
	$(GOMOD) tidy

coverage:
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

# Development targets
dev: export LISTEN_PORT=8080
dev: export LISTEN_ADDRESS=""
dev: export MAPPED_HEADERS=Authorization,User-Agent,Content-Type
dev: build run

# Docker targets
docker-build:
	docker build --build-arg GO_VERSION=$(GO_VERSION) -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-build-debug:
	docker build --build-arg GO_VERSION=$(GO_VERSION) -t $(DOCKER_IMAGE):$(DOCKER_TAG)-DEBUG -f Dockerfile.debug .

docker-run: export LISTEN_PORT=8080
docker-run: export LISTEN_ADDRESS=""
docker-run: export MAPPED_HEADERS=Authorization,User-Agent,Content-Type
docker-run: docker-build
docker-run:
	docker run -p $(LISTEN_PORT):$(LISTEN_PORT) \
		-e LISTEN_PORT=$(LISTEN_PORT) \
		-e LISTEN_ADDRESS=$(LISTEN_ADDRESS) \
		-e UPSTREAM_ADDRESS=$(UPSTREAM_ADDRESS) \
		-e UPSTREAM_HEALTHZ_PATH=$(UPSTREAM_HEALTHZ_PATH) \
		-e MAPPED_HEADERS=$(MAPPED_HEADERS) \
		--add-host=$(UPSTREAM_ADDRESS):host-gateway \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

docker-push:
	@docker push $(DOCKER_IMAGE):$(DOCKER_TAG)

docker-buildx: ## Build and push docker image for the proxy for cross-platform support
	# copy existing Dockerfile and insert --platform=${BUILDPLATFORM} into Dockerfile.cross, and preserve the original Dockerfile
	@sed -e '4 s/\(^FROM\)/FROM --platform=\$$\{BUILDPLATFORM\}/; t' -e ' 4,// s//FROM --platform=\$$\{BUILDPLATFORM\}/' Dockerfile > Dockerfile.cross
	@echo "--- Dockerfile.cross ---"
	@cat Dockerfile.cross
	@echo "---"
	- @docker buildx create --name kdex-proxy-builder
	@docker buildx use kdex-proxy-builder
	- docker buildx build \
		--push \
		--platform=$(PLATFORMS) \
		--build-arg GO_VERSION=$(GO_VERSION) \
		--tag $(DOCKER_IMAGE):$(DOCKER_TAG) \
		--tag $(DOCKER_IMAGE):latest \
		--annotation 'manifest:org.opencontainers.image.source=https://github.com/kdex-tech/kdex-proxy,\
manifest:org.opencontainers.image.description="KDex Proxy",\
manifest:org.opencontainers.image.licenses=Apache-2.0' \
		-f Dockerfile.cross .
	@rm Dockerfile.cross

# Install addlicense tool if not present
install-addlicense:
	@which addlicense > /dev/null || go install github.com/google/addlicense@latest

# Add/update license headers
license: install-addlicense
	@echo "Updating license headers..."
	@addlicense -v \
		-f LICENSE.header \
		-c "$(LICENSE_HOLDER)" \
		Dockerfile* ./cmd/* ./internal/* ./k8s/*

# Check license headers
# in CI do: make check-license || exit 1
# to fail if license headers are not present
check-license: install-addlicense
	@echo "Checking license headers..."
	@addlicense -check \
		-f LICENSE.header \
		-c "$(LICENSE_HOLDER)" \
		Dockerfile* ./cmd/* ./internal/* ./k8s/*

.DEFAULT_GOAL := all 