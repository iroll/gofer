#!/bin/bash
set -e

# Ensure we are inside the gofer project root
cd "$(dirname "$0")/.."

# Build package (dpkg will drop outputs in home dir)
dpkg-buildpackage -b

echo "gofer is go"
