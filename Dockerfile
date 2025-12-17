# ----------------------
# Stage 1: Build the app
# ----------------------
FROM golang:1.24-alpine3.23 AS builder

# Install build dependencies
RUN apk add --no-cache \
    gcc \
    musl-dev \
    sqlite-dev \
    make \
    dos2unix

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Copy and fix entrypoint line endings (for Windows compatibility)
COPY entrypoint.sh /entrypoint.sh
RUN dos2unix /entrypoint.sh && chmod +x /entrypoint.sh

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags="-s -w" -o bin/codex github.com/gary-norman/forum/cmd/server

# ----------------------
# Stage 2: Runtime image
# ----------------------
FROM alpine:3.23

# Install only runtime dependencies
RUN apk add --no-cache \
    sqlite-libs \
    ca-certificates \
    tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/codex /app/bin/codex

# Copy required files
COPY --from=builder /app/migrations /app/migrations
COPY --from=builder /app/assets /app/assets
COPY --from=builder /app/robots.txt /app/robots.txt
COPY --from=builder /app/default-images /app/default-images
COPY --from=builder /app/.env /app/.env

# Copy entrypoint (already fixed with dos2unix in builder stage)
COPY --from=builder /entrypoint.sh /entrypoint.sh

# Create required directories
RUN mkdir -p /var/lib/db-codex /app/db/userdata/images/{channel-images,user-images,post-images}

EXPOSE 8888

ENTRYPOINT ["/entrypoint.sh"]
CMD ["/app/bin/codex"]
