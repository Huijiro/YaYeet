#!/bin/sh
set -eu

if [ "$#" -ne 1 ]; then
	echo "usage: $0 VERSION" >&2
	exit 2
fi
if ! command -v dpkg-deb >/dev/null 2>&1; then
	echo "dpkg-deb is required to build the Debian package" >&2
	exit 1
fi

VERSION=${1#v}
case "$VERSION" in
	''|*[!0-9A-Za-z.+:~_-]*)
		echo "invalid Debian package version: $VERSION" >&2
		exit 2
		;;
esac

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
PACKAGE_ROOT="$ROOT/dist/deb"
PACKAGE_PATH="$ROOT/dist/yayeet_${VERSION}_amd64.deb"

rm -rf "$PACKAGE_ROOT" "$PACKAGE_PATH"
mkdir -p \
	"$PACKAGE_ROOT/DEBIAN" \
	"$PACKAGE_ROOT/usr/bin" \
	"$PACKAGE_ROOT/usr/share/applications" \
	"$PACKAGE_ROOT/usr/share/icons/hicolor/64x64/apps" \
	"$PACKAGE_ROOT/usr/share/doc/yayeet"

GOOS=linux GOARCH=amd64 go build \
	-trimpath \
	-buildmode=pie \
	-ldflags='-linkmode=external -s -w' \
	-o "$PACKAGE_ROOT/usr/bin/yayeet" \
	"$ROOT/cmd/yayeet"

install -m 644 "$ROOT/packaging/yayeet.desktop" \
	"$PACKAGE_ROOT/usr/share/applications/yayeet.desktop"
install -m 644 "$ROOT/packaging/yayeet.png" \
	"$PACKAGE_ROOT/usr/share/icons/hicolor/64x64/apps/yayeet.png"
install -m 644 "$ROOT/LICENSE" \
	"$PACKAGE_ROOT/usr/share/doc/yayeet/copyright"

INSTALLED_SIZE=$(du -sk "$PACKAGE_ROOT/usr" | cut -f1)
cat > "$PACKAGE_ROOT/DEBIAN/control" <<EOF
Package: yayeet
Version: $VERSION
Section: games
Priority: optional
Architecture: amd64
Installed-Size: $INSTALLED_SIZE
Maintainer: Gabriel Rodrigues <huijirohankei@gmail.com>
Depends: libc6, libgl1, libwayland-client0, libwayland-cursor0, libwayland-egl1, libx11-6, libxcursor1, libxi6, libxinerama1, libxkbcommon0, libxrandr2, libxxf86vm1
Suggests: wine
Homepage: https://github.com/Huijiro/YaYeet
Description: Launcher for Voices of the Void
 YaYeet installs selectable game versions and runs them through an available
 Wine or Proton installation.
EOF

dpkg-deb --root-owner-group --build "$PACKAGE_ROOT" "$PACKAGE_PATH"
rm -rf "$PACKAGE_ROOT"
