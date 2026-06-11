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

### 🧩 Plugin Chains (Anti-DPI)

Composable, per-connection transforms that obfuscate, encrypt, compress, and
disguise tunnel traffic to defeat deep-packet-inspection firewalls. Compiled in
(no `.so`, no CGO) and combinable in any order.

| Plugin | Category | Purpose |
|--------|----------|---------|
| `flate` | size | DEFLATE compression |
| `aead` | security | ChaCha20-Poly1305 / AES-GCM authenticated encryption |
| `pad` | anti-DPI | random length padding |
| `probe-guard` | active-probe | keyed tag; silently drops censor probes |
| `tls-mimic` | mimicry | disguises the wire as TLS |
| `http-mimic` | mimicry | disguises the wire as HTTP/1.1 chunked |
| `jitter` | morphing | random per-frame timing delay |
| `bucket` | morphing | normalize frame sizes to a fixed quantum |

A high-entropy encrypted tunnel that a censor blocks passes cleanly once wrapped
in `tls-mimic`. See [Plugin Chains](#plugin-chains) below, the full
[plugin reference](docs/PLUGINS.md), and [benchmarks](docs/BENCHMARKS.md).

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

## Plugin Chains

Plugin chains transform tunnel payloads to evade deep-packet inspection (DPI) and
to add security or size optimization. Plugins are stateful per connection,
compiled into the binary (no `.so`, no CGO — works on every platform), and
combine in any order.

### Enabling a chain

Add a `Plugins` field to a tunnel config. **The same chain string must be set on
both the client and the server** — the client's encode is the exact inverse of
the server's decode.

```json
{
  "secure-tunnel": {
    "ServiceMode": "client",
    "ServerType": "tcp",
    "InterfaceType": "socks",
    "Listen": "127.0.0.1:1080",
    "Connect": "vps.example.com:8443",
    "Plugins": "flate,aead?key=0123456789abcdef0123456789abcdef,tls-mimic"
  }
}
```

### Chain grammar

```
name?k=v&k2=v2,name2,name3?k=v
```

Comma-separated plugins, each with optional `?`-prefixed `&`-joined params.
`Encode` runs left→right on the way out; the peer's `Decode` runs right→left.

### Available plugins

| Plugin | Category | Params | Purpose |
|--------|----------|--------|---------|
| `flate` | size | `level` (0–9) | DEFLATE compression |
| `aead` | security | `key` (hex, **required**), `algo` (`chacha`\|`aesgcm`) | authenticated encryption, random nonce |
| `pad` | anti-DPI | `min`, `max` | random length padding |
| `probe-guard` | active-probe | `key` (hex, **required**), `taglen` (8–32) | keyed tag; drops unauthenticated probes |
| `tls-mimic` | mimicry | — | disguise the wire as TLS |
| `http-mimic` | mimicry | — | disguise the wire as HTTP/1.1 chunked |
| `jitter` | morphing | `min`, `max` (durations) | random per-frame timing delay |
| `bucket` | morphing | `size` | pad frames to a fixed size quantum |

### Ordering rules

- **Compress before encrypt** — `flate` must come before `aead` (ciphertext does
  not compress).
- **Mimicry goes last** — `tls-mimic` / `http-mimic` provide the wire framing and
  must be the rightmost plugin, or the length prefix would precede the protocol
  header and break the disguise.

### Recommended chains

```text
# fast, strong, looks like TLS
flate,aead?key=<hex>,tls-mimic

# add size-fingerprint resistance, look like HTTP
flate,aead?key=<hex>,bucket?size=512,http-mimic

# minimal obfuscation without crypto
flate,pad?min=16&max=256
```

### Why it beats DPI

A passive entropy detector flags high-entropy unknown protocols — exactly what a
naive encrypted proxy looks like. Wrapping the chain in `tls-mimic` makes the wire
open with a convincing TLS handshake, so a protocol allowlist passes it while the
body stays encrypted. `probe-guard` drops active probes (no distinguishing
response), and `pad` / `bucket` defeat packet-size fingerprinting.

The [`test/dpi/`](test/dpi/) harness demonstrates this end to end against a
simulated GFW middlebox. See the [plugin reference](docs/PLUGINS.md) and
[benchmarks](docs/BENCHMARKS.md) for details.

### Gates: authentication & port knocking

Beyond the byte-transform chain, two **connection gates** control *who* may use a
tunnel. Each has its own config field.

**Authentication** (`Auth`) — a handshake that runs inside the chain framing (so
it is disguised too) and rejects unauthorized clients:

| Authenticator | Purpose |
|---------------|---------|
| `psk` | HMAC challenge-response with a shared key |
| `jwt` | verify a presented JWT (HS256 secret / RS256 public key); identity = `sub` |
| `mtls` | mutual-TLS client certificate; identity = cert Common Name |

```json
"Auth": "jwt?alg=HS256&secret=<hex>"
```

**Port knocking** (`Knock`) — authorizes a source IP *before* it may connect, so a
scanner sees nothing on the tunnel port:

| Knocker | Purpose |
|---------|---------|
| `spa` | single encrypted UDP packet (HMAC + timestamp + nonce, anti-replay) authorizes the IP for a TTL |

```json
"Knock": "spa?key=<hex>&port=62201&ttl=10s"
```

Gates compose with the chain — a tunnel can require a knock, look like TLS,
encrypt with `aead`, and authenticate clients by JWT all at once. OAuth/LDAP
authenticators are planned. Full reference: [docs/PLUGINS.md](docs/PLUGINS.md#gates).

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
│   ├── plugin/             # Anti-DPI plugin chain system
│   ├── common/             # Shared utilities
│   └── metrics/            # Prometheus metrics system
├── app/
│   ├── desktop/            # Fyne desktop app
│   └── mobile/             # iOS/Android mobile apps
├── bindings/               # Mobile language bindings (gomobile)
├── clib/                   # C shared library wrapper
├── test/dpi/               # 3-node docker DPI evasion harness
├── docs/                   # PLUGINS.md, BENCHMARKS.md, design specs
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
