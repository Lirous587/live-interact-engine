# Multi-stage build for optimized production image
FROM golang:1.22-alpine as builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY services/danmaku-service ./services/danmaku-service
COPY shared ./shared

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o danmaku-service ./services/danmaku-service/cmd/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /build/danmaku-service .

# Non-root user for security
RUN addgroup -g 1000 appuser && adduser -D -u 1000 -G appuser appuser
USER appuser

EXPOSE 9093

CMD ["./danmaku-service"]
