#!/usr/bin/env bash
set -euo pipefail

HOST="${WORKCELL_LAB_SSH_HOST:-}"
USER="${WORKCELL_LAB_SSH_USER:-root}"
KEY="${WORKCELL_LAB_SSH_KEY:-}"
REPO_DIR="${WORKCELL_LAB_REPO_DIR:-/opt/workcell}"
RUN_BACKEND_SMOKES="${WORKCELL_LAB_PREFLIGHT_RUN_BACKEND_SMOKES:-0}"

if [[ -z "$HOST" || -z "$KEY" ]]; then
  cat <<'EOF'
usage:
  WORKCELL_LAB_SSH_HOST=<host> \
  WORKCELL_LAB_SSH_KEY=<path> \
  scripts/run-remote-lab-preflight.sh

Optional:
  WORKCELL_LAB_SSH_USER=root
  WORKCELL_LAB_REPO_DIR=/opt/workcell
  WORKCELL_LAB_PREFLIGHT_RUN_BACKEND_SMOKES=1
EOF
  exit 64
fi

ssh \
  -i "$KEY" \
  -o StrictHostKeyChecking=accept-new \
  "${USER}@${HOST}" \
  "set -euo pipefail; cd '${REPO_DIR}'; git pull --ff-only; WORKCELL_LAB_PREFLIGHT_RUN_BACKEND_SMOKES='${RUN_BACKEND_SMOKES}' scripts/lab-host-preflight.sh"
