# =============================================================================
# Stage 1: Build the Go binary
# =============================================================================
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy dependency files first for layer caching
COPY services/core/go.mod services/core/go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY services/core/ .

# Build a statically-linked binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.version=0.1.0" \
    -o /oce-core ./cmd/api

# =============================================================================
# Stage 2: Minimal runtime image
# =============================================================================
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

# Create a non-root user
RUN adduser -D -h /home/oce oce

COPY --from=builder /oce-core /usr/local/bin/oce-core

USER oce
WORKDIR /home/oce

EXPOSE 8081

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8081/health || exit 1

ENTRYPOINT ["oce-core"]
