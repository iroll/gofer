#!/usr/bin/env bash

# Build script for gofer application on macOS
# A Swift wrapper is used to create a macOS application bundle around the Go engine.
set -euo pipefail

# 0. Clean previous builds
rm -rf gofer.app
rm -f gofer_ui
rm -f gofer
echo "0. Cleaned previous builds."

# 1. Build the Go engine
cd ..
go build -o macos/gofer .
cd macos
echo "1. Gofer Go engine built."

# 2. Compile the Swift Wrapper
# This creates a binary named 'gofer_ui'
swiftc main.swift -o gofer_ui
echo "2. Swift UI wrapper built."

# 3. Create the Bundle structure
mkdir -p gofer.app/Contents/MacOS
mkdir -p gofer.app/Contents/Resources
echo "3. App bundle structure created."

# 4. Move binaries into place
# Note: In Swift apps, the 'main' binary is our UI, and 'gofer' is the helper
cp gofer_ui gofer.app/Contents/MacOS/gofer_ui
cp gofer gofer.app/Contents/MacOS/gofer
cp Info.plist gofer.app/Contents/Info.plist
cp AppIcon.icns gofer.app/Contents/Resources/AppIcon.icns
echo "4. Binaries and Info.plist copied into app bundle."

# 5. Clean up again
rm -f gofer_ui
rm -f gofer
echo "5. Temporary build files removed."

# 6. Sign binaries individually, then the bundle
codesign --force --sign - gofer.app/Contents/MacOS/gofer
codesign --force --sign - gofer.app/Contents/MacOS/gofer_ui
codesign --force --sign - gofer.app
echo "6. App signed."

# 7. Unregister old version if it exists in /Applications
if [ -d "/Applications/gofer.app" ]; then
    /System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister -u /Applications/gofer.app
fi
echo "7. Old app unregistered if it existed."

# 8. Register the new build
/System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister -f gofer.app
echo "8. New app registered with Launch Services."
