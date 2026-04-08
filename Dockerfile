# Stage 1: Build the Go app
FROM golang:1.24-bookworm AS builder

WORKDIR /app

# Copy go.mod and go.sum and install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the code
COPY . .

# Build the application
# Use the native architecture of the builder (which will match the runtime stage)
RUN CGO_ENABLED=0 go build -o main .

# Stage 2: Runtime image
FROM debian:bookworm-slim

# Install Chromium and required libraries for headless operation
RUN apt-get update && apt-get install -y \
    chromium \
    ca-certificates \
    libnss3 \
    libatk1.0-0 \
    libatk-bridge2.0-0 \
    libcups2 \
    libdrm2 \
    libxkbcommon0 \
    libxcomposite1 \
    libxdamage1 \
    libxext6 \
    libxfixes3 \
    libxrandr2 \
    libgbm1 \
    libasound2 \
    fonts-liberation \
    --no-install-recommends \
    && rm -rf /var/lib/apt/lists/*

# Set environment variables for Rod to find Chromium
# We set both common names just in case
ENV ROD_BIN=/usr/bin/chromium
ENV ROD_BROWSER_BIN=/usr/bin/chromium

WORKDIR /app

# Copy the binary from the builder
COPY --from=builder /app/main .

# Expose the API port
EXPOSE 8080

# Start the server
CMD ["./main"]
