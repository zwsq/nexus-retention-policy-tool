FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o nexus-retention-policy ./cmd

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/nexus-retention-policy .

# Create directory for logs
RUN mkdir -p /app/logs

# Run as non-root user
RUN addgroup -g 1000 nexus && \
    adduser -D -u 1000 -G nexus nexus && \
    chown -R nexus:nexus /app

USER nexus

ENTRYPOINT ["./nexus-retention-policy"]
CMD ["-config", "/app/config/config.yaml"]
