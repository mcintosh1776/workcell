#!/usr/bin/env bash
set -euo pipefail

go test ./...
go run ./cmd/workcell run --profile fake -- echo hello >/tmp/workcell-dev-smoke-job.json
jq -e '.state == "succeeded" and .backend == "fake"' /tmp/workcell-dev-smoke-job.json >/dev/null
printf 'workcell_dev_smoke=ok\n'
