#!/usr/bin/env bash
set -euo pipefail

if [[ "${1:-}" != "--yes" ]]; then
  cat <<'EOF'
usage:
  scripts/destroy-hetzner-lab-host.sh --yes

Deletes the disposable Hetzner Workcell lab host by name.

Required environment:
  HCLOUD_TOKEN

Optional environment:
  WORKCELL_LAB_NAME              default: workcell-lab-001
  WORKCELL_LAB_DELETE_FIREWALL   default: 0

Firewall deletion is optional because operators may want to reuse the SSH-only
firewall across rebuilds.
EOF
  exit 64
fi

require_env() {
  local name="$1"
  if [[ -z "${!name:-}" ]]; then
    echo "missing required env: $name" >&2
    exit 1
  fi
}

require_command() {
  local name="$1"
  if ! command -v "$name" >/dev/null 2>&1; then
    echo "missing required command: $name" >&2
    exit 1
  fi
}

require_command curl
require_command jq
require_env HCLOUD_TOKEN

NAME="${WORKCELL_LAB_NAME:-workcell-lab-001}"
FIREWALL_NAME="${NAME}-firewall"
DELETE_FIREWALL="${WORKCELL_LAB_DELETE_FIREWALL:-0}"

hcloud_get() {
  curl -sS -H "Authorization: Bearer ${HCLOUD_TOKEN}" "$1"
}

hcloud_delete() {
  curl -sS -X DELETE -H "Authorization: Bearer ${HCLOUD_TOKEN}" "$1"
}

server_id="$(
  hcloud_get "https://api.hetzner.cloud/v1/servers?name=${NAME}" |
    jq -r '.servers[0].id // empty'
)"

if [[ -z "$server_id" ]]; then
  echo "workcell_lab_server_absent name=${NAME}"
else
  response="$(hcloud_delete "https://api.hetzner.cloud/v1/servers/${server_id}")"
  if [[ "$(jq -r '.error.code // empty' <<<"$response")" != "" ]]; then
    jq -r '"workcell_lab_server_delete_failed code=" + .error.code + " message=" + .error.message' <<<"$response" >&2
    exit 1
  fi
  echo "workcell_lab_server_delete_started id=${server_id} name=${NAME}"
fi

if [[ "$DELETE_FIREWALL" == "1" ]]; then
  firewall_id="$(
    hcloud_get "https://api.hetzner.cloud/v1/firewalls?name=${FIREWALL_NAME}" |
      jq -r '.firewalls[0].id // empty'
  )"
  if [[ -z "$firewall_id" ]]; then
    echo "workcell_lab_firewall_absent name=${FIREWALL_NAME}"
  else
    response="$(hcloud_delete "https://api.hetzner.cloud/v1/firewalls/${firewall_id}")"
    if [[ "$(jq -r '.error.code // empty' <<<"$response")" != "" ]]; then
      jq -r '"workcell_lab_firewall_delete_failed code=" + .error.code + " message=" + .error.message' <<<"$response" >&2
      exit 1
    fi
    echo "workcell_lab_firewall_delete_started id=${firewall_id} name=${FIREWALL_NAME}"
  fi
fi
