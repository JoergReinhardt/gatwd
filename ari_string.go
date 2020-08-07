// Code generated by "stringer -type Ari"; DO NOT EDIT.

package gatw

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Variadic - -1]
	_ = x[Nullary-1]
	_ = x[Unary-2]
	_ = x[Binary-3]
	_ = x[Ternary-4]
	_ = x[Quaternary-5]
	_ = x[Quinary-6]
	_ = x[Senary-7]
	_ = x[Septenary-8]
	_ = x[Octonary-9]
	_ = x[Novenary-10]
	_ = x[Denary-11]
	_ = x[Multary-12]
}

const (
	_Ari_name_0 = "Variadic"
	_Ari_name_1 = "NullaryUnaryBinaryTernaryQuaternaryQuinarySenarySeptenaryOctonaryNovenaryDenaryMultary"
)

var (
	_Ari_index_1 = [...]uint8{0, 7, 12, 18, 25, 35, 42, 48, 57, 65, 73, 79, 86}
)

func (i Ari) String() string {
	switch {
	case i == -1:
		return _Ari_name_0
	case 1 <= i && i <= 12:
		i -= 1
		return _Ari_name_1[_Ari_index_1[i]:_Ari_index_1[i+1]]
	default:
		return "Ari(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}