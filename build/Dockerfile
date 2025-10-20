# ---- Build stage ----
FROM golang:1.25.3-alpine AS builder
WORKDIR /app

# Cache deps
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build static binary for linux
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o smauth ./cmd/main.go

# ---- Runtime stage ----
# Distroless for Kubernetes (has certificates + minimal attack surface)
FROM gcr.io/distroless/static:nonroot

WORKDIR /app

# Copy binary
COPY --from=builder /app/smauth .
COPY --from=builder /app/internal/web/ui/static ./internal/web/ui/static
COPY --from=builder /app/internal/database/migrations ./internal/database/migrations

# Expose
EXPOSE 3001

# Run as non-root by default
USER nonroot:nonroot

# No shell - must use ENTRYPOINT
ENTRYPOINT ["/app/smauth"]
