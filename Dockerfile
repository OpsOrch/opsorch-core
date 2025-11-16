# syntax=docker/dockerfile:1
# Multi-stage build for the OpsOrch Core binary. Supports a slim base image (core only)
# and a full image that also bundles plugin binaries.

ARG GO_VERSION=1.22
# Space-separated list of plugin directories under ./plugins to build into the image.
ARG PLUGINS="incidentmock logmock secretmock"

FROM golang:${GO_VERSION} AS builder
ARG TARGETOS=linux
ARG TARGETARCH=amd64
WORKDIR /src

COPY go.mod ./
COPY . .

ENV CGO_ENABLED=0

# Build the main API binary.
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /out/opsorch ./cmd/opsorch

# Build plugin binaries. Kept in a separate stage so the base image can omit them.
FROM builder AS plugin-builder
ARG TARGETOS
ARG TARGETARCH
ARG PLUGINS
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} sh -c "\
      mkdir -p /out/plugins && \
      for plugin in ${PLUGINS}; do \
        go build -o /out/plugins/$plugin ./plugins/$plugin; \
      done \
    "

# Base runtime image: only the core binary.
FROM gcr.io/distroless/static-debian12 AS runtime-base
WORKDIR /opt/opsorch
COPY --from=builder /out/opsorch /usr/local/bin/opsorch
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/opsorch"]

# Full runtime image: core + bundled plugins. This stays the default final stage.
FROM runtime-base AS runtime
COPY --from=plugin-builder /out/plugins ./plugins
