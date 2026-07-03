# syntax=docker/dockerfile:1
# check=error=true

ARG GO_VERSION=1.26

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION} AS builder

ARG TARGETARCH

WORKDIR /app

RUN --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=bind,source=go.mod,target=go.mod \
    --mount=type=cache,target=/go/pkg/mod/,sharing=locked \
    go mod download -x

RUN --mount=type=bind,target=. \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build -ldflags="-s -w" -trimpath -o /bin/turf ./cmd/turf

FROM gcr.io/distroless/static-debian13:nonroot

COPY --from=builder /bin/turf /bin

WORKDIR /app

ENTRYPOINT ["/bin/turf"]
