# Multi-stage build for optimized production image
FROM golang:1.22-alpine as builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY services/room-service ./services/room-service
COPY shared ./shared

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o room-service ./services/room-service/cmd/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /build/room-service .

# Non-root user for security
RUN addgroup -g 1000 appuser && adduser -D -u 1000 -G appuser appuser
USER appuser

EXPOSE 9095

CMD ["./room-service"]
