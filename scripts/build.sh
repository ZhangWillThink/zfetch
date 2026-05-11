#!/bin/bash
set -e

BIN="zfetch"
DIST="dist"

VERSION="${ZFETCH_VERSION:-}"
if [ -z "$VERSION" ]; then
	if v="$(git describe --tags --exact-match 2>/dev/null)"; then
		VERSION="$v"
	else
		VERSION="v0.0.0-dev+$(git rev-parse --short HEAD 2>/dev/null || echo unknown)"
	fi
fi

LDFLAGS="-s -w -X github.com/WillZhang/zfetch/internal/upgrade.CurrentVersion=${VERSION}"

rm -rf "$DIST"
mkdir -p "$DIST"

build() {
	local os="$1"
	local arch="$2"
	local file="${BIN}-${os}-${arch}"

	echo "Building $file... (${VERSION})"

	local ext=""
	if [ "$os" = "windows" ]; then
		ext=".exe"
	fi

	CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build -trimpath -ldflags "$LDFLAGS" -o "${DIST}/${file}${ext}" .
}

build linux amd64
build linux arm64
build darwin amd64
build darwin arm64
build windows amd64

echo ""
echo "Done. Binaries (${VERSION}):"
ls -lh "$DIST"
