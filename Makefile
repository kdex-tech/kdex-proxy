# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=proxy
MAIN_PATH=cmd/main.go

# Build flags
LDFLAGS=-ldflags "-s -w"

.PHONY: all build test clean run deps tidy

all: deps test build

build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PATH)

test:
	$(GOTEST) -v ./...

clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out

run: build
	./$(BINARY_NAME)

deps:
	$(GOMOD) download

tidy:
	$(GOMOD) tidy

coverage:
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

# Development targets
dev: export PORT=8080
dev: export SERVER=localhost:8081
dev: export MAPPED_HEADERS=Authorization,User-Agent,Content-Type
dev: build run

.DEFAULT_GOAL := all 