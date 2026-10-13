// Command gatwd-k8s is the Kubernetes graph Edition of gatwd: the full interpreter
// plus the live crawl steps, the offline manifests/kustomize steps and the pure-DSL
// relation prelude. It is pure assembly; every relation lives in .gatwd.
//
// The cluster client is built lazily, the first time a crawl step runs: `version`,
// `completion`, `check` and scripts that read manifests never look for a kubeconfig. A
// crawl with no reachable cluster is an `[ERROR] no cluster access (…)` result.
package main

import (
	"github.com/joergreinhardt/gatwd/pkg/cli"
	"github.com/joergreinhardt/gatwd/pkg/k8sgraph"
)

func main() {
	k8sgraph.ApplyEdition(cli.NewEdition("gatwd-k8s"), k8sgraph.LazyDynamicClient()).Run()
}
