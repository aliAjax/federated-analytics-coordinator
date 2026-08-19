#!/bin/sh
set -eu
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$root"
FAC_ADDRESS=:18081 go run ./cmd/coordinator >/tmp/fac-smoke.log 2>&1 &
pid=$!
cleanup(){ kill "$pid" 2>/dev/null || true; wait "$pid" 2>/dev/null || true; }
trap cleanup EXIT INT TERM
ready=false
for n in 1 2 3 4 5 6 7 8 9 10; do
  if curl -fsS http://127.0.0.1:18081/healthz >/dev/null; then ready=true; break; fi
  sleep 1
done
if [ "$ready" != true ]; then cat /tmp/fac-smoke.log >&2; exit 1; fi
go run ./scripts/smoke_client.go
