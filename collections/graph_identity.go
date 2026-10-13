package collections

import (
	"fmt"
	"strings"

	"github.com/joergreinhardt/gatwd/functions"
)

// DuplicateNodeError is returned by Graph.InsertNode when a node with the same
// identity is already declared.
//
// # Why this exists
//
// A graph identifies a node by its id alone, so two different things that were
// given the same id are *one* node as far as the graph is concerned. The
// upserting WithNode/AddNode make that the rule ("the later declaration wins"),
// which is exactly right for an incremental update and exactly wrong for
// ingest: two resources that collide on their id would silently turn into one,
// and every relation derived afterwards would be about a thing that does not
// exist. InsertNode is the strict twin — it refuses — and this error carries
// what a person needs to tell the two declarations apart: the id, and the
// color and payload of both.
//
// The identity scheme itself (what an id is made of) belongs to the domain:
// the convention is "every adapter names nodes Kind/scope/name" so that equal
// local names in different kinds or scopes are different ids and a duplicate
// really is the same thing listed twice.
//
// The fields are untyped (any) so one error type serves every Graph
// instantiation; compare with errors.As:
//
//	var dup *collections.DuplicateNodeError
//	if errors.As(err, &dup) { fmt.Println(dup.ID) }
type DuplicateNodeError struct {
	ID              any // the identity both declarations claim
	ExistingColor   any // color of the node already in the graph
	NewColor        any // color of the rejected node
	ExistingPayload any // payload of the node already in the graph
	NewPayload      any // payload of the rejected node

	// Render, when set, turns a payload into the text Error quotes. The collection
	// layer cannot know how to print an arbitrary payload (a nested map prints as a
	// pointer); the layer that does — the DSL, which can encode a payload as JSON —
	// sets it before the error is shown. The default is fmt.Sprint.
	Render func(payload any) string
}

// maxPayloadText bounds how much of a payload the error message quotes.
const maxPayloadText = 160

// Error names the id, both colors and a bounded rendering of both payloads, and
// says when the two declarations are identical (the same resource listed twice).
func (e *DuplicateNodeError) Error() string {
	show := e.Render
	if show == nil {
		show = func(p any) string { return fmt.Sprint(p) }
	}
	ex, nw := clip(show(e.ExistingPayload)), clip(show(e.NewPayload))
	what := "different"
	if e.ExistingColor == e.NewColor && functions.StructuralEqual(e.ExistingPayload, e.NewPayload) {
		what = "identical"
	}
	return fmt.Sprintf("duplicate node id %q: declared as %v (%s) and again as %v (%s) — the two declarations are %s",
		fmt.Sprint(e.ID), e.ExistingColor, ex, e.NewColor, nw, what)
}

func clip(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > maxPayloadText {
		return s[:maxPayloadText] + "…"
	}
	return s
}

// InsertNode returns a new graph with the node added, or a *DuplicateNodeError
// (and a nil graph) when a node with the same identity is already declared.
//
// It is the strict counterpart of AddNode/WithNode, which silently replace the
// payload and color of an existing node. Use InsertNode wherever a second
// declaration of an id is a mistake in the data rather than an update — above
// all when lifting a stream of records into a graph. The receiver is never
// modified, whichever way the call goes.
func (g *Graph[K, C, N, E]) InsertNode(id K, color C, payload N) (*Graph[K, C, N, E], error) {
	if existing, ok := g.GetNode(id); ok {
		return nil, &DuplicateNodeError{
			ID: id, ExistingColor: existing.Color, NewColor: color,
			ExistingPayload: existing.Payload, NewPayload: payload,
		}
	}
	return g.AddNode(id, color, payload), nil
}

// AddEdgesOnce returns a new graph with every edge added that the graph does not
// already hold: an edge Equal to an earlier edge of the same batch, or to an edge
// already in the graph, is skipped.
//
// AddEdge/Connect keep the multigraph meaning (a second identical edge is a
// second edge, so a caller can count parallel relations). A *derivation pass*
// wants the opposite: when a relation can be found twice — a ConfigMap
// referenced by a volume and by an environment variable of the same pod — the
// fact "A configures B" is still one fact. AddEdgesOnce is the set-like bulk
// insert for that case, so running a pass twice is also harmless. Edges that
// differ in color, weight, group or inferred flag are different facts and are all
// kept.
func (g *Graph[K, C, N, E]) AddEdgesOnce(edges []functions.Edge[K, C, E]) *Graph[K, C, N, E] {
	res := g.norm()
	for _, e := range edges {
		held := false
		for _, have := range edgesOf(mustList(res.out, e.From)) {
			if have.Equal(e) {
				held = true
				break
			}
		}
		if !held {
			res = res.AddEdge(e)
		}
	}
	return res
}
