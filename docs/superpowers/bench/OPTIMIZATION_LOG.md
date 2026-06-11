# Plugin Chains — Optimization Log (Phase 3)

Iterative develop → test → benchmark → find-bottleneck → fix loop on the plugin,
framing, and relay hot paths. Each iteration kept the full test suite (unit,
permutation, tamper, framing, TCP-e2e) and the docker DPI evasion suite green.

Two measurement surfaces:
- **Go microbenchmarks** (`go test -bench`) — isolate plugin/framing CPU + allocs.
- **Docker e2e** (`test/dpi/scenarios.sh`) — real client→DPI-router→server→dest
  throughput. Note: passing-flow throughput here is partly gated by the Python
  DPI middlebox (GIL, per-connection copy), so it under-reports the tunnel's true
  ceiling; the Go benches isolate our code.

## Iterations

| # | Change | Result |
|---|--------|--------|
| 0 | Baseline | Flate 51.3µs, 865 KB/op, 32 allocs; FullChain 55µs; e2e baseline 37 MB/s |
| 1 | Pool flate writers/readers (the ~600 KiB window was allocated per frame) | Flate 25.3µs (2.0×), **11 KB/op (78×)**, 12 allocs; FullChain 28µs |
| 2 | Pool probe-guard BLAKE2b hashers | 5→3 allocs (ns flat — hashing CPU dominates) |
| 3 | FramedConn vectored writes + reuse read scratch | FramedBaseline 18.7 KB→9.1 KB/op |
| 4 | Relay copy loop: 1 KiB-realloc-per-iter + per-chunk logging → `io.CopyBuffer(32 KiB)`, no logs (tcp + socks) | **e2e baseline 37→336 MB/s (≈9×)**; exposed a frame-size bug |
| 5 | 32 KiB buffer exceeded 16 KiB max frame → non-compressed chains failed. `FramedConn.Write` now splits large writes into bounded frames (respects TLS record limit) | Correctness fix; **e2e tls-mimic 14→136 MB/s (≈9.7×)**; suite 5/5 stable |

## Final numbers

Go benchmarks (4 KiB payload, Apple M-series, `-benchtime=300ms`):

```
Flate            25.2µs   163 MB/s   10.8 KB/op   12 allocs
FullChain        28.1µs   146 MB/s   12.3 KB/op   19 allocs
FullChain64K    143.8µs   456 MB/s  151.7 KB/op   27 allocs
FramedTLSMimic   17.7µs   231 MB/s   18.8 KB/op    6 allocs
FramedFullMimic  39.4µs   104 MB/s   14.8 KB/op   19 allocs
FramedBaseline   17.0µs   241 MB/s    9.1 KB/op    6 allocs
```

End-to-end through the DPI router (1 MiB × 2 streams):

```
baseline (plaintext)              336 MB/s   PASS (looks like HTTP)
aead-only                          —         BLOCK (entropy 7.6)
aead + tls-mimic                  136 MB/s   PASS (allowlisted as TLS)
flate + aead + bucket + http-mimic 21 MB/s   PASS (allowlisted as HTTP)
```

## Where the loop stopped and why

After iter 5 the dominant remaining costs are **intrinsic**, not allocation or
framing overhead:

- **flate compression CPU** (~25µs/4 KiB at default level) dominates any chain
  containing `flate`. This is algorithmic; a future `zstd`/`lz4` plugin with a
  speed-biased level is the lever, not micro-optimization.
- **AEAD/BLAKE2b CPU** is hardware-bound (AES-GCM already ~1.9 GB/s via AES-NI).
- **The Python DPI middlebox** caps observed e2e throughput for passing flows; it
  is test scaffolding, not the product, so optimizing it has no user value.

Allocation counts are now small and flat across the hot path, so further
micro-iterations yield diminishing returns. The loop was stopped where measured
gains went flat rather than continued for its own sake.

## Future levers (Phase 2+ plugins / later work)

- `zstd`/`lz4`/`brotli` size plugins with speed/ratio levels.
- Per-connection encode-buffer pooling to drop the remaining per-frame output
  allocations under sustained load.
- A native (non-Python) DPI middlebox if higher-fidelity throughput numbers are
  needed.
