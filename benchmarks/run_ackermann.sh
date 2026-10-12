#!/usr/bin/env bash
# Run the Ackermann interpreter benchmark and report wall-clock + peak RSS.
#
# Usage: benchmarks/run_ackermann.sh [script]
#   script  .gatwd file to run (default: benchmarks/ackermann.gatwd)
#
# Peak RSS is the leak-adjacent signal: for a non-leaking run it stays flat
# (~18 MB for A(3,6)) regardless of how long the recursion runs. A rising peak
# across larger inputs that *should* unwind is the first sign of a leak.
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
script="${1:-$repo_root/benchmarks/ackermann.gatwd}"
bin="$repo_root/gatwd"

echo "==> Building gatwd"
go build -o "$bin" "$repo_root/cmd/gatwd"

echo "==> Running: $script"
# GNU time (-v) reports peak RSS; fall back to the shell builtin if absent.
if command -v /usr/bin/time >/dev/null 2>&1; then
    /usr/bin/time -v "$bin" "$script" 2>&1 \
        | grep -E 'Elapsed \(wall|Maximum resident|∷|assert|FAIL|error' || true
else
    time "$bin" "$script"
fi
