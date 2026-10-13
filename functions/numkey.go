package functions

import (
	"fmt"
	"math"
	"math/big"
)

// CanonNumber maps a number to the one representative of its equivalence class under `==`.
//
// Two numbers of the tower are `==` when they denote the same value exactly (an Int and a
// Float that happen to be equal, `1r2` and `0.5`, a big integer written two ways). A hash
// table needs the same notion: all members of a class must meet in one bucket, and two
// numbers that are not `==` must never share one. CanonNumber picks the representative:
//
//   - a whole number — whatever it was written as, Int, big integer, whole Float, whole
//     Ratio, a Complex with a zero imaginary part — is the canonical Int (CanonInt);
//   - a fraction that a double represents exactly (`0.5`, `1r2`, `0.1` as the double) is
//     that float64; a fraction that no double represents (`1r3`, `1r10`) stays a *Ratio —
//     which is why `1r10` and the double `0.1` are different keys, exactly as `==` says;
//   - ±Inf stay float64; a Complex with a non-zero imaginary part stays complex128.
//
// NaN equals nothing (not even itself), so it has no class: it is an error, as is any
// value that is not a number (a String is not a number, D2).
func CanonNumber(v any) (any, error) {
	switch classOf(v) {
	case clsNone:
		return nil, fmt.Errorf("%s is not a number", KindName(v))
	case clsComplex:
		c, _ := ToComplex128(v)
		if math.IsNaN(real(c)) || math.IsNaN(imag(c)) {
			return nil, fmt.Errorf("NaN is not equal to itself, so it has no canonical form")
		}
		if imag(c) == 0 {
			return CanonNumber(real(c))
		}
		return c, nil
	}
	if f, ok := v.(float64); ok {
		if math.IsNaN(f) {
			return nil, fmt.Errorf("NaN is not equal to itself, so it has no canonical form")
		}
		if math.IsInf(f, 0) {
			return f, nil
		}
	}
	if f, ok := v.(float32); ok {
		return CanonNumber(float64(f))
	}
	if bf, ok := v.(*big.Float); ok && bf.IsInf() {
		return math.Inf(bf.Sign()), nil
	}
	r, ok := ratOperand(v)
	if !ok {
		return nil, fmt.Errorf("%s has no canonical form", KindName(v))
	}
	if r.IsInt() {
		return CanonInt(new(big.Int).Set(r.Num())), nil
	}
	if f, exact := r.Float64(); exact {
		return f, nil
	}
	return (*Ratio)(r), nil
}
