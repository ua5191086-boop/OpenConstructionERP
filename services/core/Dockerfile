FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY services/core/go.mod services/core/go.sum ./
RUN go mod download
COPY services/core/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o /oce-core ./cmd/api

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /oce-core /usr/local/bin/
EXPOSE 8081
CMD ["oce-core"]
