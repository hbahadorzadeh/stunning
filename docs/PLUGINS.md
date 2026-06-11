# Plugin Chains

Plugin chains transform tunnel traffic to evade deep-packet inspection (DPI) and
to add security or size optimization. Plugins are composable and combine in any
order inside a chain. They are compiled into the binary (no `.so`, no CGO) and
work on every platform.

## Configuring a chain

Add a `Plugins` field to a tunnel config (the same string on client and server):

```json
{
  "client": {
    "ServiceMode": "client", "ServerType": "tcp", "InterfaceType": "tcp",
    "Listen": ":1080", "Connect": "server:8443",
    "Plugins": "flate,aead?key=0123…,tls-mimic"
  }
}
```

Spec grammar: `name?k=v&k2=v2,name2,…` — comma-separated plugins, each with
optional `?`-prefixed params. `Encode` runs left→right; the peer's `Decode` runs
right→left, so the same string on both ends is a mutual inverse.

## Built-in plugins

| Plugin | Category | Params | Purpose |
|--------|----------|--------|---------|
| `flate` | size | `level` (0–9) | DEFLATE compression |
| `aead` | security | `key` (hex, required), `algo` (`chacha`\|`aesgcm`) | authenticated encryption, random-nonce |
| `pad` | anti-DPI | `min`, `max` | random length padding |
| `probe-guard` | active-probe | `key` (hex, required), `taglen` (8–32) | keyed tag; drops unauthenticated probes |
| `tls-mimic` | mimicry | — | disguises the wire as TLS (synthetic handshake + application_data records) |
| `http-mimic` | mimicry | — | disguises the wire as HTTP/1.1 chunked transfer |
| `jitter` | morphing | `min`, `max` (durations) | random per-frame timing delay |
| `bucket` | morphing | `size` | pad frames to a fixed size quantum |

## Ordering

`Encode` is applied left→right, so order matters:

- **Compress before encrypt**: `flate` must precede `aead` (ciphertext is
  incompressible).
- **Mimicry goes last**: `tls-mimic`/`http-mimic` provide the wire framing and
  must be the outermost (rightmost) plugin, or the framing length prefix would
  precede the protocol header and break the disguise.
- **Recommended full chain**: `flate,aead?key=…,bucket?size=512,tls-mimic`.

## How it defeats DPI

A passive entropy detector flags high-entropy unknown protocols — which is what a
naive `aead`-only tunnel looks like. Wrapping the chain in `tls-mimic` makes the
wire begin with a convincing TLS handshake, so a fingerprint allowlist passes it
even though the body is encrypted. `probe-guard` drops active probes (the censor
gets no distinguishing response), and `pad`/`bucket` defeat size fingerprinting.

See `test/dpi/` for an end-to-end harness that demonstrates this against a
simulated firewall.
