#!/bin/sh
# Runs the DPI relay. Tunables come from the environment (see dpi_engine.py):
#   LISTEN_PORT (default 8443), UPSTREAM, MODE, ENTROPY_THRESHOLD, MARKERS.
set -e
exec python3 /opt/dpi/dpi_engine.py
