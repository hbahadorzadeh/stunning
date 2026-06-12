# Benchmarks

Performance of the plugin-chain system: per-plugin and per-chain Go
microbenchmarks, plus end-to-end throughput measured through the docker DPI
evasion harness.

## Environment

- CPU: Apple M-series (arm64), 8 logical CPUs
- Go: 1.25
- Microbenchmarks: `go test -bench` on `core/plugin`, `-benchtime=500ms`
- End-to-end: `test/dpi/scenarios.sh`, 1 MiB × 2 concurrent streams

> Absolute ns/op is machine- and thermal-dependent; compare allocations and
> relative cost between chains. The interesting invariants are the low allocation
> counts and that `flate` dominates wherever it appears.

> The end-to-end figures pass through a Python DPI middlebox (the simulated
> firewall), whose per-connection copy partly caps passing-flow throughput. The
> Go microbenchmarks isolate the tunnel's own cost.

## Reproduce

```bash
# microbenchmarks
go test -run='^$' -bench=. -benchmem ./core/plugin/

# end-to-end through the simulated firewall
test/dpi/build.sh && test/dpi/scenarios.sh
```

## Per-plugin (4 KiB payload, encode+decode round trip)

| Plugin | ns/op | Throughput | B/op | allocs/op |
|--------|------:|-----------:|-----:|----------:|
| `profile` | 642 | 6382 MB/s | 4864 | 1 |
| `pad` | 675 | 6068 MB/s | 4866 | 2 |
| `aead` (AES-GCM) | 2193 | 1868 MB/s | 8960 | 2 |
| `aead` (ChaCha20) | 9716 | 422 MB/s | 8960 | 2 |
| `probe-guard` | 10512 | 390 MB/s | 4896 | 2 |
| `flate` | 38035 | 108 MB/s | 10648 | 13 |

AES-GCM is hardware-accelerated (AES-NI/ARM crypto extensions). `flate` is the
heaviest plugin and dominates any chain that includes it; everything else is
sub-microsecond to a few microseconds with ≤2 allocations.

## Per-chain

| Chain | ns/op | Throughput | B/op | allocs/op |
|-------|------:|-----------:|-----:|----------:|
| `flate,aead,pad,probe-guard` (4 KiB) | 40285 | 102 MB/s | 11439 | 19 |
| `flate,aead,pad,probe-guard` (64 KiB) | 156145 | 420 MB/s | 140776 | 27 |
| `aead,tls-mimic` (framed, 4 KiB) | 16203 | 253 MB/s | 13878 | 5 |
| `aead,http-mimic` (framed, 4 KiB) | 16861 | 243 MB/s | 18777 | 10 |
| `chaff,aead,tls-mimic` (framed, 4 KiB) | 16702 | 245 MB/s | 19510 | 6 |
| `flate,aead,bucket,tls-mimic` (framed, 4 KiB) | 53481 | 77 MB/s | 11979 | 19 |

`chaff` adds negligible per-frame cost over the baseline framed chain; its decoy
injection is paced by a background timer, not the data path.

## Gates (one-time, off the data path)

Authentication and port-knock gates run once per connection during setup, not per
frame, so they do not affect steady-state throughput:

- `psk`/`jwt` — a few small handshake messages (HMAC/JWT verify).
- `mtls` — one TLS handshake (then the conn is a normal TLS session).
- `oauth`/`ldap` — one HTTP introspection / LDAP bind round trip to the IdP/dir.
- `spa` knock — one UDP packet plus a short client settle delay before dialing.

## End-to-end through the DPI firewall (1 MiB × 2 streams)

| Chain | Censor verdict | Throughput | Integrity |
|-------|----------------|-----------:|-----------|
| `""` (plaintext, HTTP-shaped) | PASS — but exposed | 336 MB/s | ✓ |
| `""` + marker rule | **BLOCK** | — | — |
| `aead?key=…` | **BLOCK** (entropy 7.6) | — | — |
| `aead?key=…,tls-mimic` | PASS (looks like TLS) | 136 MB/s | ✓ |
| `flate,aead?key=…,bucket?size=512,http-mimic` | PASS (looks like HTTP) | 21 MB/s | ✓ |

The payoff: a high-entropy AEAD tunnel the censor **blocks** passes cleanly once
wrapped in `tls-mimic`.

## Optimization history

The hot path was tuned iteratively (develop → test → benchmark → fix), each step
keeping every test green:

| Step | Change | Effect |
|------|--------|--------|
| 1 | Pool `flate` writers/readers (the ~600 KiB window was allocated per frame) | flate 865 KB → **11 KB/op** (78×), 2× faster |
| 2 | Pool `probe-guard` BLAKE2b hashers | 5 → 3 allocs/op |
| 3 | FramedConn vectored writes + reuse read scratch | FramedBaseline 18.7 KB → 9.1 KB/op |
| 4 | Relay copy buffer 1 KiB→32 KiB, drop per-chunk logging | end-to-end baseline 37 → **336 MB/s** |
| 5 | Split large writes into bounded frames (was failing >16 KiB) | fix + tls-mimic 14 → **136 MB/s** |
| 6 | Direct per-conn fields (flate/probe-guard/tls-mimic) instead of `sync.Pool` | equal CPU, fewer allocs, no GC reclaim of reusable state between frames |
| 7 | `probe-guard` writes its tag in a single allocation | 3 → 2 allocs/op |

Remaining cost is intrinsic (`flate`/AEAD CPU) plus the Python test middlebox, so
the loop stopped where measured gains went flat. Raw per-iteration numbers live in
`docs/superpowers/bench/`.
