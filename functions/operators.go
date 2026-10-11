package functions

import (
	"fmt"
	"math"
	"math/big"
	"reflect"
	"strconv"
	"strings"
)

// OpKind represents the supported binary operators.
type OpKind string

const (
	OpAdd   OpKind = "+"
	OpSub   OpKind = "-"
	OpMul   OpKind = "*"
	OpDiv   OpKind = "/"   // True division -> float64 (or *big.Rat / *big.Int if BigInt)
	OpIDiv  OpKind = "//"  // Integer Euclidean division
	OpMod   OpKind = "%"   // Remainder / modulo
	OpConcat OpKind = "++" // Monoidal concatenation
	OpEq    OpKind = "=="
	OpNeq   OpKind = "!="
	OpLt    OpKind = "<"
	OpLte   OpKind = "<="
	OpGt    OpKind = ">"
	OpGte   OpKind = ">="
	OpAnd   OpKind = "&&"
	OpOr    OpKind = "||"
)

// IsNumeric checks if val is a signed integer, unsigned integer, float, or *big.Int.
func IsNumeric(val any) bool {
	switch val.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64, *big.Int:
		return true
	default:
		return false
	}
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
		if _, ok := n.SetString(v, 10); ok {
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
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
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

// AddWithPromotion adds two 64-bit signed integers with automatic overflow promotion to *big.Int.
func AddWithPromotion(a, b int64) any {
	// Check signed addition overflow
	if (b > 0 && a > math.MaxInt64-b) || (b < 0 && a < math.MinInt64-b) {
		res := new(big.Int).Add(big.NewInt(a), big.NewInt(b))
		return res
	}
	sum := a + b
	return fitInt(sum)
}

// SubWithPromotion subtracts two 64-bit signed integers with automatic overflow promotion to *big.Int.
func SubWithPromotion(a, b int64) any {
	// Check signed subtraction overflow
	if (b < 0 && a > math.MaxInt64+b) || (b > 0 && a < math.MinInt64+b) {
		res := new(big.Int).Sub(big.NewInt(a), big.NewInt(b))
		return res
	}
	diff := a - b
	return fitInt(diff)
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

func fitInt(v int64) any {
	// Keep as int if it fits standard platform int
	if v >= math.MinInt && v <= math.MaxInt {
		return int(v)
	}
	return v
}

// EvalBinaryOp evaluates an overloaded infix binary operation over two values.
// Handles automatic numeric promotion:
//   - int8 -> int16 -> int32 -> int64 -> *big.Int upon integer overflow
//   - true division (`/`) of integers promotes to float64
//   - integer division (`//`) truncates
//   - collections concatenation (`++`)
//   - comparisons (`==`, `!=`, `<`, `<=`, `>`, `>=`)
//   - boolean logic (`&&`, `||`)
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

	// Float Operations (if either operand is float, or if operator is true division OpDiv)
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
				return lf + rf, nil
			case OpSub:
				return lf - rf, nil
			case OpMul:
				return lf * rf, nil
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

	// Integer Operations (int8, int16, int32, int64, int)
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

// StructuralEqual tests equality across primitives, pointers, and types.
func StructuralEqual(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	// Direct equality
	if a == b {
		return true
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

	// Fallback to reflect.DeepEqual for composite structures
	return reflect.DeepEqual(a, b)
}
