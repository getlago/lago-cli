#!/bin/sh
# Lago CLI installer.
#
#   curl -fsSL https://getlago.github.io/lago-cli/install.sh | sh
#
# Downloads the prebuilt release archive for this machine from GitHub Releases,
# verifies its SHA-256 against the release's checksums.txt (and the cosign
# signature over that file when cosign is installed), and installs the `lago`
# binary. No Go toolchain is needed.
#
# Environment:
#   LAGO_INSTALL_VERSION  release to install, with or without the leading v
#                         (default: the latest stable release)
#   LAGO_INSTALL_DIR      directory to install into (default: /usr/local/bin)
#
# The script is published from the getlago/lago-cli repository by the
# .github/workflows/pages.yml workflow and smoke-tested on every release.
set -eu

main() {
  # Only Lago-owned repositories may be installed from. Without this the override
  # turned `curl … | sh` into an arbitrary-binary installer for anyone able to set
  # one environment variable, for example inside a compromised CI job.
  repo=${LAGO_INSTALL_REPOSITORY:-getlago/lago-cli}
  case $repo in
    getlago/*) ;;
    *) fail "refusing to install from $repo; only getlago/* is allowed" ;;
  esac
  version=${LAGO_INSTALL_VERSION:-latest}
  install_dir=${LAGO_INSTALL_DIR:-/usr/local/bin}

  need curl
  need tar

  case $(uname -s) in
    Darwin) os=darwin ;;
    Linux) os=linux ;;
    MINGW*|MSYS*|CYGWIN*|Windows_NT)
      fail "Windows is not supported by this installer. Use 'go install github.com/getlago/lago-cli/cmd/lago@latest' or the zip archive from https://github.com/$repo/releases" ;;
    *) fail "unsupported operating system: $(uname -s)" ;;
  esac

  case $(uname -m) in
    x86_64|amd64) arch=amd64 ;;
    arm64|aarch64) arch=arm64 ;;
    *) fail "unsupported architecture: $(uname -m)" ;;
  esac

  if [ "$version" = latest ]; then
    version=$(resolve_latest "$repo")
  fi
  version=${version#v}
  [ -n "$version" ] || fail "could not resolve a release version"

  archive="lago_${version}_${os}_${arch}.tar.gz"
  base="https://github.com/$repo/releases/download/v${version}"
  tmp_dir=$(mktemp -d)
  trap 'rm -rf "$tmp_dir"' EXIT HUP INT TERM

  say "Downloading lago $version for $os/$arch"
  download "$base/$archive" "$tmp_dir/$archive" || fail "no release archive at $base/$archive; check the version and https://github.com/$repo/releases"
  download "$base/checksums.txt" "$tmp_dir/checksums.txt" || fail "release v$version publishes no checksums.txt"

  verify_checksum "$tmp_dir/checksums.txt" "$tmp_dir/$archive" "$archive"
  verify_signature "$base" "$tmp_dir" "$repo"

  tar -xzf "$tmp_dir/$archive" -C "$tmp_dir" lago
  install_binary "$tmp_dir/lago" "$install_dir"

  say "Installed lago $version to $install_dir/lago"
  case ":$PATH:" in
    *":$install_dir:"*) ;;
    *) say "note: $install_dir is not on your PATH; add it to your shell profile to run 'lago' directly" ;;
  esac
  say "Run 'lago init' to connect it to your Lago organization. Shell completions: 'lago completion --help'."
}

# resolve_latest follows the GitHub "latest release" redirect instead of calling the
# REST API: the redirect is not subject to the 60 requests/hour unauthenticated API
# limit that CI runners and shared offices hit. /releases/latest never points at a
# prerelease, so `latest` is always a stable release.
resolve_latest() {
  location=$(curl -fsSLI --proto '=https' --tlsv1.2 -o /dev/null -w '%{url_effective}' "https://github.com/$1/releases/latest") ||
    fail "could not reach https://github.com/$1/releases/latest"
  tag=${location##*/}
  case $tag in
    v[0-9]*) printf '%s\n' "$tag" ;;
    *) fail "no published release found for $1" ;;
  esac
}

download() {
  curl -fsSL --proto '=https' --tlsv1.2 --retry 3 "$1" -o "$2"
}

verify_checksum() {
  checksums=$1 file=$2 name=$3
  expected=$(awk -v file="$name" '$2 == file { print $1 }' "$checksums")
  [ -n "$expected" ] || fail "checksums.txt has no entry for $name"
  if command -v sha256sum >/dev/null 2>&1; then
    actual=$(sha256sum "$file" | awk '{print $1}')
  elif command -v shasum >/dev/null 2>&1; then
    actual=$(shasum -a 256 "$file" | awk '{print $1}')
  else
    fail "neither sha256sum nor shasum is available to verify the download"
  fi
  [ "$actual" = "$expected" ] || fail "checksum verification failed for $name"
}

# The checksum file travels the same path as the archive, so on its own it proves
# integrity, not authenticity: whoever can serve a bad archive can serve a matching
# checksums.txt. The release pipeline cosign-signs checksums.txt with the identity of
# the release workflow, so when cosign is installed the signature is verified against
# that exact identity, and the script says plainly when it cannot.
verify_signature() {
  base=$1 dir=$2 repo=$3
  if ! command -v cosign >/dev/null 2>&1; then
    say "cosign not found: verified the checksum only, not the release signature (install cosign to verify it)"
    return 0
  fi
  if ! download "$base/checksums.txt.sig" "$dir/checksums.txt.sig" ||
     ! download "$base/checksums.txt.pem" "$dir/checksums.txt.pem"; then
    fail "release publishes no signature for checksums.txt"
  fi
  if ! output=$(cosign verify-blob \
      --certificate "$dir/checksums.txt.pem" \
      --signature "$dir/checksums.txt.sig" \
      --certificate-identity-regexp "^https://github.com/$repo/.github/workflows/release.yml@refs/tags/v" \
      --certificate-oidc-issuer https://token.actions.githubusercontent.com \
      "$dir/checksums.txt" 2>&1); then
    printf '%s\n' "$output" >&2
    fail "signature verification failed: checksums.txt was not signed by the $repo release workflow"
  fi
  say "Verified release signature"
}

install_binary() {
  binary=$1 dir=$2
  if [ -d "$dir" ] && [ -w "$dir" ]; then
    install -m 0755 "$binary" "$dir/lago"
  elif [ ! -e "$dir" ] && [ -w "$(dirname "$dir")" ]; then
    mkdir -p "$dir"
    install -m 0755 "$binary" "$dir/lago"
  elif command -v sudo >/dev/null 2>&1; then
    say "Installing to $dir requires sudo"
    sudo mkdir -p "$dir"
    sudo install -m 0755 "$binary" "$dir/lago"
  else
    fail "$dir is not writable and sudo is not available; set LAGO_INSTALL_DIR to a directory you own, for example LAGO_INSTALL_DIR=\$HOME/.local/bin"
  fi
}

need() {
  command -v "$1" >/dev/null 2>&1 || fail "$1 is required but not installed"
}

say() {
  printf 'lago installer: %s\n' "$1" >&2
}

fail() {
  printf 'lago installer: %s\n' "$1" >&2
  exit 1
}

# Everything runs from main so a partially downloaded script does nothing.
main "$@"
