#!/bin/sh
# Puts the mounted host-built binaries on PATH, optionally adds a static route to
# the far subnet via the DPI router, then execs the node command. Env:
#   ROUTE_NET   subnet to route via the router (e.g. 172.31.0.0/24)
#   ROUTE_VIA   router IP on this node's network (e.g. 172.30.0.10)
set -e
export PATH=/cfg/bin:$PATH
if [ -n "${ROUTE_NET:-}" ] && [ -n "${ROUTE_VIA:-}" ]; then
    ip route replace "$ROUTE_NET" via "$ROUTE_VIA"
    echo "node: route $ROUTE_NET via $ROUTE_VIA installed"
fi
exec "$@"
