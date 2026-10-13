package cli

import (
	"fmt"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/functions"
)

// This file holds the VALUE-level, carrier-polymorphic algebra behind the
// expression-level function forms (map, filter, fold, flatmap, bimap): given a
// runtime value and a callable it applies the Functor / Filterable / Foldable /
// Monad action of whichever carrier the value is — List, Chunk, Set, HAMT (over
// values), Tree, Pair (over the second component), Maybe, Either (over Right),
// Variant (over the payload) or a plain slice.
//
// It is the value-level counterpart of the stream-level mapStepOver/filterStepOver
// pipeline steps: the stream steps lift an element function over each stream node,
// these functions are what a lambda body or comprehension calls directly. Element
// types are preserved — nothing is stringified — and every callable failure is
// returned as an error (the DSL's Left channel), never swallowed.

// listOf builds an immutable cons list from items, preserving order.
func listOf(items []any) *collections.List[any] {
	res := collections.Nil[any]()
	for idx := len(items) - 1; idx >= 0; idx-- {
		res = collections.Cons(items[idx], res)
	}
	return res
}

// ---- the Char / String conversion table (D1) --------------------------------------
//
// Conversion is on demand and narrow; it happens only where a function or operator
// names the type it needs, never silently elsewhere:
//
//	wanted        given                          result
//	Char          Char                           itself
//	Char          String of exactly one char     that Char
//	String        String                         itself
//	String        List/Chunk/slice of Chars      the joined String (an empty collection too)
//	[Char]        String                         the Chars of the String
//
// Anything else — a longer String where a Char is wanted, a collection holding a
// non-Char — is an error that names the offending element.

// asChar reads v as a Char, converting a one-character String on demand.
func asChar(v any) (functions.Char, error) {
	switch c := v.(type) {
	case functions.Char:
		return c, nil
	case string:
		if ch, ok := functions.CharOfString(c); ok {
			return ch, nil
		}
		return 0, fmt.Errorf("want a Char, got the String %q (only a one-character String converts)", c)
	}
	return 0, fmt.Errorf("want a Char, got %T", v)
}

// asString reads v as a String: a String itself, or a List/Chunk/slice of Chars spelling
// one (the empty collection spells ""). A collection holding a non-Char is an error that
// names the first offending element.
func asString(v any) (string, error) {
	if s, ok := v.(string); ok {
		return s, nil
	}
	items, ok := functions.ToSliceAny(v)
	if !ok {
		return "", fmt.Errorf("want a String, got %T", v)
	}
	for idx, it := range items {
		if !functions.IsChar(it) {
			return "", fmt.Errorf("want a String or a collection of Chars; element %d is %v (%T)", idx, it, it)
		}
	}
	s, _ := functions.CharsAsString(v)
	return s, nil
}

// charList lists the Chars of a String as a List (the String → collection-of-Chars row).
func charList(s string) (*collections.List[any], error) {
	cs, err := functions.CharsOfString(s)
	if err != nil {
		return nil, err
	}
	items := make([]any, len(cs))
	for idx, c := range cs {
		items[idx] = c
	}
	return listOf(items), nil
}

// seqOf is functions.ToSliceAny extended with the String row: a String is the sequence of
// its Chars. It is for the expression function forms only — in a pipeline a String node is
// one element of the stream, so the stream steps never call it.
func seqOf(v any) ([]any, bool) {
	if s, ok := v.(string); ok {
		l, err := charList(s)
		if err != nil {
			return nil, false
		}
		return functions.ToSliceAny(l)
	}
	return functions.ToSliceAny(v)
}

// textResult gives a carrier-preserving result for an operation applied to a String: a
// sequence of Chars is spelled back as a String, so filter/reverse/tail/cons on a String
// answer with a String.
func textResult(res any) any {
	if s, ok := functions.CharsAsString(res); ok {
		return s
	}
	return res
}

// call1 applies a callable to one argument.
func call1(fn, x any) (any, error) { return functions.Apply(fn, []any{x}) }

// callPred applies a callable predicate and requires a boolean result.
func callPred(fn, x any) (bool, error) {
	r, err := call1(fn, x)
	if err != nil {
		return false, err
	}
	b, ok := r.(bool)
	if !ok {
		return false, fmt.Errorf("predicate returned %T, want bool", r)
	}
	return b, nil
}

// requireCallable reports a typed error when v cannot be applied.
func requireCallable(name string, v any) error {
	if !functions.IsCallable(v) {
		return fmt.Errorf("%s: first argument must be a function, got %T", name, v)
	}
	return nil
}

// elementsOf extracts the element sequence of any collection-like value for
// iteration: HAMT yields its values, Tree its depth-first values, Maybe/Either their
// 0-or-1 payload, anything else whatever functions.ToSliceAny recognises.
func elementsOf(v any) ([]any, bool) {
	switch c := v.(type) {
	case *collections.HAMT[string, any]:
		var out []any
		for val := range c.Values() {
			out = append(out, val)
		}
		return out, true
	case *collections.Tree[any]:
		var out []any
		for val := range c.WalkDepthFirst() {
			out = append(out, val)
		}
		return out, true
	case *collections.BST[any]:
		var out []any
		for val := range c.Seq() {
			out = append(out, val)
		}
		return out, true
	case *collections.BST[string]:
		var out []any
		for val := range c.Seq() {
			out = append(out, val)
		}
		return out, true
	case dsltypes.Variant:
		_, _, payload := c.VariantElements()
		return []any{payload}, true
	case string:
		return seqOf(c)
	}
	if mb, ok := v.(interface{ MaybeValue() (any, bool) }); ok {
		if x, isJust := mb.MaybeValue(); isJust {
			return []any{x}, true
		}
		return nil, true
	}
	if eb, ok := v.(interface{ EitherValue() (any, any, bool) }); ok {
		if _, r, isLeft := eb.EitherValue(); !isLeft {
			return []any{r}, true
		}
		return nil, true
	}
	return functions.ToSliceAny(v)
}

// notCarrierError reports that a value has no element structure for an action (a
// scalar). Stream steps use it to tell "treat the value as one element" apart from a
// genuine failure inside the element function.
type notCarrierError struct {
	what string
	v    any
}

func (e *notCarrierError) Error() string {
	return fmt.Sprintf("cannot %s a value of type %T", e.what, e.v)
}

// mapValue is the Functor action: it applies f to every element of v, rebuilding the
// same carrier (Pair maps its second component, Either its Right, a Maybe its Just).
func mapValue(v any, f func(any) (any, error)) (any, error) {
	mapAll := func(items []any) ([]any, error) {
		out := make([]any, len(items))
		for idx, it := range items {
			r, err := f(it)
			if err != nil {
				return nil, err
			}
			out[idx] = r
		}
		return out, nil
	}
	switch c := v.(type) {
	case *collections.HAMT[string, any]:
		res := collections.NewHAMT[string, any]()
		for k, val := range c.Seq() {
			r, err := f(val)
			if err != nil {
				return nil, err
			}
			res = res.Set(k, r)
		}
		return res, nil
	case *collections.Tree[any]:
		var walk func(t *collections.Tree[any]) (*collections.Tree[any], error)
		walk = func(t *collections.Tree[any]) (*collections.Tree[any], error) {
			r, err := f(t.Value())
			if err != nil {
				return nil, err
			}
			kids := collections.Nil[*collections.Tree[any]]()
			if t.Children() != nil {
				var subs []*collections.Tree[any]
				for ch := range t.Children().Seq() {
					s, err := walk(ch)
					if err != nil {
						return nil, err
					}
					subs = append(subs, s)
				}
				for idx := len(subs) - 1; idx >= 0; idx-- {
					kids = collections.Cons(subs[idx], kids)
				}
			}
			return collections.NewTree(r, kids), nil
		}
		return walk(c)
	case dsltypes.Variant:
		var ferr error
		out := c.Map(func(p any) any {
			r, err := f(p)
			if err != nil {
				ferr = err
			}
			return r
		})
		if ferr != nil {
			return nil, ferr
		}
		return out, nil
	case []any:
		return mapAll(c)
	}
	if p, ok := v.(interface{ PairElements() (any, any) }); ok {
		fst, snd := p.PairElements()
		r, err := f(snd)
		if err != nil {
			return nil, err
		}
		return functions.Pair[any, any]{Fst: fst, Snd: r}, nil
	}
	if mb, ok := v.(interface{ MaybeValue() (any, bool) }); ok {
		x, isJust := mb.MaybeValue()
		if !isJust {
			return functions.Nothing[any](), nil
		}
		r, err := f(x)
		if err != nil {
			return nil, err
		}
		return functions.Just[any](r), nil
	}
	if eb, ok := v.(interface{ EitherValue() (any, any, bool) }); ok {
		l, r, isLeft := eb.EitherValue()
		if isLeft {
			return functions.Left[any, any](l), nil
		}
		out, err := f(r)
		if err != nil {
			return nil, err
		}
		return functions.Right[any, any](out), nil
	}
	if _, ok := v.(interface{ IsSet() bool }); ok {
		items, _ := functions.ToSliceAny(v)
		mapped, err := mapAll(items)
		if err != nil {
			return nil, err
		}
		return newAnySet(mapped)
	}
	if ch, ok := v.(interface{ IsChunk() bool }); ok && ch.IsChunk() {
		items, _ := functions.ToSliceAny(v)
		mapped, err := mapAll(items)
		if err != nil {
			return nil, err
		}
		capacity := len(mapped)
		if c, ok := v.(*collections.Chunk[any]); ok {
			capacity = c.Capacity()
		}
		return newAnyChunk(capacity, mapped), nil
	}
	if _, ok := v.(interface{ IsList() bool }); ok {
		items, _ := functions.ToSliceAny(v)
		mapped, err := mapAll(items)
		if err != nil {
			return nil, err
		}
		return listOf(mapped), nil
	}
	return nil, &notCarrierError{what: "map over", v: v}
}

// filterValue is the Filterable action: it keeps the elements for which keep is true
// (HAMT filters by value, Tree prunes subtrees, Maybe becomes Nothing).
func filterValue(v any, keep func(any) (bool, error)) (any, error) {
	filterAll := func(items []any) ([]any, error) {
		var out []any
		for _, it := range items {
			k, err := keep(it)
			if err != nil {
				return nil, err
			}
			if k {
				out = append(out, it)
			}
		}
		return out, nil
	}
	var ferr error
	guard := func(x any) bool {
		k, err := keep(x)
		if err != nil {
			ferr = err
		}
		return k
	}
	switch c := v.(type) {
	case *collections.HAMT[string, any]:
		res := c.Filter(func(_ string, val any) bool { return guard(val) })
		return res, ferr
	case *collections.Tree[any]:
		res := c.Filter(guard)
		return res, ferr
	case []any:
		return filterAll(c)
	}
	if mb, ok := v.(interface{ MaybeValue() (any, bool) }); ok {
		x, isJust := mb.MaybeValue()
		if !isJust {
			return functions.Nothing[any](), nil
		}
		k, err := keep(x)
		if err != nil {
			return nil, err
		}
		if k {
			return functions.Just[any](x), nil
		}
		return functions.Nothing[any](), nil
	}
	if _, ok := v.(interface{ IsSet() bool }); ok {
		items, _ := functions.ToSliceAny(v)
		kept, err := filterAll(items)
		if err != nil {
			return nil, err
		}
		return newAnySet(kept)
	}
	if ch, ok := v.(interface{ IsChunk() bool }); ok && ch.IsChunk() {
		items, _ := functions.ToSliceAny(v)
		kept, err := filterAll(items)
		if err != nil {
			return nil, err
		}
		capacity := len(items)
		if c, ok := v.(*collections.Chunk[any]); ok {
			capacity = c.Capacity()
		}
		return newAnyChunk(capacity, kept), nil
	}
	if _, ok := v.(interface{ IsList() bool }); ok {
		items, _ := functions.ToSliceAny(v)
		kept, err := filterAll(items)
		if err != nil {
			return nil, err
		}
		return listOf(kept), nil
	}
	return nil, &notCarrierError{what: "filter", v: v}
}

// foldValue is the Foldable catamorphism: a left fold of f over the elements of v
// from seed.
func foldValue(v, seed any, f func(acc, x any) (any, error)) (any, error) {
	items, ok := elementsOf(v)
	if !ok {
		return nil, fmt.Errorf("cannot fold a value of type %T", v)
	}
	acc := seed
	for _, it := range items {
		next, err := f(acc, it)
		if err != nil {
			return nil, err
		}
		acc = next
	}
	return acc, nil
}

// bindValue is the Monad bind (flatMap): f maps each element to a value of the same
// carrier and the results are flattened (List/Chunk concatenate, Set unions, Maybe
// and Either short-circuit on Nothing/Left).
func bindValue(v any, f func(any) (any, error)) (any, error) {
	flatten := func(items []any) ([]any, error) {
		var out []any
		for _, it := range items {
			r, err := f(it)
			if err != nil {
				return nil, err
			}
			sub, ok := functions.ToSliceAny(r)
			if !ok {
				return nil, fmt.Errorf("function must return a collection, got %T", r)
			}
			out = append(out, sub...)
		}
		return out, nil
	}
	if mb, ok := v.(interface{ MaybeValue() (any, bool) }); ok {
		x, isJust := mb.MaybeValue()
		if !isJust {
			return v, nil
		}
		return f(x)
	}
	if eb, ok := v.(interface{ EitherValue() (any, any, bool) }); ok {
		_, r, isLeft := eb.EitherValue()
		if isLeft {
			return v, nil
		}
		return f(r)
	}
	if va, ok := v.(dsltypes.Variant); ok {
		_, _, payload := va.VariantElements()
		return f(payload)
	}
	items, ok := functions.ToSliceAny(v)
	if !ok {
		return nil, &notCarrierError{what: "bind over", v: v}
	}
	out, err := flatten(items)
	if err != nil {
		return nil, err
	}
	if _, ok := v.(interface{ IsSet() bool }); ok {
		return newAnySet(out)
	}
	if ch, ok := v.(interface{ IsChunk() bool }); ok && ch.IsChunk() {
		return newAnyChunk(len(out), out), nil
	}
	return listOf(out), nil
}

// newAnySet builds a Set[any] from items, converting the runtime panic Go raises for
// an unhashable element (e.g. a slice) into an error.
func newAnySet(items []any) (set *collections.Set[any], err error) {
	defer func() {
		if r := recover(); r != nil {
			set, err = nil, fmt.Errorf("set: unhashable element (%v)", r)
		}
	}()
	set = collections.NewSet[any]()
	for _, it := range items {
		set = set.Add(it)
	}
	return set, nil
}

// newAnyChunk builds a Chunk[any] of at least the given capacity holding items.
func newAnyChunk(capacity int, items []any) *collections.Chunk[any] {
	if capacity < len(items) {
		capacity = len(items)
	}
	ch := collections.NewChunk[any](capacity)
	for _, it := range items {
		ch.Append(it)
	}
	return ch
}
