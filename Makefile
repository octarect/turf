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

.PHONY: e2e
e2e: build e2e_cases.tsv
	TURF_BIN=./bin/turf ./scripts/e2e_test.sh ./e2e_cases.tsv

.PHONY: e2e-clean
e2e-clean:
	rm ./e2e_cases.tsv

bin/turf: $(GO_SRCS)
	go build -o ./bin/turf -ldflags=$(BUILD_LDFLAGS) -trimpath ./cmd/turf

e2e_cases.tsv:
	TURF_BIN=./bin/turf ./scripts/gen_e2e_cases.sh ./e2e_cases.tsv

