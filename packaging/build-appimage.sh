#!/bin/sh
set -eu

if [ "$#" -ne 1 ]; then
	echo "usage: $0 VERSION" >&2
	exit 2
fi

VERSION=${1#v}
case "$VERSION" in
	''|*[!0-9A-Za-z.+:~_-]*)
		echo "invalid AppImage version: $VERSION" >&2
		exit 2
		;;
esac

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
APPDIR="$ROOT/dist/YaYeet.AppDir"

rm -rf "$APPDIR" "$ROOT/dist"/*.AppImage
mkdir -p "$APPDIR/usr/bin" "$APPDIR/usr/share/applications"

GOOS=linux GOARCH=amd64 go build -trimpath \
	-ldflags="-s -w -X github.com/Huijiro/YaYeet/internal/buildinfo.Version=$VERSION -X github.com/Huijiro/YaYeet/internal/buildinfo.InstallMethod=appimage" \
	-o "$APPDIR/usr/bin/yayeet" "$ROOT/cmd/yayeet"
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
