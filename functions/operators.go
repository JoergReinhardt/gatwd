package functions

import (
	"fmt"
	"math"
	"math/big"
	"math/cmplx"
	"reflect"
	"strconv"
	"strings"
)

// OpKind represents the supported binary operators.
type OpKind string

const (
	OpAdd    OpKind = "+"
	OpSub    OpKind = "-"
	OpMul    OpKind = "*"
	OpDiv    OpKind = "/"   // True division -> float64 / complex128
	OpIDiv   OpKind = "//"  // Integer Euclidean division
	OpMod    OpKind = "%"   // Remainder / modulo
	OpConcat OpKind = "++" // Monoidal concatenation
	OpEq     OpKind = "=="
	OpNeq    OpKind = "!="
	OpLt     OpKind = "<"
	OpLte    OpKind = "<="
	OpGt     OpKind = ">"
	OpGte    OpKind = ">="
	OpAnd    OpKind = "&&"
	OpOr     OpKind = "||"
)

// IsNumeric checks if val is a signed integer, unsigned integer, float, complex, or *big.Int.
func IsNumeric(val any) bool {
	switch val.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64, complex64, complex128,
		*big.Int, *big.Float:
		return true
	default:
		return false
	}
}

// ToComplex128 attempts to convert numeric or complex value to complex128.
func ToComplex128(val any) (complex128, bool) {
	switch v := val.(type) {
	case complex128:
		return v, true
	case complex64:
		return complex128(v), true
	case float64:
		return complex(v, 0), true
	case float32:
		return complex(float64(v), 0), true
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
		if f, ok := ToFloat64(v); ok {
			return complex(f, 0), true
		}
	case string:
		// parse complex string like "1+2i" or "3i"
		v = strings.TrimSpace(v)
		if c, err := strconv.ParseComplex(v, 128); err == nil {
			return c, true
		}
	}
	return 0, false
}

// ToBigInt attempts to convert any integer/numeric value to *big.Int.
func ToBigInt(val any) (*big.Int, bool) {
	switch v := val.(type) {
	case *big.Int:
		return new(big.Int).Set(v), true
	case int:
		return big.NewInt(int64(v)), true
	case int64:
		return big.NewInt(v), true
	case int32:
		return big.NewInt(int64(v)), true
	case int16:
		return big.NewInt(int64(v)), true
	case int8:
		return big.NewInt(int64(v)), true
	case uint:
		return new(big.Int).SetUint64(uint64(v)), true
	case uint64:
		return new(big.Int).SetUint64(v), true
	case uint32:
		return new(big.Int).SetUint64(uint64(v)), true
	case uint16:
		return new(big.Int).SetUint64(uint64(v)), true
	case uint8:
		return new(big.Int).SetUint64(uint64(v)), true
	case string:
		n := new(big.Int)
		if _, ok := n.SetString(strings.TrimSpace(v), 10); ok {
			return n, true
		}
	}
	return nil, false
}

// ToFloat64 converts any numeric value or numeric string to float64.
func ToFloat64(val any) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case int32:
		return float64(v), true
	case int16:
		return float64(v), true
	case int8:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint64:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint8:
		return float64(v), true
	case *big.Int:
		f, _ := new(big.Float).SetInt(v).Float64()
		return f, true
	case *big.Float:
		f, _ := v.Float64()
		return f, true
	case string:
		if f, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

// TryNarrowInteger downscales an integer to the smallest fitting primitive (int8, int16, int32, int64, int)
// or promotes upward if overflow occurred.
func TryNarrowInteger(val int64) any {
	if val >= math.MinInt8 && val <= math.MaxInt8 {
		return int8(val)
	}
	if val >= math.MinInt16 && val <= math.MaxInt16 {
		return int16(val)
	}
	if val >= math.MinInt32 && val <= math.MaxInt32 {
		return int32(val)
	}
	return val
}

// TryNarrowUint downscales an unsigned integer to uint8, uint16, uint32, or uint64.
func TryNarrowUint(val uint64) any {
	if val <= math.MaxUint8 {
		return uint8(val)
	}
	if val <= math.MaxUint16 {
		return uint16(val)
	}
	if val <= math.MaxUint32 {
		return uint32(val)
	}
	return val
}

// AddWithPromotion adds two 64-bit signed integers with automatic overflow promotion to *big.Int.
func AddWithPromotion(a, b int64) any {
	if (b > 0 && a > math.MaxInt64-b) || (b < 0 && a < math.MinInt64-b) {
		return new(big.Int).Add(big.NewInt(a), big.NewInt(b))
	}
	return fitInt(a + b)
}

// SubWithPromotion subtracts two 64-bit signed integers with automatic overflow promotion to *big.Int.
func SubWithPromotion(a, b int64) any {
	if (b < 0 && a > math.MaxInt64+b) || (b > 0 && a < math.MinInt64+b) {
		return new(big.Int).Sub(big.NewInt(a), big.NewInt(b))
	}
	return fitInt(a - b)
}

// MulWithPromotion multiplies two 64-bit signed integers with automatic overflow promotion to *big.Int.
func MulWithPromotion(a, b int64) any {
	if a == 0 || b == 0 {
		return fitInt(0)
	}
	if a > 0 {
		if b > 0 {
			if a > math.MaxInt64/b {
				return new(big.Int).Mul(big.NewInt(a), big.NewInt(b))
			}
		} else {
			if b < math.MinInt64/a {
				return new(big.Int).Mul(big.NewInt(a), big.NewInt(b))
			}
		}
	} else {
		if b > 0 {
			if a < math.MinInt64/b {
				return new(big.Int).Mul(big.NewInt(a), big.NewInt(b))
			}
		} else {
			if a < math.MaxInt64/b {
				return new(big.Int).Mul(big.NewInt(a), big.NewInt(b))
			}
		}
	}
	return fitInt(a * b)
}

// AddUintWithPromotion adds two unsigned 64-bit integers with promotion to *big.Int upon overflow.
func AddUintWithPromotion(a, b uint64) any {
	if a > math.MaxUint64-b {
		return new(big.Int).Add(new(big.Int).SetUint64(a), new(big.Int).SetUint64(b))
	}
	return a + b
}

// SubUintWithPromotion subtracts two unsigned 64-bit integers; if underflow (a < b), promotes to negative signed int64 / *big.Int.
func SubUintWithPromotion(a, b uint64) any {
	if a < b {
		diff := b - a
		if diff <= math.MaxInt64 {
			return -int64(diff)
		}
		res := new(big.Int).Sub(new(big.Int).SetUint64(a), new(big.Int).SetUint64(b))
		return res
	}
	return a - b
}

// MulUintWithPromotion multiplies two unsigned 64-bit integers with promotion to *big.Int upon overflow.
func MulUintWithPromotion(a, b uint64) any {
	if a == 0 || b == 0 {
		return uint64(0)
	}
	if a > math.MaxUint64/b {
		return new(big.Int).Mul(new(big.Int).SetUint64(a), new(big.Int).SetUint64(b))
	}
	return a * b
}

// AddFloatWithPromotion adds two float64s, promoting to *big.Float if infinite overflow occurs.
func AddFloatWithPromotion(a, b float64) any {
	res := a + b
	if math.IsInf(res, 0) {
		return new(big.Float).Add(big.NewFloat(a), big.NewFloat(b))
	}
	return res
}

// SubFloatWithPromotion subtracts two float64s, promoting to *big.Float if overflow occurs.
func SubFloatWithPromotion(a, b float64) any {
	res := a - b
	if math.IsInf(res, 0) {
		return new(big.Float).Sub(big.NewFloat(a), big.NewFloat(b))
	}
	return res
}

// MulFloatWithPromotion multiplies two float64s, promoting to *big.Float if overflow occurs.
func MulFloatWithPromotion(a, b float64) any {
	res := a * b
	if math.IsInf(res, 0) && !math.IsInf(a, 0) && !math.IsInf(b, 0) {
		return new(big.Float).Mul(big.NewFloat(a), big.NewFloat(b))
	}
	return res
}

// AddComplexWithPromotion adds two complex128 numbers with component float promotion.
func AddComplexWithPromotion(a, b complex128) any {
	res := a + b
	if cmplx.IsInf(res) && !cmplx.IsInf(a) && !cmplx.IsInf(b) {
		// Component big.Float representation
		r := new(big.Float).Add(big.NewFloat(real(a)), big.NewFloat(real(b)))
		im := new(big.Float).Add(big.NewFloat(imag(a)), big.NewFloat(imag(b)))
		return fmt.Sprintf("(%s+%si)", r.Text('g', 10), im.Text('g', 10))
	}
	return res
}

func fitInt(v int64) any {
	if v >= math.MinInt && v <= math.MaxInt {
		return int(v)
	}
	return v
}

// EvalBinaryOp evaluates an overloaded infix binary operation over two values.
func EvalBinaryOp(op OpKind, left, right any) (any, error) {
	// Monoidal concatenation operator ++
	if op == OpConcat {
		if ls, ok := left.(string); ok {
			return ls + fmt.Sprintf("%v", right), nil
		}
		if rs, ok := right.(string); ok {
			return fmt.Sprintf("%v", left) + rs, nil
		}
	}

	// Boolean Logic
	if op == OpAnd || op == OpOr {
		lb, lok := left.(bool)
		rb, rok := right.(bool)
		if !lok || !rok {
			return nil, fmt.Errorf("operator %s requires boolean operands, got %T and %T", op, left, right)
		}
		if op == OpAnd {
			return lb && rb, nil
		}
		return lb || rb, nil
	}

	// Equality & Inequality (structural or scalar)
	if op == OpEq {
		return StructuralEqual(left, right), nil
	}
	if op == OpNeq {
		return !StructuralEqual(left, right), nil
	}

	// Complex numbers
	isComplex := false
	switch left.(type) {
	case complex64, complex128:
		isComplex = true
	}
	switch right.(type) {
	case complex64, complex128:
		isComplex = true
	}
	if isComplex {
		lc, lok := ToComplex128(left)
		rc, rok := ToComplex128(right)
		if lok && rok {
			switch op {
			case OpAdd:
				return AddComplexWithPromotion(lc, rc), nil
			case OpSub:
				return lc - rc, nil
			case OpMul:
				res := lc * rc
				if cmplx.IsInf(res) && !cmplx.IsInf(lc) && !cmplx.IsInf(rc) {
					// Overflow occurred
					r := new(big.Float).Sub(
						new(big.Float).Mul(big.NewFloat(real(lc)), big.NewFloat(real(rc))),
						new(big.Float).Mul(big.NewFloat(imag(lc)), big.NewFloat(imag(rc))),
					)
					im := new(big.Float).Add(
						new(big.Float).Mul(big.NewFloat(real(lc)), big.NewFloat(imag(rc))),
						new(big.Float).Mul(big.NewFloat(imag(lc)), big.NewFloat(real(rc))),
					)
					return fmt.Sprintf("(%s+%si)", r.Text('g', 10), im.Text('g', 10)), nil
				}
				return res, nil
			case OpDiv:
				if rc == 0 {
					return nil, fmt.Errorf("division by zero in complex arithmetic")
				}
				return lc / rc, nil
			}
		}
	}

	// BigFloat Operations
	_, lIsBigF := left.(*big.Float)
	_, rIsBigF := right.(*big.Float)
	if lIsBigF || rIsBigF {
		lbf := toBigFloat(left)
		rbf := toBigFloat(right)
		switch op {
		case OpAdd:
			return new(big.Float).Add(lbf, rbf), nil
		case OpSub:
			return new(big.Float).Sub(lbf, rbf), nil
		case OpMul:
			return new(big.Float).Mul(lbf, rbf), nil
		case OpDiv:
			if rbf.Sign() == 0 {
				return nil, fmt.Errorf("division by zero")
			}
			return new(big.Float).Quo(lbf, rbf), nil
		case OpLt:
			return lbf.Cmp(rbf) < 0, nil
		case OpLte:
			return lbf.Cmp(rbf) <= 0, nil
		case OpGt:
			return lbf.Cmp(rbf) > 0, nil
		case OpGte:
			return lbf.Cmp(rbf) >= 0, nil
		}
	}

	// BigInt Operations
	_, lIsBig := left.(*big.Int)
	_, rIsBig := right.(*big.Int)
	if lIsBig || rIsBig {
		lb, lok := ToBigInt(left)
		rb, rok := ToBigInt(right)
		if lok && rok {
			switch op {
			case OpAdd:
				return new(big.Int).Add(lb, rb), nil
			case OpSub:
				return new(big.Int).Sub(lb, rb), nil
			case OpMul:
				return new(big.Int).Mul(lb, rb), nil
			case OpIDiv:
				if rb.Sign() == 0 {
					return nil, fmt.Errorf("division by zero")
				}
				return new(big.Int).Quo(lb, rb), nil
			case OpDiv:
				if rb.Sign() == 0 {
					return nil, fmt.Errorf("division by zero")
				}
				lf, _ := new(big.Float).SetInt(lb).Float64()
				rf, _ := new(big.Float).SetInt(rb).Float64()
				return lf / rf, nil
			case OpMod:
				if rb.Sign() == 0 {
					return nil, fmt.Errorf("modulo by zero")
				}
				return new(big.Int).Rem(lb, rb), nil
			case OpLt:
				return lb.Cmp(rb) < 0, nil
			case OpLte:
				return lb.Cmp(rb) <= 0, nil
			case OpGt:
				return lb.Cmp(rb) > 0, nil
			case OpGte:
				return lb.Cmp(rb) >= 0, nil
			}
		}
	}

	// Float Operations (if either operand is float, or true division OpDiv)
	isFloatOp := op == OpDiv
	if !isFloatOp {
		switch left.(type) {
		case float32, float64:
			isFloatOp = true
		}
		switch right.(type) {
		case float32, float64:
			isFloatOp = true
		}
	}

	if isFloatOp {
		lf, lok := ToFloat64(left)
		rf, rok := ToFloat64(right)
		if lok && rok {
			switch op {
			case OpAdd:
				return AddFloatWithPromotion(lf, rf), nil
			case OpSub:
				return SubFloatWithPromotion(lf, rf), nil
			case OpMul:
				return MulFloatWithPromotion(lf, rf), nil
			case OpDiv:
				if rf == 0 {
					return nil, fmt.Errorf("division by zero")
				}
				return lf / rf, nil
			case OpLt:
				return lf < rf, nil
			case OpLte:
				return lf <= rf, nil
			case OpGt:
				return lf > rf, nil
			case OpGte:
				return lf >= rf, nil
			}
		}
	}

	// Unsigned Integer Operations (uint, uint64, uint32, uint16, uint8)
	isUintOp := false
	switch left.(type) {
	case uint, uint64, uint32, uint16, uint8:
		isUintOp = true
	}
	switch right.(type) {
	case uint, uint64, uint32, uint16, uint8:
		isUintOp = true
	}
	if isUintOp {
		lu, lok := toUint64(left)
		ru, rok := toUint64(right)
		if lok && rok {
			switch op {
			case OpAdd:
				return AddUintWithPromotion(lu, ru), nil
			case OpSub:
				return SubUintWithPromotion(lu, ru), nil
			case OpMul:
				return MulUintWithPromotion(lu, ru), nil
			case OpIDiv:
				if ru == 0 {
					return nil, fmt.Errorf("division by zero")
				}
				return lu / ru, nil
			case OpMod:
				if ru == 0 {
					return nil, fmt.Errorf("modulo by zero")
				}
				return lu % ru, nil
			case OpLt:
				return lu < ru, nil
			case OpLte:
				return lu <= ru, nil
			case OpGt:
				return lu > ru, nil
			case OpGte:
				return lu >= ru, nil
			}
		}
	}

	// Signed Integer Operations (int8, int16, int32, int64, int)
	li, lok := toInt64(left)
	ri, rok := toInt64(right)
	if lok && rok {
		switch op {
		case OpAdd:
			return AddWithPromotion(li, ri), nil
		case OpSub:
			return SubWithPromotion(li, ri), nil
		case OpMul:
			return MulWithPromotion(li, ri), nil
		case OpIDiv:
			if ri == 0 {
				return nil, fmt.Errorf("division by zero")
			}
			return fitInt(li / ri), nil
		case OpMod:
			if ri == 0 {
				return nil, fmt.Errorf("modulo by zero")
			}
			return fitInt(li % ri), nil
		case OpLt:
			return li < ri, nil
		case OpLte:
			return li <= ri, nil
		case OpGt:
			return li > ri, nil
		case OpGte:
			return li >= ri, nil
		}
	}

	// Relational String comparisons if both are strings
	if ls, lok := left.(string); lok {
		if rs, rok := right.(string); rok {
			switch op {
			case OpLt:
				return ls < rs, nil
			case OpLte:
				return ls <= rs, nil
			case OpGt:
				return ls > rs, nil
			case OpGte:
				return ls >= rs, nil
			}
		}
	}

	// Fallback string concatenation for + if either operand is string
	if op == OpAdd {
		if ls, lok := left.(string); lok {
			return ls + fmt.Sprintf("%v", right), nil
		}
		if rs, rok := right.(string); rok {
			return fmt.Sprintf("%v", left) + rs, nil
		}
	}

	return nil, fmt.Errorf("unsupported binary operator %q for types %T and %T", op, left, right)
}

func toBigFloat(val any) *big.Float {
	switch v := val.(type) {
	case *big.Float:
		return v
	case *big.Int:
		return new(big.Float).SetInt(v)
	case float64:
		return big.NewFloat(v)
	case int64:
		return big.NewFloat(float64(v))
	case int:
		return big.NewFloat(float64(v))
	default:
		if f, ok := ToFloat64(val); ok {
			return big.NewFloat(f)
		}
	}
	return new(big.Float)
}

func toUint64(val any) (uint64, bool) {
	switch v := val.(type) {
	case uint:
		return uint64(v), true
	case uint64:
		return v, true
	case uint32:
		return uint64(v), true
	case uint16:
		return uint64(v), true
	case uint8:
		return uint64(v), true
	case int:
		if v >= 0 {
			return uint64(v), true
		}
	case int64:
		if v >= 0 {
			return uint64(v), true
		}
	case string:
		if u, err := strconv.ParseUint(strings.TrimSpace(v), 10, 64); err == nil {
			return u, true
		}
	}
	return 0, false
}

func toInt64(val any) (int64, bool) {
	switch v := val.(type) {
	case int:
		return int64(v), true
	case int64:
		return v, true
	case int32:
		return int64(v), true
	case int16:
		return int64(v), true
	case int8:
		return int64(v), true
	case uint:
		if v <= math.MaxInt64 {
			return int64(v), true
		}
	case uint64:
		if v <= math.MaxInt64 {
			return int64(v), true
		}
	case uint32:
		return int64(v), true
	case uint16:
		return int64(v), true
	case uint8:
		return int64(v), true
	case string:
		if iv, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64); err == nil {
			return iv, true
		}
	}
	return 0, false
}

// StructuralEqual tests equality across primitives, pointers, complex, and types.
func StructuralEqual(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if a == b {
		return true
	}

	// Complex equality
	if ac, ok := ToComplex128(a); ok {
		if bc, ok := ToComplex128(b); ok {
			return ac == bc
		}
	}

	// Compare BigInts
	if ab, ok := a.(*big.Int); ok {
		if bb, ok := b.(*big.Int); ok {
			return ab.Cmp(bb) == 0
		}
		if bb, ok := ToBigInt(b); ok {
			return ab.Cmp(bb) == 0
		}
	}
	if bb, ok := b.(*big.Int); ok {
		if ab, ok := ToBigInt(a); ok {
			return ab.Cmp(bb) == 0
		}
	}

	// Numeric equality across int and float
	if af, ok := ToFloat64(a); ok {
		if bf, ok := ToFloat64(b); ok {
			return af == bf
		}
	}

	// Structural collection equality across List, Slice, and custom collections
	type elementGetter interface{ Elements() []any }
	if egA, okA := a.(elementGetter); okA {
		if egB, okB := b.(elementGetter); okB {
			elemsA := egA.Elements()
			elemsB := egB.Elements()
			if len(elemsA) != len(elemsB) {
				return false
			}
			for i := range elemsA {
				if !StructuralEqual(elemsA[i], elemsB[i]) {
					return false
				}
			}
			return true
		}
	}

	// Fallback for formatted string comparison across collection/serialized boundaries
	if strA, okA := a.(string); okA {
		if fmt.Sprintf("%v", b) == strA {
			return true
		}
	}
	if strB, okB := b.(string); okB {
		if fmt.Sprintf("%v", a) == strB {
			return true
		}
	}

	// Fallback to reflect.DeepEqual for composite structures
	return reflect.DeepEqual(a, b)
}
