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
| `profile` | morphing | `name` (`web`\|`video`\|`voip`\|`custom`), `quantum`, `min`, `max` | mimic a real protocol's size/timing distribution (quantize size + sampled delay) to defeat statistical classifiers |
| `chaff` | morphing | `min`, `max` (decoy size), `interval`, `jitter` | inject decoy/cover frames on a timer to mask volume and timing; the peer drops them. Must be the **innermost** (first) plugin |

## Ordering

`Encode` is applied left→right, so order matters:

- **Compress before encrypt**: `flate` must precede `aead` (ciphertext is
  incompressible).
- **Mimicry goes last**: `tls-mimic`/`http-mimic` provide the wire framing and
  must be the outermost (rightmost) plugin, or the framing length prefix would
  precede the protocol header and break the disguise.
- **Chaff goes first**: `chaff` tags real vs decoy frames and must be the
  innermost (leftmost) plugin so its type byte is covered by the outer
  encryption/disguise.
- **Recommended full chain**: `flate,aead?key=…,bucket?size=512,tls-mimic`, or
  with cover traffic: `chaff,flate,aead?key=…,profile?name=web,tls-mimic`.

## How it defeats DPI

A passive entropy detector flags high-entropy unknown protocols — which is what a
naive `aead`-only tunnel looks like. Wrapping the chain in `tls-mimic` makes the
wire begin with a convincing TLS handshake, so a fingerprint allowlist passes it
even though the body is encrypted. `probe-guard` drops active probes (the censor
gets no distinguishing response), and `pad`/`bucket` defeat size fingerprinting.

See `test/dpi/` for an end-to-end harness that demonstrates this against a
simulated firewall.

## Gates

Gates are connection-level controls that run *around* the byte-transform chain,
not as payload transforms. There are two kinds, each configured by its own field.

### Authentication (`Auth`)

An authenticator runs a handshake right after the connection is established —
**inside** the plugin framing, so the handshake itself is disguised — and either
lets the connection proceed or rejects it. The client proves identity; the server
verifies and records the identity.

```json
"Auth": "psk?key=0123456789abcdef"
```

| Authenticator | Params | Behavior |
|---------------|--------|----------|
| `psk` | `key` (hex, both peers) | HMAC-SHA256 challenge-response; the server sends a fresh per-connection challenge, the client returns `HMAC(key, challenge)` |
| `jwt` | client: `token`; server: `alg` (`HS256`\|`RS256`), `secret` (hex, HS256) or `pubkey` (PEM path, RS256) | Client presents a signed JWT; server verifies signature + `exp`/`nbf`; identity = `sub` claim |
| `mtls` | client: `cert`, `key`, `ca`, `servername`; server: `cert`, `key`, `clientca` | Mutual-TLS handshake over the conn; identity = client cert Common Name |
| `oauth` | client: `token`; server: `introspect` (URL), `client_id`, `client_secret`, `scope` | Client presents an OAuth 2.0 access token; server validates it via an RFC 7662 introspection endpoint; identity = `sub`/`username` |
| `ldap` | client: `user`, `password`; server: `url`, `userdn` (DN template with `%s`) | Client sends credentials; server verifies by binding to the directory as the user's DN; identity = username. Run behind `aead`/`ldaps` since the password crosses the connection |

Auth config is asymmetric: the client carries its credential (`token`, client
cert, password), the server carries the verification material (`secret`,
`pubkey`, `clientca`, `introspect`, `url`). The `oauth`/`ldap` end-to-end tests
run against a mock OIDC IdP and an OpenLDAP directory in the harness
(`test/dpi/scenarios-auth.sh`).

### Port knocking (`Knock`)

Port knocking authorizes a source IP *before* it may connect to the tunnel port,
so an unauthenticated scanner/prober sees nothing on the tunnel port at all.

```json
"Knock": "spa?key=0123456789abcdef&port=62201&ttl=10s"
```

| Knocker | Params | Behavior |
|---------|--------|----------|
| `spa` | `key` (hex, required), `port` (UDP knock port, required), `ttl` (authorized window, default 10s), `window` (timestamp tolerance, default 30s), `delay` (client settle, default 100ms) | Single-packet authorization: the client sends one authenticated UDP packet (`nonce ‖ timestamp ‖ HMAC`); the server verifies it (with replay + timestamp checks) and authorizes the source IP for `ttl`. The tunnel server drops connections from un-knocked IPs. |

### Order of operations

```
client:  knock (SPA/UDP) ──▶ dial ──▶ plugin chain framing ──▶ auth handshake ──▶ data
server:  knock listener authorizes IP ──▶ accept (gated) ──▶ chain ──▶ auth verify ──▶ data
```

Gates compose with the byte-transform chain: e.g. a tunnel can require a knock,
disguise itself as TLS, encrypt with `aead`, and authenticate clients by JWT all
at once.

## Security considerations

- **Use `aead` for confidentiality and integrity.** Without it the chain is
  obfuscation only — a network attacker can read, modify, or replay traffic. Add
  `probe-guard` for an outer authenticated gate. The recommended baseline is
  `…,aead?key=…,…`.
- **Gates send credentials over the chain.** `psk`/`jwt`/`oauth`/`ldap` exchange
  secrets during the handshake, so run them behind `aead` (the handshake is
  carried inside the chain framing) — or use `mtls`, which encrypts itself, and
  `ldaps` for the directory connection. Never run `ldap` over an unencrypted
  chain.
- **Bearer tokens are replayable until expiry.** A captured `jwt`/`oauth` token
  works until it expires; use short lifetimes, and prefer `mtls` (bound to a
  private key) where possible.
- **`psk` authenticates the client to the server only**, not the reverse; the
  shared `aead` key provides the mutual binding.
- **`mtls insecure=true` disables server verification** — testing only; it logs a
  warning at startup.
- **Keep the chain/auth/knock specs secret.** They embed key material (`key=`,
  `secret=`, `password=`); treat config files as secrets and keep them out of
  logs and shared repos.
- **Port knocking authorizes by source IP** for `ttl`; behind NAT, other hosts
  sharing that IP are also allowed during the window. Pair it with `auth` for
  per-client identity.
