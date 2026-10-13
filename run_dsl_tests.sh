#!/bin/bash
# Exhaustive DSL integration suite. A test can only fail if it can be SEEN to run, so
# this gate fails on three things, not one:
#   1. any [ERROR] in the output (a failed assert, an unknown step, a failed call);
#   2. any "(No output or node unhandled)" — a statement that produced nothing. An
#      `assert` that never sees a value reports an error, but any other silent
#      statement is still a failure here;
#   3. any structural error from `gatwd --check` (warnings are advisory).
set -e

echo "Running Exhaustive DSL Tests..."
go build -o gatwd ./cmd/gatwd
OUTPUT=$(./gatwd tests/dsl_exhaustive_test.gatwd)

echo "$OUTPUT"

status=0
if echo "$OUTPUT" | grep -q "\[ERROR\]"; then
    echo "❌ DSL Tests FAILED! Found errors in pipeline execution."
    status=1
fi
if echo "$OUTPUT" | grep -q "No output or node unhandled"; then
    echo "❌ DSL Tests FAILED! A statement produced no output, so its assert (if any) ran vacuously:"
    echo "$OUTPUT" | grep -n "No output or node unhandled"
    status=1
fi
if ! ./gatwd --check tests/dsl_exhaustive_test.gatwd >/dev/null 2>&1; then
    echo "❌ DSL Tests FAILED! gatwd --check reports structural errors:"
    ./gatwd --check tests/dsl_exhaustive_test.gatwd 2>&1 | grep "error:" || true
    status=1
fi

if [ "$status" -eq 0 ]; then
    echo "✅ All DSL Tests PASSED!"
fi
exit $status
