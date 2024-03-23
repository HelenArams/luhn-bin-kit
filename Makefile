GOBIN            := $(CURDIR)/.tools/bin
GOLANGCI_LINT    := $(GOBIN)/golangci-lint
GOLANGCI_VERSION := v1.56.2

export PATH := $(GOBIN):$(PATH)

.PHONY: all build test cover vet fmt fmt-check lint fuzz bench tidy tools install clean

all: fmt-check vet test

build:
	go build ./...

test:
	go test -race -cover ./...

cover:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

vet:
	go vet ./...

fmt:
	gofmt -s -w .

fmt-check:
	@out="$$(gofmt -s -l .)"; if [ -n "$$out" ]; then echo "gofmt needed:"; echo "$$out"; exit 1; fi

FUZZTIME ?= 30s

fuzz:
	go test -run='^$$' -fuzz=FuzzValid -fuzztime=$(FUZZTIME) ./luhn
	go test -run='^$$' -fuzz=FuzzMask -fuzztime=$(FUZZTIME) ./mask

bench:
	go test -run='^$$' -bench=. -benchmem ./luhn ./bin

lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run ./...

tools: $(GOLANGCI_LINT)

$(GOLANGCI_LINT):
	GOBIN=$(GOBIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_VERSION)

tidy:
	go mod tidy

install:
	go install ./cmd/cardkit

clean:
	rm -rf cardkit dist coverage.out
