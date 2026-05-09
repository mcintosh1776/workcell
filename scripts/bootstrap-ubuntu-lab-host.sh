#!/usr/bin/env bash
set -euo pipefail

if [[ "${1:-}" != "--yes" ]]; then
  cat <<'EOF'
usage:
  scripts/bootstrap-ubuntu-lab-host.sh --yes

Installs baseline Workcell lab-host packages on Ubuntu.

This script is intended for a disposable Hetzner VPS such as workcell-lab-001.
Do not run it on Gondor or a workstation you are trying to keep clean.
EOF
  exit 64
fi

if [[ "$(id -u)" -eq 0 ]]; then
  SUDO=""
else
  SUDO="sudo"
fi

if [[ ! -f /etc/os-release ]]; then
  echo "unsupported_host: missing /etc/os-release" >&2
  exit 1
fi

# shellcheck disable=SC1091
source /etc/os-release
if [[ "${ID:-}" != "ubuntu" ]]; then
  echo "unsupported_host: expected ubuntu, got ${ID:-unknown}" >&2
  exit 1
fi

$SUDO apt update
$SUDO apt install -y \
  build-essential \
  ca-certificates \
  curl \
  git \
  jq \
  podman

cat <<'EOF'
workcell_lab_base_packages=ok

Next manual steps:
  1. Install Go 1.22 or newer.
  2. Install and initialize Incus for containers.
  3. Clone https://github.com/mcintosh1776/workcell.git.
  4. Run scripts/lab-host-preflight.sh from the Workcell checkout.

This script intentionally does not expose any Workcell API port.
EOF
