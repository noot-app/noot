# Multi-stage Dockerfile for noot
# Produces a minimal from-scratch image with only the noot binary

# Build argument for Go version (can be overridden at build time)
ARG GO_VERSION=1.24.4

# Build stage - uses Go version from build arg (defaults to .go-version content)
FROM golang:${GO_VERSION}-bullseye AS builder

# Set working directory
WORKDIR /build

# Copy source code and vendor directory (full dependency vendoring)
COPY . .

# Prepare AI config files for embedding
RUN script/prepare-ai-embed

# Get build information for embedding into binary (similar to script/build)
RUN COMMIT_SHA="$(git rev-parse HEAD 2>/dev/null || echo 'unknown')" && \
    BUILD_TIME="$(date -u '+%Y-%m-%dT%H:%M:%SZ')" && \
    PROJECT_NAME="noot" && \
    MODULE_PATH="github.com/grantbirki/noot" && \
    \
    # Build the binary exactly as goreleaser would - using vendored dependencies
    CGO_ENABLED=0 \
    GOPROXY=off \
    GOSUMDB=off \
    go build \
        -mod=vendor \
        -trimpath \
        -v \
        -ldflags="-s -w -X ${MODULE_PATH}/internal/version.commit=${COMMIT_SHA} -X ${MODULE_PATH}/internal/version.buildTime=${BUILD_TIME}" \
        -o /build/${PROJECT_NAME} \
        ./cmd/${PROJECT_NAME}

# Runtime stage - minimal from-scratch image
FROM scratch

# Add ca-certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary from the builder stage
COPY --from=builder /build/noot /noot

# Set the binary as the entrypoint
ENTRYPOINT ["/noot"]
