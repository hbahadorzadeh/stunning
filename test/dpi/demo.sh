#!/usr/bin/env bash
# Narrated DPI-evasion demo, intended for screen recording / GIF capture.
#
# Tells the payoff story end to end against the simulated DPI censor:
#   the SAME encrypted tunnel a censor blocks passes cleanly once wrapped
#   in tls-mimic.
#
# Prereqs: docker, and the harness images built once:
#     test/dpi/build.sh
#
# Record a tight GIF (idle waits during docker churn are auto-collapsed):
#     asciinema rec --idle-time-limit 2 --overwrite \
#       --command "bash test/dpi/demo.sh" test/dpi/demo.cast
#     agg test/dpi/demo.cast test/dpi/demo.gif
#
# (Alternative recorder: vhs test/dpi/demo.tape)
set -euo pipefail
cd "$(dirname "$0")"

KEY="0123456789abcdef0123456789abcdef"
LOG="$(mktemp)"
trap 'rm -f "$LOG"' EXIT

r=$'\e[0m'; dim=$'\e[2m'; b=$'\e[1m'
red=$'\e[31m'; grn=$'\e[32m'; cyn=$'\e[36m'; yel=$'\e[33m'

say()   { printf '%s\n' "$*"; }
beat()  { sleep "${1:-1}"; }

# Run one harness scenario quietly, with a spinner so the recording is never a
# dead freeze. Sets $VERDICT (pass|block|none) and $MBPS from the real output.
run_scenario() { # chain expect
  ( MODE=enforce ./run.sh "$1" "$2" 262144 2 >"$LOG" 2>&1 || true ) &
  local pid=$! i=0 sp='|/-\'
  while kill -0 "$pid" 2>/dev/null; do
    printf '\r   %s%s sending traffic past the censor…%s' "$dim" "${sp:i++%4:1}" "$r"
    sleep 0.2
  done
  printf '\r\033[K'
  wait "$pid" 2>/dev/null || true
  VERDICT=$(grep -m1 'censor verdict:' "$LOG" | awk '{print $3}' || true)
  local raw_mbps
  raw_mbps=$(grep -m1 'metrics:' "$LOG" | grep -oE '"throughput_mbps":[[:space:]]*[0-9.]+' | grep -oE '[0-9.]+$' || true)
  MBPS=""
  if [ -n "$raw_mbps" ]; then
    MBPS=$(LC_NUMERIC=C printf '%.1f MB/s' "$raw_mbps")
  fi
}

clear
say "${b}${cyn}Stunning — anti-DPI plugin chains${r}"
say "${dim}  client ──▶ [ DPI censor ] ──▶ server ──▶ destination${r}"
say "${dim}  the censor allowlists TLS/HTTP and blocks unknown high-entropy flows${r}"
say ""
beat 1.5

# --- 1) encrypted only -> blocked --------------------------------------------
say "${b}1) Encrypted tunnel, no disguise${r}    ${dim}Plugins:${r} ${yel}aead${r}"
say "   ${dim}strong AEAD encryption — but the wire is high-entropy and unknown${r}"
beat 1
run_scenario "aead?key=${KEY}" block
say "   censor sees random-looking bytes  →  ${red}${b}■ BLOCKED${r}"
say ""
beat 2

# --- 2) + tls-mimic -> passes -------------------------------------------------
say "${b}2) Same tunnel, one more plugin${r}     ${dim}Plugins:${r} ${yel}aead,tls-mimic${r}"
say "   ${dim}tls-mimic wraps the exact same stream in a convincing TLS handshake${r}"
beat 1
run_scenario "aead?key=${KEY},tls-mimic" pass
say "   censor sees an ordinary TLS session  →  ${grn}${b}● PASSED${r}   ${dim}(${MBPS:-fast}, integrity verified)${r}"
say ""
beat 2

say "${yel}${b}Same encryption. One plugin. Blocked ➜ invisible.${r}"
say "${dim}github.com/hbahadorzadeh/stunning${r}"
beat 2
