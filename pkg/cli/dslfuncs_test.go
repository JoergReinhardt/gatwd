package cli

import (
	"testing"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/functions"
)

// TestGraphIngestFoldStructural drives the "option C" ingest primitive — a fold with
// a GRAPH accumulator whose reducer applies graph-construction functions — and
// asserts on the raw Go memory type yielded (a *collections.Graph), not a formatted
// string, per STYLEGUIDE §4. It proves the whole vertical slice: the infix language
// applies the graph_add_node builtin, the accumulator threads as a graph, and each
// list element becomes a node.
func TestGraphIngestFoldStructural(t *testing.T) {
	i := NewInterpreter()
	vals := i.EvaluateToValues(`let ids = ["a", "b", "c"] in ids |> fold (\g r -> graph_add_node "N" r g) graph_empty`)
	if len(vals) != 1 {
		t.Fatalf("want 1 result value, got %d", len(vals))
	}
	g, ok := vals[0].(*collections.Graph[string, string, any, any])
	if !ok {
		t.Fatalf("ingest fold must yield a *collections.Graph, got %T", vals[0])
	}
	if g.Order() != 3 {
		t.Errorf("want 3 nodes, got %d", g.Order())
	}
	nodes := g.Nodes()
	want := []struct{ id, color string }{{"a", "N"}, {"b", "N"}, {"c", "N"}}
	for idx, w := range want {
		if nodes[idx].ID != w.id || nodes[idx].Color != w.color {
			t.Errorf("node[%d] = (%v,%v), want (%s,%s)", idx, nodes[idx].ID, nodes[idx].Color, w.id, w.color)
		}
	}
}

// TestRecordIngestPayload proves the real ingest shape: color and id are DERIVED
// from a record payload via the hamt_get function form, and the whole record is kept
// as the node payload (so later relation rules can read its other fields).
func TestRecordIngestPayload(t *testing.T) {
	i := NewInterpreter()
	vals := i.EvaluateToValues(`let r = spawn S: x |> emit kind:Pod, name:nginx |> drop 1 |> to_hamt in [r] |> fold (\g x -> graph_add_node (hamt_get "kind" x) (hamt_get "name" x) x g) graph_empty`)
	g := vals[0].(*collections.Graph[string, string, any, any])
	nodes := g.Nodes()
	if len(nodes) != 1 || nodes[0].ID != "nginx" || nodes[0].Color != "Pod" {
		t.Fatalf("want one Pod node id=nginx, got %+v", nodes)
	}
	// The payload must be the record HAMT itself, not the id string.
	if _, ok := nodes[0].Payload.(*collections.HAMT[string, any]); !ok {
		t.Errorf("node payload must be the record HAMT, got %T", nodes[0].Payload)
	}
}

// TestDSLBuiltinFuncsArity confirms the registered builtins validate arity and
// surface mistakes as errors (kept inside the DSL's Either discipline) rather than
// panicking.
func TestDSLBuiltinFuncsArity(t *testing.T) {
	addNode := dslBuiltinFuncs["graph_add_node"].(functions.DSLFunc)
	if _, err := addNode([]any{"only-one"}); err == nil {
		t.Errorf("graph_add_node with 1 arg should error on arity")
	}
	if _, err := addNode([]any{"N", "id", "not-a-graph"}); err == nil {
		t.Errorf("graph_add_node with a non-graph accumulator should error")
	}
	// graph_empty is the empty-graph VALUE used as the fold seed.
	if _, ok := dslBuiltinFuncs["graph_empty"].(*collections.Graph[string, string, any, any]); !ok {
		t.Errorf("graph_empty must be a *collections.Graph value")
	}
}

// TestExprFallback covers the general-lookup fallback (§3B + §5.1b): a builtin
// function name resolves to its DSLFunc, a declared sum-type constructor resolves to
// its tag STRING, and an unknown identifier still misses (so a real undefined name
// surfaces an error rather than being silently swallowed).
func TestExprFallback(t *testing.T) {
	i := NewInterpreter()
	i.EvaluateLine(`type EdgeColor = Owns | Selects | RoutesTo`)

	if v, ok := i.exprFallback("hamt_get"); !ok {
		t.Errorf("hamt_get should resolve to a builtin")
	} else if _, isFn := v.(functions.DSLFunc); !isFn {
		t.Errorf("hamt_get should resolve to a DSLFunc, got %T", v)
	}
	if v, ok := i.exprFallback("Owns"); !ok || v != "Owns" {
		t.Errorf(`constructor "Owns" should resolve to the tag string "Owns", got %v (ok=%v)`, v, ok)
	}
	if _, ok := i.exprFallback("definitely_not_defined"); ok {
		t.Errorf("an unknown identifier must NOT resolve via the fallback")
	}
}
