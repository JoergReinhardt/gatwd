// Code generated by "stringer -type TyAri"; DO NOT EDIT.

package functions

import "strconv"

const _TyAri_name = "NaryNullaryUnaryBinaryTernaryQuaternaryQuinarySenarySeptenaryOctonaryNovenaryDenary"

var _TyAri_index = [...]uint8{0, 4, 11, 16, 22, 29, 39, 46, 52, 61, 69, 77, 83}

func (i TyAri) String() string {
	i -= -1
	if i < 0 || i >= TyAri(len(_TyAri_index)-1) {
		return "TyAri(" + strconv.FormatInt(int64(i+-1), 10) + ")"
	}
	return _TyAri_name[_TyAri_index[i]:_TyAri_index[i+1]]
}