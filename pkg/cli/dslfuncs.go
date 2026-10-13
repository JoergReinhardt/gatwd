package cli

import (
	"fmt"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/functions"
)

// This file registers the GENERIC, domain-free builtins that the guest DSL's infix
// expression language can *apply* (via the juxtaposition application added in
// parser/infix.go) from inside a lambda body — most importantly a `fold` reducer.
//
// # Why these exist: fold-as-ingest (the "option C" graph ingest primitive)
//
// The custom left fold `xs |> fold (\acc x -> expr) seed` already threads an
// arbitrary accumulator over a stream/list. The missing capability was letting the
// reducer *build* a value structurally rather than only combine scalars with
// operators. With function application plus the graph-construction functions below,
// the accumulator can be a graph and the reducer can grow it one element at a time:
//
//	resources
//	  |> fold (\g r -> graph_add_node (hamt_path "kind" r) (hamt_path "metadata.name" r) r g) graph_empty
//
// That is the ingest step: a stream/list of records folded into a single persistent
// graph, each record lifted to a node whose payload is the record itself. Edges are
// a second fold over the same list using graph_add_edge / graph_add_edges. No step
// is domain-specific: these are the generic graph-row and HAMT-navigation cells of
// the operator matrix, now reachable as expression-level functions. The Kubernetes
// relation rules are authored in DSL on top of exactly these primitives.
//
// # Calling convention
//
// Every graph constructor takes the graph it extends as its LAST argument, so a
// reducer reads naturally as "add ... to g". Identities and colors are coerced to
// strings (the DSL pins the generic Graph to string id/color); weights pass through
// unchanged (ShortestPath coerces numerics via ToFloat64).

// dslGraph is the single concrete instantiation the dynamically-typed DSL pins the
// generic host Graph to — string identity, string color, any payload, any weight —
// mirroring evaluator's internal alias so constructed graphs flow through the same
// graph_* pipeline steps.
type dslGraph = *collections.Graph[string, string, any, any]

// emptyDSLGraph returns a fresh empty graph; because the graph is an immutable
// persistent structure, the shared graph_empty value below can be safely reused as
// a fold seed across every evaluation.
func emptyDSLGraph() dslGraph {
	return collections.NewGraph[string, string, any, any]()
}

// asDSLGraph coerces a reducer accumulator to a dslGraph, normalising a nil graph to
// a fresh empty one so a never-seeded accumulator still accepts nodes/edges.
func asDSLGraph(v any) (dslGraph, bool) {
	g, ok := v.(dslGraph)
	if !ok {
		return nil, false
	}
	if g == nil {
		return emptyDSLGraph(), true
	}
	return g, true
}

// nodePayloads projects graph nodes to their payloads, the expression-level read
// side of the graph (graph_payloads / graph_payloads_of). It lets a derivation pass
// join across nodes with a comprehension, e.g.
// `[ … | s <- graph_payloads_of "A" g, p <- graph_payloads_of "B" g, … ]`.
func nodePayloads(nodes []functions.Node[string, string, any]) []any {
	out := make([]any, len(nodes))
	for i, n := range nodes {
		out[i] = n.Payload
	}
	return out
}

// dslStr coerces a DSL value to the string carrier the graph uses for identities and
// colors, leaving an existing string untouched and formatting anything else.
func dslStr(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

// dslBuiltinFuncs is the registry of expression-level builtins, consulted as a
// fallback by the fold reducer's and the graph lambdas' name lookups. It is built
// once and never mutated (the functions are pure; graph_empty is an immutable
// value), so it is safe to share across interpreters and goroutines.
var dslBuiltinFuncs = buildDSLBuiltinFuncs()

// buildDSLBuiltinFuncs constructs the registry. See the file header for the calling
// convention and the fold-as-ingest motivation.
func buildDSLBuiltinFuncs() map[string]any {
	m := map[string]any{}

	// graph_empty — the empty-graph VALUE, the canonical fold seed for ingest.
	m["graph_empty"] = any(emptyDSLGraph())

	// graph_add_node color id [payload] g → g with a node added.
	// 3-arg form defaults the payload to the identity (matching the pipeline step);
	// 4-arg form carries an explicit payload (the record, for ingest).
	m["graph_add_node"] = functions.DSLFunc(func(args []any) (any, error) {
		switch len(args) {
		case 3:
			g, ok := asDSLGraph(args[2])
			if !ok {
				return nil, fmt.Errorf("graph_add_node: last argument must be a graph, got %T", args[2])
			}
			id := dslStr(args[1])
			return g.AddNode(id, dslStr(args[0]), any(id)), nil
		case 4:
			g, ok := asDSLGraph(args[3])
			if !ok {
				return nil, fmt.Errorf("graph_add_node: last argument must be a graph, got %T", args[3])
			}
			return g.AddNode(dslStr(args[1]), dslStr(args[0]), args[2]), nil
		default:
			return nil, fmt.Errorf("graph_add_node: want (color id [payload] graph), got %d args", len(args))
		}
	})

	// graph_add_edge color weight from to g → g with a binary edge added.
	m["graph_add_edge"] = functions.DSLFunc(func(args []any) (any, error) {
		if len(args) != 5 {
			return nil, fmt.Errorf("graph_add_edge: want (color weight from to graph), got %d args", len(args))
		}
		g, ok := asDSLGraph(args[4])
		if !ok {
			return nil, fmt.Errorf("graph_add_edge: last argument must be a graph, got %T", args[4])
		}
		return g.Connect(dslStr(args[2]), dslStr(args[3]), dslStr(args[0]), args[1]), nil
	})

	// graph_add_inferred_edge color weight from to g → g with an inferred (assumed,
	// trust-weighted) edge added — the expression form of graph_add_edge_inferred.
	m["graph_add_inferred_edge"] = functions.DSLFunc(func(args []any) (any, error) {
		if len(args) != 5 {
			return nil, fmt.Errorf("graph_add_inferred_edge: want (color weight from to graph), got %d args", len(args))
		}
		g, ok := asDSLGraph(args[4])
		if !ok {
			return nil, fmt.Errorf("graph_add_inferred_edge: last argument must be a graph, got %T", args[4])
		}
		return g.AddInferredEdge(dslStr(args[2]), dslStr(args[3]), dslStr(args[0]), args[1]), nil
	})

	// graph_add_edges specs g → g with every edge in specs added, where specs is a
	// list of 4-element lists [color, weight, from, to]. This is the second-pass
	// edge-ingest shape: a comprehension yields a variable number of edge specs per
	// record and they are folded into the graph in one call.
	m["graph_add_edges"] = functions.DSLFunc(func(args []any) (any, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("graph_add_edges: want (specs graph), got %d args", len(args))
		}
		g, ok := asDSLGraph(args[1])
		if !ok {
			return nil, fmt.Errorf("graph_add_edges: last argument must be a graph, got %T", args[1])
		}
		specs, ok := functions.ToSliceAny(args[0])
		if !ok {
			return nil, fmt.Errorf("graph_add_edges: first argument must be a list of edge specs, got %T", args[0])
		}
		for _, spec := range specs {
			fields, ok := functions.ToSliceAny(spec)
			if !ok || len(fields) != 4 {
				return nil, fmt.Errorf("graph_add_edges: each edge spec must be [color weight from to], got %v", spec)
			}
			g = g.Connect(dslStr(fields[2]), dslStr(fields[3]), dslStr(fields[0]), fields[1])
		}
		return g, nil
	})

	// hamt_get key payload → the value at key, or an error — the expression form of
	// the hamt_get pipeline step, used by relation predicates to read a field.
	m["graph_payloads"] = functions.DSLFunc(func(args []any) (any, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("graph_payloads: want (graph), got %d args", len(args))
		}
		g, ok := asDSLGraph(args[0])
		if !ok {
			return nil, fmt.Errorf("graph_payloads: argument must be a graph, got %T", args[0])
		}
		return nodePayloads(g.Nodes()), nil
	})
	m["graph_payloads_of"] = functions.DSLFunc(func(args []any) (any, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("graph_payloads_of: want (color graph), got %d args", len(args))
		}
		g, ok := asDSLGraph(args[1])
		if !ok {
			return nil, fmt.Errorf("graph_payloads_of: last argument must be a graph, got %T", args[1])
		}
		return nodePayloads(g.NodesByColor(dslStr(args[0]))), nil
	})

	// graph_in_of / graph_out_of color id g → the identities of the nodes joined to id
	// by an incoming / outgoing edge of that color. The expression form of the
	// reverse/forward adjacency, so a comprehension can ask "is anything of this
	// color pointing at x?" (an empty list means no).
	type graphEdge = functions.Edge[string, string, any]
	for _, dir := range []struct {
		name  string
		edges func(dslGraph, string) []graphEdge
		end   func(graphEdge) string
	}{
		{"graph_in_of", func(g dslGraph, id string) []graphEdge { return g.InEdges(id) }, func(e graphEdge) string { return e.From }},
		{"graph_out_of", func(g dslGraph, id string) []graphEdge { return g.OutEdges(id) }, func(e graphEdge) string { return e.To }},
	} {
		m[dir.name] = functions.DSLFunc(func(args []any) (any, error) {
			if len(args) != 3 {
				return nil, fmt.Errorf("%s: want (color id graph), got %d args", dir.name, len(args))
			}
			g, ok := asDSLGraph(args[2])
			if !ok {
				return nil, fmt.Errorf("%s: last argument must be a graph, got %T", dir.name, args[2])
			}
			color := dslStr(args[0])
			out := []any{}
			for _, e := range dir.edges(g, dslStr(args[1])) {
				if e.Color == color {
					out = append(out, dir.end(e))
				}
			}
			return out, nil
		})
	}

	m["hamt_has_path"] = functions.DSLFunc(func(args []any) (any, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("hamt_has_path: want (path payload), got %d args", len(args))
		}
		_, err := evaluator.NavigateHAMTPath(args[1], dslStr(args[0]))
		return err == nil, nil
	})

	m["hamt_get"] = functions.DSLFunc(func(args []any) (any, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("hamt_get: want (key payload), got %d args", len(args))
		}
		key := dslStr(args[0])
		v, found, navigable := evaluator.HAMTGetKey(args[1], key)
		if !navigable {
			return nil, fmt.Errorf("hamt_get: value of type %T is not a Record/HAMT/Map", args[1])
		}
		if !found {
			return nil, fmt.Errorf("hamt_get: key '%s' not found", key)
		}
		return v, nil
	})

	// hamt_path "a.b.c" payload → the value at the dotted path, or an error — the
	// expression form of the hamt_path step (deep "functional struct field" access).
	m["hamt_path"] = functions.DSLFunc(func(args []any) (any, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("hamt_path: want (path payload), got %d args", len(args))
		}
		return evaluator.NavigateHAMTPath(args[1], dslStr(args[0]))
	})

	// subset_of sub super → bool: is sub ⊆ super? Record/HAMT containment (every
	// binding of sub holds in super) or Set/collection membership — the generic
	// primitive a relation predicate uses for "selector ⊆ labels". See
	// evaluator.IsSubsetValue for the semantics.
	m["subset_of"] = functions.DSLFunc(func(args []any) (any, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("subset_of: want (sub super), got %d args", len(args))
		}
		return evaluator.IsSubsetValue(args[0], args[1])
	})

	registerMatrixFuncs(m)
	registerGraphFuncs(m)
	registerCharFuncs(m)
	registerStrFuncs(m)
	registerBoolFuncs(m)
	registerEncodingFuncs(m)
	registerJSONFuncs(m)
	registerRegexFuncs(m)

	// Fixed-arity builtins become curried Closures, so `map (hamt_get "a") rs` or
	// `fold (graph_add_edge Owns 1 "a") …` partially apply like every user function.
	// graph_add_node is variadic (3 or 4 args) and stays a raw DSLFunc.
	for name, arity := range map[string]int{
		"graph_add_edge": 5, "graph_add_inferred_edge": 5, "graph_add_edges": 2,
		"graph_payloads": 1, "graph_payloads_of": 2, "graph_in_of": 3, "graph_out_of": 3,
		"hamt_has_path": 2, "hamt_get": 2, "hamt_path": 2, "subset_of": 2,
	} {
		raw := m[name].(functions.DSLFunc)
		params := make([]string, arity)
		for idx := range params {
			params[idx] = fmt.Sprintf("a%d", idx+1)
		}
		m[name] = functions.NewClosure(params, func(a []any) (any, error) { return raw(a) })
	}

	return m
}

// dslBuiltinLookup resolves a name against the expression-level builtin registry.
// Interpreter lookups consult it as a low-precedence fallback, so a user binding of
// the same name (a let/const/variable) always wins.
func dslBuiltinLookup(name string) (any, bool) {
	v, ok := dslBuiltinFuncs[name]
	return v, ok
}
