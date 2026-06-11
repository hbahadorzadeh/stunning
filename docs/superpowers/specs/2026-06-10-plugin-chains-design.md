# Plugin Chains for Anti-DPI Tunneling — Design

Status: approved (2026-06-10). Goal: a composable plugin system for the Stunning
tunnel that helps users defeat the Iran/China great firewalls. Plugins fall into
four categories — anti-DPI/mimicry, security/crypto, size optimization, and
active-probe resistance — and can be combined in any order inside a *chain*.

## Background

The repo already declares a plugin abstraction (`core/common/plugin_chain.go`)
based on Go's `plugin.Open` (`.so` dynamic loading). It has **no plugins**, is
**wired into no tunnel data path**, and the `.so` mechanism does not cross-compile
or run on mobile/Windows. We replace the loading mechanism with an in-process
registry while leaving the old file in place (no file removals).

## Decisions (locked)

- **Mechanism**: in-process registry of compiled-in plugins (no `.so`, no CGO).
- **Contract**: stateful, per-connection plugin instances built by factories from
  parsed params (supports AEAD nonce sequencing, handshakes, per-flow timing).
- **Framing**: length-prefixed message framing; the length prefix is XOR-masked
  with a per-connection keystream so there is no plaintext length tell. Datagram
  tunnels pass whole packets (no prefix).
- **DPI sim**: custom `nftables` + Python DPI engine container as the middle
  router node between client and server containers.
- **Sequencing**: phased. Phase 1 = foundation + harness + 4 seed plugins. Phase 2
  = mimicry/morphing. Phase 3 = 20× optimize loop. Each phase builds on the prior.

## Architecture

### `core/plugin` package

```go
type Plugin interface {
    Encode(src []byte) ([]byte, error) // clear -> wire
    Decode(src []byte) ([]byte, error) // wire -> clear
    Close() error
}
type Params map[string]string            // typed getters: String/Int/Bytes/Bool/Duration
type Factory func(p Params) (Plugin, error)
func Register(name string, f Factory)
func New(name string, p Params) (Plugin, error)
```

Plugins self-register in `init()`. A `Stateless` helper adapts pure
`func([]byte) []byte` transforms to the interface.

### Chain (`chain.go`)

- Config string: `flate,aead?key=<hex>&algo=chacha,pad?min=16&max=256`.
- `Encode` applies plugins left→right; `Decode` applies right→left.
- Built **per connection**; each plugin gets fresh state.
- Client and server use the *same* chain string → client `Encode` is the exact
  inverse of server `Decode`.

### Framing (`frame.go`)

- `FramedConn` wraps `net.Conn`. `Write(msg)` → `chain.Encode` →
  `[masked-uint16 len][payload]`. `Read` reverses. Length masked by a keystream
  seeded from the chain PSK. Max frame 16 KiB (configurable).
- `FramedPacketConn` wraps a datagram conn: whole packet through the chain, no
  length prefix.

### Wiring

- `TunnelConfig` gains `Plugins string`. `core/tunnel/common` wraps the
  established connection in `FramedConn`/`FramedPacketConn` before the interface
  layer reads/writes. Stream tunnels → `FramedConn`; datagram → `FramedPacketConn`.

## Phase 1 plugins

| Name | Category | Behavior |
|------|----------|----------|
| `flate` | size | stdlib DEFLATE compress/decompress (pure Go) |
| `aead` | security | ChaCha20-Poly1305 (default) or AES-GCM; per-frame counter nonce; key from params |
| `pad` | anti-DPI | random length padding, self-describing trailer; breaks size fingerprints |
| `probe-guard` | active-probe | keyed BLAKE2b tag prefix; server drops frames failing auth → no probe response |

Recommended order: `flate → aead → pad → probe-guard`. Tests validate arbitrary
permutations for round-trip correctness.

## DPI harness (`test/dpi`)

`docker-compose.dpi.yml`, two networks (`client-net`, `server-net`):

- **server**: stunning server + chain.
- **dpi-router**: gateway between nets; `nftables` + Python DPI engine
  (entropy threshold, protocol fingerprinting, optional active-probe), per-flow
  PASS/BLOCK verdict log.
- **client**: stunning client + traffic generator + integrity check.

Scenario runner asserts (a) end-to-end integrity, (b) DPI verdict
(chain evades where baseline is flagged), (c) throughput/latency JSON metrics.

## Testing & benchmarking

- Unit: per-plugin round-trip, param parsing, error/tamper paths.
- Property/fuzz: random buffers × random chain permutations round-trip.
- Go benchmarks: per plugin + full chain (MB/s, allocs/op).
- E2E (Go): real TCP tunnel client→server through a chain.
- E2E (docker): through the DPI router — evasion + perf.

## Later phases

- **Phase 2**: mimicry plugins (tls/http/quic/rtp), timing jitter, size morphing,
  decoy/chaff, zstd/lz4/brotli size plugins.
- **Phase 3**: ≥20 iterations of bench → profile (pprof) → fix bottleneck
  (allocs, copies, syscalls; buffer pooling, zero-copy, batching) → re-bench,
  tracking throughput/latency/evasion deltas.

## Constraints

- No file removals. Old `.so` loader stays.
- Pure Go, cross-compilable (mobile/desktop/Windows) — no CGO in plugins.
- Plugins deterministic and reversible (except deliberately random pad/timing,
  which carry self-describing metadata so Decode is exact).
