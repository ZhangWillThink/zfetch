#!/bin/sh
set -e

add_to_path() {
  dir="$1"
  shell_name="$(basename "${SHELL:-/bin/sh}")"

  case "$shell_name" in
    zsh)  rc="${ZDOTDIR:-$HOME}/.zshrc" ;;
    bash) rc="$HOME/.bashrc" ;;
    fish) rc="$HOME/.config/fish/config.fish" ;;
    *)    rc="$HOME/.profile" ;;
  esac

  if [ "$shell_name" = "fish" ]; then
    line="fish_add_path $dir"
  else
    line="export PATH=\"\$PATH:$dir\""
  fi

  if grep -qF "$line" "$rc" 2>/dev/null; then
    echo "${dir} already in PATH (${rc})"
    return
  fi

  mkdir -p "$(dirname "$rc")"
  printf '\n# Added by zfetch installer\n%s\n' "$line" >> "$rc"
  echo "Added ${dir} to PATH in ${rc}"
  echo "Run 'source ${rc}' or restart your terminal to apply."
}

REPO="ZhangWillThink/zfetch"
VERSION="${ZFETCH_VERSION:-v0.5.1}"
BASE_URL="https://github.com/${REPO}/releases/download/${VERSION}"

case "$(uname -sm)" in
  "Darwin x86_64")  FILE="zfetch-darwin-amd64" ;;
  "Darwin arm64")   FILE="zfetch-darwin-arm64" ;;
  "Linux x86_64")   FILE="zfetch-linux-amd64" ;;
  "Linux aarch64")  FILE="zfetch-linux-arm64" ;;
  *) echo "Unsupported platform: $(uname -sm)"; exit 1 ;;
esac

INSTALL_DIR="${ZFETCH_INSTALL_DIR:-/usr/local/bin}"
echo "Installing zfetch ${VERSION} for $(uname -sm)..."

if [ ! -w "$INSTALL_DIR" ]; then
  INSTALL_DIR="${HOME}/.local/bin"
  mkdir -p "$INSTALL_DIR"
fi

curl -fsSL "${BASE_URL}/${FILE}" -o "${INSTALL_DIR}/zfetch"
chmod +x "${INSTALL_DIR}/zfetch"

echo "zfetch installed to ${INSTALL_DIR}/zfetch"

case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *) add_to_path "$INSTALL_DIR" ;;
esac
