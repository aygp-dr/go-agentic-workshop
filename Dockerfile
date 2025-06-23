# Multi-stage build for workshop environment
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make gcc musl-dev

# Set working directory
WORKDIR /workshop

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build all binaries
RUN make build

# Runtime image
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    curl \
    bash \
    jq \
    postgresql-client \
    aws-cli

# Create non-root user
RUN addgroup -g 1000 workshop && \
    adduser -D -u 1000 -G workshop workshop

# Copy binaries from builder
COPY --from=builder /workshop/bin /usr/local/bin

# Copy workshop materials
COPY --chown=workshop:workshop . /workshop

# Set working directory
WORKDIR /workshop

# Switch to non-root user
USER workshop

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD /usr/local/bin/agent --health || exit 1

# Default command
CMD ["/usr/local/bin/agent"]