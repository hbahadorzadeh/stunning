<p align="center">
  <img src="./stunning.png" width="160" alt="Stunning Logo"/>
</p>

<h1 align="center">Stunning - Network Tunneling Engine</h1>

<p align="center">
  <strong>Production-ready tunneling with 10 protocols, 4 interfaces, and real-time monitoring</strong>
</p>

<p align="center">
  <a href="https://github.com/hbahadorzadeh/stunning/releases/tag/v1.0.0-beta"><img alt="Release" src="https://img.shields.io/badge/release-v1.0.0--beta-blue.svg"></a>
  <a href="https://github.com/hbahadorzadeh/stunning/actions"><img alt="Build Status" src="https://img.shields.io/badge/build-passing-brightgreen.svg"></a>
  <a href="#license"><img alt="License" src="https://img.shields.io/badge/license-MIT-green.svg"></a>
  <a href="https://golang.org"><img alt="Go" src="https://img.shields.io/badge/go-1.25-blue.svg"></a>
</p>

<p align="center">
  <a href="#features"><strong>Features</strong></a> •
  <a href="#installation"><strong>Installation</strong></a> •
  <a href="#quick-start"><strong>Quick Start</strong></a> •
  <a href="#cli-usage"><strong>CLI Usage</strong></a> •
  <a href="#api-usage"><strong>API Usage</strong></a> •
  <a href="#monitoring"><strong>Monitoring</strong></a> •
  <a href="#platforms"><strong>Platforms</strong></a>
</p>

---

## 🚀 Latest Release

**v1.0.0-beta** is now available! [Download](https://github.com/hbahadorzadeh/stunning/releases/tag/v1.0.0-beta) pre-built binaries for Linux and macOS, or the C library for embedding.

- ✅ 20 unit tests passing
- ✅ Race detection enabled
- ✅ Security scanning verified
- ✅ Full CI/CD pipeline

---

## Overview

**Stunning** is a modern tunneling engine that securely forwards network traffic through various protocols and interfaces. It's a production-ready replacement for stunnel, providing flexible multi-protocol support with real-time metrics, process-based management, and comprehensive monitoring.

### Use Cases

- 🔒 Secure legacy service access with TLS/HTTPS
- 📊 Load balancing and traffic routing
- 🌐 Protocol translation and bridging
- 🔐 VPN alternatives with custom protocols  
- 📡 Private network tunneling
- 🚀 Microservice gateway and mesh integration

---

## Features

### 🔄 Tunnel Protocols (10 Types)

| Protocol | Type | Use Case |
|----------|------|----------|
| **TCP** | Standard | Direct TCP forwarding |
| **UDP** | Datagram | DNS, VoIP, gaming |
| **UDPS** | Secure | Encrypted UDP tunneling |
| **TLS** | Encrypted | Secure socket layer tunneling |
| **HTTP** | Web | HTTP transparent proxy |
| **HTTPS** | Web | HTTPS/SSL proxy |
| **H2** | Modern | HTTP/2 multiplexed tunneling |
| **WS** | WebSocket | WebSocket tunneling (HTTP upgrade) |
| **DNS** | Query | DNS over custom protocol |
| **ICMP** | Echo | Stealth tunneling via ICMP |

### 🎯 Interface Types (4 Modes)

- **TCP Socket** — Standard network socket
- **SOCKS5 Proxy** — SOCKS5 protocol support
- **TUN Device** — Virtual network interface for VPN
- **Serial** — Serial port communication

### 🛠️ Management & Monitoring

- **Process-based** — Each tunnel runs as independent background process
- **Prometheus Metrics** — Real-time metrics in standard format
- **HTTP API** — JSON metrics and health endpoints
- **Uptime Tracking** — Connection count, bytes transferred, error rates
- **Auto-restart** — Automatic tunnel recovery on failure

### 📦 Multiple Distributions

- **CLI Tool** — Command-line tunnel manager
- **Desktop App** — Cross-platform GUI (Linux, macOS)
- **Mobile Apps** — iOS and Android VPN clients
- **C Library** — Embed tunneling in other applications
- **Go Library** — Use as Go package

---

## Installation

### From Releases

Download pre-built binaries:

```bash
# Linux
wget https://github.com/hbahadorzadeh/stunning/releases/download/v1.0.0-beta/stunning-linux-amd64
chmod +x ./stunning-linux-amd64
./stunning-linux-amd64 help

# macOS (Intel)
wget https://github.com/hbahadorzadeh/stunning/releases/download/v1.0.0-beta/stunning-darwin-amd64
chmod +x ./stunning-darwin-amd64
./stunning-darwin-amd64 help

# macOS (Apple Silicon)
wget https://github.com/hbahadorzadeh/stunning/releases/download/v1.0.0-beta/stunning-darwin-arm64
chmod +x ./stunning-darwin-arm64
./stunning-darwin-arm64 help
```

### C Library

Download the C library for embedding in C/C++ projects:

```bash
# Shared library (Linux)
wget https://github.com/hbahadorzadeh/stunning/releases/download/v1.0.0-beta/libstunning.so
wget https://github.com/hbahadorzadeh/stunning/releases/download/v1.0.0-beta/libstunning.h

# Static archive (all platforms)
wget https://github.com/hbahadorzadeh/stunning/releases/download/v1.0.0-beta/libstunning.a

# Compile with C library
gcc -o myapp myapp.c -L. -lstunning
```

### From Source

```bash
git clone https://github.com/hbahadorzadeh/stunning.git
cd stunning
go build -o ./stunning .
./stunning help
```

### Desktop App

```bash
# Install Fyne first
go install fyne.io/fyne/v2/cmd/fyne@latest

# Linux
go build -o ./Stunning ./app/desktop/

# macOS
fyne package -os darwin -appID com.stunning.tunnel \
  -name Stunning -icon app/desktop/assets/icon.png \
  -sourceDir ./app/desktop
```

### As Go Library

```bash
go get github.com/hbahadorzadeh/stunning
```

---

## Quick Start

### 1. Create Configuration

Create `tunnels.json`:

```json
{
  "secure-web": {
    "ServiceMode": "server",
    "ServerType": "tls",
    "InterfaceType": "tcp",
    "Listen": "127.0.0.1:443",
    "Connect": "example.com:443",
    "Cert": "/path/to/cert.pem",
    "Key": "/path/to/key.pem"
  },
  
  "http-proxy": {
    "ServiceMode": "server",
    "ServerType": "http",
    "InterfaceType": "socks",
    "Listen": "127.0.0.1:8080",
    "Connect": "upstream-proxy.local:3128"
  }
}
```

### 2. Start Tunnel

```bash
# Start in background
./stunning start secure-web

# Or run in foreground (debugging)
./stunning fg secure-web
```

### 3. Check Status

```bash
./stunning status
```

Output:
```
╔════════════════════════════════════════════════════════════════════════════════╗
║                         Tunnel Status                                          ║
╠════════════════════════════════════════════════════════════════════════════════╣
║ Name                 │ Status     │ Listen                    │ PID             │
╠════════════════════════════════════════════════════════════════════════════════╣
║ secure-web           │ ✓ Running  │ 127.0.0.1:443             │ 12345           │
║ http-proxy           │ ✗ Stopped  │ 127.0.0.1:8080            │ -               │
╚════════════════════════════════════════════════════════════════════════════════╝

Metrics available at: http://localhost:9090/metrics
```

---

## CLI Usage

### Commands

```bash
# Start tunnel in background
./stunning start <name>

# Run tunnel in foreground (for debugging/testing)
./stunning fg <name>

# Stop a running tunnel
./stunning stop <name>

# Show status of all tunnels
./stunning status

# List all configured tunnels
./stunning list

# View Prometheus metrics
./stunning metrics

# Show help
./stunning help
```

### Options

```bash
-config <file>       Config file (default: tunnels.json)
-metrics-port <port> Metrics HTTP port (default: 9090)
```

### Examples

```bash
# Start tunnel from custom config
./stunning -config tunnels-prod.json start my-tunnel

# Start on different metrics port
./stunning -metrics-port 9091 start my-tunnel

# Stop tunnel
./stunning stop my-tunnel

# View metrics directly
curl http://localhost:9090/metrics
```

---

## API Usage (Go Library)

### Create Tunnel Programmatically

```go
package main

import (
	"github.com/hbahadorzadeh/stunning/core"
)

func main() {
	config := core.TunnelConfig{
		ServiceMode:   "server",
		ServerType:    "tcp",
		InterfaceType: "tcp",
		Listen:        "127.0.0.1:8080",
		Connect:       "127.0.0.1:9090",
	}

	// Create tunnel
	tunnel := core.TunnelFactory("my-tunnel", config)

	// Start tunnel (blocking)
	go tunnel.ListenAndServer()

	// Check if alive
	if tunnel.IsAlive() {
		println("Tunnel is running")
	}

	// Access metrics
	metrics := tunnel.GetMetrics()
	println("Bytes sent:", metrics.BytesSent.Load())
	println("Bytes received:", metrics.BytesReceived.Load())
}
```

### Get Metrics

```go
// Export Prometheus format
prometheus := tunnel.GetMetrics().Export("my-tunnel")
println(prometheus)

// Export JSON format
json := tunnel.GetMetrics().ExportJSON("my-tunnel")
println(json)
```

---

## Monitoring

### Prometheus Metrics

Metrics are automatically exported at `http://localhost:9090/metrics`:

```prometheus
tunnel_uptime_seconds{tunnel="my-tunnel"} 3600
tunnel_bytes_received_total{tunnel="my-tunnel"} 1048576
tunnel_bytes_sent_total{tunnel="my-tunnel"} 2097152
tunnel_connections_total{tunnel="my-tunnel"} 125
tunnel_connections_current{tunnel="my-tunnel"} 3
tunnel_errors_total{tunnel="my-tunnel"} 2
```

### JSON API

Get metrics as JSON:

```bash
curl http://localhost:9090/api/metrics

curl http://localhost:9090/api/metrics/my-tunnel
```

### Health Check

```bash
curl http://localhost:9090/health
```

### Prometheus Scraping

Add to Prometheus `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'stunning-tunnels'
    static_configs:
      - targets: ['localhost:9090']
```

---

## Platforms

<p align="center">

| Platform | CLI | Desktop | Mobile | Library |
|----------|:---:|:-------:|:------:|:-------:|
| **Linux** | ✓ | ✓ | - | ✓ |
| **macOS** | ✓ | ✓ | - | ✓ |
| **Windows** | ✓ | - | - | ✓ |
| **iOS** | - | - | ✓ | ✓ |
| **Android** | - | - | ✓ | ✓ |

</p>

### Architecture Support

- Linux: x86_64, ARM64
- macOS: Intel, Apple Silicon (M1/M2/M3)
- Windows: x86_64
- iOS: ARM64
- Android: ARM64

---

## Configuration Guide

### TLS/HTTPS Tunnel

For TLS or HTTPS protocols, provide certificate and key:

```json
{
  "my-tls": {
    "ServiceMode": "server",
    "ServerType": "tls",
    "InterfaceType": "tcp",
    "Listen": "0.0.0.0:443",
    "Connect": "backend-server:8080",
    "Cert": "/etc/certs/cert.pem",
    "Key": "/etc/certs/key.pem"
  }
}
```

### TUN Device Interface

For VPN-like functionality, use TUN interface:

```json
{
  "vpn": {
    "ServiceMode": "server",
    "ServerType": "tcp",
    "InterfaceType": "tun",
    "Listen": "10.0.0.1",
    "Connect": "vpn-gateway.local",
    "DeviceName": "tun0",
    "Mtu": "1500"
  }
}
```

### SOCKS Proxy Interface

For SOCKS5 proxy:

```json
{
  "socks-proxy": {
    "ServiceMode": "server",
    "ServerType": "tcp",
    "InterfaceType": "socks",
    "Listen": "127.0.0.1:1080",
    "Connect": "upstream-proxy.local:3128"
  }
}
```

---

## Docker Build

Build for all platforms using Docker:

```bash
# Build everything
docker-compose run build-all

# Run tests
docker-compose run test

# Interactive shell
docker-compose run shell
```

See [DOCKER.md](DOCKER.md) for detailed Docker guide.

---

## Project Structure

```
.
├── core/                    # Core tunneling library
│   ├── tunnel/             # 10 tunnel protocol implementations
│   ├── interface/          # 4 interface implementations
│   ├── common/             # Shared utilities
│   └── metrics/            # Prometheus metrics system
├── app/
│   ├── desktop/            # Fyne desktop app
│   └── mobile/             # iOS/Android mobile apps
├── bindings/               # Mobile language bindings (gomobile)
├── clib/                   # C shared library wrapper
├── main.go                 # CLI tool
└── README.md              # This file
```

---

## License

MIT License - See [LICENSE](LICENSE) file

---

## Contributing

Contributions welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `go test -race ./...`
5. Submit a pull request

---

## Support

- 📧 Email: h.bahadorzadeh@gmail.com
- 🐛 Issues: [GitHub Issues](https://github.com/hbahadorzadeh/stunning/issues)
- 📖 Docs: Check project README and inline code comments

---

<p align="center">
  <strong>Made with ❤️ in Go</strong>
</p>
