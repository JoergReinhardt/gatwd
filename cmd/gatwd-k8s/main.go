// Command gatwd-k8s is the Kubernetes graph Edition of gatwd: the full interpreter
// plus the live crawl steps, the offline manifests/kustomize steps and the pure-DSL
// relation prelude. It is pure assembly; every relation lives in .gatwd.
package main

import (
	"fmt"
	"os"

	"github.com/joergreinhardt/gatwd/pkg/cli"
	"github.com/joergreinhardt/gatwd/pkg/k8sgraph"
)

func main() {
	client, err := k8sgraph.DynamicClientFromKubeconfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "gatwd-k8s: no cluster access (%v); running offline (manifests/kustomize only)\n", err)
		client = nil
	}
	k8sgraph.ApplyEdition(cli.NewEdition("gatwd-k8s"), client).Run()
}
