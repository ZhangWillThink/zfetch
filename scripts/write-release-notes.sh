#!/bin/bash
# Writes dist/RELEASE_NOTES.md for GitHub Releases (semver-oriented layout).
set -euo pipefail

: "${GITHUB_REPOSITORY:?Set GITHUB_REPOSITORY (e.g. owner/repo)}"
: "${VERSION:?Set VERSION (e.g. v0.6.0)}"

OUT="${1:-dist/RELEASE_NOTES.md}"
mkdir -p "$(dirname "$OUT")"

# Previous semver-like v* tag (exclude current).
PREV=""
while IFS= read -r t; do
	if [[ "$t" != "$VERSION" ]]; then
		PREV="$t"
		break
	fi
done < <(git tag -l 'v*' | sort -V -r || true)

COMPARE=""
if [[ -n "$PREV" ]]; then
	COMPARE="https://github.com/${GITHUB_REPOSITORY}/compare/${PREV}...${VERSION}"
fi

cat >"$OUT" <<EOF
## zfetch ${VERSION}

### Highlights

See [\`CHANGELOG.md\`](https://github.com/${GITHUB_REPOSITORY}/blob/${VERSION}/CHANGELOG.md) for a full categorized list.

### Install

**One-line installer (Linux & macOS):** follows the [\`releases/latest/download/install.sh\`](https://docs.github.com/en/repositories/releasing-projects-on-github/managing-releases-in-a-repository) pattern.

Latest release:

\`\`\`bash
curl -fsSL https://github.com/${GITHUB_REPOSITORY}/releases/latest/download/install.sh | bash
\`\`\`

**Pinned (${VERSION}):**

\`\`\`bash
curl -fsSL https://github.com/${GITHUB_REPOSITORY}/releases/download/${VERSION}/install.sh | bash
curl -fsSL https://github.com/${GITHUB_REPOSITORY}/releases/download/${VERSION}/install.sh | ZFETCH_VERSION=${VERSION} bash
\`\`\`

**Manual binary**

| Platform | Asset |
|:--|:--|
| Linux x86_64 | [\`zfetch-linux-amd64\`](https://github.com/${GITHUB_REPOSITORY}/releases/download/${VERSION}/zfetch-linux-amd64) |
| Linux arm64 | [\`zfetch-linux-arm64\`](https://github.com/${GITHUB_REPOSITORY}/releases/download/${VERSION}/zfetch-linux-arm64) |
| macOS x86_64 | [\`zfetch-darwin-amd64\`](https://github.com/${GITHUB_REPOSITORY}/releases/download/${VERSION}/zfetch-darwin-amd64) |
| macOS arm64 | [\`zfetch-darwin-arm64\`](https://github.com/${GITHUB_REPOSITORY}/releases/download/${VERSION}/zfetch-darwin-arm64) |
| Windows x86_64 | [\`zfetch-windows-amd64.exe\`](https://github.com/${GITHUB_REPOSITORY}/releases/download/${VERSION}/zfetch-windows-amd64.exe) |

After download (Unix-like):

\`\`\`bash
chmod +x zfetch-*
sudo mv zfetch-linux-amd64 /usr/local/bin/zfetch   # adjust OS/arch in the filename
\`\`\`

Windows: place \`zfetch-windows-amd64.exe\` on your \`PATH\` (rename to \`zfetch.exe\` if you prefer).

### Upgrade

If you already have zfetch on your \`PATH\`:

\`\`\`bash
zfetch upgrade
\`\`\`

This downloads the matching asset for the current OS/arch from **GitHub Releases** and replaces the running binary. When \`SHA256SUMS\` is attached to the release, integrity is checked before install.

### Uninstall

\`\`\`bash
zfetch uninstall
\`\`\`

Optional (config and presets):

\`\`\`bash
rm -rf ~/.config/zfetch
\`\`\`

### Verify checksums

Release assets include \`SHA256SUMS\` (GNU \`sha256sum\` format).

\`\`\`bash
curl -fsSL -O https://github.com/${GITHUB_REPOSITORY}/releases/download/${VERSION}/SHA256SUMS
curl -fsSL -O https://github.com/${GITHUB_REPOSITORY}/releases/download/${VERSION}/zfetch-linux-amd64
grep ' zfetch-linux-amd64\$' SHA256SUMS | sha256sum -c
\`\`\`

(On systems without GNU \`sha256sum\`/pipe patterns, verify the checksum line manually against \`openssl dgst -sha256 zfetch-*\`.)

EOF

if [[ -n "$COMPARE" ]]; then
	cat >>"$OUT" <<EOF

**Full changelog (compare):** ${COMPARE}
EOF
fi
