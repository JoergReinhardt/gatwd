package cli

import (
	"fmt"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/functions"
)

// registerSourceFuncs adds the function forms of value provenance (collections.Source): the
// origin a record was read from, carried beside the record and never inside it.
//
//	with_source origin document h   h, remembering it was read from origin (document is its
//	                                1-based position in origin, 0 when origin holds one)
//	source_of v                     Just "origin, document N" when v remembers a source,
//	                                Nothing otherwise
//
// The source is invisible to everything else: `==`, `hamt_get`, `hamt_keys`, `json_encode`
// and the printed form of h are exactly what they are without it. Only the duplicate-node
// error (graph_insert_node) and source_of look. A HAMT is the only carrier today — it is the
// shape of a record — so any other value is an error from with_source and Nothing from
// source_of. Both are ordinary function forms, so `h |> source_of` works as a step too.
func registerSourceFuncs(m map[string]any) {
	defn(m, "with_source", "origin document h", func(a []any) (any, error) {
		origin, ok := a[0].(string)
		if !ok {
			return nil, fmt.Errorf("origin must be a String, got %s", functions.KindName(a[0]))
		}
		doc, err := strictInt(a[1])
		if err != nil {
			return nil, fmt.Errorf("document: %w", err)
		}
		if doc < 0 {
			return nil, fmt.Errorf("document must be 0 (the origin holds one) or a 1-based position, got %d", doc)
		}
		h, err := asAnyHAMT(a[2])
		if err != nil {
			return nil, err
		}
		return h.WithSource(collections.Source{Origin: origin, Document: doc}), nil
	})
	defn(m, "source_of", "v", func(a []any) (any, error) {
		if src, ok := collections.SourceOf(a[0]); ok {
			return functions.Just[any](src.String()), nil
		}
		return functions.Nothing[any](), nil
	})
}
