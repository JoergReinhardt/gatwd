package functions

import "fmt"

// This file defines the two atoms of the generic graph abstraction: the typed
// ("colored") Node and the typed, weighted, directional Edge. They live in
// Layer 0 (functions/) because they are pure algebraic data types with no
// dependency on the persistent collections that assemble them into a Graph
// (collections/graph.go, Layer 1).
//
// # Why "color"?
//
// A plain graph has anonymous vertices and edges. A *typed* (or "colored")
// graph tags every node and edge with a discriminator so that the same
// structure can hold heterogeneous relations — a Service node, a Pod node, a
// "selects" edge, an "owns" edge — and be queried by type. Mathematically the
// color is an element of some finite labelling set; here it is simply a generic
// comparable type parameter C.
//
// Keeping the color generic (rather than hard-coding the DSL's runtime Variant)
// preserves the host/guest split required by the project: the core engine is
// expressed in Go generics, and the dynamically-typed DSL layer instantiates C
// with whatever tag representation it uses (a Variant tag string, say) when it
// wires the graph through to Layer 5. Unit colors (a tag with a nil payload) are
// the common case and are always safely comparable.

// Node is a single typed vertex of a graph: an identity, a color, and a payload.
//
//   - K is the node's identity type and must be comparable, because the Graph
//     collection keys its adjacency maps (HAMTs) by node identity.
//   - C is the node's color/type tag and must be comparable, so nodes can be
//     indexed and grouped by color (e.g. "all Pod nodes").
//   - N is the free-form payload carried at the node (labels, spec, metadata);
//     it is unconstrained because the graph never needs to compare payloads to
//     maintain its structure — only the caller does, through Equal.
//
// A node's "type" is nothing more than its Color: there is no separate enum.
type Node[K comparable, C comparable, N any] struct {
	ID      K // identity; unique within a graph
	Color   C // type/color discriminator
	Payload N // free-form associated data
}

// NewNode constructs a colored node with the given identity, color, and payload.
func NewNode[K comparable, C comparable, N any](id K, color C, payload N) Node[K, C, N] {
	return Node[K, C, N]{ID: id, Color: color, Payload: payload}
}

// Equal reports whether two nodes are structurally equal. Identity and color are
// compared with Go's native == (both are comparable); the payload is compared
// with StructuralEqual so that nested collections and ADTs compare by value
// rather than by pointer.
func (n Node[K, C, N]) Equal(other Node[K, C, N]) bool {
	if n.ID != other.ID || n.Color != other.Color {
		return false
	}
	return StructuralEqual(n.Payload, other.Payload)
}

// NodeElements exposes the node's identity, color, and payload as untyped values
// for reflection-free structural inspection by the operator layer.
func (n Node[K, C, N]) NodeElements() (any, any, any) {
	return n.ID, n.Color, n.Payload
}

// IsNode is a marker method for type introspection by the dynamic DSL layer.
func (n Node[K, C, N]) IsNode() bool { return true }

// String implements fmt.Stringer, rendering the node as color#id.
func (n Node[K, C, N]) String() string {
	return fmt.Sprintf("%v#%v", n.Color, n.ID)
}

// Edge is a single typed, weighted, directional connection between two nodes.
//
//   - K is the endpoint identity type (matching the nodes it connects).
//   - C is the edge's color/type tag (e.g. "selects", "owns", "routes-to").
//   - E is the weight/payload type. The default weight 1 (int) makes unweighted
//     walks behave correctly; richer E (a cost, a struct) enables weighted paths.
//
// Edges are directional by construction (From → To). An undirected relation is
// modelled as two opposing directed edges — this keeps every graph algorithm
// (BFS/DFS/Dijkstra/SCC) defined on the standard binary digraph.
//
// Two small fields fall out of the cardinality design:
//
//   - Group ties together the several binary edges that together form one
//     logical n-ary relation (e.g. a Service that "selects" three Pods is three
//     Selects edges sharing one group id). An empty Group means an ungrouped 1:1
//     edge.
//   - Inferred marks an edge that was *assumed* rather than read from the source
//     data (e.g. "this Pod consumes that Service because a ConfigMap value equals
//     the Service name"). Walks and compliance folds can trust-weight, down-rank,
//     or exclude inferred edges explicitly.
type Edge[K comparable, C comparable, E any] struct {
	From     K      // source endpoint identity
	To       K      // destination endpoint identity
	Color    C      // type/color discriminator
	Weight   E      // weight/payload (default 1 for unweighted graphs)
	Group    string // optional: groups the binary edges of one n-ary relation
	Inferred bool   // true for assumed edges not read from the source data
}

// NewEdge constructs a directed, colored, weighted edge with no group and the
// inferred flag cleared.
func NewEdge[K comparable, C comparable, E any](from, to K, color C, weight E) Edge[K, C, E] {
	return Edge[K, C, E]{From: from, To: to, Color: color, Weight: weight}
}

// NewGroupedEdge constructs a directed edge tagged with a group id, used to tie
// together the binary edges that make up one logical fan-out / fan-in relation.
func NewGroupedEdge[K comparable, C comparable, E any](from, to K, color C, weight E, group string) Edge[K, C, E] {
	return Edge[K, C, E]{From: from, To: to, Color: color, Weight: weight, Group: group}
}

// NewInferredEdge constructs a directed edge marked as inferred/assumed (see the
// Edge.Inferred field) with the given weight.
func NewInferredEdge[K comparable, C comparable, E any](from, to K, color C, weight E) Edge[K, C, E] {
	return Edge[K, C, E]{From: from, To: to, Color: color, Weight: weight, Inferred: true}
}

// Equal reports whether two edges are structurally equal across all of their
// fields. Endpoints and color compare with native ==; the weight compares with
// StructuralEqual (so numeric and nested weights compare by value).
func (e Edge[K, C, E]) Equal(other Edge[K, C, E]) bool {
	if e.From != other.From || e.To != other.To || e.Color != other.Color {
		return false
	}
	if e.Group != other.Group || e.Inferred != other.Inferred {
		return false
	}
	return StructuralEqual(e.Weight, other.Weight)
}

// Compare gives edges a total, deterministic ordering so that the Graph
// collection can canonicalise its adjacency lists (making order-independent
// equality and reproducible iteration possible). Edges order lexicographically
// by (From, To, Color, Weight), falling back to a formatted-string comparison
// for any field whose dynamic type the numeric/string Compare cannot order.
func (e Edge[K, C, E]) Compare(other Edge[K, C, E]) int {
	if c := compareField(e.From, other.From); c != 0 {
		return c
	}
	if c := compareField(e.To, other.To); c != 0 {
		return c
	}
	if c := compareField(e.Color, other.Color); c != 0 {
		return c
	}
	if c := compareField(e.Weight, other.Weight); c != 0 {
		return c
	}
	if c := compareField(e.Group, other.Group); c != 0 {
		return c
	}
	switch {
	case !e.Inferred && other.Inferred:
		return -1
	case e.Inferred && !other.Inferred:
		return 1
	}
	return 0
}

// compareField orders two untyped values using the numeric/string-aware Compare,
// degrading to a stable formatted-string comparison when Compare cannot order
// them (e.g. two values of an exotic color type).
func compareField(a, b any) int {
	if c, err := Compare(a, b); err == nil {
		return c
	}
	sa, sb := fmt.Sprintf("%v", a), fmt.Sprintf("%v", b)
	switch {
	case sa < sb:
		return -1
	case sa > sb:
		return 1
	}
	return 0
}

// EdgeElements exposes the edge's fields as untyped values for reflection-free
// structural inspection by the operator layer.
func (e Edge[K, C, E]) EdgeElements() (any, any, any, any, string, bool) {
	return e.From, e.To, e.Color, e.Weight, e.Group, e.Inferred
}

// IsEdge is a marker method for type introspection by the dynamic DSL layer.
func (e Edge[K, C, E]) IsEdge() bool { return true }

// String implements fmt.Stringer, rendering the edge as from -color-> to, with a
// trailing ~ for inferred edges.
func (e Edge[K, C, E]) String() string {
	s := fmt.Sprintf("%v -%v-> %v", e.From, e.Color, e.To)
	if e.Inferred {
		s += "~"
	}
	return s
}
