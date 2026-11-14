#!/usr/bin/env bash

# Build script for gofer application on macOS
# A Swift wrapper is used to create a macOS application bundle around the Go engine.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$SCRIPT_DIR"

# 0. Clean previous builds
rm -rf "$SCRIPT_DIR/gofer.app"
rm -f "$SCRIPT_DIR/gofer_ui" "$SCRIPT_DIR/gofer"
echo "0. Cleaned previous builds."

# 1. Build the Go engine
(cd "$REPO_ROOT" && go build -o "$SCRIPT_DIR/gofer" .)
echo "1. Gofer Go engine built."

# 2. Compile the Swift Wrapper
# This creates a binary named 'gofer_ui'
swiftc "$SCRIPT_DIR/main.swift" -o "$SCRIPT_DIR/gofer_ui"
echo "2. Swift UI wrapper built."

# 3. Create the Bundle structure
mkdir -p "$SCRIPT_DIR/gofer.app/Contents/MacOS"
mkdir -p "$SCRIPT_DIR/gofer.app/Contents/Resources"
echo "3. App bundle structure created."

# 4. Move binaries into place
# Note: In Swift apps, the 'main' binary is our UI, and 'gofer' is the helper
cp "$SCRIPT_DIR/gofer_ui" "$SCRIPT_DIR/gofer.app/Contents/MacOS/gofer_ui"
cp "$SCRIPT_DIR/gofer" "$SCRIPT_DIR/gofer.app/Contents/MacOS/gofer"
cp "$SCRIPT_DIR/Info.plist" "$SCRIPT_DIR/gofer.app/Contents/Info.plist"
cp "$SCRIPT_DIR/AppIcon.icns" "$SCRIPT_DIR/gofer.app/Contents/Resources/AppIcon.icns"
echo "4. Binaries and Info.plist copied into app bundle."

# 5. Clean up again
rm -f "$SCRIPT_DIR/gofer_ui" "$SCRIPT_DIR/gofer"
echo "5. Temporary build files removed."

# 6. Sign binaries individually, then the bundle
codesign --force --sign - "$SCRIPT_DIR/gofer.app/Contents/MacOS/gofer"
codesign --force --sign - "$SCRIPT_DIR/gofer.app/Contents/MacOS/gofer_ui"
codesign --force --sign - "$SCRIPT_DIR/gofer.app"
echo "6. App signed."

# 7. Package the app bundle for release
ARCH="$(uname -m)"
case "$ARCH" in
    arm64|x86_64) ;;
    *)
        echo "Unsupported macOS architecture: $ARCH" >&2
        exit 1
        ;;
esac
VERSION=$(/usr/libexec/PlistBuddy -c 'Print :CFBundleShortVersionString' "$SCRIPT_DIR/Info.plist")
PACKAGE="$SCRIPT_DIR/gofer-macos-$ARCH-$VERSION.zip"
rm -f "$PACKAGE"
ditto -c -k --keepParent "$SCRIPT_DIR/gofer.app" "$PACKAGE"
echo "7. Release package created: $PACKAGE"
