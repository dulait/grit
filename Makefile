APP_NAME := grit
VERSION := $(shell git describe --tags --always --dirty)
COMMIT := $(shell git rev-parse --short HEAD)
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -ldflags "-X github.com/dulait/grit/internal/cli.Version=$(VERSION) -X github.com/dulait/grit/internal/cli.CommitSHA=$(COMMIT) -X github.com/dulait/grit/internal/cli.BuildDate=$(BUILD_DATE)"

.PHONY: build install clean test fmt vet

build:
	go build $(LDFLAGS) -o bin/$(APP_NAME) ./cmd/grit

install:
	go install $(LDFLAGS) ./cmd/grit

clean:
	rm -rf bin/

test:
	go test ./...

fmt:
	go fmt ./...

vet:
	go vet ./...
