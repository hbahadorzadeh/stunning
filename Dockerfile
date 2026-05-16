# Dockerfile for building Stunning components
# Handles CLI, C library, desktop app, and tests with all required dependencies
FROM golang:1.25-bookworm

# Install all build dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    gcc \
    g++ \
    pkg-config \
    libx11-dev \
    libxcursor-dev \
    libxrandr-dev \
    libxinerama-dev \
    libxi-dev \
    libxext-dev \
    libxfixes-dev \
    libgl1-mesa-dev \
    libxkbcommon-dev \
    git \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy source code
COPY . .

# Download Go dependencies
RUN go mod download

# Default command runs bash
CMD ["bash"]
