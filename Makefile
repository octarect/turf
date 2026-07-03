BUILD_LDFLAGS="-s -w"
GO_SRCS := $(shell find . -name '*.go' -not -path './vendor/*')

.PHONY: all
all: clean build

.PHONY: build
build: bin/turf

.PHONY: clean
clean:
	rm -rf ./bin
	go clean

.PHONY: test
test:
	go test -v ./...

bin/turf: $(GO_SRCS)
	go build -o ./bin/turf -ldflags=$(BUILD_LDFLAGS) -trimpath ./cmd/turf

