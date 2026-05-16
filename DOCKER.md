# Docker Build Guide

This guide explains how to build Stunning components using Docker, which resolves graphics library compilation issues on systems without X11/OpenGL development headers.

## Why Docker?

When running `go build ./...` locally, the build system attempts to compile all packages including graphics-dependent Fyne desktop and mobile apps. These require:

- X11 development libraries (libx11-dev, libxcursor-dev, etc.)
- OpenGL libraries (libgl1-mesa-dev)
- Platform-specific build tools

Docker provides a pre-configured environment with all dependencies installed, enabling complete builds on any system.

## Quick Start

### Build Everything

```bash
docker-compose run build-all
```

This builds:
- ✓ CLI tool (`build/stunning`)
- ✓ C shared library (`build/libstunning.so`)
- ✓ Desktop app (`build/stunning-desktop`)
- ✓ All unit tests with race detection

### Run Tests Only

```bash
docker-compose run test
```

Runs comprehensive tests:
- Core tunnel protocols (TCP, UDP, UDPS, TLS, H2, WS, DNS, ICMP)
- Bindings layer
- C library
- Race detection enabled

### Interactive Shell

```bash
docker-compose run shell
```

Opens an interactive bash shell inside the build container with all tools available:

```bash
# Inside container
cd /app
go test -v ./core/tunnel/...
go build -o /tmp/stunning .
```

## Services

### `build-all`

Complete build pipeline:
1. Compiles core library
2. Builds CLI tool (pure Go, no CGO)
3. Builds C shared library (with CGO)
4. Builds desktop app (requires Fyne)
5. Runs full test suite with race detection

**Usage:**
```bash
docker-compose run build-all
```

**Output:**
- `build/stunning` - CLI binary for Linux amd64
- `build/libstunning.so` - C shared library for Linux
- `build/libstunning.h` - Auto-generated C header
- `build/stunning-desktop` - Desktop app binary

### `test`

Runs all tests with race detection:

```bash
docker-compose run test
```

Tests included:
- `./core/tunnel/...` - All tunnel protocol implementations
- `./bindings/...` - Mobile bindings
- `./clib/...` - C library interface

### `shell`

Interactive development environment:

```bash
docker-compose run shell
```

Useful for:
- Debugging build issues
- Running custom go commands
- Exploring build artifacts
- Testing different configurations

## Build Details

### Dockerfile

The `Dockerfile` provides a Linux (Debian) build environment with:

- **Base:** golang:1.25-bookworm
- **Build tools:** gcc, g++, pkg-config, git
- **Graphics libraries:** libx11-dev, libxcursor-dev, libxrandr-dev, libxinerama-dev, libxi-dev, libxext-dev, libxfixes-dev, libgl1-mesa-dev, libxkbcommon-dev
- **Runtime:** ca-certificates for HTTPS

### Volume Mounts

- `.:/app` - Project source code
- `go-cache:/go/pkg/mod` - Go module cache (persistent across runs)

The module cache dramatically speeds up subsequent builds by avoiding re-downloading dependencies.

## Advanced Usage

### Build Specific Components

#### CLI Tool Only

```bash
docker-compose run build-all bash -c "CGO_ENABLED=0 go build -v -o ./build/stunning ."
```

#### C Library Only

```bash
docker-compose run build-all bash -c "CGO_ENABLED=1 go build -v -buildmode=c-shared -o ./build/libstunning.so ./clib/"
```

#### Custom Test Subset

```bash
docker-compose run test bash -c "go test -v ./core/tunnel/tcp ./core/tunnel/tls"
```

### Cross-Platform Builds

Modify `build-all` service environment variables in `docker-compose.yml`:

```yaml
environment:
  - GOOS=linux          # target OS
  - GOARCH=amd64        # target architecture
  - CGO_ENABLED=1
```

Supported combinations:
- `linux/amd64` - Linux 64-bit Intel
- `linux/arm64` - Linux 64-bit ARM
- `darwin/amd64` - macOS Intel (requires macOS host for some components)
- `darwin/arm64` - macOS Apple Silicon

## Troubleshooting

### Docker Not Available

If Docker isn't installed:

```bash
# macOS with Homebrew
brew install docker

# Or use Docker Desktop
# https://www.docker.com/products/docker-desktop
```

### Module Cache Issues

Clear the Go module cache:

```bash
docker volume rm <project>_go-cache
docker-compose run build-all
```

### Build Failures

Check the full build output:

```bash
docker-compose run build-all --no-ansi
```

Inspect the running container:

```bash
docker-compose run shell
# Then inside:
ls -la /app/build/
go env
```

## Local vs Docker Builds

| Component | Local | Docker | Notes |
|-----------|-------|--------|-------|
| CLI | ✓ | ✓ | Pure Go, no dependencies |
| C Library | ✓ | ✓ | Requires CGO + C compiler |
| Desktop App | ✗ | ✓ | Requires X11/OpenGL headers |
| Mobile Apps | ✗ | ✗ | Requires Android SDK/iOS Xcode |
| Tests | ✓ | ✓ | All tests work locally/Docker |

**Local builds work for:**
- `go build ./core/...` - Core library
- `go build ./bindings/...` - Bindings
- `go test ./...` - Tests
- `go build .` - CLI tool

**Docker required for:**
- Desktop app with Fyne
- Complete `go build ./...` (all packages)

## Performance Tips

1. **Reuse containers:** Docker-compose caches images and volumes
2. **Parallel builds:** Use `docker-compose run build-all` for full pipeline
3. **Module cache:** Persists in `go-cache` volume, dramatically speeds up repeated builds
4. **Incremental compilation:** Go's build cache works across container runs

First build: ~60s (downloading images and dependencies)
Subsequent builds: ~15s (with cached dependencies)

## CI/CD Integration

The Docker setup mirrors the `.github/workflows/release.yml` CI/CD pipeline:

```bash
# Local docker build
docker-compose run build-all

# Same steps run in GitHub Actions
# (see .github/workflows/release.yml)
```

Both produce identical artifacts through the same build steps.
