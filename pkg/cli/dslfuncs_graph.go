package cli

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/functions"
)

// payloadJSON renders a node payload for an error message: JSON when the value
// encodes (resources are nested maps), otherwise its default text.
func payloadJSON(v any) string {
	var b strings.Builder
	if err := encodeJSON(&b, v, "payload"); err != nil {
		return fmt.Sprint(v)
	}
	return b.String()
}

// anyTreeFromGraph reboxes a *collections.Tree[string] (collections.Graph.Unroll's
// result type) as *collections.Tree[any], the single carrier every tree_* DSL
// builtin shares (dslfuncs_matrix.go), so graph_unroll's result interops with
// tree_value, tree_children, visualize and the rest without a special case.
func anyTreeFromGraph(t *collections.Tree[string]) *collections.Tree[any] {
	kids := t.Children().Elements()
	list := collections.Nil[*collections.Tree[any]]()
	for i := len(kids) - 1; i >= 0; i-- {
		list = collections.Cons(anyTreeFromGraph(kids[i].(*collections.Tree[string])), list)
	}
	return collections.NewTree[any](t.Value(), list)
}

// registerGraphFuncs adds the READ-side and inferred-edge function forms of the graph
// row of the operator-type matrix, complementing the constructors and accessors in
// dslfuncs.go. They are the expression-level twins of the graph_* pipeline steps and
// take the graph as their LAST argument, like every graph function:
//
//	graph_add_inferred_edges specs g   twin of graph_add_edge_inferred, over a list of
//	                                   [color weight from to] specs (assumed, trust-weighted)
//	graph_insert_node c id [p] g       graph_add_node that REFUSES an id already in g (duplicate-id error)
//	graph_has_node id g                is id a node of g?
//	graph_nodes g                      node identities, sorted
//	graph_order g / graph_size g       node / edge counts
//	graph_edge_pairs color g           [from to] pairs of every edge of that color
//	graph_reachable id g               ids reachable from id (including id), sorted
//	graph_reachable_to id g            ids that can reach id (including id), sorted — Reachable's mirror
//	graph_out_degree id g              number of edges leaving id
//	graph_in_degree id g               number of edges entering id
//	graph_degree id g                  id's total degree (in + out)
//	graph_unroll id g                  id's reachable subgraph rendered as a Tree, duplicating a node
//	                                   every time a path (or a cycle) reaches it again
//	graph_shortest_path from to g      node ids of a least-weight path, [] when none
//	graph_topo_sort g                  a dependency order, [] when the graph is cyclic
//	graph_scc g                        strongly connected components (lists of ids)
//
// With these a rule can run a whole-graph algorithm INSIDE a comprehension —
// `[ c | c <- graph_scc g, length c > 1 ]` finds dependency cycles — rather than
// only at the end of a pipeline.
func registerGraphFuncs(m map[string]any) {
	registerSourceFuncs(m) // provenance of the payloads graph_insert_node reports on (source_funcs.go)
	graphArg := func(v any) (dslGraph, error) {
		g, ok := asDSLGraph(v)
		if !ok {
			return nil, fmt.Errorf("last argument must be a graph, got %T", v)
		}
		return g, nil
	}
	idList := func(ids []string) []any {
		out := make([]any, len(ids))
		for idx, id := range ids {
			out[idx] = id
		}
		return out
	}

	defn(m, "graph_add_inferred_edges", "specs g", func(a []any) (any, error) {
		g, err := graphArg(a[1])
		if err != nil {
			return nil, err
		}
		specs, ok := functions.ToSliceAny(a[0])
		if !ok {
			return nil, fmt.Errorf("first argument must be a list of edge specs, got %T", a[0])
		}
		edges := make([]functions.Edge[string, string, any], 0, len(specs))
		for _, spec := range specs {
			f, ok := functions.ToSliceAny(spec)
			if !ok || len(f) != 4 {
				return nil, fmt.Errorf("each edge spec must be [color weight from to], got %v", spec)
			}
			ids, err := idStrs(f[2], f[3], f[0])
			if err != nil {
				return nil, err
			}
			edges = append(edges, functions.NewInferredEdge[string, string, any](ids[0], ids[1], ids[2], f[1]))
		}
		return g.AddEdgesOnce(edges), nil // identical specs once, like graph_add_edges
	})
	// graph_insert_node color id [payload] g — the strict twin of graph_add_node: an id
	// that is already a node is an error naming both declarations, not an overwrite.
	// Like graph_add_node it takes 3 (payload = id) or 4 arguments and stays a raw
	// DSLFunc, so the step form `graph_insert_node COLOR ID` and the function form agree.
	m["graph_insert_node"] = functions.DSLFunc(func(args []any) (any, error) {
		if len(args) != 3 && len(args) != 4 {
			return nil, fmt.Errorf("graph_insert_node: want (color id [payload] graph), got %d args", len(args))
		}
		g, err := graphArg(args[len(args)-1])
		if err != nil {
			return nil, fmt.Errorf("graph_insert_node: %w", err)
		}
		ids, err := idStrs(args[0], args[1])
		if err != nil {
			return nil, fmt.Errorf("graph_insert_node: %w", err)
		}
		color, id := ids[0], ids[1]
		var payload any = id
		if len(args) == 4 {
			payload = args[2]
		}
		res, err := g.InsertNode(id, color, payload)
		var dup *collections.DuplicateNodeError
		if errors.As(err, &dup) {
			dup.Render = payloadJSON // a nested-map payload prints as JSON, not as a pointer
		}
		if err != nil {
			return nil, fmt.Errorf("graph_insert_node: %w", err)
		}
		return res, nil
	})
	defn(m, "graph_has_node", "id g", func(a []any) (any, error) {
		g, err := graphArg(a[1])
		if err != nil {
			return nil, err
		}
		id, err := idStr(a[0])
		if err != nil {
			return nil, err
		}
		return g.HasNode(id), nil
	})
	defn(m, "graph_nodes", "g", func(a []any) (any, error) {
		g, err := graphArg(a[0])
		if err != nil {
			return nil, err
		}
		ids := g.NodeIDs()
		sort.Strings(ids)
		return listOf(idList(ids)), nil
	})
	defn(m, "graph_order", "g", func(a []any) (any, error) {
		g, err := graphArg(a[0])
		if err != nil {
			return nil, err
		}
		return g.Order(), nil
	})
	defn(m, "graph_size", "g", func(a []any) (any, error) {
		g, err := graphArg(a[0])
		if err != nil {
			return nil, err
		}
		return g.Size(), nil
	})
	defn(m, "graph_edge_pairs", "color g", func(a []any) (any, error) {
		g, err := graphArg(a[1])
		if err != nil {
			return nil, err
		}
		var out []any
		color, err := idStr(a[0])
		if err != nil {
			return nil, err
		}
		for _, e := range g.EdgesByColor(color) {
			out = append(out, listOf([]any{e.From, e.To}))
		}
		return listOf(out), nil
	})
	defn(m, "graph_reachable", "id g", func(a []any) (any, error) {
		g, err := graphArg(a[1])
		if err != nil {
			return nil, err
		}
		var ids []string
		from, err := idStr(a[0])
		if err != nil {
			return nil, err
		}
		for _, id := range g.Reachable(from).Elements() {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		return listOf(idList(ids)), nil
	})
	defn(m, "graph_reachable_to", "id g", func(a []any) (any, error) {
		g, err := graphArg(a[1])
		if err != nil {
			return nil, err
		}
		target, err := idStr(a[0])
		if err != nil {
			return nil, err
		}
		var ids []string
		for _, id := range g.ReachableTo(target).Elements() {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		return listOf(idList(ids)), nil
	})
	defn(m, "graph_out_degree", "id g", func(a []any) (any, error) {
		g, err := graphArg(a[1])
		if err != nil {
			return nil, err
		}
		id, err := idStr(a[0])
		if err != nil {
			return nil, err
		}
		return g.OutDegree(id), nil
	})
	defn(m, "graph_in_degree", "id g", func(a []any) (any, error) {
		g, err := graphArg(a[1])
		if err != nil {
			return nil, err
		}
		id, err := idStr(a[0])
		if err != nil {
			return nil, err
		}
		return g.InDegree(id), nil
	})
	defn(m, "graph_degree", "id g", func(a []any) (any, error) {
		g, err := graphArg(a[1])
		if err != nil {
			return nil, err
		}
		id, err := idStr(a[0])
		if err != nil {
			return nil, err
		}
		return g.Degree(id), nil
	})
	defn(m, "graph_unroll", "id g", func(a []any) (any, error) {
		g, err := graphArg(a[1])
		if err != nil {
			return nil, err
		}
		id, err := idStr(a[0])
		if err != nil {
			return nil, err
		}
		return anyTreeFromGraph(g.Unroll(id)), nil
	})
	defn(m, "graph_shortest_path", "from to g", func(a []any) (any, error) {
		g, err := graphArg(a[2])
		if err != nil {
			return nil, err
		}
		ends, err := idStrs(a[0], a[1])
		if err != nil {
			return nil, err
		}
		path, _, ok := g.ShortestPath(ends[0], ends[1])
		if !ok {
			return listOf(nil), nil
		}
		return listOf(idList(path)), nil
	})
	defn(m, "graph_topo_sort", "g", func(a []any) (any, error) {
		g, err := graphArg(a[0])
		if err != nil {
			return nil, err
		}
		order, isOrder := g.TopoSort().UnwrapRight()
		if !isOrder {
			return listOf(nil), nil
		}
		return listOf(idList(order)), nil
	})
	defn(m, "graph_scc", "g", func(a []any) (any, error) {
		g, err := graphArg(a[0])
		if err != nil {
			return nil, err
		}
		var out []any
		for _, comp := range g.SCC() {
			out = append(out, listOf(idList(comp)))
		}
		return listOf(out), nil
	})
}
