# Stage 1: Build the Go app
FROM golang:1.24-bullseye AS builder

WORKDIR /app

# Copy go.mod and go.sum and install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Stage 2: Runtime image
FROM debian:bullseye-slim

# Install Chromium and required libraries
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

# Set environment variable for Rod to find Chromium
ENV ROD_BIN=/usr/bin/chromium

WORKDIR /app

# Copy the binary from the builder
COPY --from=builder /app/main .

# Expose the API port
EXPOSE 8080

# Start the server
CMD ["./main"]
