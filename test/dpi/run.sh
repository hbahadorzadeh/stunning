#!/usr/bin/env bash
# DPI harness scenario runner.
#
#   run.sh "<chain-spec>" [expect] [size] [streams]
#
# expect: pass | block | none   (asserted against the DPI verdict; default none)
# Brings up client -> dpi-router -> server -> dest, drives traffic from inside the
# client container to its local tunnel interface, and reports throughput + the
# censor's verdict. Exit non-zero if the assertion fails or traffic is corrupted.
set -euo pipefail
cd "$(dirname "$0")"

CHAIN="${1:-}"
EXPECT="${2:-none}"
SIZE="${3:-4194304}"
STREAMS="${4:-4}"
export MODE="${MODE:-enforce}"
export ENTROPY_THRESHOLD="${ENTROPY_THRESHOLD:-7.4}"
export MARKERS="${MARKERS:-}"

# Pick docker compose flavor.
if docker compose version >/dev/null 2>&1; then
  DC="docker compose"
elif command -v docker-compose >/dev/null 2>&1; then
  DC="docker-compose"
else
  echo "ERROR: neither 'docker compose' nor 'docker-compose' is available" >&2
  exit 2
fi
COMPOSE="$DC -f docker-compose.dpi.yml -p stunning_dpi"

mkdir -p gen/dpi-log
: > gen/dpi-log/verdicts.jsonl

cat > gen/server.json <<JSON
{
  "server": {
    "ServiceMode": "server",
    "ServerType": "tcp",
    "InterfaceType": "tcp",
    "Listen": ":8443",
    "Connect": "172.31.0.30:9000",
    "Plugins": "${CHAIN}"
  }
}
JSON

cat > gen/client.json <<JSON
{
  "client": {
    "ServiceMode": "client",
    "ServerType": "tcp",
    "InterfaceType": "tcp",
    "Listen": ":1080",
    "Connect": "172.30.0.10:8443",
    "Plugins": "${CHAIN}"
  }
}
JSON

cleanup() { $COMPOSE down -v --remove-orphans >/dev/null 2>&1 || true; }
trap cleanup EXIT

echo ">>> scenario: chain='${CHAIN:-<none>}' expect=${EXPECT} mode=${MODE}"
# Images are pre-built via test/dpi/build.sh (compose build needs a credential
# helper that isn't present in every environment), so bring the stack up as-is.
$COMPOSE up -d --no-build >/dev/null

# Wait for the client tunnel interface listener to come up. This is a connect-only
# probe so it succeeds regardless of the censor's verdict (which applies further
# downstream), letting block scenarios reach the measured workload.
ready=0
for i in $(seq 1 30); do
  if $COMPOSE exec -T client sh -c '/cfg/bin/tools probe -connect 127.0.0.1:1080 -timeout 2s' >/dev/null 2>&1; then
    ready=1; break
  fi
  sleep 1
done
if [ "$ready" != "1" ]; then
  echo "ERROR: client interface never became ready" >&2
  $COMPOSE logs --no-color client server dpi-router | tail -40 >&2
  exit 1
fi

# Drive the measured workload.
set +e
RESULT=$($COMPOSE exec -T client sh -c "/cfg/bin/tools gen -connect 127.0.0.1:1080 -size ${SIZE} -streams ${STREAMS} -timeout 60s")
GEN_RC=$?
set -e
echo "metrics: $RESULT"

# Let the DPI engine flush its async verdict line before we read it.
sleep 1

# Summarize the censor's view.
VERDICT="none"
if [ -s gen/dpi-log/verdicts.jsonl ]; then
  if grep -q '"verdict": "BLOCK"' gen/dpi-log/verdicts.jsonl; then VERDICT="block"; else VERDICT="pass"; fi
  echo "dpi verdicts:"; cat gen/dpi-log/verdicts.jsonl
fi
echo "censor verdict: ${VERDICT}"

rc=0
case "$EXPECT" in
  pass)
    [ "$VERDICT" = "pass" ] || { echo "FAIL: expected censor PASS, got ${VERDICT}"; rc=1; }
    [ "$GEN_RC" = "0" ] || { echo "FAIL: traffic did not complete intact"; rc=1; } ;;
  block)
    [ "$VERDICT" = "block" ] || { echo "FAIL: expected censor BLOCK, got ${VERDICT}"; rc=1; } ;;
  none) : ;;
  *) echo "unknown expect '$EXPECT'"; rc=2 ;;
esac
[ "$rc" = "0" ] && echo "OK: scenario passed"
exit $rc
