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

// The Op* constants enumerate every binary operator the evaluator understands,
// each tagged with the concrete source token it is parsed from. They form the
// closed set of OpKind values dispatched by EvalBinaryOp.
const (
	// OpAdd is numeric/string addition ("+").
	OpAdd OpKind = "+"
	// OpSub is numeric subtraction ("-").
	OpSub OpKind = "-"
	// OpMul is numeric multiplication ("*").
	OpMul OpKind = "*"
	// OpDiv is true division ("/"), promoting to float64 or complex128.
	OpDiv OpKind = "/"
	// OpIDiv is integer Euclidean division ("//").
	OpIDiv OpKind = "//"
	// OpMod is the remainder/modulo operator ("%").
	OpMod OpKind = "%"
	// OpPow is exponentiation ("**"), right-associative.
	OpPow OpKind = "**"
	// OpConcat is the monoidal concatenation operator ("++"), used to combine
	// strings, lists, and other monoids.
	OpConcat OpKind = "++"
	// OpEq is the equality comparison ("==").
	OpEq OpKind = "=="
	// OpNeq is the inequality comparison ("!=").
	OpNeq OpKind = "!="
	// OpLt is the less-than comparison ("<").
	OpLt OpKind = "<"
	// OpLte is the less-than-or-equal comparison ("<=").
	OpLte OpKind = "<="
	// OpGt is the greater-than comparison (">").
	OpGt OpKind = ">"
	// OpGte is the greater-than-or-equal comparison (">=").
	OpGte OpKind = ">="
	// OpAnd is the short-circuiting logical conjunction ("&&").
	OpAnd OpKind = "&&"
	// OpOr is the short-circuiting logical disjunction ("||").
	OpOr OpKind = "||"
)

// IsNumeric checks if val is a signed integer, unsigned integer, float, complex, *big.Int or *Ratio.
func IsNumeric(val any) bool {
	switch val.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64, complex64, complex128,
		*big.Int, *big.Float, *Ratio:
		return true
	default:
		return false
	}
}

// ToComplex128 converts a number to complex128. Text is never a number (D2): a String is
// not parsed, use str_to_float or parse explicitly.
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
	case *Ratio:
		if f, ok := ratioFloat(v.rat()); ok {
			return complex(f, 0), true
		}
	case *big.Int:
		if f, ok := floatOperand(v); ok {
			return complex(f, 0), true
		}
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
		if f, ok := ToFloat64(v); ok {
			return complex(f, 0), true
		}
	}
	return 0, false
}

// ToBigInt converts any integer value to a fresh *big.Int; a Float, a String or anything
// else is not an integer (false).
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
	}
	return nil, false
}

// ToFloat64 converts a number to float64. Text is never a number (D2): a numeric String is
// not parsed.
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
	case *Ratio:
		return ratioFloat(v.rat())
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

// orderingHolds applies one of the ordering operators to a three-way comparison result.
func orderingHolds(op OpKind, cmp int) bool {
	switch op {
	case OpLt:
		return cmp < 0
	case OpLte:
		return cmp <= 0
	case OpGt:
		return cmp > 0
	default:
		return cmp >= 0
	}
}

// EvalBinaryOp evaluates an overloaded infix binary operation over two values.
func EvalBinaryOp(op OpKind, left, right any) (any, error) {
	// Monoidal concatenation operator ++
	if op == OpConcat {
		// Bytes ++ Bytes is a fresh Bytes (neither operand is aliased).
		if lb, ok := AsBytes(left); ok {
			if rb, ok := AsBytes(right); ok {
				return append(lb, rb...), nil
			}
		}

		// Set ++ Set (Set Union)
		if _, isSetA := left.(interface{ IsSet() bool }); isSetA {
			if _, isSetB := right.(interface{ IsSet() bool }); isSetB {
				rvA := reflect.ValueOf(left)
				mUnion := rvA.MethodByName("Union")
				// Two sets of one carrier unite by the carrier's own Union; a Set of another
				// carrier is not an argument it accepts (it would panic), so it is no union.
				if mUnion.IsValid() && reflect.TypeOf(right).AssignableTo(mUnion.Type().In(0)) {
					out := mUnion.Call([]reflect.Value{reflect.ValueOf(right)})
					return out[0].Interface(), nil
				}
			}
		}

		// HAMT ++ HAMT (Map Merge)
		if _, isHamtA := left.(interface{ IsHAMT() bool }); isHamtA {
			if _, isHamtB := right.(interface{ IsHAMT() bool }); isHamtB {
				rvA := reflect.ValueOf(left)
				mMerge := rvA.MethodByName("Merge")
				if mMerge.IsValid() {
					out := mMerge.Call([]reflect.Value{reflect.ValueOf(right)})
					return out[0].Interface(), nil
				}
			}
		}

		// Chunk ++ Chunk
		if _, isChunkA := left.(interface{ IsChunk() bool }); isChunkA {
			if _, isChunkB := right.(interface{ IsChunk() bool }); isChunkB {
				rvA := reflect.ValueOf(left)
				mConcat := rvA.MethodByName("Concat")
				if mConcat.IsValid() {
					out := mConcat.Call([]reflect.Value{reflect.ValueOf(right)})
					return out[0].Interface(), nil
				}
			}
		}

		// List ++ List
		if _, isListA := left.(interface{ IsList() bool }); isListA {
			if _, isListB := right.(interface{ IsList() bool }); isListB {
				rvA := reflect.ValueOf(left)
				mConcat := rvA.MethodByName("Concat")
				if mConcat.IsValid() {
					out := mConcat.Call([]reflect.Value{reflect.ValueOf(right)})
					return out[0].Interface(), nil
				}
			}
		}

		// Tree ++ Tree (attach child)
		if _, isTreeA := left.(interface{ IsTree() bool }); isTreeA {
			if _, isTreeB := right.(interface{ IsTree() bool }); isTreeB {
				rvA := reflect.ValueOf(left)
				mAddChild := rvA.MethodByName("AddChild")
				if mAddChild.IsValid() {
					out := mAddChild.Call([]reflect.Value{reflect.ValueOf(right)})
					return out[0].Interface(), nil
				}
			}
		}

		// Graph ++ Graph (graph monoid: node + edge set union)
		if _, isGraphA := left.(interface{ IsGraph() bool }); isGraphA {
			if _, isGraphB := right.(interface{ IsGraph() bool }); isGraphB {
				rvA := reflect.ValueOf(left)
				mUnion := rvA.MethodByName("Union")
				if mUnion.IsValid() {
					out := mUnion.Call([]reflect.Value{reflect.ValueOf(right)})
					return out[0].Interface(), nil
				}
			}
		}

		// Pair ++ Pair (element-wise concatenation)
		if pa, okA := left.(interface{ PairElements() (any, any) }); okA {
			if pb, okB := right.(interface{ PairElements() (any, any) }); okB {
				fA, sA := pa.PairElements()
				fB, sB := pb.PairElements()
				combF, errF := EvalBinaryOp(OpConcat, fA, fB)
				if errF == nil {
					combS, errS := EvalBinaryOp(OpConcat, sA, sB)
					if errS == nil {
						return Pair[any, any]{Fst: combF, Snd: combS}, nil
					}
				}
			}
		}

		// Maybe ++ Maybe
		if ma, okA := left.(interface{ MaybeValue() (any, bool) }); okA {
			if mb, okB := right.(interface{ MaybeValue() (any, bool) }); okB {
				vA, isJA := ma.MaybeValue()
				vB, isJB := mb.MaybeValue()
				if !isJA {
					return right, nil
				}
				if !isJB {
					return left, nil
				}
				comb, err := EvalBinaryOp(OpConcat, vA, vB)
				if err == nil {
					return Just(comb), nil
				}
				return Just(vA), nil
			}
		}

		// Sequences / Slices
		seqA, okA := ToSliceAny(left)
		seqB, okB := ToSliceAny(right)
		if okA && okB {
			combined := append(seqA, seqB...)
			return combined, nil
		}

		// A String and a collection of Chars concatenate as Strings (D1: the String side
		// demands a String, so the Char collection converts).
		if ls, ok := left.(string); ok {
			if rs, isChars := CharsAsString(right); isChars {
				return ls + rs, nil
			}
		}
		if rs, ok := right.(string); ok {
			if ls, isChars := CharsAsString(left); isChars {
				return ls + rs, nil
			}
		}
		if ls, ok := left.(string); ok {
			return ls + fmt.Sprintf("%v", right), nil
		}
		if rs, ok := right.(string); ok {
			return fmt.Sprintf("%v", left) + rs, nil
		}
	}

	// Duration, Time and Date (timeops.go): points and vectors, never numbers.
	if res, handled, err := temporalBinaryOp(op, left, right); handled {
		return res, err
	}

	// Arithmetic and ordering never read a String as a number (D2 extended): `"5" - 1` and
	// `"5" < 10` are errors, and two Strings never do arithmetic. Two Strings still ORDER, as
	// text (see Compare). A Char next to a String is the D1 on-demand conversion, not a number.
	switch op {
	case OpSub, OpMul, OpDiv, OpIDiv, OpMod, OpPow, OpLt, OpLte, OpGt, OpGte:
		_, lStr := left.(string)
		_, rStr := right.(string)
		ordering := op == OpLt || op == OpLte || op == OpGt || op == OpGte
		switch {
		case lStr && rStr && !ordering:
			return nil, fmt.Errorf("operator '%s' is not defined for String and String; use str_to_int / str_to_float to parse", op)
		case lStr != rStr && !IsChar(left) && !IsChar(right):
			return nil, fmt.Errorf("operator '%s' is not defined for %s and %s; use str_to_int / str_to_float to parse", op, KindName(left), KindName(right))
		}
	}

	// `+` is numeric addition only (D2): a String or Char operand is a type error, never
	// a concatenation and never a numeric parse. `++` concatenates; str_to_int parses.
	if op == OpAdd {
		_, lStr := left.(string)
		_, rStr := right.(string)
		if lStr || rStr || IsChar(left) || IsChar(right) {
			hint := "use ++ to concatenate"
			if lStr && rStr {
				hint = "use ++ to concatenate, or str_to_int / str_to_float to parse"
			}
			return nil, fmt.Errorf("operator '+' is not defined for %s and %s; %s", KindName(left), KindName(right), hint)
		}
	}

	// Exponentiation has its own exact/float rules (see Pow).
	if op == OpPow {
		return Pow(left, right)
	}

	// Relational ordering across all comparable types (<, <=, >, >=)
	if op == OpLt || op == OpLte || op == OpGt || op == OpGte {
		// Numbers order exactly; NaN is unordered, so every ordering with it is false.
		if cmp, ordered, isReal := CompareNumbers(left, right); isReal {
			if !ordered {
				return false, nil
			}
			return orderingHolds(op, cmp), nil
		}
		if cmp, err := Compare(left, right); err == nil {
			switch op {
			case OpLt:
				return cmp < 0, nil
			case OpLte:
				return cmp <= 0, nil
			case OpGt:
				return cmp > 0, nil
			case OpGte:
				return cmp >= 0, nil
			}
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

	// Numbers: the promotion lattice Int ⊂ Float ⊂ Complex (see arith in numeric.go).
	if res, handled, err := arith(op, left, right); handled {
		return res, err
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

// ParseNumberLiteral reads the text of a numeric literal exactly as the expression lexer
// does: an integer (an int64 narrowed to the smallest Int, or a *big.Int beyond int64), a
// float (`1.5`) or a complex number (`2i`), with an optional sign. It reports false for
// anything else, including numeric-looking text with spaces or a trailing garbage rune.
func ParseNumberLiteral(text string) (any, bool) {
	if text == "" {
		return nil, false
	}
	if strings.Contains(text, "r") {
		return parseSignedRatioLiteral(text)
	}
	switch {
	case strings.HasSuffix(text, "i"):
		c, err := strconv.ParseComplex(text, 128)
		return c, err == nil
	case strings.Contains(text, "."):
		f, err := strconv.ParseFloat(text, 64)
		return f, err == nil
	}
	if n, err := strconv.ParseInt(text, 10, 64); err == nil {
		return TryNarrowInteger(n), true
	}
	if b, ok := new(big.Int).SetString(strings.TrimPrefix(text, "+"), 10); ok && strings.Trim(strings.TrimPrefix(strings.TrimPrefix(text, "+"), "-"), "0123456789") == "" {
		return b, true
	}
	return nil, false
}

// parseSignedRatioLiteral reads a ratio literal with an optional single sign.
func parseSignedRatioLiteral(text string) (any, bool) {
	body := strings.TrimLeft(text, "+-")
	if len(text)-len(body) > 1 {
		return nil, false
	}
	v, err := parseRatioLiteral(body)
	if err != nil {
		return nil, false
	}
	if text[0] == '-' {
		neg, err := Negate(v)
		return neg, err == nil
	}
	return v, true
}

// ParseCanonicalNumber is ParseNumberLiteral restricted to text that prints back unchanged
// (`10`, `-5`, `1.5`; not `007`, `+5`, `1.50`, `2i`), so turning a token into a number
// never loses information about how it was written.
func ParseCanonicalNumber(text string) (any, bool) {
	n, ok := ParseNumberLiteral(text)
	if !ok {
		return nil, false
	}
	switch x := n.(type) {
	case float64:
		return n, strconv.FormatFloat(x, 'f', -1, 64) == text
	case complex128:
		return nil, false
	case *big.Int:
		return n, x.String() == text
	case *Ratio:
		return n, x.String() == text
	}
	i, _ := toInt64(n)
	return n, strconv.FormatInt(i, 10) == text
}

// LiteralToken is the value a bare literal token denotes — in `emit a, 10`, `const N = 5`
// and `-D N=5`: a canonical number is that number, anything else is the text. One rule for
// every place a script writes a value without quotes.
func LiteralToken(text string) any {
	if n, ok := ParseCanonicalNumber(text); ok {
		return n
	}
	return text
}

// OperandValue turns the literal token of a pipeline operator section (`|> + 5`,
// `math_add 10`) into the value the operator is applied to: a numeric literal becomes the
// number it spells, anything else stays the text. `++` always keeps the text.
func OperandValue(op OpKind, text string) any {
	if op != OpConcat {
		if n, ok := ParseNumberLiteral(text); ok {
			return n
		}
	}
	return text
}

// KindName names the DSL type of a runtime value for error messages: Int, Float, String,
// Char, Bool, … — and the Go type for anything else.
func KindName(v any) string {
	switch v.(type) {
	case string:
		return "String"
	case Char:
		return "Char"
	case bool:
		return "Bool"
	case float32, float64:
		return "Float"
	case complex64, complex128:
		return "Complex"
	case *big.Int:
		return "BigInt"
	case *Ratio:
		return "Ratio"
	case Duration:
		return "Duration"
	case Time:
		return "Time"
	case Date:
		return "Date"
	case Bytes, []byte:
		return "Bytes"
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return "Int"
	}
	for _, c := range []struct {
		name  string
		match func(any) bool
	}{
		{"List", func(v any) bool { l, ok := v.(interface{ IsList() bool }); return ok && l.IsList() }},
		{"Chunk", func(v any) bool { l, ok := v.(interface{ IsChunk() bool }); return ok && l.IsChunk() }},
		{"Set", func(v any) bool { l, ok := v.(interface{ IsSet() bool }); return ok && l.IsSet() }},
		{"HAMT", func(v any) bool { l, ok := v.(interface{ IsHAMT() bool }); return ok && l.IsHAMT() }},
		{"Tree", func(v any) bool { l, ok := v.(interface{ IsTree() bool }); return ok && l.IsTree() }},
		{"Graph", func(v any) bool { l, ok := v.(interface{ IsGraph() bool }); return ok && l.IsGraph() }},
		{"Pair", func(v any) bool { _, ok := v.(interface{ PairElements() (any, any) }); return ok }},
		{"Maybe", func(v any) bool { _, ok := v.(interface{ MaybeValue() (any, bool) }); return ok }},
		{"Either", func(v any) bool { _, ok := v.(interface{ EitherValue() (any, any, bool) }); return ok }},
		{"Variant", func(v any) bool { _, ok := v.(interface{ VariantElements() (string, uint64, any) }); return ok }},
		{"Function", IsCallable},
	} {
		if c.match(v) {
			return c.name
		}
	}
	if v == nil {
		return "Null"
	}
	return fmt.Sprintf("%T", v)
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
	}
	return 0, false
}

// ToSliceAny converts any supported sequence, collection, or Go slice to []any.
func ToSliceAny(val any) ([]any, bool) {
	if val == nil {
		return nil, false
	}
	if eg, ok := val.(interface{ Elements() []any }); ok {
		return eg.Elements(), true
	}
	if sa, ok := val.(interface{ ElementsAny() []any }); ok {
		return sa.ElementsAny(), true
	}
	if p, ok := val.(interface{ PairElements() (any, any) }); ok {
		f, s := p.PairElements()
		return []any{f, s}, true
	}
	rv := reflect.ValueOf(val)
	if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
		n := rv.Len()
		res := make([]any, n)
		for i := 0; i < n; i++ {
			res[i] = rv.Index(i).Interface()
		}
		return res, true
	}
	return nil, false
}

// Compare performs total ordering comparison across scalars, strings, pairs, and sequences.
// Returns -1 for a < b, 0 for a == b, 1 for a > b.
func Compare(a, b any) (int, error) {
	if a == nil && b == nil {
		return 0, nil
	}
	if a == nil {
		return -1, nil
	}
	if b == nil {
		return 1, nil
	}
	// Time values order within one type only (timeops.go).
	if IsTemporal(a) || IsTemporal(b) {
		return temporalCompare(a, b)
	}
	// Numbers order exactly across Int, big integers and Floats (never through a double).
	if cmp, ordered, isReal := CompareNumbers(a, b); isReal {
		if !ordered {
			return 0, fmt.Errorf("NaN is not ordered")
		}
		return cmp, nil
	}
	if StructuralEqual(a, b) {
		return 0, nil
	}

	// Two Strings order as text, never as numbers: "10" < "9".
	if as, ok := a.(string); ok {
		if bs, ok := b.(string); ok {
			return strings.Compare(as, bs), nil
		}
	}

	// Chars order by code point; a one-character String converts on demand (D1).
	if ca, cb, ok := charPair(a, b); ok {
		switch {
		case ca < cb:
			return -1, nil
		case ca > cb:
			return 1, nil
		}
		return 0, nil
	}

	// Strings
	if as, okA := a.(string); okA {
		if bs, okB := b.(string); okB {
			return strings.Compare(as, bs), nil
		}
	}

	// Booleans
	if ab, okA := a.(bool); okA {
		if bb, okB := b.(bool); okB {
			if !ab && bb {
				return -1, nil
			}
			if ab && !bb {
				return 1, nil
			}
			return 0, nil
		}
	}

	// Pairs (lexicographical)
	if pa, okA := a.(interface{ PairElements() (any, any) }); okA {
		if pb, okB := b.(interface{ PairElements() (any, any) }); okB {
			fA, sA := pa.PairElements()
			fB, sB := pb.PairElements()
			cmpF, errF := Compare(fA, fB)
			if errF != nil {
				return 0, errF
			}
			if cmpF != 0 {
				return cmpF, nil
			}
			return Compare(sA, sB)
		}
	}

	// Maybes (Nothing < Just(x))
	if ma, okA := a.(interface{ MaybeValue() (any, bool) }); okA {
		if mb, okB := b.(interface{ MaybeValue() (any, bool) }); okB {
			vA, isJA := ma.MaybeValue()
			vB, isJB := mb.MaybeValue()
			if !isJA && !isJB {
				return 0, nil
			}
			if !isJA && isJB {
				return -1, nil
			}
			if isJA && !isJB {
				return 1, nil
			}
			return Compare(vA, vB)
		}
	}

	// Eithers (Left(x) < Right(y))
	if ea, okA := a.(interface{ EitherValue() (any, any, bool) }); okA {
		if eb, okB := b.(interface{ EitherValue() (any, any, bool) }); okB {
			lA, rA, isLA := ea.EitherValue()
			lB, rB, isLB := eb.EitherValue()
			if isLA && !isLB {
				return -1, nil
			}
			if !isLA && isLB {
				return 1, nil
			}
			if isLA {
				return Compare(lA, lB)
			}
			return Compare(rA, rB)
		}
	}

	// Variants (tag comparison, then payload comparison)
	if va, okA := a.(interface{ VariantElements() (string, uint64, any) }); okA {
		if vb, okB := b.(interface{ VariantElements() (string, uint64, any) }); okB {
			tA, _, valA := va.VariantElements()
			tB, _, valB := vb.VariantElements()
			if cmpTag := strings.Compare(tA, tB); cmpTag != 0 {
				return cmpTag, nil
			}
			return Compare(valA, valB)
		}
	}

	// Sequences (lexicographical: List, Chunk, []any, etc.)
	sliceA, isSliceA := ToSliceAny(a)
	sliceB, isSliceB := ToSliceAny(b)
	if isSliceA && isSliceB {
		minLen := len(sliceA)
		if len(sliceB) < minLen {
			minLen = len(sliceB)
		}
		for i := 0; i < minLen; i++ {
			cmpElem, err := Compare(sliceA[i], sliceB[i])
			if err != nil {
				return 0, err
			}
			if cmpElem != 0 {
				return cmpElem, nil
			}
		}
		if len(sliceA) < len(sliceB) {
			return -1, nil
		}
		if len(sliceA) > len(sliceB) {
			return 1, nil
		}
		return 0, nil
	}

	hint := ""
	_, aStr := a.(string)
	_, bStr := b.(string)
	if (aStr && IsNumber(b)) || (bStr && IsNumber(a)) {
		hint = " (a String is not a number; use str_to_int / str_to_float to parse)"
	}
	return 0, fmt.Errorf("incomparable types: %s and %s%s", KindName(a), KindName(b), hint)
}

// StructuralEqual tests equality across primitives, pointers, complex, and types.
func StructuralEqual(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	typA := reflect.TypeOf(a)
	typB := reflect.TypeOf(b)
	if typA == typB && typA.Comparable() {
		if a == b {
			return true
		}
	}

	// Char equality; a one-character String converts on demand (D1).
	if ca, cb, ok := charPair(a, b); ok {
		return ca == cb
	}

	// Duration, Time and Date equal only a value of their own type (a Time by instant).
	if IsTemporal(a) || IsTemporal(b) {
		c, err := temporalCompare(a, b) // an error: two different types, never equal
		return err == nil && c == 0
	}

	// Numbers: exact across Int, big integers and Floats (a NaN equals nothing); Complex
	// numbers compare by both parts.
	if ca, cb := classOf(a), classOf(b); ca != clsNone && cb != clsNone {
		if ca == clsComplex || cb == clsComplex {
			ac, aok := ToComplex128(a)
			bc, bok := ToComplex128(b)
			return aok && bok && ac == bc
		}
		cmp, ordered, _ := CompareNumbers(a, b)
		return ordered && cmp == 0
	}

	// Pairs
	if pa, okA := a.(interface{ PairElements() (any, any) }); okA {
		if pb, okB := b.(interface{ PairElements() (any, any) }); okB {
			fA, sA := pa.PairElements()
			fB, sB := pb.PairElements()
			return StructuralEqual(fA, fB) && StructuralEqual(sA, sB)
		}
	}

	// Maybes
	if ma, okA := a.(interface{ MaybeValue() (any, bool) }); okA {
		if mb, okB := b.(interface{ MaybeValue() (any, bool) }); okB {
			vA, isJA := ma.MaybeValue()
			vB, isJB := mb.MaybeValue()
			if isJA != isJB {
				return false
			}
			if !isJA {
				return true
			}
			return StructuralEqual(vA, vB)
		}
	}

	// Eithers
	if ea, okA := a.(interface{ EitherValue() (any, any, bool) }); okA {
		if eb, okB := b.(interface{ EitherValue() (any, any, bool) }); okB {
			lA, rA, isLA := ea.EitherValue()
			lB, rB, isLB := eb.EitherValue()
			if isLA != isLB {
				return false
			}
			if isLA {
				return StructuralEqual(lA, lB)
			}
			return StructuralEqual(rA, rB)
		}
	}

	// Variants
	if va, okA := a.(interface{ VariantElements() (string, uint64, any) }); okA {
		if vb, okB := b.(interface{ VariantElements() (string, uint64, any) }); okB {
			tA, fA, valA := va.VariantElements()
			tB, fB, valB := vb.VariantElements()
			if tA != tB || fA != fB {
				return false
			}
			return StructuralEqual(valA, valB)
		}
	}

	// Graphs (structural node + edge equality, order-independent on edges)
	if ga, okA := a.(interface {
		IsGraph() bool
		GraphEqual(any) bool
	}); okA {
		if _, okB := b.(interface{ IsGraph() bool }); okB {
			return ga.GraphEqual(b)
		}
	}

	// Sets (order-independent mathematical equality)
	type setChecker interface {
		IsSet() bool
		Len() int
		ElementsAny() []any
	}
	if sa, okA := a.(setChecker); okA {
		if sb, okB := b.(setChecker); okB {
			if sa.Len() != sb.Len() {
				return false
			}
			elemsA := sa.ElementsAny()
			elemsB := sb.ElementsAny()
			for _, ea := range elemsA {
				found := false
				for _, eb := range elemsB {
					if StructuralEqual(ea, eb) {
						found = true
						break
					}
				}
				if !found {
					return false
				}
			}
			return true
		}
	}

	// HAMTs (key-value structural equality)
	type hamtChecker interface {
		IsHAMT() bool
		Len() int
	}
	if ha, okA := a.(hamtChecker); okA {
		if hb, okB := b.(hamtChecker); okB {
			if ha.Len() != hb.Len() {
				return false
			}
			if ga, okGA := a.(interface{ Graph() any }); okGA {
				if gb, okGB := b.(interface{ Graph() any }); okGB {
					return StructuralEqual(ga.Graph(), gb.Graph())
				}
			}
			return reflect.DeepEqual(a, b)
		}
	}

	// Sequences (order-dependent: List, Chunk, Slice)
	type setIgnorer interface{ IsSet() bool }
	if _, isSetA := a.(setIgnorer); !isSetA {
		if _, isSetB := b.(setIgnorer); !isSetB {
			if seqA, okA := ToSliceAny(a); okA {
				if seqB, okB := ToSliceAny(b); okB {
					if len(seqA) != len(seqB) {
						return false
					}
					for i := range seqA {
						if !StructuralEqual(seqA[i], seqB[i]) {
							return false
						}
					}
					return true
				}
			}
		}
	}

	// There is deliberately no printed-form rule: a String equals only a String (and a
	// one-character String a Char, D1). `"[2 1]" == [2, 1]` is false; a collection used as a
	// HAMT key is carried by value (collections key encoding, evaluator.EncodeKey), not by text.

	// Fallback to reflect.DeepEqual for composite structures
	return reflect.DeepEqual(a, b)
}
