// Code generated by "stringer -type=TyFnc"; DO NOT EDIT.

package functions

import "strconv"

const _TyFnc_name = "TypeNativeDataExpressionCallArityCallPropertysListVectorTupleRecordEnumSetPairApplicableOperatorFunctorMonadUndecidedFalseTrueEqualLesserGreaterJustNoneCaseSwitchEitherOrIfElseDoWhileHigherOrder"

var _TyFnc_map = map[TyFnc]string{
	1:          _TyFnc_name[0:4],
	2:          _TyFnc_name[4:10],
	4:          _TyFnc_name[10:14],
	8:          _TyFnc_name[14:24],
	16:         _TyFnc_name[24:33],
	32:         _TyFnc_name[33:46],
	64:         _TyFnc_name[46:50],
	128:        _TyFnc_name[50:56],
	256:        _TyFnc_name[56:61],
	512:        _TyFnc_name[61:67],
	1024:       _TyFnc_name[67:71],
	2048:       _TyFnc_name[71:74],
	4096:       _TyFnc_name[74:78],
	8192:       _TyFnc_name[78:88],
	16384:      _TyFnc_name[88:96],
	32768:      _TyFnc_name[96:103],
	65536:      _TyFnc_name[103:108],
	131072:     _TyFnc_name[108:117],
	262144:     _TyFnc_name[117:122],
	524288:     _TyFnc_name[122:126],
	1048576:    _TyFnc_name[126:131],
	2097152:    _TyFnc_name[131:137],
	4194304:    _TyFnc_name[137:144],
	8388608:    _TyFnc_name[144:148],
	16777216:   _TyFnc_name[148:152],
	33554432:   _TyFnc_name[152:156],
	67108864:   _TyFnc_name[156:162],
	134217728:  _TyFnc_name[162:168],
	268435456:  _TyFnc_name[168:170],
	536870912:  _TyFnc_name[170:172],
	1073741824: _TyFnc_name[172:176],
	2147483648: _TyFnc_name[176:178],
	4294967296: _TyFnc_name[178:183],
	8589934592: _TyFnc_name[183:194],
}

func (i TyFnc) String() string {
	if str, ok := _TyFnc_map[i]; ok {
		return str
	}
	return "TyFnc(" + strconv.FormatInt(int64(i), 10) + ")"
}
