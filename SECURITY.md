# Security Policy

## Supported versions

Security fixes are applied to the latest minor release. Older lines are not
maintained.

| Version | Supported |
| ------- | --------- |
| 1.1.x   | ✅ |
| 1.0.x   | ⚠️ critical fixes only |
| < 1.0   | ❌ |

## Reporting a vulnerability

**Do not open public issues for security vulnerabilities.** Report privately by
emailing [h.bahadorzadeh@gmail.com](mailto:h.bahadorzadeh@gmail.com) with:

1. **Description** — what the vulnerability is
2. **Location** — affected file(s) and line numbers
3. **Impact** — what an attacker can achieve
4. **Reproduction** — steps or a minimal proof of concept
5. **Suggested fix** — if you have one

### Disclosure timeline

| Stage | Target |
| ----- | ------ |
| Acknowledge receipt | 24–48 hours |
| Initial assessment | 3–5 business days |
| Fix developed & tested | varies by severity |
| Patch released | as a patch version (e.g. 1.1.1) |
| Public disclosure | after the fix ships |

## Security model

Stunning provides transport security and access control directly, through three
composable mechanisms. See [docs/PLUGINS.md](docs/PLUGINS.md) for full details.

### Confidentiality & integrity — the `aead` plugin

Use the `aead` plugin (ChaCha20-Poly1305 or AES-GCM, random nonce per frame) for
authenticated encryption. **Without `aead` the plugin chain is obfuscation only**
— a network attacker can read, modify, or replay traffic. The recommended baseline
chain is `…,aead?key=…,…`.

### Access control — gates

- **Authentication (`Auth`)** — `psk`, `jwt`, `mtls`, `oauth`, `ldap`. Runs a
  handshake inside the (disguised) chain framing and rejects unauthorized clients.
- **Port knocking (`Knock`)** — `spa` single-packet authorization; an
  un-knocked source IP sees nothing on the tunnel port.

### TLS configuration

- Certificate verification is **enabled by default** (`InsecureSkipVerify: false`).
- Custom TLS configs for development/testing are available via
  `GetTlsDialerWithConfig()`.
- `mtls insecure=true` disables server verification — **testing only**; it logs a
  warning at startup.
- Never disable certificate verification in production.

### Operational guidance

- **Treat config files as secrets.** Chain/auth/knock specs embed key material
  (`key=`, `secret=`, `password=`). Keep them out of logs and shared repos.
- **Bearer tokens are replayable until expiry.** Use short `jwt`/`oauth`
  lifetimes; prefer `mtls` where possible.
- **Run credential-carrying auth behind `aead`** (`psk`/`jwt`/`oauth`/`ldap`
  exchange secrets); never run `ldap` over an unencrypted chain (use `ldaps`).
- Isolate tunnel endpoints with firewall rules; expose only what is needed.
- Monitor connection counts and set timeouts to limit resource exhaustion.
- Run with least privilege.

## CI security tooling

Every push and pull request runs (see [.github/CI_CD.md](.github/CI_CD.md)):

| Tool | Purpose |
| ---- | ------- |
| **gosec** | Go SAST; fails on HIGH/CRITICAL findings |
| **nancy** | vulnerable-dependency scanner |
| **staticcheck** | advanced static analysis |
| **race detector** | data-race detection (`go test -race`) |

## Dependencies

Direct dependencies, vetted and kept current by Dependabot (weekly):

- `github.com/getlantern/go-socks5` — SOCKS5 proxy interface
- `github.com/go-ldap/ldap/v3` — LDAP authenticator
- `github.com/gorilla/websocket` — WebSocket tunnel
- `github.com/jacobsa/go-serial` — serial interface
- `github.com/pion/dtls/v3` — secure UDP (UDPS) tunnel
- `github.com/songgao/water` — TUN device interface
- `golang.org/x/crypto` — AEAD, key derivation, primitives
- `golang.org/x/net` — extended networking (HTTP/2, etc.)

## Questions

For non-vulnerability security questions, email
[h.bahadorzadeh@gmail.com](mailto:h.bahadorzadeh@gmail.com).
