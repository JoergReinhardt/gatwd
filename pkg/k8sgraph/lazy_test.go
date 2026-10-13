package k8sgraph_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"k8s.io/client-go/dynamic"

	"github.com/joergreinhardt/gatwd/pkg/cli"
	"github.com/joergreinhardt/gatwd/pkg/k8sgraph"
	"github.com/joergreinhardt/gatwd/pkg/k8sgraph/k8sgraphtest"
)

// The lazy client builds nothing until a crawl asks for a resource, builds exactly once, and
// then behaves like the real client (here a fake cluster).
func TestLazyClientBuildsOnFirstUseOnly(t *testing.T) {
	built := 0
	fake := k8sgraphtest.FakeClusterFromFile(t, "testdata/stack.yaml")
	edition := k8sgraph.ApplyEdition(cli.NewEdition("lazy"), k8sgraph.LazyClient(func() (dynamic.Interface, error) {
		built++
		return fake, nil
	}))

	in := edition.BuildInterpreter()
	if built != 0 {
		t.Fatalf("assembling the edition built the client %d time(s); it must wait for a crawl", built)
	}
	// An offline-only command never touches it.
	if vals := in.EvaluateToValues(`spawn StrVal: testdata/stack.yaml |> manifests |> collect_list`); len(vals) == 0 || built != 0 {
		t.Fatalf("offline pipeline: %v (built %d)", vals, built)
	}
	for i := 0; i < 2; i++ {
		vals := in.EvaluateToValues(`spawn StrVal: default |> k8s_pods |> collect_list`)
		if len(vals) != 1 || fmt.Sprint(vals[0]) == "[]" {
			t.Fatalf("crawl through the lazy client: %v", vals)
		}
	}
	if built != 1 {
		t.Fatalf("client built %d times, want exactly once", built)
	}
}

// A client that cannot be built is an error where the crawl runs, naming the cause — not a
// step that quietly does not exist, and not an empty cluster.
func TestLazyClientFailureIsLoudAtTheCrawl(t *testing.T) {
	cause := errors.New("kubeconfig: no such file")
	edition := k8sgraph.ApplyEdition(cli.NewEdition("lazy"), k8sgraph.LazyClient(func() (dynamic.Interface, error) { return nil, cause }))
	in := edition.BuildInterpreter()

	// (No `collect_list` here: it drops error nodes from the list it builds, a pre-existing
	// silent failure recorded in the plan, so the error would be invisible through it.)
	vals := in.EvaluateToValues(`spawn StrVal: default |> k8s_pods`)
	if len(vals) == 0 {
		t.Fatal("a crawl with no cluster must report an error, got nothing")
	}
	out := fmt.Sprint(vals)
	if !strings.Contains(out, "no cluster access") || !strings.Contains(out, "kubeconfig: no such file") {
		t.Fatalf("crawl error = %s, want one naming 'no cluster access' and the cause", out)
	}
	// the offline steps still work with the same edition
	if vals := in.EvaluateToValues(`spawn StrVal: testdata/stack.yaml |> manifests |> collect_list`); len(vals) == 0 || strings.Contains(fmt.Sprint(vals), "ERROR") {
		t.Fatalf("manifests must be unaffected: %v", vals)
	}
}
