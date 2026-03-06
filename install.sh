#!/usr/bin/env sh
set -e

REPO="danielmrdev/dtasks-cli"
BINARY="dtasks"

# ── Detect OS ────────────────────────────────────────────────────────────────
os="$(uname -s)"
case "$os" in
  Linux)  platform="linux"  ;;
  Darwin) platform="macos"  ;;
  *)
    echo "Unsupported OS: $os"
    exit 1
    ;;
esac

# ── Detect arch ──────────────────────────────────────────────────────────────
arch="$(uname -m)"
case "$arch" in
  x86_64)          arch="amd64" ;;
  arm64 | aarch64) arch="arm64" ;;
  *)
    echo "Unsupported architecture: $arch"
    exit 1
    ;;
esac

asset="${BINARY}-${platform}-${arch}"

# ── Resolve latest version ───────────────────────────────────────────────────
if command -v curl >/dev/null 2>&1; then
  fetch() { curl -fsSL "$1"; }
  download() { curl -fsSL -o "$2" "$1"; }
elif command -v wget >/dev/null 2>&1; then
  fetch() { wget -qO- "$1"; }
  download() { wget -qO "$2" "$1"; }
else
  echo "curl or wget is required"
  exit 1
fi

api="https://api.github.com/repos/${REPO}/releases/latest"
version="$(fetch "$api" | grep '"tag_name"' | sed 's/.*"tag_name": *"\(.*\)".*/\1/')"

if [ -z "$version" ]; then
  echo "Could not determine latest version"
  exit 1
fi

url="https://github.com/${REPO}/releases/download/${version}/${asset}"

# ── Pick install dir ─────────────────────────────────────────────────────────
if [ -w /usr/local/bin ]; then
  install_dir="/usr/local/bin"
elif [ "$(id -u)" -eq 0 ]; then
  install_dir="/usr/local/bin"
else
  install_dir="${HOME}/.local/bin"
  mkdir -p "$install_dir"
fi

dest="${install_dir}/${BINARY}"

# ── Download & install ───────────────────────────────────────────────────────
echo "Downloading dtasks ${version} (${platform}/${arch})…"
tmp="$(mktemp)"
download "$url" "$tmp"
chmod +x "$tmp"

# Verify checksum if shasum/sha256sum is available
checksum_url="https://github.com/${REPO}/releases/download/${version}/checksums.txt"
if command -v sha256sum >/dev/null 2>&1 || command -v shasum >/dev/null 2>&1; then
  echo "Verifying checksum…"
  checksums="$(fetch "$checksum_url")"
  expected="$(echo "$checksums" | grep "$asset" | awk '{print $1}')"
  if [ -n "$expected" ]; then
    if command -v sha256sum >/dev/null 2>&1; then
      actual="$(sha256sum "$tmp" | awk '{print $1}')"
    else
      actual="$(shasum -a 256 "$tmp" | awk '{print $1}')"
    fi
    if [ "$actual" != "$expected" ]; then
      echo "Checksum mismatch — aborting"
      rm -f "$tmp"
      exit 1
    fi
    echo "Checksum OK"
  fi
fi

mv "$tmp" "$dest"
echo "Installed dtasks ${version} → ${dest}"

# ── PATH hint ────────────────────────────────────────────────────────────────
case ":${PATH}:" in
  *":${install_dir}:"*) ;;
  *)
    echo ""
    echo "Add ${install_dir} to your PATH:"
    echo "  export PATH=\"\$PATH:${install_dir}\""
    ;;
esac

# ── Shell completions ─────────────────────────────────────────────────────────
install_completions() {
    # Skip in non-interactive (pipe/CI) environments
    [ -t 0 ] || return 0

    # Detect shell from $SHELL env var
    shell_name="$(basename "${SHELL:-}")"
    if [ -z "$shell_name" ]; then
        return 0
    fi

    printf "Install shell completions for %s? [y/N] " "$shell_name"
    read -r answer
    case "$answer" in
        [Yy]*) ;;
        *) return 0 ;;
    esac

    case "$shell_name" in
        bash)
            comp_dir="${HOME}/.local/share/bash-completion/completions"
            mkdir -p "$comp_dir"
            "${install_dir}/${BINARY}" completion bash > "${comp_dir}/${BINARY}"
            echo "Completions installed for bash: ${comp_dir}/${BINARY}"
            ;;
        zsh)
            comp_dir="${HOME}/.zsh/completions"
            mkdir -p "$comp_dir"
            "${install_dir}/${BINARY}" completion zsh > "${comp_dir}/_${BINARY}"
            echo "Completions installed for zsh: ${comp_dir}/_${BINARY}"
            echo "Ensure your ~/.zshrc contains: fpath=(~/.zsh/completions \$fpath) && autoload -U compinit && compinit"
            ;;
        fish)
            comp_dir="${HOME}/.config/fish/completions"
            mkdir -p "$comp_dir"
            "${install_dir}/${BINARY}" completion fish > "${comp_dir}/${BINARY}.fish"
            echo "Completions installed for fish: ${comp_dir}/${BINARY}.fish"
            ;;
        *)
            echo "Shell '${shell_name}' not supported for auto-install. Run 'dtasks completion --help' to install manually."
            ;;
    esac
}

install_completions
