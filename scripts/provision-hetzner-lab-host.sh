#!/usr/bin/env bash
set -euo pipefail

if [[ "${1:-}" != "--yes" ]]; then
  cat <<'EOF'
usage:
  scripts/provision-hetzner-lab-host.sh --yes

Creates a disposable Hetzner Cloud Workcell lab host and an SSH-only firewall.

Required environment:
  HCLOUD_TOKEN
  WORKCELL_LAB_SSH_KEY_ID
  WORKCELL_LAB_SSH_SOURCE_CIDR

Optional environment:
  WORKCELL_LAB_NAME            default: workcell-lab-001
  WORKCELL_LAB_SERVER_TYPE     default: cx23
  WORKCELL_LAB_IMAGE           default: ubuntu-24.04
  WORKCELL_LAB_LOCATION        default: fsn1

This script intentionally does not expose any Workcell API port.
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
require_env WORKCELL_LAB_SSH_KEY_ID
require_env WORKCELL_LAB_SSH_SOURCE_CIDR

NAME="${WORKCELL_LAB_NAME:-workcell-lab-001}"
SERVER_TYPE="${WORKCELL_LAB_SERVER_TYPE:-cx23}"
IMAGE="${WORKCELL_LAB_IMAGE:-ubuntu-24.04}"
LOCATION="${WORKCELL_LAB_LOCATION:-fsn1}"
FIREWALL_NAME="${NAME}-firewall"

hcloud_get() {
  curl -sS -H "Authorization: Bearer ${HCLOUD_TOKEN}" "$1"
}

hcloud_post() {
  local url="$1"
  local body="$2"
  curl -sS -X POST \
    -H "Authorization: Bearer ${HCLOUD_TOKEN}" \
    -H "Content-Type: application/json" \
    -d "$body" \
    "$url"
}

existing_server="$(
  hcloud_get "https://api.hetzner.cloud/v1/servers?name=${NAME}" |
    jq -r '.servers[0].id // empty'
)"
if [[ -n "$existing_server" ]]; then
  echo "workcell_lab_server_exists id=${existing_server} name=${NAME}"
  exit 0
fi

firewall_id="$(
  hcloud_get "https://api.hetzner.cloud/v1/firewalls?name=${FIREWALL_NAME}" |
    jq -r '.firewalls[0].id // empty'
)"

if [[ -z "$firewall_id" ]]; then
  firewall_body="$(
    jq -n \
      --arg name "$FIREWALL_NAME" \
      --arg source "$WORKCELL_LAB_SSH_SOURCE_CIDR" \
      '{
        name: $name,
        rules: [
          {
            direction: "in",
            protocol: "tcp",
            port: "22",
            source_ips: [$source]
          }
        ]
      }'
  )"
  firewall_response="$(hcloud_post "https://api.hetzner.cloud/v1/firewalls" "$firewall_body")"
  if [[ "$(jq -r '.error.code // empty' <<<"$firewall_response")" != "" ]]; then
    jq -r '"workcell_lab_firewall_create_failed code=" + .error.code + " message=" + .error.message' <<<"$firewall_response" >&2
    exit 1
  fi
  firewall_id="$(jq -r '.firewall.id' <<<"$firewall_response")"
  echo "workcell_lab_firewall_created id=${firewall_id} name=${FIREWALL_NAME}"
else
  echo "workcell_lab_firewall_exists id=${firewall_id} name=${FIREWALL_NAME}"
fi

server_body="$(
  jq -n \
    --arg name "$NAME" \
    --arg serverType "$SERVER_TYPE" \
    --arg image "$IMAGE" \
    --arg location "$LOCATION" \
    --argjson sshKeyId "$WORKCELL_LAB_SSH_KEY_ID" \
    --argjson firewallId "$firewall_id" \
    '{
      name: $name,
      server_type: $serverType,
      image: $image,
      location: $location,
      ssh_keys: [$sshKeyId],
      firewalls: [{firewall: $firewallId}],
      labels: {
        project: "workcell",
        purpose: "lab",
        environment: "dev"
      },
      start_after_create: true
    }'
)"
server_response="$(hcloud_post "https://api.hetzner.cloud/v1/servers" "$server_body")"

if [[ "$(jq -r '.error.code // empty' <<<"$server_response")" != "" ]]; then
  jq -r '"workcell_lab_server_create_failed code=" + .error.code + " message=" + .error.message' <<<"$server_response" >&2
  exit 1
fi

jq -r '
  "workcell_lab_server_created id=" + (.server.id | tostring)
    + " name=" + .server.name
    + " status=" + .server.status
    + " type=" + .server.server_type.name
    + " location=" + .server.datacenter.location.name
    + " ipv4=" + .server.public_net.ipv4.ip
' <<<"$server_response"
