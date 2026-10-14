package k8sgraph_test

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/functions"
	"github.com/joergreinhardt/gatwd/pkg/cli"
	"github.com/joergreinhardt/gatwd/pkg/k8sgraph"
	"github.com/joergreinhardt/gatwd/pkg/k8sgraph/k8sgraphtest"
)

const parityManifestPath = "testdata/parity_manifest.yaml"

// parityRows runs a manifest ⊆ live parity report and renders it as sorted
// "id:status" strings. The report is built with a top-level `let ... in`
// expression, not a reusable task def: see lib/k8s/parity.gatwd's comment on
// k8s_parity_report for why (a lambda chaining more than one `|>`, or calling
// another `|>`-bodied def by name, is not reliably resolved by this
// interpreter's task-closure path — found and worked around while building
// this feature, docs/manifest_live_parity_design.md).
func parityRows(t *testing.T, interp *cli.Interpreter, manifestPath, ns string) []string {
	t.Helper()
	stmt := fmt.Sprintf(
		`let m = (%q |> manifests |> k8s_graph) in let l = (%q |> k8s_resources |> k8s_graph) in k8s_parity_report m l`,
		manifestPath, ns)
	vals := interp.EvaluateToValues(stmt)
	if len(vals) != 1 {
		t.Fatalf("%s: expected one value, got %d: %v", stmt, len(vals), vals)
	}
	rows, ok := functions.ToSliceAny(vals[0])
	if !ok {
		t.Fatalf("%s: expected a list, got %T (%v)", stmt, vals[0], vals[0])
	}
	out := make([]string, len(rows))
	for i, r := range rows {
		pair, ok := functions.ToSliceAny(r)
		if !ok || len(pair) != 2 {
			t.Fatalf("row %d: expected an [id, status] pair, got %T (%v)", i, r, r)
		}
		out[i] = fmt.Sprintf("%v:%v", pair[0], pair[1])
	}
	sort.Strings(out)
	return out
}

// TestParityReportsOkDriftAndMissing is Phase 2/3 of
// docs/manifest_live_parity_design.md, over a fake live cluster
// (k8sgraphtest.FakeClusterFromYAML — never a real cluster): one resource is
// identical live plus extra server-managed noise (ok — the positive case),
// one has a genuinely different declared field live (drift), and one was
// never applied at all (missing — the negative fixture pairing of Mandate 6:
// a report that is ALWAYS "ok" looks identical to a report that never ran).
func TestParityReportsOkDriftAndMissing(t *testing.T) {
	live := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: default
  resourceVersion: "123"
  uid: abc-123
  labels:
    extra: added-by-someone
data:
  FOO: bar
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
  namespace: default
  resourceVersion: "456"
spec:
  replicas: 5
  strategy:
    type: RollingUpdate
status:
  availableReplicas: 5
`
	interp := k8sgraph.ApplyEdition(cli.NewEdition("gatwd-k8s-parity-test"),
		k8sgraphtest.FakeClusterFromYAML(t, []byte(live))).BuildInterpreter()

	got := parityRows(t, interp, parityManifestPath, "default")
	want := []string{
		"ConfigMap/default/app-config:ok",
		"Deployment/default/app:drift",
		"Service/default/never-applied:missing",
	}
	if len(got) != len(want) {
		t.Fatalf("parity report = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("parity report = %v, want %v", got, want)
			break
		}
	}
}

// TestParityCommand drives the actual `parity` subcommand (ParityCommand,
// wired in cmd/gatwd-k8s/main.go) end to end over the fake cluster, using
// editionRun (kustomize_edition_test.go) the same way every other CLI-level
// test in this package does.
func TestParityCommand(t *testing.T) {
	live := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: default
  resourceVersion: "123"
data:
  FOO: bar
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
  namespace: default
spec:
  replicas: 5
`
	newEdition := func() *cli.Edition {
		ed := k8sgraph.ApplyEdition(cli.NewEdition("gatwd-k8s-parity-cmd-test"), k8sgraphtest.FakeClusterFromYAML(t, []byte(live)))
		return ed.WithCommand(k8sgraph.ParityCommand(ed))
	}

	code, out, errOut := editionRun(t, newEdition(), "parity", "-D", "NS=default", "--", parityManifestPath)
	if code != 0 {
		t.Fatalf("parity: exit %d\nstdout:\n%s\nstderr:\n%s", code, out, errOut)
	}
	for _, want := range []string{"ConfigMap/default/app-config", "ok", "Deployment/default/app", "drift", "Service/default/never-applied", "missing"} {
		if !strings.Contains(out, want) {
			t.Errorf("parity output %q lacks %q", out, want)
		}
	}

	// Negative control, on a FRESH command instance (a *cobra.Command's own flag
	// state persists across repeated Execute() calls on the same instance, so
	// reusing the one above would just see the previous run's -D NS=default):
	// no -D NS at all must fail with a usage error, not run with a silent
	// default — the `-D OVERLAY`/`-D PATTERN` convention this mirrors (easyk8s's
	// AGENTS.md) exists specifically so a forgotten flag is loud, not a quietly
	// wrong (and here, maybe reassuring-looking) report.
	code, _, errOut = editionRun(t, newEdition(), "parity", "--", parityManifestPath)
	if code == 0 || !strings.Contains(errOut, "NS") {
		t.Fatalf("parity without -D NS: exit %d, stderr %q, want a usage error naming NS", code, errOut)
	}
}

// Negative control, pinned the way the rest of this codebase pins one: the
// SAME manifest compared against ITSELF (as the "live" side) must report
// every resource "ok" — if this ever reported anything else, the comparison
// logic itself would be broken, not the fixture.
func TestParityOfIdenticalGraphsIsAllOK(t *testing.T) {
	interp := k8sgraph.ApplyEdition(cli.NewEdition("gatwd-k8s-parity-identity-test"),
		k8sgraphtest.FakeClusterFromFile(t, parityManifestPath)).BuildInterpreter()

	got := parityRows(t, interp, parityManifestPath, "default")
	for _, row := range got {
		if row[len(row)-2:] != "ok" {
			t.Errorf("identical manifest/live comparison: %q is not ok", row)
		}
	}
	if len(got) != 3 {
		t.Fatalf("parity report = %v, want 3 rows", got)
	}
}
