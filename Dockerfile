# syntax=docker/dockerfile:experimental

ARG GO_VERSION=1.25.5

# ========================================
# build stage: compile static go binary
# ========================================
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine AS build

WORKDIR /src

# copy deps
COPY go.mod go.sum ./

# download deps with buildkit cache mount
# GHA will cachce /go/pkg/mod across builds
RUN --mount=type=cache,target=/go/pkg/mod \
  go mod download

# copy src only
COPY cmd ./cmd
COPY internal ./internal

# build args for cross-platform compilation
ARG TARGETOS
ARG TARGETARCH

# build static binary with optimizations
# cache mount for go-build speeds up repeated builds
RUN --mount=type=cache,target=/root/.cache/go-build \
  CGO_ENABLED=0 \
  GOOS=${TARGETOS:-linux} \
  GOARCH=${TARGETARCH:-amd64} \
  go build \
  -trimpath \
  -ldflags="-s -w" \
  -o /out/ipecho-api \
  ./cmd/server

# ========================================
# runtime stage: minimal distroless image
# ========================================
FROM gcr.io/distroless/static-debian12:nonroot

# OCI labels for supply chain tracking
ARG VERSION=dev
ARG VCS_REF=unknown
ARG BUILD_DATE=unknown

LABEL org.opencontainers.image.title="ipecho-api" \
  org.opencontainers.image.description="Minimal IP echo API" \
  org.opencontainers.image.vendor="Code After Dark" \
  org.opencontainers.image.source="https://github.org/jacob-lineberry/ipecho-api" \
  org.opencontainers.image.version="${VERSION}" \
  org.opencontainers.image.revision="${VCS_REF}" \
  org.opencontainers.image.created="${BUILD_DATE}" \
  org.opencontainers.image.licenses="MIT"

# copy only the compiled binary from build stage
COPY --from=build /out/ipecho-api /ipecho-api

# document port (GCR will set the $PORT env var)
EXPOSE 8080

# run as non-root user (distroless default: 65532:65532)
USER nonroot:nonroot

# execute the binary
ENTRYPOINT ["/ipecho-api"]

