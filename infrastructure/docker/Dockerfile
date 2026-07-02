# Stage 1: Build Go binary
FROM golang:1.22-alpine AS builder

WORKDIR /build
COPY services/core/go.mod services/core/go.sum ./
RUN go mod download

COPY services/core/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o oce-core.bin ./cmd/api/

# Stage 2: Runtime
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /build/oce-core.bin .
EXPOSE 8081
CMD ["./oce-core.bin"]
