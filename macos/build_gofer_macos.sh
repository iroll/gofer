#!/usr/bin/env bash
set -euo pipefail

# =========================================================
# build_gofer_macos.sh
# Minimal macOS build stub for Gofer
# AppleScript-first, no Cocoa assumptions
# =========================================================

APP_BUNDLE="gofer.app"
WRAPPER_SCRIPT="GoferMacOSWrapper.scpt"
ENGINE_NAME="gofer"

echo "==> Building Gofer (macOS)"

# --- Step 0: clean slate ---
echo "-> Removing old app bundle"
rm -rf "${APP_BUNDLE}"

# --- Step 1: compile AppleScript wrapper ---
echo "-> Compiling AppleScript wrapper"
osacompile -x -o "${APP_BUNDLE}" "${WRAPPER_SCRIPT}"

ICON_FILE="./AppIcon.icns"
cp "$ICON_FILE" "${APP_BUNDLE}/Contents/Resources/applet.icns"

# --- Step 2: register gopher:// URL scheme handler ---
echo "-> Registering gopher:// URL scheme in Info.plist"
# Create the CFBundleURLTypes array (ignore error if it already exists)
/usr/libexec/PlistBuddy -c "Add :CFBundleURLTypes array" "${APP_BUNDLE}/Contents/Info.plist" 2>/dev/null || true
# Add the URL handler configuration
/usr/libexec/PlistBuddy -c "Add :CFBundleURLTypes:0 dict" "${APP_BUNDLE}/Contents/Info.plist"
/usr/libexec/PlistBuddy -c "Add :CFBundleURLTypes:0:CFBundleURLName string 'Gopher URL'" "${APP_BUNDLE}/Contents/Info.plist"
/usr/libexec/PlistBuddy -c "Add :CFBundleURLTypes:0:CFBundleURLSchemes array" "${APP_BUNDLE}/Contents/Info.plist"
/usr/libexec/PlistBuddy -c "Add :CFBundleURLTypes:0:CFBundleURLSchemes:0 string 'gopher'" "${APP_BUNDLE}/Contents/Info.plist"

# --- Step 3: set app metadata in Info.plist ---
echo "-> Setting app metadata (name, version, copyright, bundle ID)"

# Set bundle identifier first - important for URL handling
/usr/libexec/PlistBuddy -c "Set :CFBundleIdentifier 'com.iroll.gofer'" "${APP_BUNDLE}/Contents/Info.plist" 2>/dev/null || \
  /usr/libexec/PlistBuddy -c "Add :CFBundleIdentifier string 'com.iroll.gofer'" "${APP_BUNDLE}/Contents/Info.plist"

/usr/libexec/PlistBuddy -c "Set :CFBundleName 'gofer'" "${APP_BUNDLE}/Contents/Info.plist" 2>/dev/null || \
  /usr/libexec/PlistBuddy -c "Add :CFBundleName string 'gofer'" "${APP_BUNDLE}/Contents/Info.plist"

/usr/libexec/PlistBuddy -c "Set :CFBundleDisplayName 'gofer'" "${APP_BUNDLE}/Contents/Info.plist" 2>/dev/null || \
  /usr/libexec/PlistBuddy -c "Add :CFBundleDisplayName string 'gofer'" "${APP_BUNDLE}/Contents/Info.plist"

/usr/libexec/PlistBuddy -c "Set :CFBundleShortVersionString '0.9'" "${APP_BUNDLE}/Contents/Info.plist" 2>/dev/null || \
  /usr/libexec/PlistBuddy -c "Add :CFBundleShortVersionString string '0.9'" "${APP_BUNDLE}/Contents/Info.plist"

/usr/libexec/PlistBuddy -c "Set :CFBundleVersion '0.9'" "${APP_BUNDLE}/Contents/Info.plist" 2>/dev/null || \
  /usr/libexec/PlistBuddy -c "Add :CFBundleVersion string '0.9'" "${APP_BUNDLE}/Contents/Info.plist"

/usr/libexec/PlistBuddy -c "Set :NSHumanReadableCopyright 'A Gopher protocol handler for macOS'" "${APP_BUNDLE}/Contents/Info.plist" 2>/dev/null || \
  /usr/libexec/PlistBuddy -c "Add :NSHumanReadableCopyright string 'A Gopher protocol handler for macOS'" "${APP_BUNDLE}/Contents/Info.plist"

# --- Step 3.9: move to gofer-app folder
cd ~/gofer-app  

# --- Step 4.0: build Go engine ---
echo "-> Building Go engine"
GOOS=darwin GOARCH=arm64 go build -o ./macos/"${ENGINE_NAME}" .

# --- Step 4.1 move back to macOS folder ---
cd ~/gofer-app/macOS 

# --- Step 5: install Go engine into bundle ---
# Safe even if wrapper doesn't call it yet
echo "-> Installing Go engine into app bundle"
mkdir -p "${APP_BUNDLE}/Contents/MacOS"
mv -f "${ENGINE_NAME}" "${APP_BUNDLE}/Contents/MacOS/${ENGINE_NAME}"
chmod +x "${APP_BUNDLE}/Contents/MacOS/${ENGINE_NAME}"

# --- Step 6: unregister old version in /Applications ---
echo "-> Unregistering old version from /Applications (if it exists)"
if [ -d "/Applications/${APP_BUNDLE}" ]; then
    /System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister -u /Applications/"${APP_BUNDLE}"
fi

echo "==> Build complete:"
echo "   ${APP_BUNDLE}"