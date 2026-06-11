# DPI Evasion Test Harness

A 3-node docker test bed that proves plugin chains evade a GFW-style deep-packet
inspection middlebox.

```
client (client-net) ──▶ dpi-router (both nets) ──▶ server ──▶ echo dest (server-net)
                         │ entropy + fingerprint + marker detectors
                         └ verdict log (PASS/BLOCK) per flow
```

- **dpi-router** (`dpi/dpi_engine.py`): a transparent TCP relay that reassembles
  each flow's opening bytes and classifies it like a censor — allowlists real-
  looking TLS/HTTP, blocks known cleartext markers, and blocks high-entropy
  unknown protocols (a naive encrypted proxy). `MODE=enforce` drops blocked
  flows; `MODE=monitor` logs but relays.
- **server/client**: the real `stunning` tunnel (tcp tunnel + tcp interface) with
  a configurable plugin chain.
- **dest / tools**: a Go echo destination and a load generator that verifies
  integrity and reports throughput/latency as JSON.

## Run

```bash
test/dpi/build.sh                 # one-time: cross-compile binaries + build images
test/dpi/scenarios.sh             # run the canonical evasion assertion suite
test/dpi/run.sh "<chain>" pass    # single scenario; expect pass|block|none
```

Binaries are cross-compiled on the host and mounted into static images, so code
changes only need a recompile (`build.sh` re-runs both), not slow image rebuilds.

## What the suite asserts

| chain | censor verdict |
|-------|----------------|
| `""` (plaintext, HTTP-shaped) | PASS (but exposed) |
| `""` + marker detector | BLOCK |
| `aead?key=…` | BLOCK (high entropy) |
| `aead?key=…,tls-mimic` | PASS (disguised as TLS) |
| `flate,aead?key=…,bucket?size=512,http-mimic` | PASS (disguised as HTTP) |

The payoff: a high-entropy AEAD tunnel that a censor *blocks* sails through once
wrapped in `tls-mimic`/`http-mimic`.

## CI

`.github/workflows/ci.yml` runs this suite (`e2e-dpi` job) and the plugin
benchmarks (`benchmarks` job) on every push/PR, uploading the verdict log and
benchmark results as artifacts.
