# =============================================================================
# DOCKERFILE - Matrix MUD
# =============================================================================
# Multi-stage build for minimal, secure production image
# Build: docker build -t matrix-mud .
# Run:   docker run -p 2323:2323 -p 8080:8080 -v mud-data:/app/data matrix-mud
# =============================================================================

# --- Build Stage ---
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY *.go ./
COPY pkg/ ./pkg/

# Build static binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.Version=1.31.0" \
    -o matrix-mud .

# --- Runtime Stage ---
FROM alpine:3.19

# Security: Run as non-root user
RUN addgroup -g 1000 mud && \
    adduser -u 1000 -G mud -s /bin/sh -D mud

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/matrix-mud .

# Copy default data (will be overwritten by volume mount)
COPY data/ ./data/

# Set ownership
RUN chown -R mud:mud /app

# Switch to non-root user
USER mud

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -q --spider http://localhost:8080/health || exit 1

# Expose ports
# 2323 - Telnet MUD
# 8080 - Web client
# 9090 - Admin panel (localhost only by default)
EXPOSE 2323 8080 9090

# Environment defaults
ENV TELNET_PORT=2323 \
    WEB_PORT=8080 \
    ADMIN_PORT=9090 \
    ADMIN_BIND_ADDR=127.0.0.1:9090

# Run
ENTRYPOINT ["./matrix-mud"]
