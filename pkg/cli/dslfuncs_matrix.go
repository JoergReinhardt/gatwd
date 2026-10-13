package cli

import (
	"cmp"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/functions"
	"github.com/joergreinhardt/gatwd/parser"
)

// This file fills the operator-type matrix (STYLEGUIDE §7) with EXPRESSION-LEVEL
// function forms — Workstream E of docs/language_roadmap.md. Historically an
// operation was reachable only as a pipeline STEP whose arguments are literal string
// tokens (`hamt_set k v`); the function forms below take COMPUTED, real runtime
// values (`hamt_set (hamt_get "k" r) 42 m`) and compose inside lambdas, folds and
// comprehensions. Both forms are kept behaviorally identical.
//
// # Conventions
//
//   - Data-last: the collection/value operated on is the LAST argument, so
//     `x |> f a` and `f a x` agree and partial application (`map (add 1)`,
//     `set_add 3`) yields a reusable function.
//   - Curried: every function here is a functions.Closure, so supplying fewer
//     arguments than the arity returns a function awaiting the rest.
//   - Types are preserved, never stringified: element values pass through untouched.
//     Only the HAMT's key (a string by construction) is coerced, via dslStr.
//   - Errors are values: a type mismatch or bad callable returns an error, which the
//     evaluator surfaces on the Left channel — never a panic, never a silent no-op.
//
// # Cells (rows are carriers, see the STYLEGUIDE matrix)
//
//	generic Functor/Foldable/Filterable/Monad   map filter reject fold flatmap bimap
//	sequences                                   length is_empty head tail nth reverse cons concat range member to_list to_set
//	Set                                         set_empty set_add set_remove set_contains set_union set_intersect set_diff set_elements
//	HAMT                                        hamt_empty hamt_from hamt_set hamt_del hamt_has hamt_get_or hamt_keys hamt_values hamt_pairs hamt_merge hamt_diff hamt_intersect
//	Tree                                        tree_leaf tree_node tree_add_child tree_value tree_children tree_size tree_depth tree_flatten tree_member
//	BST                                         bst_empty bst_from bst_insert bst_member
//	Pair                                        pair fst snd pair_swap
//	Maybe / Either                              just nothing is_just is_nothing maybe_default maybe_fold either_left either_right is_left is_right either_default either_fold
//	Variant                                     variant variant_tag variant_payload is_variant
//	primitives                                  min max abs negate compare

// defn registers a curried function value: Params names the arguments (for
// diagnostics and arity), and impl receives exactly len(Params) evaluated arguments.
// The function name prefixes any error impl returns.
func defn(m map[string]any, name, params string, impl func(a []any) (any, error)) {
	m[name] = functions.NewClosure(strings.Fields(params), func(a []any) (any, error) {
		res, err := impl(a)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		return res, nil
	})
}

func asAnyHAMT(v any) (*collections.HAMT[string, any], error) {
	h, ok := v.(*collections.HAMT[string, any])
	if !ok {
		return nil, fmt.Errorf("want a HAMT, got %T", v)
	}
	return h, nil
}

func asAnyTree(v any) (*collections.Tree[any], error) {
	t, ok := v.(*collections.Tree[any])
	if !ok || t == nil {
		return nil, fmt.Errorf("want a Tree, got %T", v)
	}
	return t, nil
}

// asAnySet normalises any set carrier (Set[any], Set[string], …) to a Set[any].
func asAnySet(v any) (*collections.Set[any], error) {
	if s, ok := v.(*collections.Set[any]); ok {
		return s, nil
	}
	if _, ok := v.(interface{ IsSet() bool }); ok {
		items, _ := functions.ToSliceAny(v)
		return newAnySet(items)
	}
	return nil, fmt.Errorf("want a Set, got %T", v)
}

func asInt(v any) (int, error) {
	f, ok := functions.ToFloat64(v)
	if !ok || f != float64(int(f)) {
		return 0, fmt.Errorf("want an integer, got %v (%T)", v, v)
	}
	return int(f), nil
}

func structuralMember(x any, items []any) bool {
	for _, it := range items {
		if functions.StructuralEqual(x, it) {
			return true
		}
	}
	return false
}

// safeSetOp runs a Set operation, converting the runtime panic Go raises for an
// unhashable element into an error.
func safeSetOp(f func() *collections.Set[any]) (res any, err error) {
	defer func() {
		if r := recover(); r != nil {
			res, err = nil, fmt.Errorf("unhashable element (%v)", r)
		}
	}()
	return f(), nil
}

func compareAny(a, b any) (int, error) {
	c, err := functions.Compare(a, b)
	if err != nil {
		return 0, err
	}
	return c, nil
}

// textAsChars lifts a String argument to the List of its Chars so the collection function
// forms (map, filter, flatmap) treat it as the collection row it is; any other value passes
// through. Stream steps do not use it: in a pipeline a String node is one element.
func textAsChars(v any) (any, error) {
	if s, ok := v.(string); ok {
		return charList(s)
	}
	return v, nil
}

// registerMatrixFuncs adds every function form of the matrix to m.
func registerMatrixFuncs(m map[string]any) {
	// ---- generic Functor / Filterable / Foldable / Monad -------------------------
	defn(m, "map", "f x", func(a []any) (any, error) {
		if err := requireCallable("first argument", a[0]); err != nil {
			return nil, err
		}
		x, err := textAsChars(a[1])
		if err != nil {
			return nil, err
		}
		return mapValue(x, func(e any) (any, error) { return call1(a[0], e) })
	})
	for _, neg := range []bool{false, true} {
		name := "filter"
		if neg {
			name = "reject"
		}
		defn(m, name, "p x", func(a []any) (any, error) {
			if err := requireCallable("first argument", a[0]); err != nil {
				return nil, err
			}
			x, err := textAsChars(a[1])
			if err != nil {
				return nil, err
			}
			res, err := filterValue(x, func(e any) (bool, error) {
				ok, err := callPred(a[0], e)
				return ok != neg, err
			})
			if _, wasText := a[1].(string); wasText && err == nil {
				return textResult(res), nil
			}
			return res, err
		})
	}
	defn(m, "fold", "f seed x", func(a []any) (any, error) {
		if err := requireCallable("first argument", a[0]); err != nil {
			return nil, err
		}
		return foldValue(a[2], a[1], func(acc, e any) (any, error) {
			return functions.Apply(a[0], []any{acc, e})
		})
	})
	defn(m, "flatmap", "f x", func(a []any) (any, error) {
		if err := requireCallable("first argument", a[0]); err != nil {
			return nil, err
		}
		x, err := textAsChars(a[1])
		if err != nil {
			return nil, err
		}
		return bindValue(x, func(e any) (any, error) { return call1(a[0], e) })
	})
	defn(m, "bimap", "f g x", func(a []any) (any, error) {
		if err := requireCallable("first argument", a[0]); err != nil {
			return nil, err
		}
		if err := requireCallable("second argument", a[1]); err != nil {
			return nil, err
		}
		if p, ok := a[2].(interface{ PairElements() (any, any) }); ok {
			fst, snd := p.PairElements()
			x, err := call1(a[0], fst)
			if err != nil {
				return nil, err
			}
			y, err := call1(a[1], snd)
			if err != nil {
				return nil, err
			}
			return functions.Pair[any, any]{Fst: x, Snd: y}, nil
		}
		if eb, ok := a[2].(interface{ EitherValue() (any, any, bool) }); ok {
			l, r, isLeft := eb.EitherValue()
			if isLeft {
				x, err := call1(a[0], l)
				if err != nil {
					return nil, err
				}
				return functions.Left[any, any](x), nil
			}
			y, err := call1(a[1], r)
			if err != nil {
				return nil, err
			}
			return functions.Right[any, any](y), nil
		}
		return nil, fmt.Errorf("want a Pair or Either, got %T", a[2])
	})

	// ---- sequences -----------------------------------------------------------------
	defn(m, "length", "x", func(a []any) (any, error) {
		switch c := a[0].(type) {
		case string:
			return utf8.RuneCountInString(c), nil
		case *collections.Tree[any]:
			return c.Size(), nil
		case *collections.BST[any]:
			return c.Size(), nil
		case *collections.BST[string]:
			return c.Size(), nil
		}
		if l, ok := a[0].(interface{ Len() int }); ok {
			return l.Len(), nil
		}
		items, ok := elementsOf(a[0])
		if !ok {
			return nil, fmt.Errorf("cannot take the length of %T", a[0])
		}
		return len(items), nil
	})
	defn(m, "is_empty", "x", func(a []any) (any, error) {
		n, err := call1(m["length"], a[0])
		if err != nil {
			return nil, err
		}
		return n == 0, nil
	})
	defn(m, "head", "xs", func(a []any) (any, error) {
		items, ok := elementsOf(a[0])
		if !ok || len(items) == 0 {
			return nil, fmt.Errorf("empty or non-sequence %T", a[0])
		}
		return items[0], nil
	})
	defn(m, "tail", "xs", func(a []any) (any, error) {
		switch c := a[0].(type) {
		case *collections.List[any]:
			t, ok := c.Tail()
			if !ok {
				return nil, fmt.Errorf("empty list")
			}
			return t, nil
		case *collections.Chunk[any]:
			if c.Len() == 0 {
				return nil, fmt.Errorf("empty chunk")
			}
			return c.Slice(1, c.Len()), nil
		}
		items, ok := seqOf(a[0])
		if !ok || len(items) == 0 {
			return nil, fmt.Errorf("empty or non-sequence %T", a[0])
		}
		if _, wasText := a[0].(string); wasText {
			return textResult(listOf(items[1:])), nil
		}
		return listOf(items[1:]), nil
	})
	defn(m, "nth", "i xs", func(a []any) (any, error) {
		idx, err := asInt(a[0])
		if err != nil {
			return nil, err
		}
		items, ok := seqOf(a[1])
		if !ok {
			return nil, fmt.Errorf("not a sequence: %T", a[1])
		}
		if idx < 0 || idx >= len(items) {
			return nil, fmt.Errorf("index %d out of range [0,%d)", idx, len(items))
		}
		return items[idx], nil
	})
	defn(m, "reverse", "xs", func(a []any) (any, error) {
		switch c := a[0].(type) {
		case *collections.List[any]:
			return c.Reverse(), nil
		case *collections.Chunk[any]:
			return c.Reverse(), nil
		}
		items, ok := seqOf(a[0])
		if !ok {
			return nil, fmt.Errorf("not a sequence: %T", a[0])
		}
		out := make([]any, len(items))
		for idx, it := range items {
			out[len(items)-1-idx] = it
		}
		if _, wasText := a[0].(string); wasText {
			return textResult(listOf(out)), nil
		}
		return listOf(out), nil
	})
	defn(m, "cons", "x xs", func(a []any) (any, error) {
		if l, ok := a[1].(*collections.List[any]); ok {
			return collections.Cons(a[0], l), nil
		}
		items, ok := seqOf(a[1])
		if !ok {
			return nil, fmt.Errorf("not a sequence: %T", a[1])
		}
		if _, wasText := a[1].(string); wasText {
			c, err := asChar(a[0])
			if err != nil {
				return nil, fmt.Errorf("cannot cons onto a String: %w", err)
			}
			return textResult(listOf(append([]any{c}, items...))), nil
		}
		return listOf(append([]any{a[0]}, items...)), nil
	})
	defn(m, "concat", "a b", func(a []any) (any, error) { return parser.ConcatValues(a[0], a[1]) })
	defn(m, "range", "lo hi", func(a []any) (any, error) {
		lo, err := asInt(a[0])
		if err != nil {
			return nil, err
		}
		hi, err := asInt(a[1])
		if err != nil {
			return nil, err
		}
		var out []any
		for n := lo; n <= hi; n++ {
			out = append(out, functions.TryNarrowInteger(int64(n)))
		}
		return listOf(out), nil
	})
	member := func(a []any) (any, error) {
		if s, ok := a[1].(*collections.Set[any]); ok {
			return safeSetOp2(func() bool { return s.Contains(a[0]) })
		}
		items, ok := elementsOf(a[1])
		if !ok {
			return nil, fmt.Errorf("not a collection: %T", a[1])
		}
		return structuralMember(a[0], items), nil
	}
	defn(m, "member", "x c", member)
	defn(m, "to_list", "x", func(a []any) (any, error) {
		items, ok := elementsOf(a[0])
		if !ok {
			return nil, fmt.Errorf("not a collection: %T", a[0])
		}
		return listOf(items), nil
	})
	defn(m, "to_set", "x", func(a []any) (any, error) {
		items, ok := elementsOf(a[0])
		if !ok {
			return nil, fmt.Errorf("not a collection: %T", a[0])
		}
		return newAnySet(items)
	})

	// ---- Set -----------------------------------------------------------------------
	m["set_empty"] = any(collections.NewSet[any]())
	defn(m, "set_add", "x s", func(a []any) (any, error) {
		s, err := asAnySet(a[1])
		if err != nil {
			return nil, err
		}
		return safeSetOp(func() *collections.Set[any] { return s.Add(a[0]) })
	})
	defn(m, "set_remove", "x s", func(a []any) (any, error) {
		s, err := asAnySet(a[1])
		if err != nil {
			return nil, err
		}
		return safeSetOp(func() *collections.Set[any] { return s.Remove(a[0]) })
	})
	defn(m, "set_contains", "x s", func(a []any) (any, error) {
		s, err := asAnySet(a[1])
		if err != nil {
			return nil, err
		}
		return safeSetOp2(func() bool { return s.Contains(a[0]) })
	})
	for name, op := range map[string]func(x, y *collections.Set[any]) *collections.Set[any]{
		"set_union":     func(x, y *collections.Set[any]) *collections.Set[any] { return x.Union(y) },
		"set_intersect": func(x, y *collections.Set[any]) *collections.Set[any] { return x.Intersect(y) },
		"set_diff":      func(x, y *collections.Set[any]) *collections.Set[any] { return x.Diff(y) },
	} {
		op := op
		defn(m, name, "a b", func(a []any) (any, error) {
			x, err := asAnySet(a[0])
			if err != nil {
				return nil, err
			}
			y, err := asAnySet(a[1])
			if err != nil {
				return nil, err
			}
			return safeSetOp(func() *collections.Set[any] { return op(x, y) })
		})
	}
	defn(m, "set_elements", "s", func(a []any) (any, error) {
		s, err := asAnySet(a[0])
		if err != nil {
			return nil, err
		}
		return listOf(s.ElementsAny()), nil
	})

	// ---- HAMT ----------------------------------------------------------------------
	m["hamt_empty"] = any(collections.NewHAMT[string, any]())
	defn(m, "hamt_from", "pairs", func(a []any) (any, error) {
		items, ok := functions.ToSliceAny(a[0])
		if !ok {
			return nil, fmt.Errorf("want a collection of [key value] pairs, got %T", a[0])
		}
		h := collections.NewHAMT[string, any]()
		for _, it := range items {
			kv, ok := functions.ToSliceAny(it)
			if !ok || len(kv) != 2 {
				return nil, fmt.Errorf("each entry must be a [key value] pair, got %v", it)
			}
			h = h.Set(dslStr(kv[0]), kv[1])
		}
		return h, nil
	})
	defn(m, "hamt_set", "k v m", func(a []any) (any, error) {
		h, err := asAnyHAMT(a[2])
		if err != nil {
			return nil, err
		}
		return h.Set(dslStr(a[0]), a[1]), nil
	})
	defn(m, "hamt_del", "k m", func(a []any) (any, error) {
		h, err := asAnyHAMT(a[1])
		if err != nil {
			return nil, err
		}
		return h.Delete(dslStr(a[0])), nil
	})
	defn(m, "hamt_get_or", "default k m", func(a []any) (any, error) {
		h, err := asAnyHAMT(a[2])
		if err != nil {
			return nil, err
		}
		if v, ok := h.Get(dslStr(a[1])); ok {
			return v, nil
		}
		return a[0], nil
	})
	defn(m, "hamt_has", "k m", func(a []any) (any, error) {
		h, err := asAnyHAMT(a[1])
		if err != nil {
			return nil, err
		}
		_, ok := h.Get(dslStr(a[0]))
		return ok, nil
	})
	defn(m, "hamt_keys", "m", func(a []any) (any, error) {
		h, err := asAnyHAMT(a[0])
		if err != nil {
			return nil, err
		}
		var out []any
		for k := range h.Keys() {
			out = append(out, k)
		}
		return listOf(out), nil
	})
	defn(m, "hamt_values", "m", func(a []any) (any, error) {
		h, err := asAnyHAMT(a[0])
		if err != nil {
			return nil, err
		}
		var out []any
		for v := range h.Values() {
			out = append(out, v)
		}
		return listOf(out), nil
	})
	defn(m, "hamt_pairs", "m", func(a []any) (any, error) {
		h, err := asAnyHAMT(a[0])
		if err != nil {
			return nil, err
		}
		var out []any
		for k, v := range h.Seq() {
			out = append(out, functions.Pair[any, any]{Fst: k, Snd: v})
		}
		return listOf(out), nil
	})
	for name, op := range map[string]func(x, y *collections.HAMT[string, any]) *collections.HAMT[string, any]{
		"hamt_merge":     func(x, y *collections.HAMT[string, any]) *collections.HAMT[string, any] { return x.Merge(y) },
		"hamt_diff":      func(x, y *collections.HAMT[string, any]) *collections.HAMT[string, any] { return x.Diff(y) },
		"hamt_intersect": func(x, y *collections.HAMT[string, any]) *collections.HAMT[string, any] { return x.Intersect(y) },
	} {
		op := op
		defn(m, name, "a b", func(a []any) (any, error) {
			x, err := asAnyHAMT(a[0])
			if err != nil {
				return nil, err
			}
			y, err := asAnyHAMT(a[1])
			if err != nil {
				return nil, err
			}
			return op(x, y), nil
		})
	}

	// ---- Tree ----------------------------------------------------------------------
	defn(m, "tree_leaf", "x", func(a []any) (any, error) { return collections.Leaf[any](a[0]), nil })
	defn(m, "tree_node", "x children", func(a []any) (any, error) {
		kids, ok := functions.ToSliceAny(a[1])
		if !ok {
			return nil, fmt.Errorf("children must be a list of trees, got %T", a[1])
		}
		list := collections.Nil[*collections.Tree[any]]()
		for idx := len(kids) - 1; idx >= 0; idx-- {
			t, err := asAnyTree(kids[idx])
			if err != nil {
				return nil, err
			}
			list = collections.Cons(t, list)
		}
		return collections.NewTree[any](a[0], list), nil
	})
	defn(m, "tree_add_child", "child t", func(a []any) (any, error) {
		child, err := asAnyTree(a[0])
		if err != nil {
			return nil, err
		}
		t, err := asAnyTree(a[1])
		if err != nil {
			return nil, err
		}
		return t.AddChild(child), nil
	})
	defn(m, "tree_value", "t", func(a []any) (any, error) {
		t, err := asAnyTree(a[0])
		if err != nil {
			return nil, err
		}
		return t.Value(), nil
	})
	defn(m, "tree_children", "t", func(a []any) (any, error) {
		t, err := asAnyTree(a[0])
		if err != nil {
			return nil, err
		}
		var out []any
		if t.Children() != nil {
			for c := range t.Children().Seq() {
				out = append(out, c)
			}
		}
		return listOf(out), nil
	})
	defn(m, "tree_size", "t", func(a []any) (any, error) {
		t, err := asAnyTree(a[0])
		if err != nil {
			return nil, err
		}
		return t.Size(), nil
	})
	defn(m, "tree_depth", "t", func(a []any) (any, error) {
		t, err := asAnyTree(a[0])
		if err != nil {
			return nil, err
		}
		return t.Depth(), nil
	})
	defn(m, "tree_flatten", "t", func(a []any) (any, error) {
		t, err := asAnyTree(a[0])
		if err != nil {
			return nil, err
		}
		return t.Flatten(), nil
	})
	defn(m, "tree_member", "x t", func(a []any) (any, error) {
		t, err := asAnyTree(a[1])
		if err != nil {
			return nil, err
		}
		for v := range t.WalkDepthFirst() {
			if functions.StructuralEqual(a[0], v) {
				return true, nil
			}
		}
		return false, nil
	})

	// ---- BST (ordered by functions.Compare for any, lexicographic for string) -----
	m["bst_empty"] = any((*collections.BST[any])(nil))
	bstInsert := func(tree any, x any) (any, error) {
		switch t := tree.(type) {
		case *collections.BST[string]:
			return t.Insert(dslStr(x), strings.Compare), nil
		case *collections.BST[any]:
			var cerr error
			res := t.Insert(x, func(p, q any) int {
				c, err := compareAny(p, q)
				if err != nil {
					cerr = err
				}
				return c
			})
			return res, cerr
		}
		return nil, fmt.Errorf("want a BST, got %T", tree)
	}
	defn(m, "bst_insert", "x t", func(a []any) (any, error) { return bstInsert(a[1], a[0]) })
	defn(m, "bst_from", "xs", func(a []any) (any, error) {
		items, ok := functions.ToSliceAny(a[0])
		if !ok {
			return nil, fmt.Errorf("not a collection: %T", a[0])
		}
		var tree any = (*collections.BST[any])(nil)
		for _, it := range items {
			var err error
			if tree, err = bstInsert(tree, it); err != nil {
				return nil, err
			}
		}
		return tree, nil
	})
	defn(m, "bst_member", "x t", func(a []any) (any, error) {
		switch t := a[1].(type) {
		case *collections.BST[string]:
			return t.Contains(dslStr(a[0]), strings.Compare), nil
		case *collections.BST[any]:
			var cerr error
			ok := t.Contains(a[0], func(p, q any) int {
				c, err := compareAny(p, q)
				if err != nil {
					cerr = err
				}
				return c
			})
			return ok, cerr
		}
		return nil, fmt.Errorf("want a BST, got %T", a[1])
	})

	// ---- Pair ----------------------------------------------------------------------
	defn(m, "pair", "a b", func(a []any) (any, error) { return functions.Pair[any, any]{Fst: a[0], Snd: a[1]}, nil })
	pairParts := func(v any) (any, any, error) {
		p, ok := v.(interface{ PairElements() (any, any) })
		if !ok {
			return nil, nil, fmt.Errorf("want a Pair, got %T", v)
		}
		x, y := p.PairElements()
		return x, y, nil
	}
	defn(m, "fst", "p", func(a []any) (any, error) { x, _, err := pairParts(a[0]); return x, err })
	defn(m, "snd", "p", func(a []any) (any, error) { _, y, err := pairParts(a[0]); return y, err })
	defn(m, "pair_swap", "p", func(a []any) (any, error) {
		x, y, err := pairParts(a[0])
		if err != nil {
			return nil, err
		}
		return functions.Pair[any, any]{Fst: y, Snd: x}, nil
	})

	// ---- Maybe / Either --------------------------------------------------------------
	defn(m, "just", "x", func(a []any) (any, error) { return functions.Just[any](a[0]), nil })
	m["nothing"] = any(functions.Nothing[any]())
	maybeParts := func(v any) (any, bool, error) {
		mb, ok := v.(interface{ MaybeValue() (any, bool) })
		if !ok {
			return nil, false, fmt.Errorf("want a Maybe, got %T", v)
		}
		x, isJust := mb.MaybeValue()
		return x, isJust, nil
	}
	defn(m, "is_just", "m", func(a []any) (any, error) { _, j, err := maybeParts(a[0]); return j, err })
	defn(m, "is_nothing", "m", func(a []any) (any, error) { _, j, err := maybeParts(a[0]); return !j, err })
	defn(m, "maybe_default", "d m", func(a []any) (any, error) {
		x, j, err := maybeParts(a[1])
		if err != nil || !j {
			return a[0], err
		}
		return x, nil
	})
	defn(m, "maybe_fold", "d f m", func(a []any) (any, error) {
		if err := requireCallable("second argument", a[1]); err != nil {
			return nil, err
		}
		x, j, err := maybeParts(a[2])
		if err != nil {
			return nil, err
		}
		if !j {
			return a[0], nil
		}
		return call1(a[1], x)
	})
	defn(m, "either_left", "x", func(a []any) (any, error) { return functions.Left[any, any](a[0]), nil })
	defn(m, "either_right", "x", func(a []any) (any, error) { return functions.Right[any, any](a[0]), nil })
	eitherParts := func(v any) (l, r any, isLeft bool, err error) {
		eb, ok := v.(interface{ EitherValue() (any, any, bool) })
		if !ok {
			return nil, nil, false, fmt.Errorf("want an Either, got %T", v)
		}
		l, r, isLeft = eb.EitherValue()
		return l, r, isLeft, nil
	}
	defn(m, "is_left", "e", func(a []any) (any, error) { _, _, l, err := eitherParts(a[0]); return l, err })
	defn(m, "is_right", "e", func(a []any) (any, error) { _, _, l, err := eitherParts(a[0]); return !l, err })
	defn(m, "either_default", "d e", func(a []any) (any, error) {
		_, r, isLeft, err := eitherParts(a[1])
		if err != nil || isLeft {
			return a[0], err
		}
		return r, nil
	})
	defn(m, "either_fold", "f g e", func(a []any) (any, error) {
		if err := requireCallable("first argument", a[0]); err != nil {
			return nil, err
		}
		if err := requireCallable("second argument", a[1]); err != nil {
			return nil, err
		}
		l, r, isLeft, err := eitherParts(a[2])
		if err != nil {
			return nil, err
		}
		if isLeft {
			return call1(a[0], l)
		}
		return call1(a[1], r)
	})

	// ---- Variant (the interpreter-aware `variant` lives in exprFallback) ------------
	variantParts := func(v any) (dsltypes.Variant, error) {
		va, ok := v.(dsltypes.Variant)
		if !ok {
			return va, fmt.Errorf("want a Variant, got %T", v)
		}
		return va, nil
	}
	defn(m, "variant_tag", "v", func(a []any) (any, error) {
		va, err := variantParts(a[0])
		return va.Tag, err
	})
	defn(m, "variant_payload", "v", func(a []any) (any, error) {
		va, err := variantParts(a[0])
		return va.Value, err
	})
	defn(m, "is_variant", "tag v", func(a []any) (any, error) {
		va, err := variantParts(a[1])
		return va.Is(dslStr(a[0])), err
	})

	// ---- primitives ------------------------------------------------------------------
	defn(m, "min", "a b", func(a []any) (any, error) {
		c, err := compareAny(a[0], a[1])
		if err != nil || c <= 0 {
			return a[0], err
		}
		return a[1], nil
	})
	defn(m, "max", "a b", func(a []any) (any, error) {
		c, err := compareAny(a[0], a[1])
		if err != nil || c >= 0 {
			return a[0], err
		}
		return a[1], nil
	})
	defn(m, "compare", "a b", func(a []any) (any, error) {
		c, err := compareAny(a[0], a[1])
		return cmp.Compare(c, 0), err
	})
	negate := func(v any) (any, error) {
		return functions.EvalBinaryOp(functions.OpSub, 0, v)
	}
	defn(m, "negate", "n", func(a []any) (any, error) { return negate(a[0]) })
	defn(m, "abs", "n", func(a []any) (any, error) {
		c, err := compareAny(a[0], 0)
		if err != nil {
			return nil, err
		}
		if c < 0 {
			return negate(a[0])
		}
		return a[0], nil
	})
}

// safeSetOp2 is safeSetOp for boolean-valued Set queries.
func safeSetOp2(f func() bool) (res any, err error) {
	defer func() {
		if r := recover(); r != nil {
			res, err = nil, fmt.Errorf("unhashable element (%v)", r)
		}
	}()
	return f(), nil
}

// variantConstructor is the Variant row's constructor `variant "Tag" payload`. It is
// interpreter-bound (not in the shared registry) because boxing a tagged value needs
// the parser's sum-type registry: a tag declared by `type T = A | B` is boxed with
// its branch BitFlag, exactly as `spawn T: A` does, so `match` and `is_variant`
// discriminate on it. An undeclared tag is an error rather than a silent new type.
func (i *Interpreter) variantConstructor() any {
	return functions.NewClosure([]string{"tag", "payload"}, func(a []any) (any, error) {
		tag := dslStr(a[0])
		typeName, _, ok := i.Parser.SumTypeOf(tag)
		if !ok {
			return nil, fmt.Errorf("variant: %q is not a constructor of any declared sum type", tag)
		}
		v, _ := i.Parser.MakeVariant(typeName, tag)
		v.Value = a[1]
		return v, nil
	})
}
