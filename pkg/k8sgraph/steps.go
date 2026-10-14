package k8sgraph

import (
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/joergreinhardt/gatwd/adapters"
	"github.com/joergreinhardt/gatwd/pkg/cli"
)

// Commonly-crawled core GVRs, provided as a convenience so an edition's assembly
// reads declaratively (CrawlStep("k8s_pods", client, PodsGVR)). Any GVR works —
// including CRDs discovered at runtime via EnumerateSteps — these are just the
// built-ins an edition is most likely to register by hand.
var (
	PodsGVR         = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	ServicesGVR     = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}
	ConfigMapsGVR   = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}
	SecretsGVR      = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}
	EndpointsGVR    = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "endpoints"}
	DeploymentsGVR  = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	StatefulSetsGVR = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "statefulsets"}
	ReplicaSetsGVR  = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "replicasets"}
	IngressesGVR    = schema.GroupVersionResource{Group: "networking.k8s.io", Version: "v1", Resource: "ingresses"}

	PVCsGVR            = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "persistentvolumeclaims"}
	ServiceAccountsGVR = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "serviceaccounts"}
	RolesGVR           = schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "roles"}
	RoleBindingsGVR    = schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "rolebindings"}
)

// CoreCrawls lists the live source steps an edition registers, keyword → GVR, in
// the order the merged `k8s_resources` stream (lib/k8s/sources.gatwd) visits them.
// Keeping the list as data lets ApplyEdition, the live/offline parity test and the
// prelude's source definition stay in step with one another.
var CoreCrawls = []struct {
	Keyword string
	GVR     schema.GroupVersionResource
}{
	{"k8s_pods", PodsGVR},
	{"k8s_services", ServicesGVR},
	{"k8s_configmaps", ConfigMapsGVR},
	{"k8s_secrets", SecretsGVR},
	{"k8s_endpoints", EndpointsGVR},
	{"k8s_deployments", DeploymentsGVR},
	{"k8s_statefulsets", StatefulSetsGVR},
	{"k8s_replicasets", ReplicaSetsGVR},
	{"k8s_ingresses", IngressesGVR},
	{"k8s_pvcs", PVCsGVR},
	{"k8s_serviceaccounts", ServiceAccountsGVR},
	{"k8s_roles", RolesGVR},
	{"k8s_rolebindings", RoleBindingsGVR},
}

// CrawlStep exposes a configured Kubernetes resource crawl as a named `|>`
// pipeline step. It is the entire domain-specific half of the adapter→DSL bridge:
// it constructs the structured (nested-HAMT) resource adapter for a single GVR
// over a client-go dynamic client, then hands it to the generic, domain-free
// cli.AdapterStep, which does all the engine lifting (EitherT + StepAdapter).
//
// There is deliberately no graph or relation logic here — only client-go wiring.
// The resulting step, used in the DSL as `spawn StrVal: <namespace> |> <keyword>`,
// yields one nested-HAMT payload per resource (navigable with hamt_get /
// hamt_path), ready to be lifted into typed graph nodes by the DSL prelude.
//
// Example:
//
//	cli.NewEdition("gatwd-k8s").
//	    WithStep(k8sgraph.CrawlStep("k8s_pods", client, k8sgraph.PodsGVR)).
//	    WithStep(k8sgraph.CrawlStep("k8s_services", client, k8sgraph.ServicesGVR)).
//	    Run()
func CrawlStep(keyword string, client dynamic.Interface, gvr schema.GroupVersionResource) cli.StepHandler {
	adapter := adapters.NativeKubernetesResourceAdapter(client, gvr)
	return cli.AdapterStep(keyword, adapter)
}

// ManifestStep is the offline twin of CrawlStep: it exposes the filesystem
// manifest reader (adapters.ManifestAdapter) as a named `|>` pipeline step, so an
// edition can crawl rendered or raw manifests with the SAME DSL it uses against a
// live cluster. The step takes the path from the spawned node value:
//
//	spawn StrVal: ./overlays/prod |> manifests |> hamt_path spec.selector
//
// Because the manifest adapter yields the identical nested-HAMT payload as the live
// crawl, every downstream graph/relation rule is source-agnostic (live vs. offline).
//
// With no options the step reads plain files and directories only. ApplyEdition
// registers it with InProcessRenderer and a switch, which is what makes `manifests`
// render a kustomization root on its own (see ApplyEdition); pass
// adapters.ManifestOption values to wire the same behavior by hand.
func ManifestStep(keyword string, opts ...adapters.ManifestOption) cli.StepHandler {
	return cli.AdapterStep(keyword, adapters.ManifestAdapter(opts...))
}

// KustomizeStep is the explicit, always-rendering ManifestStep: when the spawned
// path is a kustomization root it is built in process (InProcessRenderer — no
// external binary) and the output is treated as any other YAML; plain files and
// directories are read directly. Unlike `manifests`, it ignores `--no-kustomize`: a
// script that names the step asked for the render.
//
//	spawn StrVal: ./overlays/prod |> kustomize |> hamt_path spec.selector
//
// To render with an external tool instead (e.g. `kubectl kustomize`), build the
// adapter directly: cli.AdapterStep(keyword,
// adapters.ManifestAdapter(adapters.WithKustomizeCommand("kubectl", "kustomize"))).
func KustomizeStep(keyword string) cli.StepHandler {
	return cli.AdapterStep(keyword, adapters.ManifestAdapter(adapters.WithRenderer(InProcessRenderer())))
}

// RestConfigFromKubeconfig loads a *rest.Config for talking to the cluster,
// preferring the in-cluster service-account config (so the same binary runs as an
// Appliance inside the cluster) and falling back to a kubeconfig file — $KUBECONFIG
// if set, else ~/.kube/config — for local/Edition use. This is the only genuine
// I/O in the shim.
func RestConfigFromKubeconfig() (*rest.Config, error) {
	if cfg, err := rest.InClusterConfig(); err == nil {
		return cfg, nil
	}
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = filepath.Join(homedir.HomeDir(), ".kube", "config")
	}
	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}

// DynamicClientFromKubeconfig builds the schema-agnostic dynamic client used by
// CrawlStep. Being schema-agnostic is what lets a single client crawl any
// resource, including CRDs, without compile-time types.
func DynamicClientFromKubeconfig() (dynamic.Interface, error) {
	cfg, err := RestConfigFromKubeconfig()
	if err != nil {
		return nil, err
	}
	return dynamic.NewForConfig(cfg)
}

// DiscoveryClientFromKubeconfig builds the discovery client used to enumerate the
// full GVR set (incl. CRDs) at runtime — feed its result to adapters.EnumerateResources
// and register a CrawlStep per GVR to crawl "everything in the cluster".
func DiscoveryClientFromKubeconfig() (discovery.DiscoveryInterface, error) {
	cfg, err := RestConfigFromKubeconfig()
	if err != nil {
		return nil, err
	}
	return discovery.NewDiscoveryClientForConfig(cfg)
}
