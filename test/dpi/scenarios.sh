#!/usr/bin/env bash
# Runs the canonical DPI evasion assertion suite. Each scenario brings up the
# client -> dpi-router -> server -> dest stack and asserts the censor's verdict.
# Exits non-zero if any scenario fails. Requires images built (test/dpi/build.sh).
set -uo pipefail
cd "$(dirname "$0")"

KEY="0123456789abcdef0123456789abcdef"
PASS=0
FAIL=0

run() { # desc, expect-env..., chain, expect
  desc="$1"; shift
  echo
  echo "==================================================================="
  echo "# $desc"
  echo "==================================================================="
  if env "$@"; then
    PASS=$((PASS + 1)); echo "RESULT: PASS"
  else
    FAIL=$((FAIL + 1)); echo "RESULT: FAIL"
  fi
}

# 1. Plaintext tunnel that looks like HTTP -> censor allows it (but it is exposed).
run "baseline plaintext (HTTP-shaped) -> PASS" \
  ./run.sh "" pass 524288 2

# 2. Plaintext tunnel carrying a known marker -> censor blocks it.
run "plaintext + marker detector -> BLOCK" \
  MARKERS=STUNNING-PROBE ./run.sh "" block 131072 1

# 3. Naive encryption (aead only) is high-entropy and unknown -> censor blocks it.
run "aead-only (high entropy) -> BLOCK" \
  ./run.sh "aead?key=${KEY}" block 131072 1

# 4. The payoff: aead disguised as TLS evades the entropy detector -> PASS, intact.
run "aead + tls-mimic (TLS disguise) -> PASS" \
  ./run.sh "aead?key=${KEY},tls-mimic" pass 524288 2

# 5. Full chain disguised as HTTP -> PASS, intact.
run "flate+aead+bucket + http-mimic -> PASS" \
  ./run.sh "flate,aead?key=${KEY},bucket?size=512,http-mimic" pass 524288 2

echo
echo "==================================================================="
echo "SUITE: ${PASS} passed, ${FAIL} failed"
echo "==================================================================="
[ "$FAIL" -eq 0 ]
