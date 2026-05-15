<p align="center">
  <img src="./stunning.png" width="128" alt="Stunning Logo"/>
</p>

# Stunning - Network Tunneling Library

[![CI](https://github.com/hbahadorzadeh/stunning/workflows/CI/badge.svg)](https://github.com/hbahadorzadeh/stunning/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/hbahadorzadeh/stunning)](https://goreportcard.com/report/github.com/hbahadorzadeh/stunning)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Downloads

[![Download Linux](https://img.shields.io/badge/Download-Linux-blue?logo=linux)](https://github.com/hbahadorzadeh/stunning/releases)
[![Download macOS](https://img.shields.io/badge/Download-macOS-black?logo=apple)](https://github.com/hbahadorzadeh/stunning/releases)
[![Download Android](https://img.shields.io/badge/Download-Android-3DDC84?logo=android)](https://github.com/hbahadorzadeh/stunning/releases)
[![Download iOS](https://img.shields.io/badge/Download-iOS-000000?logo=apple)](https://github.com/hbahadorzadeh/stunning/releases)

**Stunning** is a production-ready Go library for tunneling different types of network traffic. It's a modern alternative to stunnel, providing flexible tunneling capabilities with multiple protocols and interfaces.

### Applications

- **CLI Tool** - Command-line tunnel manager with JSON configuration
- **Desktop App** - Cross-platform GUI (Linux, macOS) with Fyne
- **Mobile Apps** - iOS and Android VPN clients with native integration
- **C Library** - Shared library (.so/.dylib) for embedding in other languages
- **Go Library** - Core library for programmatic use

## Features

### Supported Tunnel Protocols
- ✓ **TCP** - TCP connections
- ✓ **UDP** - UDP datagrams
- ✓ **TLS** - Secure TLS connections
- ✓ **HTTP** - HTTP tunneling
- ✓ **HTTPS** - Secure HTTPS tunneling

### Supported Interfaces
- ✓ **TCP Socket** - Standard TCP socket interface
- ✓ **SOCKS5** - SOCKS5 proxy protocol
- ✓ **TUN Device** - Virtual network interface (TUN)
- 🔄 **UDP Socket** - UDP socket interface (in progress)
- 🔄 **Serial** - Serial port communication (in progress)

## Installation

```bash
go get github.com/hbahadorzadeh/stunning
```

## Quick Start

### Basic TCP Tunnel

```go
package main

import (
    "github.com/hbahadorzadeh/stunning/tunnel/tcp"
    "github.com/hbahadorzadeh/stunning/interface/tcp"
    "log"
)

func main() {
    // Create TCP tunnel server
    server, err := tcptun.StartTcpServer("127.0.0.1:9000")
    if err != nil {
        log.Fatal(err)
    }
    defer server.Close()

    // Set interface server
    iface := tcpiface.GetTcpServer("127.0.0.1:8080")
    server.SetServer(iface)

    // Start listening for connections
    server.WaitingForConnection()
}
```

### Configuration File

Create a JSON configuration file to define tunnels:

```json
{
  "tunnel1": {
    "ServiceMode": "server",
    "ServerType": "tcp",
    "InterfaceType": "tcp",
    "Listen": "127.0.0.1:9000",
    "Connect": "127.0.0.1:8080"
  },
  "secure_tunnel": {
    "ServiceMode": "server",
    "ServerType": "https",
    "InterfaceType": "socks",
    "Listen": "127.0.0.1:443",
    "Cert": "/path/to/cert.pem",
    "Key": "/path/to/key.pem"
  }
}
```

Run with configuration:

```bash
./stunning --config=config.json
# or
./stunning -c config.json
```

## Architecture

### Visual Overview

#### Core Capabilities & Roadmap
![Architecture Overview](./docs/images/architecture-overview.png)

#### Network Architecture
![Network Architecture](./docs/images/network-architecture.png)

### Two-Layer Design

1. **Tunnel Layer** - Transport mechanism (TCP, UDP, TLS, HTTP, HTTPS)
2. **Interface Layer** - User-facing endpoint (TCP Socket, SOCKS5, TUN Device)

This modular architecture allows flexible combinations of any tunnel with any interface.

### Data Flow

```
Client → Interface Layer → Tunnel Layer → Remote Server
         (SOCKS5, TCP)   (TLS, HTTPS)   (Destination)
```

## Testing

### Run All Tests

```bash
# Run with race detection
go test -v -race -coverprofile=coverage.out ./...

# Generate coverage report
go tool cover -html=coverage.out
```

### Run Tests

```bash
# Run all tests
go test -v ./... -timeout=60s

# Run with race detection (recommended for finding data races)
go test -v ./... -race -timeout=60s

# Run specific component tests
go test -v ./bindings -timeout=10s          # Mobile bindings
go test -v ./clib -timeout=10s              # C library
go test -v ./app/desktop -timeout=10s       # Desktop app
go test -v ./app/desktop/ui -timeout=10s    # Desktop UI
go test -v ./core/tunnel/... -timeout=30s   # Tunnel protocols

# Run tests excluding ICMP (requires root/CAP_NET_RAW privileges)
go test -v ./... -skip TestStartIcmpServer -timeout=60s
```

**Test Suite**:
- ✓ **Unit tests**: bindings, clib, desktop/mobile components, tunnel protocols
- ✓ **Integration tests**: Socket, SOCKS, TUN interfaces with protocols  
- ✓ **Thread-safety**: Race detection via `-race` flag
- ✓ **Coverage**: All tunnel types (TCP, TLS, H2, HTTP/HTTPS, WS, DNS, UDP/UDPS, ICMP)

### Local Security Check

```bash
# SAST security scan
gosec ./...

# Dependency vulnerability check
nancy sleuth

# Code quality
golangci-lint run
```

## Development & Contributing

### Building from Source

```bash
# Clone the repository
git clone https://github.com/hbahadorzadeh/stunning.git
cd stunning

# Install dependencies
go mod download

# Build
go build -o stunning

# Run tests
go test -v -race ./...
```

### Code Quality

The project maintains high code quality standards:

- **Linting**: 50+ Go linters via golangci-lint
- **Security**: SAST scanning with gosec
- **Testing**: Comprehensive unit and integration tests with race detection
- **Coverage**: Codecov integration
- **Race Detection**: Go race detector enabled

Before submitting PRs:

```bash
# Format code
gofmt -s -w .

# Run linters
golangci-lint run --fix

# Run tests
go test -race ./...

# Security scan
gosec ./...
```

## CI/CD Pipeline

Automated GitHub Actions pipeline with:

- ✓ Code formatting and linting
- ✓ SAST security scanning (HIGH/CRITICAL severity blocks merge)
- ✓ Unit tests with race detection
- ✓ Integration tests across all tunnel protocols and interfaces
- ✓ Code quality metrics
- ✓ Dependency vulnerability scanning
- ✓ Automated dependency updates (Dependabot)

See [`.github/CI_CD.md`](.github/CI_CD.md) for detailed pipeline documentation.

## Security

### Security Policy

- TLS certificate verification is **enabled by default**
- For custom TLS configurations, use `GetTlsDialerWithConfig()`
- Vulnerability reporting: See [SECURITY.md](SECURITY.md)

### Secure by Default

- No insecure defaults
- SAST scanning catches common issues
- Race condition detection
- Dependency vulnerability scanning

## Project Status

### Completed
- ✓ TCP, UDP, TLS, HTTP, HTTPS tunnels
- ✓ TCP Socket, SOCKS5, TUN Device interfaces
- ✓ Go module support (go.mod/go.sum)
- ✓ Comprehensive testing suite (unit + integration + race detection)
- ✓ 65+ critical bugs fixed in 5-round review
- ✓ Modern GitHub Actions CI/CD pipeline
- ✓ Complete security policy
- ✓ Production-ready codebase

### In Progress / Planned
- 🔄 UDP Socket interface
- 🔄 Serial port interface
- 🔄 Terminal monitoring interface
- 🔄 Performance optimizations
- 🔄 Extended documentation

## Dependencies

Core dependencies:
- `github.com/getlantern/go-socks5` - SOCKS5 protocol
- `github.com/jacobsa/go-serial` - Serial communication
- `github.com/rainycape/dl` - Dynamic library loading
- `github.com/songgao/water` - TUN/TAP interface
- `github.com/yuin/gopher-lua` - Lua scripting
- `golang.org/x/net` - Extended network utilities

All dependencies are automatically scanned for vulnerabilities via nancy.

## Documentation

- **[SECURITY.md](SECURITY.md)** - Security policy and vulnerability reporting
- **[REVIEW_SUMMARY.md](REVIEW_SUMMARY.md)** - 5-round bug fix summary
- **[.github/CI_CD.md](.github/CI_CD.md)** - CI/CD pipeline documentation
- **[.github/README.md](.github/README.md)** - GitHub Actions setup guide

## Examples

Example applications are in the `example/` directory:

- `tcp_example.go` - TCP tunnel example
- `tun_example.go` - TUN device example
- `socks_example.go` - SOCKS5 proxy example

Run examples:

```bash
go run example/tcp_example.go
```

## License

MIT License - See LICENSE file for details

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Ensure tests pass: `go test -race ./...`
5. Run linters: `golangci-lint run`
6. Submit a pull request

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

## Code Quality Metrics

### Automated Quality Checks
- **Linting**: golangci-lint with 50+ rules
- **Security**: gosec SAST scanner
- **Dependencies**: nancy vulnerability scanner
- **Testing**: 100+ test cases
- **Coverage**: Codecov integration
- **Race Detection**: Go race detector

### Recent Improvements (5-Round Review)

#### Round 1: Module Setup & Compilation
- Created go.mod with all dependencies
- Fixed 9 compilation issues

#### Round 2: Value-Receiver Mutations
- Fixed systematic bug affecting 10+ files
- Ensured proper state management

#### Round 3: Concurrency Bugs
- Eliminated 4 critical deadlocks
- Fixed goroutine resource leaks
- Fixed shared connection races

#### Round 4: Interface Compliance
- Added missing interface methods
- Fixed logic bugs and divide-by-zero issues

#### Round 5: Security Hardening
- Enabled TLS verification by default
- Isolated HTTP mux per-server
- Fixed all resource leaks

Total: **65+ bugs fixed**, production-ready code

## Multi-Platform Releases

Stunning is built and released for multiple platforms automatically via GitHub Actions:

### Supported Platforms

- **Linux** - CLI, Desktop GUI, C Library
- **macOS** - CLI, Desktop GUI, C Library (Intel + Apple Silicon)
- **Android** - Mobile VPN app
- **iOS** - Mobile VPN app

### Building for Your Platform

```bash
# CLI tool (pure Go, no CGO required)
go build -o stunning .

# C Shared Library (Linux/macOS)
# Note: Requires CGO_ENABLED=1 and C compiler (gcc/clang)
CGO_ENABLED=1 go build -buildmode=c-shared -o libstunning.so ./clib/    # Linux
CGO_ENABLED=1 go build -buildmode=c-shared -o libstunning.dylib ./clib/ # macOS

# Desktop App (requires Fyne and CGO)
# Note: Fyne requires CGO_ENABLED=1 for native platform integration
go install fyne.io/fyne/v2/cmd/fyne@latest
CGO_ENABLED=1 go build -o stunning-desktop ./app/desktop/

# Mobile bindings (requires gomobile)
# Note: Android and iOS builds have platform-specific requirements
go install golang.org/x/mobile/cmd/gomobile@latest
gomobile init

# Android APK (requires Android SDK)
CGO_ENABLED=1 fyne package -os android -appID com.stunning.tunnel -name Stunning ./app/mobile/

# iOS IPA (requires Xcode on macOS)
CGO_ENABLED=1 fyne package -os ios -appID com.stunning.tunnel -name Stunning ./app/mobile/

# Generate bindings for native embedding
gomobile bind -target android -o libstunning.aar ./bindings/    # Android
gomobile bind -target ios -o Stunning.xcframework ./bindings/   # iOS
```

### Automated Release Pipeline

Triggered automatically when pushing a version tag to GitHub (e.g., `v1.0.0`), the CI/CD workflow in `.github/workflows/release.yml` handles:

1. **CLI Binaries** - Compiled for Linux/macOS (amd64, arm64)
2. **Libraries** - C shared libraries (.so/.dylib) with auto-generated headers
3. **Desktop Apps** - Linux binaries and macOS app bundles via Fyne
4. **Mobile Apps** - Android APK + AAR bindings, iOS IPA + xcframework
5. **GitHub Release** - Automatic release notes with all downloadable artifacts

**To Create a Release**:
```bash
git tag v1.0.0
git push origin v1.0.0
# GitHub Actions automatically builds all platforms and publishes release
```

**Artifacts Published** (one per platform):
- CLI: `stunning-linux-amd64`, `stunning-linux-arm64`, `stunning-darwin-amd64`, `stunning-darwin-arm64`
- Libraries: `libstunning.so` / `libstunning.h` (Linux), `libstunning.dylib` / `libstunning.h` (macOS)
- Desktop: `stunning-desktop` (Linux), `Stunning.app` (macOS)
- Mobile: `Stunning.apk` + `libstunning.aar` (Android), `Stunning.ipa` + `Stunning.xcframework` (iOS)

## Performance

- **Concurrent Connections**: Fully concurrent with race detection
- **Memory Safety**: Race detector enabled in CI
- **Resource Management**: Proper cleanup and timeout handling
- **Network Efficiency**: Optimized connection pooling

## Support

- 📧 Email: h.bahadorzadeh@gmail.com
- 🐛 Issues: [GitHub Issues](https://github.com/hbahadorzadeh/stunning/issues)
- 🔒 Security: See [SECURITY.md](SECURITY.md) for vulnerability reporting

## Acknowledgments

Thanks to all contributors and the Go community for tools and libraries that make this project possible.

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for detailed version history.

---

**Made with ❤️ in Go**