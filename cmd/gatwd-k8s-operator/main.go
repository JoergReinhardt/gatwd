// Command gatwd-k8s-operator is the Kubernetes compliance Appliance: the gatwd-k8s
// edition plus embedded use-case rules (rules/compliance.gatwd) that hot-reload
// from a mounted ConfigMap, layered on top of the embedded ones. It checks offline
// manifests (--manifests) or, when none are given and a cluster is reachable, the live
// cluster (--namespace). It is pure assembly: the command line is the generic
// cli.ApplianceTree, the rules live in .gatwd, and this file only names the data source.
//
//	gatwd-k8s-operator check --manifests DIR      check once
//	gatwd-k8s-operator run --namespace NS --interval 5m
//
// A kustomization root under --manifests is rendered in process; --no-kustomize reads it as
// the plain files it contains, like the gatwd-k8s edition's flag of the same name.
package main

import (
	"embed"
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/client-go/dynamic"

	"github.com/joergreinhardt/gatwd/pkg/cli"
	"github.com/joergreinhardt/gatwd/pkg/k8sgraph"
)

//go:embed rules/*.gatwd
var rulesFS embed.FS

// rulesPath is the embedded rules file inside rulesFS.
const rulesPath = "rules/compliance.gatwd"

func main() {
	operatorTree().Exit()
}

// operatorTree is the operator's command line. The data source is the only domain flag: with
// --manifests the rules run over offline manifests and no cluster client exists; without it,
// the live cluster of the kubeconfig (or the in-cluster service account) is crawled.
func operatorTree() *cli.ApplianceTree {
	return cli.NewApplianceTree(cli.ApplianceSpec{
		Name:      "gatwd-k8s-operator",
		RulesPath: "/etc/gatwd/rules.gatwd",
		Bind: func(cmd *cobra.Command) cli.ApplianceSetup {
			manifests := cmd.PersistentFlags().String("manifests", "", "manifest file or directory to check; omit to check the live cluster")
			namespace := cmd.PersistentFlags().String("namespace", "default", "namespace to crawl when checking the live cluster (\"\" = all)")
			noKustomize := cmd.PersistentFlags().Bool("no-kustomize", false, "read a kustomization root under --manifests as plain files instead of rendering it")
			// The write steps (k8s_apply, k8s_rollout_*) exist in the rules' interpreter but are denied
			// unless --allow-cluster-apply / --allow-cluster-rollout are given; --dry-run (default true),
			// --force-conflicts, --field-manager, --out and --apply-namespace shape them.
			writes := k8sgraph.DefaultWriteFlags()
			writes.Bind(cmd.PersistentFlags())
			return func(s cli.ApplianceSettings) (*cli.Appliance, string, error) {
				return setup(s, *manifests, *namespace, *noKustomize, writes)
			}
		},
		Embedded:    embeddedRules,
		LegacyFlags: []string{"manifests", "namespace", "rules", "rules-mode", "interval"},
	})
}

// embeddedRules is the embedded rule file, for `rules cat`.
func embeddedRules() (string, error) {
	b, err := rulesFS.ReadFile(rulesPath)
	return string(b), err
}

// setup picks the source. Offline manifests never build a cluster client; the live path
// builds it here, once, and its failure is "no source" — a malformed invocation (exit 2).
// --no-kustomize is the edition's switch (k8sgraph.WithoutKustomize) and means something only
// for manifests: with the live source it would be silently ignored, so it is a usage error.
func setup(s cli.ApplianceSettings, manifests, namespace string, noKustomize bool, writes *k8sgraph.WriteFlags) (*cli.Appliance, string, error) {
	opts := []k8sgraph.Option{k8sgraph.WithWriteFlags(writes)}
	if manifests != "" {
		if noKustomize {
			opts = append(opts, k8sgraph.WithoutKustomize())
		}
		app, err := newApplianceWith(nil, opts, s.Rules, s.Mode, cli.WithPolicy(s.Policy), s.BoundOptions())
		return app, checkPipeline(manifests), err
	}
	if noKustomize {
		return nil, "", cli.UsageError(fmt.Errorf("--no-kustomize applies to --manifests; the live cluster has no kustomization roots"))
	}
	client, err := k8sgraph.DynamicClientFromKubeconfig()
	if err != nil {
		return nil, "", cli.UsageError(fmt.Errorf("no --manifests given and no cluster reachable (%v)", err))
	}
	app, err := newApplianceWith(client, opts, s.Rules, s.Mode, cli.WithPolicy(s.Policy), s.BoundOptions())
	return app, livePipeline(namespace), err
}

// newAppliance assembles the operator: the k8s edition, the embedded rules, and the override
// (a mounted ConfigMap) combined with them according to mode. `compliance` is the entry point
// every rule set must define. extra carries the execution policy (--allow-*) and the per-check limits (--timeout, --max-output); without the policy the
// rules, the override above all, may only compute.
func newAppliance(client dynamic.Interface, override string, mode cli.OverrideMode, extra ...cli.ApplianceOption) (*cli.Appliance, error) {
	return newApplianceWith(client, nil, override, mode, extra...)
}

// newApplianceWith is newAppliance with edition options (k8sgraph.WithoutKustomize).
func newApplianceWith(client dynamic.Interface, opts []k8sgraph.Option, override string, mode cli.OverrideMode, extra ...cli.ApplianceOption) (*cli.Appliance, error) {
	return cli.NewAppliance(append([]cli.ApplianceOption{
		cli.WithEdition(k8sgraph.ApplyEdition(cli.NewEdition("gatwd-k8s-operator"), client, opts...)),
		cli.WithEmbeddedFS(rulesFS, rulesPath),
		cli.WithConfigMapOverride(override),
		cli.WithOverrideMode(mode),
		cli.WithRequiredDefs("compliance"),
	}, extra...)...)
}

// checkPipeline runs the compliance rules over offline manifests.
func checkPipeline(manifests string) string {
	return "spawn StrVal: " + manifests + " |> manifests |> compliance"
}

// livePipeline runs the SAME rules over a live cluster: k8s_resources (lib/k8s/sources.gatwd)
// merges every crawl stream of the namespace into the one stream k8s_graph consumes.
func livePipeline(namespace string) string {
	return "spawn StrVal: " + namespace + " |> k8s_resources |> compliance"
}
