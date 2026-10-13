package k8sgraph

import (
	"embed"
	"io/fs"

	"github.com/spf13/pflag"
	"k8s.io/client-go/dynamic"

	"github.com/joergreinhardt/gatwd/adapters"
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

// Option adjusts ApplyEdition.
type Option func(*editionConfig)

type editionConfig struct {
	noKustomize bool
	write       *WriteFlags
	resolver    adapters.ResourceResolver
}

// WithoutKustomize makes `manifests` read a kustomization root as the plain files it
// contains, as it did before auto-render; it is the programmatic form of the
// `--no-kustomize` flag (which can still be given on the command line, and overrides
// nothing: the flag's default is this option's value). An Appliance that has its own
// flag set uses it to offer the same switch.
func WithoutKustomize() Option { return func(c *editionConfig) { c.noKustomize = true } }

// ApplyEdition assembles the Kubernetes graph edition onto e: the offline
// `manifests` and `kustomize` source steps, one live crawl step per commonly
// crawled core resource (k8s_pods, k8s_services, ...) when client is non-nil, and
// the embedded DSL prelude that defines k8s_ingest, the relation rules, the
// k8s_graph pipeline and — for a live cluster — k8s_resources, the merge of every
// crawl stream into one resource stream.
//
// `manifests` renders a kustomization root automatically — in process, with the
// kustomize library, no external binary (InProcessRenderer) — and reads every
// other path (a file, a directory without a kustomization file) exactly as before.
// The edition gains one persistent flag, `--no-kustomize`, which turns that off so a
// kustomization root is read as the plain files it contains; the explicit
// `kustomize` step always renders. A render error is an `[ERROR]` that names the root.
//
// It also registers the write steps (`k8s_apply`, `k8s_rollout_*`, `k8s_scale`, see write.go),
// denied unless the policy grants cli.CapClusterApply / cli.CapClusterRollout: the edition adds
// the flags `--allow-cluster-apply`, `--allow-cluster-rollout`, `--dry-run` (default true, a
// server dry run), `--force-conflicts`, `--field-manager`, `--out` and `--apply-namespace`;
// a host with its own command line passes WithWriteFlags.
//
// It is pure assembly: every relation and graph algorithm comes from the
// embedded .gatwd library and the generic graph core. Pass a nil client to build
// an offline-only edition (manifests/kustomize), e.g. for CI or tests.
//
// Example:
//
//	k8sgraph.ApplyEdition(cli.NewEdition("gatwd-k8s"), client).Run()
func ApplyEdition(e *cli.Edition, client dynamic.Interface, opts ...Option) *cli.Edition {
	var cfg editionConfig
	for _, o := range opts {
		o(&cfg)
	}
	// noKustomize is read by the step when it runs; the flag is parsed after the steps
	// are registered, so it must bind to the variable rather than to a copied value.
	noKustomize := cfg.noKustomize
	e = e.WithFlags(func(f *pflag.FlagSet) {
		f.BoolVar(&noKustomize, "no-kustomize", cfg.noKustomize,
			"read a kustomization root as plain files instead of rendering it (the `manifests` step)")
	})
	e = e.WithStep(ManifestStep("manifests",
		adapters.WithRenderer(InProcessRenderer()),
		adapters.WithKustomizeSwitch(func() bool { return !noKustomize }))).
		WithStep(KustomizeStep("kustomize"))
	// The write steps are always registered and always gated: a script needs the explicit
	// capability (--allow-cluster-apply) and writes nothing but a server dry run until
	// --dry-run=false. With no client they fail loudly per object (nothing to write to).
	write := cfg.write
	if write == nil {
		write = DefaultWriteFlags()
		e = e.WithFlags(write.Bind).
			WithCapabilityFlag("allow-cluster-apply", cli.CapClusterApply, "let scripts create and update objects in the cluster (k8s_apply)").
			WithCapabilityFlag("allow-cluster-rollout", cli.CapClusterRollout, "let scripts restart, undo, pause, resume and scale workloads (k8s_rollout_*, k8s_scale)")
	}
	applier := newApplier(client, write, cfg.resolver)
	e = e.WithStep(ApplyStep("k8s_apply", applier, func() string { return write.Out }))
	for _, h := range RolloutSteps(applier) {
		e = e.WithStep(h)
	}
	if client != nil {
		for _, c := range CoreCrawls {
			e = e.WithStep(CrawlStep(c.Keyword, client, c.GVR))
		}
	}
	return e.WithPreludeFS(preludeFS, PreludePath)
}
