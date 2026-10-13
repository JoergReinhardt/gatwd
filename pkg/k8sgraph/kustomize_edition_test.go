package k8sgraph_test

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/pkg/cli"
	"github.com/joergreinhardt/gatwd/pkg/k8sgraph"
)

// editionRun runs the edition's command line and returns exit code, stdout and stderr.
func editionRun(t *testing.T, ed *cli.Edition, argv ...string) (int, string, string) {
	t.Helper()
	var out, errb bytes.Buffer
	code := ed.Command().Execute(argv, strings.NewReader(""), &out, &errb)
	return code, out.String(), errb.String()
}

func offlineEdition(opts ...k8sgraph.Option) *cli.Edition {
	return k8sgraph.ApplyEdition(cli.NewEdition("gatwd-k8s-test"), nil, opts...)
}

// baselineInterp is the edition as it was before auto-render: `manifests` is the plain
// reader, nothing renders. The parity tests compare against it.
func baselineInterp() *cli.Interpreter {
	return cli.NewEdition("baseline").
		WithStep(k8sgraph.ManifestStep("manifests")).
		WithPreludeFS(k8sgraph.PreludeFS(), k8sgraph.PreludePath).
		BuildInterpreter()
}

func countOf(t *testing.T, ed *cli.Edition, root, step string, flags ...string) string {
	t.Helper()
	argv := append(append([]string{}, flags...), "eval", "-e", "spawn StrVal: "+root+" |> "+step+" |> collect_list |> length")
	code, out, errb := editionRun(t, ed, argv...)
	if code != 0 {
		t.Fatalf("%v: exit %d\nstdout: %s\nstderr: %s", argv, code, out, errb)
	}
	return strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(out), "∷"))
}

// TestManifestsAutoRendersAKustomizationRoot is the headline of the in-process work: the
// plain `manifests` step, pointed at an overlay, hands the graph the *rendered* resources,
// and that graph is the one the hand-written plain manifests build.
func TestManifestsAutoRendersAKustomizationRoot(t *testing.T) {
	root, plain := kdir+"/overlays/prod", kdir+"/expected-prod"
	if got := countOf(t, offlineEdition(), root, "manifests"); got != "4" {
		t.Fatalf("`manifests` on the overlay yielded %s resources, want the 4 rendered ones", got)
	}

	interp := offlineEdition().BuildInterpreter()
	graph := func(path, tail string) []string {
		return edgesOf(t, interp, `spawn StrVal: `+path+` |> manifests |> k8s_graph |> `+tail+` |> collect_list`)
	}
	ids, edges := graph(root, "k8s_apply_order"), graph(root, "graph_edges")
	if len(ids) != 4 || len(edges) != 3 {
		t.Fatalf("rendered overlay graph: %d nodes %v, %d edges %v; want 4 and 3", len(ids), ids, len(edges), edges)
	}
	for _, id := range ids {
		if !strings.Contains(id, "prod-web") {
			t.Errorf("node %q was not renamed by the overlay's prefix: the base was read, not rendered", id)
		}
	}
	for _, rel := range []string{"Selects", "RoutesTo", "Configures"} {
		found := false
		for _, e := range edges {
			found = found || strings.Contains(e, rel)
		}
		if !found {
			t.Errorf("no %s edge in %v (references must have followed the renamed targets)", rel, edges)
		}
	}
	// Parity: the rendered overlay and its hand-written plain equivalent are the same graph.
	if !reflect.DeepEqual(ids, graph(plain, "k8s_apply_order")) || !reflect.DeepEqual(edges, graph(plain, "graph_edges")) {
		t.Fatalf("rendered overlay and plain equivalent build different graphs:\n%v %v\n%v %v",
			ids, edges, graph(plain, "k8s_apply_order"), graph(plain, "graph_edges"))
	}
}

// TestNoKustomizeSwitchesAutoRenderOff pins the off switch in every spelling: the flag,
// the explicit `=false`, and the programmatic option. The explicit `kustomize` step is not
// affected by it.
func TestNoKustomizeSwitchesAutoRenderOff(t *testing.T) {
	root := kdir + "/overlays/prod"
	cases := []struct {
		name  string
		ed    *cli.Edition
		flags []string
		step  string
		want  string
	}{
		// 3 = kustomization.yaml, ingress.yaml and pod-patch.yaml read as plain files.
		{"default renders", offlineEdition(), nil, "manifests", "4"},
		{"--no-kustomize reads plain files", offlineEdition(), []string{"--no-kustomize"}, "manifests", "3"},
		{"--no-kustomize=false renders", offlineEdition(), []string{"--no-kustomize=false"}, "manifests", "4"},
		{"WithoutKustomize()", offlineEdition(k8sgraph.WithoutKustomize()), nil, "manifests", "3"},
		{"WithoutKustomize() then --no-kustomize=false", offlineEdition(k8sgraph.WithoutKustomize()), []string{"--no-kustomize=false"}, "manifests", "4"},
		{"the kustomize step ignores the switch", offlineEdition(), []string{"--no-kustomize"}, "kustomize", "4"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := countOf(t, c.ed, root, c.step, c.flags...); got != c.want {
				t.Fatalf("%s %v: %s resources, want %s", c.step, c.flags, got, c.want)
			}
		})
	}

	if code, out, _ := editionRun(t, offlineEdition(), "--help"); code != 0 || !strings.Contains(out, "--no-kustomize") {
		t.Errorf("--help must document --no-kustomize, got exit %d:\n%s", code, out)
	}
}

// TestPlainManifestsBehaveAsBefore is the parity guard: for every path that is not a
// kustomization root — a file, a directory of plain manifests — the edition's `manifests`
// yields exactly what the plain reader (the pre-auto-render step) does, with the switch
// on or off.
func TestPlainManifestsBehaveAsBefore(t *testing.T) {
	base := baselineInterp()
	for _, path := range []string{stackPath, fullStackPath, kdir + "/expected-prod"} {
		for _, flags := range [][]string{nil, {"--no-kustomize"}} {
			for _, tail := range []string{"k8s_graph |> graph_edges", "k8s_graph |> k8s_apply_order"} {
				pipeline := `spawn StrVal: ` + path + ` |> manifests |> ` + tail + ` |> collect_list`
				want := edgesOf(t, base, pipeline)
				if len(want) == 0 {
					t.Fatalf("%s: the baseline yielded nothing (vacuous comparison)", pipeline)
				}
				code, out, errb := editionRun(t, offlineEdition(), append(append([]string{}, flags...), "eval", "-e", pipeline)...)
				if code != 0 {
					t.Fatalf("%v %s: exit %d: %s %s", flags, pipeline, code, out, errb)
				}
				got := edgesOf(t, offlineEdition().BuildInterpreter(), pipeline)
				if !reflect.DeepEqual(got, want) {
					t.Errorf("%s (flags %v): got %v, want the baseline %v", path, flags, got, want)
				}
			}
		}
	}
}

// TestManifestsRenderErrorsAreLoud is the negative fixture of the edition: a broken or
// refused kustomization root is an [ERROR] that names the root and the cause and fails
// the command; it is never an empty graph. (The pipeline deliberately has no
// `collect_list`: that step drops error nodes and prints `[]` — a pre-existing silent
// failure recorded in docs/hardening_plan.md §8.)
func TestManifestsRenderErrorsAreLoud(t *testing.T) {
	cases := []struct {
		dir  string
		want []string
	}{
		{"broken", []string{"overlays/broken", "does-not-exist.yaml"}},
		{"remote", []string{"overlays/remote", "remote reference", "example.invalid"}},
		{"exec", []string{"overlays/exec", "plugin"}},
	}
	for _, c := range cases {
		for _, step := range []string{"manifests", "kustomize"} {
			t.Run(c.dir+"/"+step, func(t *testing.T) {
				code, out, _ := editionRun(t, offlineEdition(), "eval", "-e",
					"spawn StrVal: "+kdir+"/overlays/"+c.dir+" |> "+step+" |> k8s_graph |> graph_edges")
				if code == 0 || !strings.Contains(out, "[ERROR]") {
					t.Fatalf("expected exit != 0 and an [ERROR], got exit %d: %s", code, out)
				}
				for _, w := range c.want {
					if !strings.Contains(out, w) {
						t.Errorf("output %q lacks %q", out, w)
					}
				}
			})
		}
	}
}
