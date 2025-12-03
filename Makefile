# Make targets for OpsOrch Core
# Uses local caches to work in sandboxed environments and keeps Go version configurable.

GO ?= go
GOCACHE ?= $(PWD)/.gocache
GOMODCACHE ?= $(PWD)/.gocache/mod
CACHE_ENV = GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE)
IMAGE ?= opsorch-core:latest
PLUGINS ?= incidentmock logmock secretmock
MOCKS_IMAGE ?= opsorch-core-mocks:latest

.PHONY: all fmt test tidy clean run build docker-build docker-build-mocks

all: test

fmt:
	$(GO)fmt -w .

# Runs unit tests with a local build cache to avoid sandbox permission issues.
test:
	$(CACHE_ENV) $(GO) test ./...

tidy:
	$(GO) mod tidy

build:
	$(CACHE_ENV) $(GO) build ./...

docker-build:
	docker build -t $(IMAGE) --build-arg PLUGINS="" .

docker-build-mocks:
	docker build -t $(MOCKS_IMAGE) --build-arg PLUGINS="$(PLUGINS)" .

run:
	$(CACHE_ENV) $(GO) run ./cmd/opsorch

clean:
	rm -rf $(GOCACHE)
