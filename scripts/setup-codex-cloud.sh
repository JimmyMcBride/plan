#!/bin/sh

set -eu

PROJECT_DIR="${1:-.}"
SCRIPT_DIR="$(CDPATH= cd -- "$(dirname "$0")" && pwd)"
ROOT_DIR="$(CDPATH= cd -- "${SCRIPT_DIR}/.." && pwd)"
PROJECT_DIR_ABS="$(CDPATH= cd -- "${PROJECT_DIR}" && pwd)"
BIN_DIR="${ROOT_DIR}/.codex/bin"
BRAIN_BIN="${BIN_DIR}/brain"
BRAIN_INSTALLER_URL="https://raw.githubusercontent.com/JimmyMcBride/brain/main/scripts/install.sh"

die() {
  printf 'setup-codex-cloud: %s\n' "$1" >&2
  exit 1
}

need_cmd() {
  command -v "$1" >/dev/null 2>&1
}

fetch() {
  url="$1"
  out="$2"

  if need_cmd curl; then
    curl -fsSL "$url" -o "$out"
    return
  fi

  if need_cmd wget; then
    wget -qO "$out" "$url"
    return
  fi

  die "need curl or wget to download Brain installer"
}

mkdir -p "${BIN_DIR}"

TMPDIR="$(mktemp -d)"
trap 'rm -rf "${TMPDIR}"' EXIT INT TERM

INSTALLER_PATH="${TMPDIR}/brain-install.sh"
fetch "${BRAIN_INSTALLER_URL}" "${INSTALLER_PATH}"

printf 'Installing Brain into %s\n' "${BIN_DIR}"
BRAIN_INSTALL_DIR="${BIN_DIR}" sh "${INSTALLER_PATH}"

[ -x "${BRAIN_BIN}" ] || die "Brain binary not found at ${BRAIN_BIN} after install"

printf 'Installing repo-local Brain skill for Codex\n'
"${BRAIN_BIN}" skills install --scope local --agent codex --project "${PROJECT_DIR_ABS}"

printf 'Verifying Brain install\n'
"${BRAIN_BIN}" --help >/dev/null
"${BRAIN_BIN}" skills targets --scope local --agent codex --project "${PROJECT_DIR_ABS}" >/dev/null

if [ -d "${PROJECT_DIR_ABS}/.brain" ]; then
  printf 'Brain workspace detected at %s/.brain\n' "${PROJECT_DIR_ABS}"
else
  printf 'No .brain workspace detected. Brain installed as optional context tool only.\n'
fi

printf 'Codex cloud setup complete\n'
