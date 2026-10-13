package cli

import (
	"fmt"
	"strings"
	"testing"
)

// graphOf evaluates one statement and returns the *collections.Graph it yields.
func graphOf(t *testing.T, in *Interpreter, stmt string) dslGraph {
	t.Helper()
	vals := in.EvaluateToValues(stmt)
	if len(vals) != 1 {
		t.Fatalf("%q: want one value, got %v", stmt, vals)
	}
	g, ok := asDSLGraph(vals[0])
	if !ok {
		t.Fatalf("%q yielded %T, want a graph", stmt, vals[0])
	}
	return g
}

// TestGraphInsertNodeForms: the strict node declaration exists as a step and as a
// function form, and both accept qualified ids; the same local name under another
// kind or scope is a different node.
func TestGraphInsertNodeForms(t *testing.T) {
	in := NewInterpreter()

	step := graphOf(t, in, `spawn S: x |> graph_empty |> graph_insert_node Pod Pod/default/web |> graph_insert_node Service Service/default/web |> graph_insert_node Pod Pod/prod/web`)
	if got := step.NodeIDs(); len(got) != 3 {
		t.Fatalf("step form: nodes = %v, want 3 distinct ids", got)
	}

	fn := graphOf(t, in, `[["Pod", "Pod/default/web"], ["Service", "Service/default/web"], ["Pod", "Pod/prod/web"]] |> fold (\g c -> graph_insert_node (head c) (nth 1 c) g) graph_empty`)
	if !step.Equal(fn) {
		t.Errorf("step form and function form disagree: %v vs %v", step.NodeIDs(), fn.NodeIDs())
	}

	withPayload := graphOf(t, in, `[1] |> fold (\g r -> graph_insert_node "N" "k" r g) graph_empty`)
	if n, _ := withPayload.GetNode("k"); fmt.Sprint(n.Payload) != "1" {
		t.Errorf("4-argument form lost the payload: %v", n)
	}
}

// TestGraphInsertNodeDuplicateIsLoud: a repeated id is an [ERROR] naming the id and
// both declarations, in the step form and in the function form inside a fold, and it
// is not swallowed by a following collect_list.
func TestGraphInsertNodeDuplicateIsLoud(t *testing.T) {
	rows := []struct{ id, stmt string }{
		{"step", `spawn S: x |> graph_empty |> graph_insert_node Pod Pod/default/web |> graph_insert_node Pod Pod/default/web`},
		{"function in fold", `["Pod/default/web", "Pod/default/web"] |> fold (\g r -> graph_insert_node "Pod" r g) graph_empty`},
		{"through a listing", `["Pod/default/web", "Pod/default/web"] |> fold (\g r -> graph_insert_node "Pod" r g) graph_empty |> graph_nodes |> collect_list`},
	}
	for _, row := range rows {
		t.Run(row.id, func(t *testing.T) {
			got := strings.Join(NewInterpreter().evaluatePipelineStr(row.stmt), "\n")
			for _, want := range []string{"[ERROR]", "duplicate node id", "Pod/default/web"} {
				if !strings.Contains(got, want) {
					t.Errorf("%s: %q does not contain %q", row.id, got, want)
				}
			}
		})
	}
}

// TestGraphInsertNodeErrorShowsPayloads: the function form renders structured
// payloads as JSON (the only way to tell two resources apart), not as a pointer.
func TestGraphInsertNodeErrorShowsPayloads(t *testing.T) {
	in := NewInterpreter()
	in.EvaluateLine(`let a = spawn S: x |> emit kind:Pod, name:web |> drop 1 |> to_hamt`)
	in.EvaluateLine(`let b = spawn S: x |> emit kind:Pod, name:other |> drop 1 |> to_hamt`)
	got := strings.Join(in.evaluatePipelineStr(`[a, b] |> fold (\g r -> graph_insert_node "Pod" "Pod/default/web" r g) graph_empty`), "\n")
	for _, want := range []string{`"name":"web"`, `"name":"other"`, "different"} {
		if !strings.Contains(got, want) {
			t.Errorf("error %q does not contain %q", got, want)
		}
	}
	if strings.Contains(got, "0x") {
		t.Errorf("a payload printed as a pointer: %q", got)
	}
}

// TestGraphInsertNodeBadArguments: errors as values.
func TestGraphInsertNodeBadArguments(t *testing.T) {
	got := strings.Join(NewInterpreter().evaluatePipelineStr(`graph_insert_node "N" "a" 5`), "\n")
	if !strings.Contains(got, "[ERROR]") || !strings.Contains(got, "graph") {
		t.Errorf("graph_insert_node on a non-graph: %q", got)
	}
}

// TestGraphAdjacencyFunctionsIgnoreUndeclared (F13): graph_in_of/graph_out_of agree
// with graph_nodes — an edge to an id that is no node is not reported.
func TestGraphAdjacencyFunctionsIgnoreUndeclared(t *testing.T) {
	in := NewInterpreter()
	in.EvaluateScript(`def g_ghost = fold (\g n -> graph_add_node "N" n g) graph_empty |> fold (\_ g -> graph_add_edges [["E", 1, "a", "b"], ["E", 1, "a", "ghost"], ["E", 1, "phantom", "a"]] g) graph_empty`)
	for _, row := range []struct{ stmt, want string }{
		{`["a", "b"] |> g_ghost |> fold (\_ g -> graph_out_of "E" "a" g) 0`, "[b]"},
		{`["a", "b"] |> g_ghost |> fold (\_ g -> graph_in_of "E" "a" g) 0`, "[]"},
		{`["a", "b"] |> g_ghost |> fold (\_ g -> graph_in_of "E" "b" g) 0`, "[a]"},
	} {
		got := in.evaluatePipelineStr(row.stmt)
		if len(got) != 1 || got[0] != row.want {
			t.Errorf("%s = %v, want %s", row.stmt, got, row.want)
		}
	}
	// the stored edges are still there: the graph is not rewritten
	if g := graphOf(t, in, `["a", "b"] |> g_ghost`); g.Size() != 3 {
		t.Errorf("Size = %d, want 3 stored edges", g.Size())
	}
}

// TestGraphAddEdgesDeduplicates (F14): one pass that finds the same fact twice adds
// it once; the inferred twin does the same; the single-edge step stays a multigraph.
func TestGraphAddEdgesDeduplicates(t *testing.T) {
	in := NewInterpreter()
	g := graphOf(t, in, `["a", "b"] |> fold (\g n -> graph_add_node "N" n g) graph_empty |> fold (\_ g -> graph_add_edges [["Configures", 1, "a", "b"], ["Configures", 1, "a", "b"], ["Owns", 1, "a", "b"]] g) graph_empty`)
	if g.Size() != 2 {
		t.Errorf("graph_add_edges: Size = %d, want 2 (Configures once, Owns once)", g.Size())
	}
	gi := graphOf(t, in, `["a", "b"] |> fold (\g n -> graph_add_node "N" n g) graph_empty |> fold (\_ g -> graph_add_inferred_edges [["Calls", 0.5, "a", "b"], ["Calls", 0.5, "a", "b"]] g) graph_empty`)
	if gi.Size() != 1 || !gi.Edges()[0].Inferred {
		t.Errorf("graph_add_inferred_edges: %v", gi.Edges())
	}
	multi := graphOf(t, in, `spawn S: x |> graph_empty |> graph_add_edge E 1 a b |> graph_add_edge E 1 a b`)
	if multi.Size() != 2 {
		t.Errorf("the single-edge step must stay a multigraph insert, Size = %d", multi.Size())
	}
}
