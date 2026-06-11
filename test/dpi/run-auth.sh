#!/usr/bin/env bash
# Auth e2e runner. Brings up client -> server -> dest with a mock OIDC IdP and an
# OpenLDAP directory, and asserts that authorized clients connect and that bad
# credentials are rejected.
#
#   run-auth.sh <oauth-good|oauth-bad|ldap-good|ldap-bad>
#
# Exit 0 if the scenario behaves as expected.
set -uo pipefail
cd "$(dirname "$0")"

SCN="${1:-oauth-good}"
KEY="0123456789abcdef0123456789abcdef"

if docker compose version >/dev/null 2>&1; then DC="docker compose"; else DC="docker-compose"; fi
COMPOSE="$DC -f docker-compose.auth.yml -p stunning_auth"

case "$SCN" in
  oauth-good) SVC_AUTH="oauth?introspect=http://idp:8080/introspect"; CLI_AUTH="oauth?token=valid-token"; EXPECT=pass ;;
  oauth-bad)  SVC_AUTH="oauth?introspect=http://idp:8080/introspect"; CLI_AUTH="oauth?token=nope";        EXPECT=fail ;;
  ldap-good)  SVC_AUTH="ldap?url=ldap://ldap:389&userdn=uid=%s,ou=users,dc=example,dc=org"; CLI_AUTH="ldap?user=alice&password=s3cret"; EXPECT=pass ;;
  ldap-bad)   SVC_AUTH="ldap?url=ldap://ldap:389&userdn=uid=%s,ou=users,dc=example,dc=org"; CLI_AUTH="ldap?user=alice&password=wrong";  EXPECT=fail ;;
  *) echo "unknown scenario $SCN"; exit 2 ;;
esac

mkdir -p gen
cat > gen/server.json <<JSON
{ "server": { "ServiceMode":"server","ServerType":"tcp","InterfaceType":"tcp",
  "Listen":":8443","Connect":"dest:9000","Plugins":"aead?key=${KEY}","Auth":"${SVC_AUTH}" } }
JSON
cat > gen/client.json <<JSON
{ "client": { "ServiceMode":"client","ServerType":"tcp","InterfaceType":"tcp",
  "Listen":":1080","Connect":"server:8443","Plugins":"aead?key=${KEY}","Auth":"${CLI_AUTH}" } }
JSON

cleanup() { $COMPOSE down -v --remove-orphans >/dev/null 2>&1 || true; }
trap cleanup EXIT

echo ">>> auth scenario: ${SCN} expect=${EXPECT}"
case "$SCN" in
  oauth-*) SERVICES="dest idp server client" ;;
  ldap-*)  SERVICES="dest ldap server client" ;;
esac
$COMPOSE up -d --no-build $SERVICES >/dev/null

# LDAP needs time to seed; the IdP is instant.
case "$SCN" in ldap-*) sleep 25 ;; *) sleep 5 ;; esac

ready=0
for i in $(seq 1 30); do
  if $COMPOSE exec -T client sh -c '/cfg/bin/tools probe -connect 127.0.0.1:1080 -timeout 2s' >/dev/null 2>&1; then
    ready=1; break
  fi
  sleep 1
done
[ "$ready" = 1 ] || { echo "client interface not ready"; $COMPOSE logs --no-color client server | tail -30; exit 1; }

set +e
RESULT=$($COMPOSE exec -T client sh -c "/cfg/bin/tools gen -connect 127.0.0.1:1080 -size 65536 -streams 1 -timeout 20s")
RC=$?
set -e
echo "metrics: $RESULT"

rc=0
if [ "$EXPECT" = pass ]; then
  [ "$RC" = 0 ] || { echo "FAIL: authorized client should have connected"; $COMPOSE logs --no-color server | tail -20; rc=1; }
else
  [ "$RC" != 0 ] || { echo "FAIL: bad credentials should have been rejected"; rc=1; }
fi
[ "$rc" = 0 ] && echo "OK: ${SCN} behaved as expected"
exit $rc
