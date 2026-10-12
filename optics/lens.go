package optics

import (
	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/functions"
)

// Lens provides functional getters and setters for deep, immutable structures.
//
// In a purely functional architecture (like our immutable AST Tree), you cannot
// mutate a field directly (e.g., node.Value = X). A Lens encapsulates the logic
// to "focus" on a specific field A inside a structure S.
//
// It provides:
// 1. Get: Extract the focused value A from S.
// 2. Set: Return a BRAND NEW copy of S with the focused value updated to A.
type Lens[S, A any] struct {
	Get func(S) A
	Set func(S, A) S
}

// ComposeLenses mathematically composes two lenses.
// If you have a Lens from S to A, and a Lens from A to B, 
// you can compose them into a Lens from S to B.
// This allows you to chain deep modifications safely!
func ComposeLenses[S, A, B any](l1 Lens[S, A], l2 Lens[A, B]) Lens[S, B] {
	return Lens[S, B]{
		Get: func(s S) B {
			return l2.Get(l1.Get(s))
		},
		Set: func(s S, b B) S {
			// Get the inner A, modify it with l2, then set the new A into S with l1
			a := l1.Get(s)
			newA := l2.Set(a, b)
			return l1.Set(s, newA)
		},
	}
}

// RecordLens creates a pure functional Lens focusing on a specific field of a Record (map[string]any).
// Setting a value returns a brand new shallow copy with the field updated, preserving immutability.
func RecordLens(field string) Lens[map[string]any, any] {
	return Lens[map[string]any, any]{
		Get: func(m map[string]any) any {
			if m == nil {
				return nil
			}
			return m[field]
		},
		Set: func(m map[string]any, val any) map[string]any {
			copied := make(map[string]any, len(m)+1)
			for k, v := range m {
				copied[k] = v
			}
			copied[field] = val
			return copied
		},
	}
}

// MapLens creates a pure functional Lens focusing on a key K in a standard map[K]V.
func MapLens[K comparable, V any](key K) Lens[map[K]V, V] {
	return Lens[map[K]V, V]{
		Get: func(m map[K]V) V {
			if m == nil {
				var zero V
				return zero
			}
			return m[key]
		},
		Set: func(m map[K]V, val V) map[K]V {
			copied := make(map[K]V, len(m)+1)
			for k, v := range m {
				copied[k] = v
			}
			copied[key] = val
			return copied
		},
	}
}

// HAMTLens creates a pure functional Lens focusing on key K in a persistent collections.HAMT[K, V].
// Updating a field returns a brand new HAMT with O(log32 N) structural sharing.
func HAMTLens[K comparable, V any](key K) Lens[*collections.HAMT[K, V], functions.Maybe[V]] {
	return Lens[*collections.HAMT[K, V], functions.Maybe[V]]{
		Get: func(h *collections.HAMT[K, V]) functions.Maybe[V] {
			if h == nil {
				return functions.Nothing[V]()
			}
			val, ok := h.Get(key)
			if !ok {
				return functions.Nothing[V]()
			}
			return functions.Just(val)
		},
		Set: func(h *collections.HAMT[K, V], mVal functions.Maybe[V]) *collections.HAMT[K, V] {
			if h == nil {
				h = collections.NewHAMT[K, V]()
			}
			if mVal.IsJust() {
				v, _ := mVal.Unwrap()
				return h.Set(key, v)
			}
			return h.Delete(key)
		},
	}
}
