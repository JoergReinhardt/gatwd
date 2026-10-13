package evaluator

import (
	"errors"
	"fmt"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/functions"
)

// This file is the DSL's Set: a *collections.Setoid[any] whose notion of "the same member"
// is the language's own `==`, carried by the key design of the HAMT (EncodeKey).
//
// What is the concept? A set needs one relation that says which elements are the same.
// For gatwd it is `==`, exactly as for a HAMT key: `1` and `1.0` are one member, `[1, 2]`
// and `[1, 2]` are one member (two lists are two Go pointers, one value), `1` and "1"
// are two, a Time seen in two zones is one instant and so one member.
//
// Why is it needed? The Set used to key its elements by their printed text (the `to_set`
// and `distinct` steps) or by Go's == on the interface value (the function form), and the two
// disagreed with each other and with `==`: `[1, 1.0, "1"] |> distinct` merged all three,
// `to_set [[1, 2], [1, 2]]` kept two lists. A single representation, built here, is what
// every producer (to_set, set_*, distinct, map/filter/bind over a Set) and every consumer
// (member, ==, json_encode, EncodeKey of a Set) now agrees on.
//
// How is it implemented? The Setoid stores the elements as themselves in a
// HAMT[string, any] under their EncodeKey; no element is ever decoded from text. The
// helpers below fix the key function (ValueSetKey: EncodeKey of the element with a Time
// reduced to its UTC instant, which is also the form a stored Time takes) so that every set
// in a program shares it, as Setoid.Union and friends require.

// ValueSetKey is the key function of the DSL's Set: the member's EncodeKey, with a Time
// reduced to its instant. An element that has no key (a function, a Graph, NaN) is an
// error that names the member's kind.
func ValueSetKey(v any) (string, error) {
	k, err := EncodeKey(functions.CanonicalKey(v))
	var u *unkeyable
	if errors.As(err, &u) {
		return "", fmt.Errorf("%s", u.describe("a Set member", "Set members are"))
	}
	return k, err
}

// NewValueSet builds the DSL's Set from items: equal elements (`==`) are one member, the
// first to arrive stays, and a Time is stored as its UTC instant. An element without a key
// is an error.
func NewValueSet(items ...any) (*collections.Setoid[any], error) {
	s, err := collections.NewSetoid(ValueSetKey)
	if err != nil {
		return nil, err
	}
	return ValueSetAdd(s, items...)
}

// ValueSetAdd adds the items to s the way NewValueSet does (the stored form of a Time is
// its UTC instant, whatever zone it was written in).
func ValueSetAdd(s *collections.Setoid[any], items ...any) (*collections.Setoid[any], error) {
	for _, it := range items {
		var err error
		if s, err = s.Add(functions.CanonicalKey(it)); err != nil {
			return nil, err
		}
	}
	return s, nil
}

// EmptyValueSet is the empty DSL Set (`set_empty`).
func EmptyValueSet() *collections.Setoid[any] {
	s, _ := collections.NewSetoid(ValueSetKey) // the key function is not nil
	return s
}

// AsValueSet reads v as a DSL Set. A *collections.Setoid[any] is itself; any other set
// carrier a Go caller may hand over (a *collections.Set[string], a *collections.Set[any])
// is rebuilt from its elements under the DSL's equality; anything else is an error.
func AsValueSet(v any) (*collections.Setoid[any], error) {
	if s, ok := v.(*collections.Setoid[any]); ok && s != nil {
		return s, nil
	}
	if _, ok := v.(interface{ IsSet() bool }); ok {
		items, _ := functions.ToSliceAny(v)
		return NewValueSet(items...)
	}
	return nil, fmt.Errorf("want a Set, got %s", functions.KindName(v))
}
