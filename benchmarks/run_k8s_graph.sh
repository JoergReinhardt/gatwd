#!/usr/bin/env bash
# Benchmark the Kubernetes graph pipeline (decode -> ingest -> every relation pass)
# on synthetic clusters of growing size.
#
# Usage: benchmarks/run_k8s_graph.sh [benchtime]
#   benchtime  go test -benchtime value (default: 3x)
#
# Read the scaling across the svc*_pods* sub-benchmarks: the Selects pass is
# index-based, so doubling the cluster should roughly double the time. A return to
# quadratic growth means the label index stopped being used.
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"
go test ./pkg/k8sgraph -run XXX -bench K8sGraph -benchtime "${1:-3x}" -benchmem
