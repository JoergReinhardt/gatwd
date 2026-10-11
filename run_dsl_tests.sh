#!/bin/bash
set -e

echo "Running Exhaustive DSL Tests..."
go build -o gatwd ./cmd/gatwd
OUTPUT=$(./gatwd tests/dsl_exhaustive_test.gatwd)

echo "$OUTPUT"

if echo "$OUTPUT" | grep -q "\[ERROR\]"; then
    echo "❌ DSL Tests FAILED! Found errors in pipeline execution."
    exit 1
else
    echo "✅ All DSL Tests PASSED!"
    exit 0
fi
