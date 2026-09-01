#!/bin/sh
set -eu

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
APPDIR="$ROOT/dist/JiroLauncher.AppDir"

rm -rf "$APPDIR" "$ROOT/dist"/*.AppImage
mkdir -p "$APPDIR/usr/bin" "$APPDIR/usr/share/applications"

GOOS=linux GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o "$APPDIR/usr/bin/jirolauncher" "$ROOT/cmd/jirolauncher"
cp "$ROOT/packaging/AppRun" "$APPDIR/AppRun"
cp "$ROOT/packaging/jirolauncher.desktop" "$APPDIR/usr/share/applications/jirolauncher.desktop"
chmod +x "$APPDIR/AppRun" "$APPDIR/usr/bin/jirolauncher"

if command -v linuxdeploy >/dev/null 2>&1; then
	NO_STRIP=1 linuxdeploy \
		--appdir "$APPDIR" \
		--desktop-file "$ROOT/packaging/jirolauncher.desktop" \
		--icon-file "$ROOT/packaging/jirolauncher.png" \
		--icon-filename jirolauncher \
		--output appimage
	mv "$ROOT/JiroLauncher-x86_64.AppImage" "$ROOT/dist/"
else
	echo "linuxdeploy is required to build the AppImage" >&2
	exit 1
fi
