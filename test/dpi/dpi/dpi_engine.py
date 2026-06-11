#!/usr/bin/env python3
"""GFW-style DPI middlebox for the Stunning test harness.

A transparent TCP relay that sits between the tunnel client and server. For each
connection it reassembles the opening bytes and classifies the flow the way a
censoring middlebox might:

  * Protocol fingerprint allowlist -- traffic that looks like real TLS or HTTP is
    allowed regardless of entropy (mimicry passes).
  * Shannon entropy detector -- a flow whose opening bytes are high-entropy and do
    NOT match an allowed fingerprint is treated as an unknown encrypted/obfuscated
    proxy and blocked (this is what catches a naive aead-only tunnel).
  * Plaintext marker detector -- a flow carrying known cleartext markers is blocked
    (baseline: an unobfuscated tunnel is trivially identified).

Every decision is appended as one JSON object to VERDICT_LOG. In MODE=enforce a
BLOCK verdict tears the connection down; in MODE=monitor the verdict is logged but
traffic is relayed anyway (useful for measuring throughput of a chain the censor
*would* have blocked).

Stdlib only -- no scapy/nfqueue, so it builds on a plain python3 image.
"""
import json
import math
import os
import socket
import sys
import threading
import time
from collections import defaultdict

LISTEN_HOST = os.environ.get("LISTEN_HOST", "0.0.0.0")
LISTEN_PORT = int(os.environ.get("LISTEN_PORT", "8443"))
UPSTREAM = os.environ.get("UPSTREAM", "172.31.0.20:8443")
ENTROPY_THRESHOLD = float(os.environ.get("ENTROPY_THRESHOLD", "7.4"))
INSPECT_BYTES = int(os.environ.get("INSPECT_BYTES", "512"))
MIN_DECISION_BYTES = int(os.environ.get("MIN_DECISION_BYTES", "16"))
MODE = os.environ.get("MODE", "enforce")  # enforce | monitor
VERDICT_LOG = os.environ.get("VERDICT_LOG", "/var/log/dpi/verdicts.jsonl")
MARKERS = [m.encode() for m in os.environ.get("MARKERS", "").split(",") if m]

os.makedirs(os.path.dirname(VERDICT_LOG), exist_ok=True)
_logf = open(VERDICT_LOG, "a", buffering=1)
_loglock = threading.Lock()


def shannon_entropy(data: bytes) -> float:
    if not data:
        return 0.0
    counts = defaultdict(int)
    for b in data:
        counts[b] += 1
    n = len(data)
    return -sum((c / n) * math.log2(c / n) for c in counts.values())


def looks_like_tls(data: bytes) -> bool:
    # TLS record: type=22 (handshake), version 0x03xx, ClientHello (0x01).
    return len(data) >= 6 and data[0] == 0x16 and data[1] == 0x03 and data[5] == 0x01


def looks_like_http(data: bytes) -> bool:
    verbs = (b"GET ", b"POST ", b"HEAD ", b"PUT ", b"OPTIONS ", b"CONNECT ", b"HTTP/")
    return any(data.startswith(v) for v in verbs)


def classify(data: bytes):
    """Return (allow, reason, entropy).

    A censor blocks known-bad cleartext markers even inside otherwise-allowed
    protocols, so markers are checked before the fingerprint allowlist.
    """
    for mk in MARKERS:
        if mk in data:
            return False, "plaintext-marker", shannon_entropy(data)
    if looks_like_tls(data):
        return True, "fingerprint:tls", shannon_entropy(data)
    if looks_like_http(data):
        return True, "fingerprint:http", shannon_entropy(data)
    ent = shannon_entropy(data)
    if len(data) >= MIN_DECISION_BYTES and ent >= ENTROPY_THRESHOLD:
        return False, "high-entropy-unknown", ent
    return True, "below-threshold", ent


def log_verdict(flow, allow, reason, entropy):
    rec = {
        "ts": round(time.time(), 3),
        "flow": flow,
        "verdict": "PASS" if allow else "BLOCK",
        "reason": reason,
        "entropy": round(entropy, 3),
        "mode": MODE,
    }
    with _loglock:
        _logf.write(json.dumps(rec) + "\n")


def pipe(src, dst):
    try:
        while True:
            chunk = src.recv(65536)
            if not chunk:
                break
            dst.sendall(chunk)
    except OSError:
        pass
    finally:
        for s in (src, dst):
            try:
                s.shutdown(socket.SHUT_RDWR)
            except OSError:
                pass


def handle(client, peer):
    flow = f"{peer[0]}:{peer[1]}"
    head = b""
    client.settimeout(10)
    try:
        while len(head) < MIN_DECISION_BYTES and len(head) < INSPECT_BYTES:
            chunk = client.recv(INSPECT_BYTES - len(head))
            if not chunk:
                break
            head += chunk
    except OSError:
        client.close()
        return
    client.settimeout(None)

    allow, reason, entropy = classify(head[:INSPECT_BYTES])
    log_verdict(flow, allow, reason, entropy)

    if not allow and MODE == "enforce":
        try:
            client.close()
        except OSError:
            pass
        return

    host, port = UPSTREAM.rsplit(":", 1)
    try:
        upstream = socket.create_connection((host, int(port)), timeout=10)
    except OSError as e:
        sys.stderr.write(f"dpi: upstream dial failed: {e}\n")
        client.close()
        return

    if head:
        try:
            upstream.sendall(head)
        except OSError:
            client.close()
            upstream.close()
            return

    threading.Thread(target=pipe, args=(client, upstream), daemon=True).start()
    pipe(upstream, client)


def main():
    srv = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    srv.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    srv.bind((LISTEN_HOST, LISTEN_PORT))
    srv.listen(128)
    sys.stderr.write(
        f"dpi-proxy: listen {LISTEN_HOST}:{LISTEN_PORT} -> {UPSTREAM} "
        f"mode={MODE} entropy>={ENTROPY_THRESHOLD} markers={len(MARKERS)}\n"
    )
    sys.stderr.flush()
    while True:
        try:
            client, peer = srv.accept()
        except OSError:
            continue
        threading.Thread(target=handle, args=(client, peer), daemon=True).start()


if __name__ == "__main__":
    main()
