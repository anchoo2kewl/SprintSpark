# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev sqlite-dev

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build static binary with CGO enabled for SQLite
RUN CGO_ENABLED=1 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -tags sqlite_omit_load_extension \
    -o sprintspark \
    ./cmd/api

# Production stage
FROM alpine:3.19

# Install CA certificates and create non-root user
RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -g 1001 -S sprintspark && \
    adduser -u 1001 -S sprintspark -G sprintspark

# Create data directory with correct permissions
RUN mkdir -p /data && chown sprintspark:sprintspark /data

WORKDIR /app

# Copy binary from builder
COPY --from=builder --chown=sprintspark:sprintspark /build/sprintspark .

# Copy OpenAPI spec (needed for /api/openapi endpoint)
COPY --chown=sprintspark:sprintspark openapi.yaml .

# Copy database migrations
COPY --chown=sprintspark:sprintspark internal/db/migrations ./internal/db/migrations

# Switch to non-root user
USER sprintspark

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/health || exit 1

# Run the binary
CMD ["./sprintspark"]
