#!/bin/sh
set -eu

repo=${LAGO_INSTALL_REPOSITORY:-getlago/lago-cli}
version=${LAGO_INSTALL_VERSION:-latest}
install_dir=${LAGO_INSTALL_DIR:-/usr/local/bin}

case $(uname -s) in
  Darwin) os=darwin ;;
  Linux) os=linux ;;
  *) echo "lago installer: unsupported operating system" >&2; exit 1 ;;
esac

case $(uname -m) in
  x86_64|amd64) arch=amd64 ;;
  arm64|aarch64) arch=arm64 ;;
  *) echo "lago installer: unsupported architecture" >&2; exit 1 ;;
esac

if [ "$version" = latest ]; then
  version=$(curl -fsSL --proto '=https' --tlsv1.2 "https://api.github.com/repos/$repo/releases/latest" | sed -n 's/.*"tag_name":[[:space:]]*"v\{0,1\}\([^"]*\)".*/\1/p' | head -1)
fi
[ -n "$version" ] || { echo "lago installer: could not resolve a release version" >&2; exit 1; }
version=${version#v}

archive="lago_${version}_${os}_${arch}.tar.gz"
base="https://github.com/$repo/releases/download/v${version}"
tmp_dir=$(mktemp -d)
trap 'rm -rf "$tmp_dir"' EXIT HUP INT TERM

curl -fsSL --proto '=https' --tlsv1.2 "$base/$archive" -o "$tmp_dir/$archive"
curl -fsSL --proto '=https' --tlsv1.2 "$base/checksums.txt" -o "$tmp_dir/checksums.txt"
expected=$(awk -v file="$archive" '$2 == file { print $1 }' "$tmp_dir/checksums.txt")
[ -n "$expected" ] || { echo "lago installer: release checksum is missing" >&2; exit 1; }

if command -v sha256sum >/dev/null 2>&1; then
  actual=$(sha256sum "$tmp_dir/$archive" | awk '{print $1}')
else
  actual=$(shasum -a 256 "$tmp_dir/$archive" | awk '{print $1}')
fi
[ "$actual" = "$expected" ] || { echo "lago installer: checksum verification failed" >&2; exit 1; }

tar -xzf "$tmp_dir/$archive" -C "$tmp_dir" lago
mkdir -p "$install_dir"
if [ -w "$install_dir" ]; then
  install -m 0755 "$tmp_dir/lago" "$install_dir/lago"
else
  sudo install -m 0755 "$tmp_dir/lago" "$install_dir/lago"
fi
echo "Installed lago $version to $install_dir/lago"
