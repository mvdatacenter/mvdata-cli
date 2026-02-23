BINARY := mvdata
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -ldflags "-X github.com/mvdatacenter/mvdata-cli/internal/cmd.version=$(VERSION)"

.PHONY: build test vet lint clean

build:
	go build $(LDFLAGS) -o $(BINARY) .

test:
	go test ./...

vet:
	go vet ./...

lint: vet
	@echo "lint passed"

clean:
	rm -f $(BINARY)
