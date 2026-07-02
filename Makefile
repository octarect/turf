BUILD_LDFLAGS="-s -w"

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

bin/turf:
	go build -o ./bin/turf -ldflags=$(BUILD_LDFLAGS) -trimpath ./cmd/turf
