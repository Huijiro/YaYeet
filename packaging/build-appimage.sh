#!/bin/sh
set -eu

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
APPDIR="$ROOT/dist/YaYeet.AppDir"

rm -rf "$APPDIR" "$ROOT/dist"/*.AppImage
mkdir -p "$APPDIR/usr/bin" "$APPDIR/usr/share/applications"

GOOS=linux GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o "$APPDIR/usr/bin/yayeet" "$ROOT/cmd/yayeet"
cp "$ROOT/packaging/AppRun" "$APPDIR/AppRun"
cp "$ROOT/packaging/yayeet.desktop" "$APPDIR/usr/share/applications/yayeet.desktop"
chmod +x "$APPDIR/AppRun" "$APPDIR/usr/bin/yayeet"

if command -v linuxdeploy >/dev/null 2>&1; then
	NO_STRIP=1 linuxdeploy \
		--appdir "$APPDIR" \
		--desktop-file "$ROOT/packaging/yayeet.desktop" \
		--icon-file "$ROOT/packaging/yayeet.png" \
		--icon-filename yayeet \
		--output appimage
	mv "$ROOT/YaYeet-x86_64.AppImage" "$ROOT/dist/"
else
	echo "linuxdeploy is required to build the AppImage" >&2
	exit 1
fi
