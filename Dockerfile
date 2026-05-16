# Dockerfile for building Stunning components
# Handles CLI, C library, desktop app, and tests with all required dependencies
FROM golang:1.25-bookworm

# Install all build dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    ca-certificates \
    g++ \
    gcc \
    git \
    libgl1-mesa-dev \
    libx11-dev \
    libxcursor-dev \
    libxext-dev \
    libxfixes-dev \
    libxi-dev \
    libxinerama-dev \
    libxkbcommon-dev \
    libxrandr-dev \
    libxvmc-dev \
    pkg-config \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy source code
COPY . .

# Download Go dependencies
RUN go mod download

# Default command runs bash
CMD ["bash"]
