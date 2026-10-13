package functions

import (
	"math"
	"math/big"
	"testing"
)

// CanonNumber picks one representative per `==` class: whole numbers are the canonical Int,
// a fraction a double holds exactly is that float64, any other fraction a *Ratio.
func TestCanonNumber(t *testing.T) {
	r12, _ := NewRatio(big.NewInt(1), big.NewInt(2))
	r13, _ := NewRatio(big.NewInt(1), big.NewInt(3))
	r41, _ := NewRatio(big.NewInt(4), big.NewInt(1)) // demotes to the Int 4
	huge, _ := new(big.Int).SetString("100000000000000000000", 10)
	for _, tc := range []struct {
		name string
		in   any
		want any
	}{
		{"int", 3, 3},
		{"int8", int8(3), 3},
		{"whole float", 3.0, 3},
		{"negative zero", math.Copysign(0, -1), 0},
		{"whole ratio", r41, 4},
		{"complex with zero imaginary", complex(2, 0), 2},
		{"float half", 0.5, 0.5},
		{"ratio half is the double", r12, 0.5},
		{"float32", float32(0.25), 0.25},
		{"ratio third stays a ratio", r13, r13},
		{"infinity", math.Inf(-1), math.Inf(-1)},
		{"complex", complex(1, 2), complex(1, 2)},
		{"big integer", huge, huge},
		{"whole float beyond int64", 1e20, huge},
	} {
		got, err := CanonNumber(tc.in)
		if err != nil {
			t.Errorf("%s: %v", tc.name, err)
			continue
		}
		if KindName(got) != KindName(tc.want) || !StructuralEqual(got, tc.want) {
			t.Errorf("%s: CanonNumber(%#v) = %#v (%s), want %#v (%s)", tc.name, tc.in, got, KindName(got), tc.want, KindName(tc.want))
		}
	}
	for name, in := range map[string]any{
		"NaN": math.NaN(), "String": "1", "Bool": true, "nil": nil, "NaN complex": complex(math.NaN(), 0),
	} {
		if got, err := CanonNumber(in); err == nil {
			t.Errorf("%s: want an error, got %#v", name, got)
		}
	}
}
