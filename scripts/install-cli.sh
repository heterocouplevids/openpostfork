#!/bin/sh
set -eu

REPO="rodrgds/openpost"
BIN_NAME="openpost"

fail() {
  printf 'openpost CLI install failed: %s\n' "$1" >&2
  exit 1
}

need() {
  command -v "$1" >/dev/null 2>&1 || fail "missing required command: $1"
}

detect_os() {
  case "$(uname -s)" in
    Linux) printf '%s\n' linux ;;
    Darwin) printf '%s\n' darwin ;;
    *) fail "unsupported platform: $(uname -s)" ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64 | amd64) printf '%s\n' amd64 ;;
    arm64 | aarch64) printf '%s\n' arm64 ;;
    *) fail "unsupported architecture: $(uname -m)" ;;
  esac
}

install_file() {
  src=$1
  dst=$2

  if [ "$(id -u)" -eq 0 ]; then
    mkdir -p "$(dirname "$dst")"
    cp "$src" "$dst"
    chmod 755 "$dst"
    return
  fi

  if mkdir -p "$(dirname "$dst")" 2>/dev/null && cp "$src" "$dst" 2>/dev/null; then
    chmod 755 "$dst"
    return
  fi

  if command -v sudo >/dev/null 2>&1; then
    sudo mkdir -p "$(dirname "$dst")"
    sudo cp "$src" "$dst"
    sudo chmod 755 "$dst"
    return
  fi

  fail "cannot write $dst and sudo is not available"
}

need curl
need uname
need mktemp

OS=$(detect_os)
ARCH=$(detect_arch)
URL="https://github.com/${REPO}/releases/latest/download/openpost-cli-${OS}-${ARCH}"

if [ "$(id -u)" -eq 0 ]; then
  INSTALL_DIR="/usr/local/bin"
else
  INSTALL_DIR="${HOME:-}/.local/bin"
  [ -n "$INSTALL_DIR" ] || fail "HOME is not set"
fi

TMP_FILE=$(mktemp "${TMPDIR:-/tmp}/openpost-cli.XXXXXX")
trap 'rm -f "$TMP_FILE"' EXIT HUP INT TERM

printf 'Downloading %s\n' "$URL"
if ! curl -fsSL "$URL" -o "$TMP_FILE"; then
  fail "download failed for ${URL}"
fi

if [ "$(id -u)" -ne 0 ] && ! mkdir -p "$INSTALL_DIR" 2>/dev/null; then
  INSTALL_DIR="/usr/local/bin"
fi

install_file "$TMP_FILE" "${INSTALL_DIR}/${BIN_NAME}"

printf 'Installed OpenPost CLI to %s\n' "${INSTALL_DIR}/${BIN_NAME}"
printf 'Now run: openpost auth login <instance>\n'
