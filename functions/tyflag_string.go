// Code generated by "stringer -type TyFlag"; DO NOT EDIT.

package functions

import "strconv"

const (
	_TyFlag_name_0 = "Flag_BitFlagFlag_NativeFlag_FunctionalFlag_DataConsFlag_ArityFlag_Prop"
	_TyFlag_name_1 = "Flag_Def"
)

var (
	_TyFlag_index_0 = [...]uint8{0, 12, 23, 38, 51, 61, 70}
)

func (i TyFlag) String() string {
	switch {
	case 0 <= i && i <= 5:
		return _TyFlag_name_0[_TyFlag_index_0[i]:_TyFlag_index_0[i+1]]
	case i == 255:
		return _TyFlag_name_1
	default:
		return "TyFlag(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}
