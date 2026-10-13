// Command gatwd-k8s-operator is the Kubernetes compliance Appliance: the gatwd-k8s
// edition plus embedded use-case rules (rules/compliance.gatwd) that hot-reload
// from a mounted ConfigMap. It is pure assembly; the rules live in .gatwd.
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
	manifests := flag.String("manifests", "", "manifest file or directory to check (required)")
	override := flag.String("rules", "/etc/gatwd/rules.gatwd", "mounted ConfigMap rules that override the embedded ones")
	interval := flag.Duration("interval", 0, "re-check (and hot-reload rules) every interval; 0 checks once")
	flag.Parse()
	if *manifests == "" {
		fmt.Fprintln(os.Stderr, "gatwd-k8s-operator: -manifests is required")
		os.Exit(2)
	}

	client, err := k8sgraph.DynamicClientFromKubeconfig()
	if err != nil {
		client = nil
	}
	app, err := newAppliance(client, *override)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gatwd-k8s-operator: %v\n", err)
		os.Exit(1)
	}

	for {
		for _, v := range app.Execute(checkPipeline(*manifests)) {
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

func newAppliance(client dynamic.Interface, override string) (*cli.Appliance, error) {
	return cli.NewAppliance(
		cli.WithEdition(k8sgraph.ApplyEdition(cli.NewEdition("gatwd-k8s-operator"), client)),
		cli.WithEmbeddedFS(rulesFS, "rules/compliance.gatwd"),
		cli.WithConfigMapOverride(override),
	)
}

func checkPipeline(manifests string) string {
	return "spawn StrVal: " + manifests + " |> manifests |> compliance"
}
