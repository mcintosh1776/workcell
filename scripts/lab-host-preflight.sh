#!/usr/bin/env bash
set -euo pipefail

RUN_BACKEND_SMOKES="${WORKCELL_LAB_PREFLIGHT_RUN_BACKEND_SMOKES:-0}"

require_command() {
  local name="$1"
  if ! command -v "$name" >/dev/null 2>&1; then
    echo "fail: missing command: $name"
    exit 1
  fi
  echo "ok: $name available"
}

require_command git
require_command curl
require_command jq
require_command go
require_command gofmt
require_command podman
require_command incus

go version
incus version
podman --version

if [[ "$RUN_BACKEND_SMOKES" == "1" ]]; then
  podman run --rm docker.io/library/alpine:3.20 echo podman-ok

  incus launch images:ubuntu/24.04 workcell-preflight >/tmp/workcell-incus-launch.out
  trap 'incus delete --force workcell-preflight >/dev/null 2>&1 || true' EXIT
  incus exec workcell-preflight -- echo incus-ok
  incus delete --force workcell-preflight
  trap - EXIT
else
  echo "skip: backend smoke runs disabled; set WORKCELL_LAB_PREFLIGHT_RUN_BACKEND_SMOKES=1 to run Podman and Incus jobs"
fi

go test ./...
scripts/dev-smoke.sh

echo "workcell_lab_preflight=ok"
