#!/bin/bash
set -e

BIN="zfetch"
DIST="dist"

rm -rf "$DIST"
mkdir -p "$DIST"

build() {
    local os="$1"
    local arch="$2"
    local file="${BIN}-${os}-${arch}"

    echo "Building $file..."

    if [ "$os" = "linux" ] && [ "$arch" = "arm64" ]; then
        CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build -o "${DIST}/${file}" .
    else
        GOOS="$os" GOARCH="$arch" go build -o "${DIST}/${file}" .
    fi
}

build linux amd64
build linux arm64
build darwin amd64
build darwin arm64

echo ""
echo "Done. Binaries:"
ls -lh "$DIST"
