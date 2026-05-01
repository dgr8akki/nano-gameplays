BINARY := bin/nano-gameplays
PKG := ./cmd/nano-gameplays

.PHONY: build run tidy vet fmt

build:
	go build -o $(BINARY) $(PKG)

run: build
	$(BINARY)

tidy:
	go mod tidy

vet:
	go vet ./...

fmt:
	go fmt ./...
