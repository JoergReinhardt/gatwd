// Command gatwd-k8s-operator is the Kubernetes compliance Appliance: the gatwd-k8s
// edition plus embedded use-case rules (rules/compliance.gatwd) that hot-reload
// from a mounted ConfigMap. It checks offline manifests (-manifests) or, when none
// are given and a cluster is reachable, the live cluster (-namespace). It is pure
// assembly; the rules live in .gatwd.
package main

import (
	"embed"
	"flag"
	"fmt"
	"os"
	"time"

	"k8s.io/client-go/dynamic"

	"github.com/joergreinhardt/gatwd/pkg/cli"
	"github.com/joergreinhardt/gatwd/pkg/k8sgraph"
)

//go:embed rules/*.gatwd
var rulesFS embed.FS

func main() {
	manifests := flag.String("manifests", "", "manifest file or directory to check; omit to check the live cluster")
	namespace := flag.String("namespace", "default", "namespace to crawl when checking the live cluster")
	override := flag.String("rules", "/etc/gatwd/rules.gatwd", "mounted ConfigMap rules that override the embedded ones")
	rulesMode := flag.String("rules-mode", "layer", "how the override combines with the embedded rules: layer (on top of them) or replace (instead of them)")
	interval := flag.Duration("interval", 0, "re-check (and hot-reload rules) every interval; 0 checks once")
	flag.Parse()

	mode, err := cli.ParseOverrideMode(*rulesMode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gatwd-k8s-operator: %v\n", err)
		os.Exit(2)
	}
	client, err := k8sgraph.DynamicClientFromKubeconfig()
	if err != nil {
		client = nil
	}
	var pipeline string
	switch {
	case *manifests != "":
		pipeline = checkPipeline(*manifests)
	case client != nil:
		pipeline = livePipeline(*namespace)
	default:
		fmt.Fprintln(os.Stderr, "gatwd-k8s-operator: no -manifests given and no cluster reachable")
		os.Exit(2)
	}
	app, err := newAppliance(client, *override, mode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gatwd-k8s-operator: %v\n", err)
		os.Exit(1)
	}
	for _, line := range app.Report().Lines() {
		fmt.Fprintln(os.Stderr, "gatwd-k8s-operator: "+line)
	}

	for {
		for _, v := range app.Execute(pipeline) {
			fmt.Println(v)
		}
		if *interval == 0 {
			return
		}
		time.Sleep(*interval)
		if err := app.Reload(); err != nil {
			fmt.Fprintf(os.Stderr, "gatwd-k8s-operator: reload: %v\n", err)
		}
	}
}

// newAppliance assembles the operator: the k8s edition, the embedded rules, and the override
// (a mounted ConfigMap) combined with them according to mode. `compliance` is the entry point
// every rule set must define.
func newAppliance(client dynamic.Interface, override string, mode cli.OverrideMode) (*cli.Appliance, error) {
	return cli.NewAppliance(
		cli.WithEdition(k8sgraph.ApplyEdition(cli.NewEdition("gatwd-k8s-operator"), client)),
		cli.WithEmbeddedFS(rulesFS, "rules/compliance.gatwd"),
		cli.WithConfigMapOverride(override),
		cli.WithOverrideMode(mode),
		cli.WithRequiredDefs("compliance"),
	)
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
