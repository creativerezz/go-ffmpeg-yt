# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.22-bookworm AS builder
WORKDIR /app

# Leverage caching
COPY go.mod .
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/server

# Runtime stage
FROM debian:bookworm-slim AS runtime

# Install ffmpeg and yt-dlp
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
       ca-certificates \
       ffmpeg \
       python3 \
       python3-pip \
    && pip3 install --no-cache-dir yt-dlp \
    && rm -rf /var/lib/apt/lists/*

# Non-root user
RUN useradd -m -u 10001 appuser

WORKDIR /app
COPY --from=builder /app/server /app/server

ENV PORT=8080
EXPOSE 8080
USER appuser

ENTRYPOINT ["/app/server"]

