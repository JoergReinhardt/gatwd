package parser_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/parser"
)

// A comprehension with a guard but no generator loops over nothing. It used to evaluate
// to [] (even when the guard held); it is now rejected at parse time (Mandate 6).
func TestComprehensionWithoutGeneratorIsAParseError(t *testing.T) {
	for _, src := range []string{
		`[ 1 | true ]`,
		`[x | x > 1]`,
		`[ 1 | 2 > 1, 3 > 2 ]`,
		`[ 1 | ]`,
		`[1, 2] ++ [ 3 | true ]`,
		`(\x -> [ x | x > 0 ]) 3`,
		`[ [ 1 | true ] | y <- [1] ]`, // the inner group is the offender
	} {
		_, err := parser.ParseInfixExpr(src)
		if err == nil {
			t.Errorf("ParseInfixExpr(%q) succeeded; want ErrNoGenerator", src)
			continue
		}
		if !errors.Is(err, parser.ErrNoGenerator) {
			t.Errorf("ParseInfixExpr(%q) = %v; want ErrNoGenerator", src, err)
		}
	}
}

// A hand-built AST (no parser) must fail loud too, not yield Nil.
func TestComprehensionWithoutGeneratorEvalIsAnError(t *testing.T) {
	e := parser.ListComprehensionExpr{
		YieldExpr: parser.LiteralExpr{Value: 1},
		Filters:   []parser.Expr{parser.LiteralExpr{Value: true}},
	}
	got, err := e.Eval(nil)
	if !errors.Is(err, parser.ErrNoGenerator) {
		t.Fatalf("Eval = %v, %v; want ErrNoGenerator", got, err)
	}
	for _, want := range []string{"no generator", "_ <- [0]"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("message %q lacks %q", err.Error(), want)
		}
	}
}

// Every comprehension that has a generator keeps working, guards before or after it.
func TestComprehensionWithGeneratorStillWorks(t *testing.T) {
	cases := map[string]int{
		`[ x | x <- [1, 2, 3] ]`:                    3,
		`[ x | x <- [1, 2, 3], x > 1 ]`:             2,
		`[ x | x <- [1, 2, 3], x > 5 ]`:             0,
		`[ 7 | _ <- [0], 2 > 1 ]`:                   1, // the documented replacement for a guard-only form
		`[ 7 | _ <- [0], 2 < 1 ]`:                   0,
		`[ 1 | true_ <- [1], 2 > 1 ]`:               1,
		`[ [x, y] | x <- [1, 2], 1 < 2, y <- [3] ]`: 2,
	}
	for src, n := range cases {
		v, err := parser.EvalInfixString(src, nil)
		if err != nil {
			t.Errorf("%q: %v", src, err)
			continue
		}
		l, ok := v.(*collections.List[any])
		if !ok || l.Len() != n {
			t.Errorf("%q = %v (%T); want a list of %d", src, v, v, n)
		}
	}
}

// CheckComprehensions is lexical and must have no false positives: every non-offending
// shape below is something the language really accepts.
func TestCheckComprehensions(t *testing.T) {
	bad := []string{
		`let a = [ 1 | true ]`,
		`def g = \x -> [ x | x > 1 ]`,
		`let f = \x -> [ x | x > 1 ] in f 3`,
		`spawn S: x |> emit [ 1 | true ]`,
		`[1, 2] |> map (\x -> [ x | x > 1 ]) |> length`,
		`head ([ 1 | true ] ++ [9])`,
		`[ [ 1 | true ] | y <- [1] ]`,
		"let r = 1\n[ r | r > 0 ]", // a multi-line statement
	}
	for _, src := range bad {
		if err := parser.CheckComprehensions(src); !errors.Is(err, parser.ErrNoGenerator) {
			t.Errorf("CheckComprehensions(%q) = %v; want ErrNoGenerator", src, err)
		}
	}
	good := []string{
		`[ x | x <- [1, 2] ]`,
		`[ x | x <- [1, 2], x > 1 ]`,
		`[ [a, b] | [a, b] <- [[1, 2]], a < b ]`,
		`[ x | y <- [1], x <- [y | z <- [2]] ]`,
		`[1, 2, 3]`,
		`[1, 2] |> map (\x -> x + 1)`,
		`[1, 2] ||> map (\x -> x + 1)`,
		`[1, 2] |> filter (\x -> x > 1) |> length`,
		`let a = [ x | x <- [1] ]`,
		`def f = \g -> [ e | e <- g, e > 0 ]`,
		`x |> match { | A -> 1 | B -> 2 }`,
		`match v { | Just y -> [y] | Nothing -> [] }`,
		`echo "[ not a | comprehension ]"`,
		`sh -c "[ -f x ] | grep y"`,
		`'|' ++ "[" ++ "|"`,
		`[ '|', ']' ]`,
		`[1] -- [ x | guard ] in a trailing comment`,
		`(a | b)`, // outside any bracket there is no comprehension to judge
		``,
	}
	for _, src := range good {
		if err := parser.CheckComprehensions(src); err != nil {
			t.Errorf("CheckComprehensions(%q) = %v; want nil", src, err)
		}
	}
}
