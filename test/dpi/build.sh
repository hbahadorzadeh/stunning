#!/usr/bin/env bash
# Prepares the DPI harness: cross-compiles the Go binaries for the container
# platform into gen/bin (mounted into the nodes at /cfg/bin) and builds the two
# static runtime images. The images bake no Go code, so after this one-time build
# you iterate by re-running this script's compile step only -- no image rebuilds,
# which avoids flaky container-storage rebuilds.
set -euo pipefail
cd "$(dirname "$0")/../.."

# Match the container architecture to the host so the mounted binaries run.
ARCH="$(uname -m)"
case "$ARCH" in
  arm64|aarch64) GOARCH=arm64 ;;
  x86_64|amd64)  GOARCH=amd64 ;;
  *) echo "unsupported arch $ARCH" >&2; exit 1 ;;
esac

echo "==> cross-compiling stunning + tools for linux/${GOARCH}"
mkdir -p test/dpi/gen/bin
GOOS=linux GOARCH=$GOARCH CGO_ENABLED=0 go build -o test/dpi/gen/bin/stunning .
GOOS=linux GOARCH=$GOARCH CGO_ENABLED=0 go build -o test/dpi/gen/bin/tools ./test/dpi/tools
chmod +x test/dpi/gen/bin/*

echo "==> building static images (one-time)"
docker build -f test/dpi/Dockerfile.stunning -t stunning-dpi-node:latest .
docker build -f test/dpi/Dockerfile.router  -t stunning-dpi-router:latest .
echo "ready: stunning-dpi-node, stunning-dpi-router, gen/bin/{stunning,tools}"
