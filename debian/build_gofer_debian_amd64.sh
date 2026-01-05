#!/bin/bash
set -e

# Ensure we are inside the debian build sandbox
cd "$(dirname "$0")"

# Clean previous build
fakeroot ./debian/rules clean

# Build package (dpkg will drop outputs in parent dir)
dpkg-buildpackage -us -uc -b -a amd64

# Collect all build artifacts locally
mkdir -p artifacts
mv ../gofer_*.deb artifacts/ || true
mv ../gofer_*.buildinfo artifacts/ || true
mv ../gofer_*.changes artifacts/ || true
mv ../gofer-dbgsym_*.deb artifacts/ || true

echo "gofer is go"