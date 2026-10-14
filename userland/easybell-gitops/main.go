// Command gatwd-easybell-gitops is a user-space Gatwd Edition: it reads a kustomize-based
// Kubernetes GitOps repository offline, builds the resource graph with the generic k8s
// prelude and adds the Ingress -> Service -> Deployment view (hosts, IPs, port names).
// It is pure assembly; all logic lives in lib/*.gatwd.
package main

import (
	"embed"

	"github.com/joergreinhardt/gatwd/pkg/cli"
	"github.com/joergreinhardt/gatwd/pkg/k8sgraph"
)

//go:embed lib/*.gatwd
var libFS embed.FS

// edition assembles the offline k8s edition plus this library's prelude.
func edition() *cli.Edition {
	return k8sgraph.ApplyEdition(cli.NewEdition("gatwd-easybell-gitops"), nil).
		WithPreludeFS(libFS, "lib/prelude.gatwd")
}

func main() { edition().Run() }
