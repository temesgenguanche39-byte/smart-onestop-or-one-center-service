# ==============================================================================
# Multi-Stage Production Dockerfile for Smart One-Stop Platform
# ==============================================================================

# STAGE 1: Compiler & Builder
FROM golang:alpine AS builder

WORKDIR /src

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Download modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source tree
COPY . .

# Compile optimized static binaries
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /bin/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /bin/worker ./cmd/worker

# STAGE 2: Minimal Secure Runtime
FROM alpine:3.20 AS runner

WORKDIR /app

# Add unprivileged system user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Install runtime certificates and timezone data
RUN apk add --no-cache ca-certificates tzdata curl

# Copy binaries from builder
COPY --from=builder /bin/api /app/api
COPY --from=builder /bin/worker /app/worker

# Copy static frontend assets and documentation
COPY --from=builder /src/web /app/web
COPY --from=builder /src/openapi.yaml /app/openapi.yaml

# Ownership & Security
RUN chown -R appuser:appgroup /app
USER appuser

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=5s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/healthz || exit 1

CMD ["/app/api"]
