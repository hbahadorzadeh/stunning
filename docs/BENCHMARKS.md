# Benchmarks

Performance of the plugin-chain system: per-plugin and per-chain Go
microbenchmarks, plus end-to-end throughput measured through the docker DPI
evasion harness.

## Environment

- CPU: Apple M-series (arm64), 8 logical CPUs
- Go: 1.26
- Microbenchmarks: `go test -bench` on `core/plugin`, `-benchtime=300ms`
- End-to-end: `test/dpi/scenarios.sh`, 1 MiB × 2 concurrent streams

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
| `pad` | 682 | 6002 MB/s | 4866 | 2 |
| `aead` (AES-GCM) | 2111 | 1940 MB/s | 8960 | 2 |
| `aead` (ChaCha20) | 9847 | 416 MB/s | 8960 | 2 |
| `probe-guard` | 10586 | 387 MB/s | 4899 | 3 |
| `flate` | 25183 | 163 MB/s | 10840 | 12 |

AES-GCM is hardware-accelerated (AES-NI/ARM crypto extensions). `flate` is the
heaviest plugin and dominates any chain that includes it.

## Per-chain

| Chain | ns/op | Throughput | B/op | allocs/op |
|-------|------:|-----------:|-----:|----------:|
| `flate,aead,pad,probe-guard` (4 KiB) | 28057 | 146 MB/s | 12341 | 19 |
| `flate,aead,pad,probe-guard` (64 KiB) | 143830 | 456 MB/s | 151733 | 27 |
| `aead,tls-mimic` (framed, 4 KiB) | 17739 | 231 MB/s | 18756 | 6 |
| `aead,http-mimic` (framed, 4 KiB) | 17924 | 229 MB/s | 18800 | 10 |
| `flate,aead,bucket,tls-mimic` (framed, 4 KiB) | 39358 | 104 MB/s | 14836 | 19 |

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

Remaining cost is intrinsic (`flate`/AEAD CPU) plus the Python test middlebox, so
the loop stopped where measured gains went flat. Raw per-iteration numbers live in
`docs/superpowers/bench/`.
