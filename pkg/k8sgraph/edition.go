package k8sgraph

import (
	"embed"
	"io/fs"

	"k8s.io/client-go/dynamic"

	"github.com/joergreinhardt/gatwd/pkg/cli"
)

// preludeFS embeds the pure-DSL Kubernetes relation library. The relations
// (Owns, Selects, RoutesTo, ...) live in these .gatwd files, never in Go.
//
//go:embed lib/k8s/*.gatwd
var preludeFS embed.FS

// PreludePath is the embedded entry point that imports the rest of the library.
const PreludePath = "lib/k8s/prelude.gatwd"

// PreludeFS returns the embedded Kubernetes prelude library, rooted so that
// PreludePath resolves inside it. Use it to load the library into an interpreter
// that is not built through an Edition.
func PreludeFS() fs.FS { return preludeFS }

// ApplyEdition assembles the Kubernetes graph edition onto e: the offline
// `manifests` and `kustomize` source steps, one live crawl step per commonly
// crawled core resource (k8s_pods, k8s_services, ...) when client is non-nil, and
// the embedded DSL prelude that defines k8s_ingest, the relation rules, the
// k8s_graph pipeline and — for a live cluster — k8s_resources, the merge of every
// crawl stream into one resource stream.
//
// It is pure assembly: every relation and graph algorithm comes from the
// embedded .gatwd library and the generic graph core. Pass a nil client to build
// an offline-only edition (manifests/kustomize), e.g. for CI or tests.
//
// Example:
//
//	k8sgraph.ApplyEdition(cli.NewEdition("gatwd-k8s"), client).Run()
func ApplyEdition(e *cli.Edition, client dynamic.Interface) *cli.Edition {
	e = e.WithStep(ManifestStep("manifests")).WithStep(KustomizeStep("kustomize"))
	if client != nil {
		for _, c := range CoreCrawls {
			e = e.WithStep(CrawlStep(c.Keyword, client, c.GVR))
		}
	}
	return e.WithPreludeFS(preludeFS, PreludePath)
}
