# Stunning - Network Tunneling Library

[![CI](https://github.com/hbahadorzadeh/stunning/workflows/CI/badge.svg)](https://github.com/hbahadorzadeh/stunning/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/hbahadorzadeh/stunning)](https://goreportcard.com/report/github.com/hbahadorzadeh/stunning)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Stunning** is a production-ready Go library for tunneling different types of network traffic. It's a modern alternative to stunnel, providing flexible tunneling capabilities with multiple protocols and interfaces.

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

### Run Specific Tests

```bash
# Unit tests only
go test -v ./... -skip E2E

# Functional tests
go test -v -timeout=30s -run 'TestTunnel|TestInterface|TestRecovery' ./...

# E2E tests
go test -v -timeout=60s -run 'TestE2E' ./...
```

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
- **Testing**: Unit, functional, and E2E tests
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
- ✓ Functional and E2E tests
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
- ✓ Comprehensive testing suite (unit + functional + E2E)
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