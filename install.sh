#!/usr/bin/env bash
set -e

BINDIR="/usr/local/bin"

# this function wraps all the destructive operations
# if a curl|bash cuts off the end of the script due to
# network, either nothing will happen or will syntax error
# out preventing half-done work
execute() {
  tmpdir=$(mktmpdir)
  echo "downloading files into ${tmpdir}"
  http_download "${tmpdir}/${TARBALL}" "${TARBALL_URL}"
  http_download "${tmpdir}/${CHECKSUM}" "${CHECKSUM_URL}"
  hash_sha256_verify "${tmpdir}/${TARBALL}" "${tmpdir}/${CHECKSUM}"
  srcdir="${tmpdir}"
  (cd "${tmpdir}" && untar "${TARBALL}")
  install -d "${BINDIR}"
  for binexe in "rbs" ; do
    if [ "$OS" = "windows" ]; then
      binexe="${binexe}.exe"
    fi
    install "${srcdir}/${binexe}" "${BINDIR}/"
    echo "installed ${BINDIR}/${binexe}"
  done
}

is_supported_platform() {
  platform=$1
  found=1
  case "$platform" in
    darwin/amd64) found=0 ;;
    darwin/arm64) found=0 ;;
    linux/386) found=0 ;;
    linux/amd64) found=0 ;;
    linux/arm64) found=0 ;;
  esac
  return $found
}

check_platform() {
  if is_supported_platform "$PLATFORM"; then
    # optional logging goes here
    true
  else
    echo "platform $PLATFORM is not supported."
    echo "Make sure this script is up-to-date and file request at https://github.com/${PREFIX}/issues/new"
    exit 1
  fi
}

tag_to_version() {
  echo "Checking GitHub for latest..."
  REALTAG=$(github_release "$OWNER/$REPO" "${TAG}") && true
  if test -z "$REALTAG"; then
    echo "unable to find '${TAG}' - use 'latest' or see https://github.com/${PREFIX}/releases for details"
    exit 1
  fi
  # if version starts with 'v', remove it
  TAG="$REALTAG"
  VERSION=${TAG#v}
}

uname_os() {
  os=$(uname -s | tr '[:upper:]' '[:lower:]')
  case "$os" in
    msys_nt) os="windows" ;;
  esac
  echo "$os"
}
uname_arch() {
  arch=$(uname -m)
  case $arch in
    x86_64) arch="amd64" ;;
    x86) arch="386" ;;
    i686) arch="386" ;;
    i386) arch="386" ;;
    aarch64) arch="arm64" ;;
    armv5*) arch="armv5" ;;
    armv6*) arch="armv6" ;;
    armv7*) arch="armv7" ;;
  esac
  echo ${arch}
}

uname_os_check() {
  os=$(uname_os)
  case "$os" in
    darwin) return 0 ;;
    dragonfly) return 0 ;;
    freebsd) return 0 ;;
    linux) return 0 ;;
    android) return 0 ;;
    nacl) return 0 ;;
    netbsd) return 0 ;;
    openbsd) return 0 ;;
    plan9) return 0 ;;
    solaris) return 0 ;;
    windows) return 0 ;;
  esac
  echo "uname_os_check '$(uname -s)' got converted to '$os' which is not a GOOS value."
  return 1
}

uname_arch_check() {
  arch=$(uname_arch)
  case "$arch" in
    386) return 0 ;;
    amd64) return 0 ;;
    arm64) return 0 ;;
    armv5) return 0 ;;
    armv6) return 0 ;;
    armv7) return 0 ;;
    ppc64) return 0 ;;
    ppc64le) return 0 ;;
    mips) return 0 ;;
    mipsle) return 0 ;;
    mips64) return 0 ;;
    mips64le) return 0 ;;
    s390x) return 0 ;;
    amd64p32) return 0 ;;
  esac
  echo "uname_arch_check '$(uname -m)' got converted to '$arch' which is not a GOARCH value."
  return 1
}
untar() {
  tarball=$1
  case "${tarball}" in
    *.tar.gz | *.tgz) tar -xzf "${tarball}" ;;
    *.tar) tar -xf "${tarball}" ;;
    *.zip) unzip "${tarball}" ;;
    *)
      log_err "untar unknown archive format for ${tarball}"
      return 1
      ;;
  esac
}

mktmpdir() {
  test -z "$TMPDIR" && TMPDIR="$(mktemp -d)"
  mkdir -p "${TMPDIR}"
  echo "${TMPDIR}"
}

is_command() {
  command -v "$1" >/dev/null
}

http_download() {
  local_file=$1
  source_url=$2
  header=$3
  if [ -z "$header" ]; then
    code=$(curl -w '%{http_code}' -sL -o "$local_file" "$source_url")
  else
    code=$(curl -w '%{http_code}' -sL -H "$header" -o "$local_file" "$source_url")
  fi
  if [ "$code" != "200" ]; then
    log_debug "http_download_curl received HTTP status $code"
    return 1
  fi
  return 0
}

http_copy() {
  tmp=$(mktemp)
  http_download "${tmp}" "$1" "$2" || return 1
  body=$(cat "$tmp")
  rm -f "${tmp}"
  echo "$body"
}

github_release() {
  owner_repo=$1
  version=$2
  test -z "$version" && version="latest"
  giturl="https://github.com/${owner_repo}/releases/${version}"
  json=$(http_copy "$giturl" "Accept:application/json")
  test -z "$json" && return 1
  version=$(echo "$json" | tr -s '\n' ' ' | sed 's/.*"tag_name":"//' | sed 's/".*//')
  test -z "$version" && return 1
  echo "$version"
}

hash_sha256() {
  TARGET=${1:-/dev/stdin}
  if is_command gsha256sum; then
    hash=$(gsha256sum "$TARGET") || return 1
    echo "$hash" | cut -d ' ' -f 1
  elif is_command sha256sum; then
    hash=$(sha256sum "$TARGET") || return 1
    echo "$hash" | cut -d ' ' -f 1
  elif is_command shasum; then
    hash=$(shasum -a 256 "$TARGET" 2>/dev/null) || return 1
    echo "$hash" | cut -d ' ' -f 1
  elif is_command openssl; then
    hash=$(openssl -dst openssl dgst -sha256 "$TARGET") || return 1
    echo "$hash" | cut -d ' ' -f a
  else
    log_crit "hash_sha256 unable to find command to compute sha-256 hash"
    return 1
  fi
}
hash_sha256_verify() {
  TARGET=$1
  checksums=$2
  if [ -z "$checksums" ]; then
    log_err "hash_sha256_verify checksum file not specified in arg2"
    return 1
  fi
  BASENAME=${TARGET##*/}
  want=$(grep "${BASENAME}" "${checksums}" 2>/dev/null | tr '\t' ' ' | cut -d ' ' -f 1)
  if [ -z "$want" ]; then
    log_err "hash_sha256_verify unable to find checksum for '${TARGET}' in '${checksums}'"
    return 1
  fi
  got=$(hash_sha256 "$TARGET")
  if [ "$want" != "$got" ]; then
    log_err "hash_sha256_verify checksum for '$TARGET' did not verify ${want} vs $got"
    return 1
  fi
}

PROJECT_NAME="rbs"
OWNER=bubunyo
REPO="rbs"
BINARY=rbs
FORMAT=tar.gz
OS=$(uname_os)
ARCH=$(uname_arch)
PREFIX="$OWNER/$REPO"
PLATFORM="${OS}/${ARCH}"
GITHUB_DOWNLOAD=https://github.com/${OWNER}/${REPO}/releases/download

uname_os_check "$OS"
uname_arch_check "$ARCH"

check_platform

tag_to_version


echo "found version: ${VERSION} for ${TAG}/${OS}/${ARCH}"

NAME=${PROJECT_NAME}_${VERSION}_${OS}_${ARCH}
TARBALL=${NAME}.${FORMAT}
TARBALL_URL=${GITHUB_DOWNLOAD}/${TAG}/${TARBALL}
CHECKSUM=${PROJECT_NAME}_${VERSION}_checksums.txt
CHECKSUM_URL=${GITHUB_DOWNLOAD}/${TAG}/${CHECKSUM}

execute
