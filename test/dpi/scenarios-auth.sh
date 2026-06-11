#!/usr/bin/env bash
# Runs the auth e2e suite against the mock OIDC IdP and the OpenLDAP directory.
# Asserts authorized clients connect and bad credentials are rejected. Requires
# images built (test/dpi/build.sh).
set -uo pipefail
cd "$(dirname "$0")"

PASS=0
FAIL=0
for scn in oauth-good oauth-bad ldap-good ldap-bad; do
  echo
  echo "=================================================================="
  if bash ./run-auth.sh "$scn"; then
    PASS=$((PASS + 1))
  else
    FAIL=$((FAIL + 1))
  fi
done

echo
echo "AUTH SUITE: ${PASS} passed, ${FAIL} failed"
[ "$FAIL" -eq 0 ]
