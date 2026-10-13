package functions_test

import (
	"fmt"
	"math"
	"math/big"
	"math/cmplx"
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/functions"
)

// The operator × type-pair matrix of the numeric tower (Mandate 2 for numerics). It is the
// executable form of the table in docs/dsl_guide.md §"The numeric tower": for every binary
// operator and every ordered pair of operand kinds it states what must happen, and the
// numeric cells are checked against an independent oracle built on math/big (exact
// fractions) and math (floats), not against the implementation's own rules.
//
// Operand kinds. Each has two distinct samples so that `a op a` demotions (1r2 + 1r2 is an
// Int) do not hide the result class: the left operand takes the first sample, the right the second.

type operandKind struct {
	name    string
	a, b    any
	class   int // 0 Int, 1 Ratio, 2 Float, 3 Complex, -1 not a number
	isBool  bool
	isText  bool // String
	isChar  bool
	display string
}

func matrixKinds(t *testing.T) []operandKind {
	t.Helper()
	return []operandKind{
		{name: "Int", a: functions.TryNarrowInteger(7), b: functions.TryNarrowInteger(2), class: 0},
		{name: "BigInt", a: bigOf("1180591620717411303424"), b: bigOf("-9223372036854775809"), class: 0},
		{name: "Ratio", a: rat(t, 3, 2), b: rat(t, 5, 3), class: 1},
		{name: "Float", a: 2.5, b: 0.5, class: 2},
		{name: "Complex", a: complex(1, 2), b: complex(0, 1), class: 3},
		{name: "String", a: "s", b: "t", class: -1, isText: true},
		{name: "Char", a: functions.Char('c'), b: functions.Char('d'), class: -1, isChar: true},
		{name: "Bool", a: true, b: false, class: -1, isBool: true},
	}
}

// exactValue reads an Int or Ratio sample as a big.Rat.
func exactValue(v any) *big.Rat {
	switch x := v.(type) {
	case *functions.Ratio:
		return x.Rat()
	case *big.Int:
		return new(big.Rat).SetInt(x)
	}
	b, _ := functions.ToBigInt(v)
	return new(big.Rat).SetInt(b)
}

func floatValue(v any) float64 {
	if c, ok := v.(complex128); ok {
		return real(c)
	}
	f, _ := functions.ToFloat64(v)
	return f
}

func complexValue(v any) complex128 {
	if c, ok := v.(complex128); ok {
		return c
	}
	return complex(floatValue(v), 0)
}

func floorRat(r *big.Rat) *big.Int {
	return new(big.Int).Div(r.Num(), r.Denom()) // positive denominator: Euclidean == floor
}

// resultClass names the class of a result for comparison with the lattice.
func resultClass(v any) string {
	switch v.(type) {
	case *functions.Ratio:
		return "Ratio"
	case float64:
		return "Float"
	case complex128:
		return "Complex"
	case bool:
		return "Bool"
	}
	if _, ok := functions.ToBigInt(v); ok {
		return "Int"
	}
	return fmt.Sprintf("%T", v)
}

var classNames = []string{"Int", "Ratio", "Float", "Complex"}

func TestNumericOperatorMatrix(t *testing.T) {
	kinds := matrixKinds(t)
	arith := []functions.OpKind{functions.OpAdd, functions.OpSub, functions.OpMul, functions.OpDiv, functions.OpIDiv, functions.OpMod, functions.OpPow}
	order := []functions.OpKind{functions.OpLt, functions.OpLte, functions.OpGt, functions.OpGte}
	eq := []functions.OpKind{functions.OpEq, functions.OpNeq}

	cells := 0
	for _, lk := range kinds {
		for _, rk := range kinds {
			l, r := lk.a, rk.b
			numeric := lk.class >= 0 && rk.class >= 0
			class := lk.class
			if rk.class > class {
				class = rk.class
			}

			// ---- arithmetic ----
			for _, o := range arith {
				cells++
				got, err := functions.EvalBinaryOp(o, l, r)
				cell := fmt.Sprintf("%s %s %s (%v %s %v)", lk.name, o, rk.name, l, o, r)
				if !numeric {
					if err == nil {
						t.Errorf("%s = %#v, want an error (not a number)", cell, got)
					}
					continue
				}
				if class == 3 && (o == functions.OpIDiv || o == functions.OpMod) {
					if err == nil {
						t.Errorf("%s = %#v, want an error (no floor on a plane)", cell, got)
					}
					continue
				}
				if o == functions.OpPow && class <= 1 && rk.name == "BigInt" && lk.class <= 1 {
					// an exact power with an exponent beyond int64 is refused, loudly
					if err == nil || !strings.Contains(err.Error(), "too large") {
						t.Errorf("%s = %#v, %v; want a 'too large' error", cell, got, err)
					}
					continue
				}
				if err != nil {
					t.Errorf("%s: unexpected error %v", cell, err)
					continue
				}
				checkArithCell(t, cell, o, l, r, class, got)
			}

			// ---- ordering ----
			for _, o := range order {
				cells++
				got, err := functions.EvalBinaryOp(o, l, r)
				cell := fmt.Sprintf("%s %s %s", lk.name, o, rk.name)
				switch {
				case numeric && class == 3:
					if err == nil {
						t.Errorf("%s = %#v, want an error (Complex has no order)", cell, got)
					}
				case numeric:
					want := orderOracle(o, l, r)
					if err != nil || got != want {
						t.Errorf("%s = %#v, %v; want %v", cell, got, err, want)
					}
				case lk.isText && rk.isText, lk.isChar && rk.isChar, lk.isBool && rk.isBool, lk.isText && rk.isChar, lk.isChar && rk.isText:
					if _, isB := got.(bool); err != nil || !isB {
						t.Errorf("%s = %#v, %v; want a Bool (these order)", cell, got, err)
					}
				default:
					if err == nil {
						t.Errorf("%s = %#v, want an error (a String is not a number, nothing orders across kinds)", cell, got)
					}
				}
			}

			// ---- equality: never an error; exact; text never equals a number ----
			for _, o := range eq {
				cells++
				got, err := functions.EvalBinaryOp(o, l, r)
				cell := fmt.Sprintf("%s %s %s", lk.name, o, rk.name)
				if err != nil {
					t.Errorf("%s: equality must never be an error, got %v", cell, err)
					continue
				}
				// l is the first sample of its kind and r the second of its kind: equal only if
				// the two kinds are numeric and the values coincide, which these samples avoid.
				if got != (o == functions.OpNeq) {
					t.Errorf("%s = %v, want %v (distinct samples are never equal)", cell, got, o == functions.OpNeq)
				}
				// and with the SAME sample on both sides: equal exactly when they can be
				same, err := functions.EvalBinaryOp(o, l, l)
				if err != nil || same != (o == functions.OpEq) {
					t.Errorf("%s %s %s (same sample) = %v, %v", lk.name, o, lk.name, same, err)
				}
			}
		}
	}
	if cells < 8*8*13 {
		t.Fatalf("matrix visited %d cells, want %d", cells, 8*8*13)
	}
}

// orderOracle orders two real numbers exactly through big.Rat (a Float by its binary value).
func orderOracle(o functions.OpKind, l, r any) bool {
	toRat := func(v any) *big.Rat {
		if f, ok := v.(float64); ok {
			return new(big.Rat).SetFloat64(f)
		}
		return exactValue(v)
	}
	c := toRat(l).Cmp(toRat(r))
	switch o {
	case functions.OpLt:
		return c < 0
	case functions.OpLte:
		return c <= 0
	case functions.OpGt:
		return c > 0
	default:
		return c >= 0
	}
}

// checkArithCell checks one numeric arithmetic cell: its class by the lattice and its value
// by an independent oracle.
func checkArithCell(t *testing.T, cell string, o functions.OpKind, l, r any, class int, got any) {
	t.Helper()
	switch class {
	case 3: // Complex
		lc, rc := complexValue(l), complexValue(r)
		var want complex128
		switch o {
		case functions.OpAdd:
			want = lc + rc
		case functions.OpSub:
			want = lc - rc
		case functions.OpMul:
			want = lc * rc
		case functions.OpDiv:
			want = lc / rc
		case functions.OpPow:
			want = cmplx.Pow(lc, rc)
		}
		gc, ok := got.(complex128)
		if !ok {
			// a Complex product beyond the double range is reported as text (existing behaviour)
			if s, isStr := got.(string); isStr && strings.HasPrefix(s, "(") {
				return
			}
			t.Errorf("%s = %#v, want a Complex", cell, got)
			return
		}
		if cmplx.Abs(gc-want) > 1e-9*math.Max(1, cmplx.Abs(want)) {
			t.Errorf("%s = %v, want %v", cell, gc, want)
		}
	case 2: // Float
		lf, rf := floatValue(l), floatValue(r)
		if _, ok := l.(*functions.Ratio); ok {
			lf, _ = exactValue(l).Float64()
		}
		if _, ok := r.(*functions.Ratio); ok {
			rf, _ = exactValue(r).Float64()
		}
		if lb, ok := l.(*big.Int); ok {
			lf, _ = new(big.Float).SetInt(lb).Float64()
		}
		if rb, ok := r.(*big.Int); ok {
			rf, _ = new(big.Float).SetInt(rb).Float64()
		}
		switch o {
		case functions.OpIDiv:
			q := math.Floor(lf / rf)
			bq, _ := new(big.Float).SetFloat64(q).Int(nil)
			if resultClass(got) != "Int" || !functions.StructuralEqual(got, functions.CanonInt(bq)) {
				t.Errorf("%s = %#v, want the Int %v (floor of a Float quotient)", cell, got, bq)
			}
			return
		case functions.OpMod:
			want := math.Mod(lf, rf)
			if want != 0 && (want < 0) != (rf < 0) {
				want += rf
			}
			if gf, ok := got.(float64); !ok || gf != want {
				t.Errorf("%s = %#v, want the Float %v", cell, got, want)
			}
			return
		}
		var want float64
		switch o {
		case functions.OpAdd:
			want = lf + rf
		case functions.OpSub:
			want = lf - rf
		case functions.OpMul:
			want = lf * rf
		case functions.OpDiv:
			want = lf / rf
		case functions.OpPow:
			want = math.Pow(lf, rf)
		}
		gf, ok := got.(float64)
		if !ok {
			if bf, isBig := got.(*big.Float); isBig { // a Float that overflowed promotes to *big.Float
				gf, _ = bf.Float64()
				ok = true
			}
		}
		if !ok || (gf != want && math.Abs(gf-want) > 1e-12*math.Max(1, math.Abs(want))) {
			t.Errorf("%s = %#v, want the Float %v", cell, got, want)
		}
	default: // Int or Ratio: exact
		lr, rr := exactValue(l), exactValue(r)
		var want *big.Rat
		wantClassFloat := false
		switch o {
		case functions.OpAdd:
			want = new(big.Rat).Add(lr, rr)
		case functions.OpSub:
			want = new(big.Rat).Sub(lr, rr)
		case functions.OpMul:
			want = new(big.Rat).Mul(lr, rr)
		case functions.OpDiv:
			want = new(big.Rat).Quo(lr, rr)
			// D11: Int / Int that does not divide is a Float (an exact answer only for Ratio operands)
			if class == 0 && !want.IsInt() {
				wantClassFloat = true
			}
		case functions.OpIDiv:
			q := new(big.Rat).Quo(lr, rr)
			want = new(big.Rat).SetInt(floorRat(q))
		case functions.OpMod:
			q := new(big.Rat).Quo(lr, rr)
			want = new(big.Rat).Sub(lr, new(big.Rat).Mul(rr, new(big.Rat).SetInt(floorRat(q))))
		case functions.OpPow:
			// an exact power only for an integer exponent that is not too wild; the samples
			// are small, a non-integer or negative exponent promotes to Float
			if !rr.IsInt() || rr.Sign() < 0 || rr.Num().BitLen() > 8 {
				if class == 0 && rr.IsInt() && rr.Sign() < 0 {
					if _, ok := got.(float64); !ok {
						t.Errorf("%s = %#v, want a Float (negative power of an Int)", cell, got)
					}
					return
				}
				if rr.IsInt() && rr.Sign() < 0 { // Ratio ** negative Int stays exact
					e := new(big.Int).Neg(rr.Num())
					inv := new(big.Rat).Inv(lr)
					want = ratPowOracle(inv, e)
					break
				}
				if _, ok := got.(float64); !ok {
					t.Errorf("%s = %#v, want a Float (non-integer exponent)", cell, got)
				}
				return
			}
			if lr.Num().BitLen()*int(rr.Num().Int64()) > 70000 {
				return
			}
			want = ratPowOracle(lr, rr.Num())
		}
		if wantClassFloat {
			f, _ := want.Float64()
			if gf, ok := got.(float64); !ok || gf != f {
				t.Errorf("%s = %#v, want the Float %v (Int / Int that does not divide, D11)", cell, got, f)
			}
			return
		}
		wantClass := "Ratio"
		if want.IsInt() {
			wantClass = "Int"
		}
		if resultClass(got) != wantClass {
			t.Errorf("%s: result class %s (%#v), want %s (canonical form)", cell, resultClass(got), got, wantClass)
		}
		if want.IsInt() {
			if b, ok := functions.ToBigInt(got); !ok || b.Cmp(want.Num()) != 0 {
				t.Errorf("%s = %#v, want %v", cell, got, want)
			}
			requireCanonical(t, cell, got)
		} else if gr, ok := got.(*functions.Ratio); !ok || gr.Rat().Cmp(want) != 0 {
			t.Errorf("%s = %#v, want %v", cell, got, want)
		}
	}
}

func ratPowOracle(base *big.Rat, e *big.Int) *big.Rat {
	n := new(big.Int).Exp(base.Num(), e, nil)
	d := new(big.Int).Exp(base.Denom(), e, nil)
	return new(big.Rat).SetFrac(n, d)
}
