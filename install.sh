#!/bin/sh

set -eu

repository="futrx-com/remote.futrx-cli"
version=${REMOTE_VERSION:-latest}
release_base_url=${REMOTE_RELEASE_BASE_URL:-"https://github.com/${repository}/releases"}

fail() {
	printf 'remote installer: %s\n' "$1" >&2
	exit 1
}

require_command() {
	command -v "$1" >/dev/null 2>&1 || fail "$1 is required"
}

download() {
	url=$1
	destination=$2
	case "$url" in
	https://*)
		curl --proto '=https' --tlsv1.2 -fsSL --retry 3 --output "$destination" "$url"
		;;
	*)
		curl -fsSL --retry 3 --output "$destination" "$url"
		;;
	esac
}

require_command curl
require_command install

case "$(uname -s)" in
Linux) os=linux ;;
Darwin) os=darwin ;;
*) fail "unsupported operating system: $(uname -s)" ;;
esac

case "$(uname -m)" in
x86_64 | amd64) arch=amd64 ;;
aarch64 | arm64) arch=arm64 ;;
*) fail "unsupported architecture: $(uname -m)" ;;
esac

case "$version" in
latest)
	download_root="${release_base_url}/latest/download"
	;;
'' | *[!A-Za-z0-9._-]*)
	fail "invalid version: $version"
	;;
*)
	download_root="${release_base_url}/download/${version}"
	;;
esac

if [ -n "${REMOTE_INSTALL_DIR:-}" ]; then
	install_dir=$REMOTE_INSTALL_DIR
else
	[ -n "${HOME:-}" ] || fail "HOME or REMOTE_INSTALL_DIR is required"
	install_dir="${HOME}/.local/bin"
fi

asset="remote-${os}-${arch}"
temporary_dir=$(mktemp -d)
trap 'rm -rf "$temporary_dir"' 0
trap 'exit 1' HUP INT TERM

download "${download_root}/${asset}" "${temporary_dir}/${asset}"
download "${download_root}/checksums.txt" "${temporary_dir}/checksums.txt"

expected_checksum=$(awk -v asset="$asset" '$2 == asset || $2 == "*" asset { print $1; exit }' "${temporary_dir}/checksums.txt")
case "$expected_checksum" in
'' | *[!0-9a-fA-F]*) fail "checksums.txt has no valid checksum for ${asset}" ;;
esac
[ "${#expected_checksum}" -eq 64 ] || fail "checksums.txt has no valid checksum for ${asset}"

if command -v sha256sum >/dev/null 2>&1; then
	actual_checksum=$(sha256sum "${temporary_dir}/${asset}" | awk '{ print $1 }')
elif command -v shasum >/dev/null 2>&1; then
	actual_checksum=$(shasum -a 256 "${temporary_dir}/${asset}" | awk '{ print $1 }')
else
	fail "sha256sum or shasum is required"
fi

[ "$actual_checksum" = "$expected_checksum" ] || fail "checksum verification failed for ${asset}"

mkdir -p "$install_dir"
install -m 0755 "${temporary_dir}/${asset}" "${install_dir}/remote"

printf 'Installed remote to %s\n' "${install_dir}/remote"
case ":${PATH}:" in
*":${install_dir}:"*) ;;
*) printf 'Add %s to PATH to run remote from any directory.\n' "$install_dir" ;;
esac
