package cli

import (
	"fmt"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/functions"
	"github.com/joergreinhardt/gatwd/parser"
)

// orderingOf maps the sign of a three-way comparison to the Ordering value LT, EQ or GT.
func orderingOf(c int) dsltypes.Variant {
	tag := "EQ"
	switch {
	case c < 0:
		tag = "LT"
	case c > 0:
		tag = "GT"
	}
	v, _ := parser.OrderingVariant(tag)
	return v
}

// orderingRank is the sign each Ordering tag stands for.
var orderingRank = map[string]int{"LT": -1, "EQ": 0, "GT": 1}

// asOrdering reads v as an Ordering: the Variant LT/EQ/GT, or its bare tag name.
func asOrdering(v any) (string, error) {
	switch x := v.(type) {
	case dsltypes.Variant:
		if x.Tag == "LT" || x.Tag == "EQ" || x.Tag == "GT" {
			return x.Tag, nil
		}
	case string:
		if x == "LT" || x == "EQ" || x == "GT" {
			return x, nil
		}
	}
	return "", fmt.Errorf("want an Ordering (LT, EQ or GT), got %v (%T)", v, v)
}

func asBool(v any) (bool, error) {
	b, ok := v.(bool)
	if !ok {
		return false, fmt.Errorf("want a Bool, got %v (%T)", v, v)
	}
	return b, nil
}

// registerBoolFuncs adds the Bool and Ordering function forms (hardening W2). The Bool
// functions are strict (both arguments are evaluated, which is what lets `fold and true`
// work) and accept only Bools: a number is never "truthy". `&&` and `||` stay the
// short-circuiting operators.
//
//	not b · xor a b · and a b · or a b
//	ordering_to_int o (-1, 0, 1) · ordering_from_int n (its sign) · ordering_reverse o · ordering_then a b
func registerBoolFuncs(m map[string]any) {
	defn(m, "not", "b", func(a []any) (any, error) { return functions.Not(a[0]) })
	for name, f := range map[string]func(x, y bool) bool{
		"and": func(x, y bool) bool { return x && y },
		"or":  func(x, y bool) bool { return x || y },
		"xor": func(x, y bool) bool { return x != y },
	} {
		defn(m, name, "a b", func(a []any) (any, error) {
			x, err := asBool(a[0])
			if err != nil {
				return nil, fmt.Errorf("first argument: %w", err)
			}
			y, err := asBool(a[1])
			if err != nil {
				return nil, fmt.Errorf("second argument: %w", err)
			}
			return f(x, y), nil
		})
	}

	defn(m, "ordering_to_int", "o", func(a []any) (any, error) {
		tag, err := asOrdering(a[0])
		if err != nil {
			return nil, err
		}
		return orderingRank[tag], nil
	})
	defn(m, "ordering_from_int", "n", func(a []any) (any, error) {
		n, err := strictInt(a[0])
		if err != nil {
			return nil, err
		}
		return orderingOf(n), nil
	})
	defn(m, "ordering_reverse", "o", func(a []any) (any, error) {
		tag, err := asOrdering(a[0])
		if err != nil {
			return nil, err
		}
		return orderingOf(-orderingRank[tag]), nil
	})
	// ordering_then a b: a unless it is EQ, then b — the lexicographic combination, so
	// `ordering_then (compare x1 y1) (compare x2 y2)` orders by x1 then x2.
	defn(m, "ordering_then", "a b", func(a []any) (any, error) {
		x, err := asOrdering(a[0])
		if err != nil {
			return nil, fmt.Errorf("first argument: %w", err)
		}
		y, err := asOrdering(a[1])
		if err != nil {
			return nil, fmt.Errorf("second argument: %w", err)
		}
		if x == "EQ" {
			return orderingOf(orderingRank[y]), nil
		}
		return orderingOf(orderingRank[x]), nil
	})
}
