// Code generated by "stringer -type=TyFnc"; DO NOT EDIT.

package functions

import "strconv"

const _TyFnc_name = "TypeDataValueClassLambdaGeneratorAccumulatorPropertyArgumentPatternElementLexicalReturnSymbolArityIndexKeyTrueFalseUndecidedEqualLesserGreaterMinMaxSwitchCaseThenElseJustNoneEitherOrNumbersLettersBytesTextListVectorSetPairEnumTupleRecordStateIOParametric"

var _TyFnc_map = map[TyFnc]string{
	1:              _TyFnc_name[0:4],
	2:              _TyFnc_name[4:8],
	4:              _TyFnc_name[8:13],
	8:              _TyFnc_name[13:18],
	16:             _TyFnc_name[18:24],
	32:             _TyFnc_name[24:33],
	64:             _TyFnc_name[33:44],
	128:            _TyFnc_name[44:52],
	256:            _TyFnc_name[52:60],
	512:            _TyFnc_name[60:67],
	1024:           _TyFnc_name[67:74],
	2048:           _TyFnc_name[74:81],
	4096:           _TyFnc_name[81:87],
	8192:           _TyFnc_name[87:93],
	16384:          _TyFnc_name[93:98],
	32768:          _TyFnc_name[98:103],
	65536:          _TyFnc_name[103:106],
	131072:         _TyFnc_name[106:110],
	262144:         _TyFnc_name[110:115],
	524288:         _TyFnc_name[115:124],
	1048576:        _TyFnc_name[124:129],
	2097152:        _TyFnc_name[129:135],
	4194304:        _TyFnc_name[135:142],
	8388608:        _TyFnc_name[142:145],
	16777216:       _TyFnc_name[145:148],
	33554432:       _TyFnc_name[148:154],
	67108864:       _TyFnc_name[154:158],
	134217728:      _TyFnc_name[158:162],
	268435456:      _TyFnc_name[162:166],
	536870912:      _TyFnc_name[166:170],
	1073741824:     _TyFnc_name[170:174],
	2147483648:     _TyFnc_name[174:180],
	4294967296:     _TyFnc_name[180:182],
	8589934592:     _TyFnc_name[182:189],
	17179869184:    _TyFnc_name[189:196],
	34359738368:    _TyFnc_name[196:201],
	68719476736:    _TyFnc_name[201:205],
	137438953472:   _TyFnc_name[205:209],
	274877906944:   _TyFnc_name[209:215],
	549755813888:   _TyFnc_name[215:218],
	1099511627776:  _TyFnc_name[218:222],
	2199023255552:  _TyFnc_name[222:226],
	4398046511104:  _TyFnc_name[226:231],
	8796093022208:  _TyFnc_name[231:237],
	17592186044416: _TyFnc_name[237:242],
	35184372088832: _TyFnc_name[242:244],
	70368744177664: _TyFnc_name[244:254],
}

func (i TyFnc) String() string {
	if str, ok := _TyFnc_map[i]; ok {
		return str
	}
	return "TyFnc(" + strconv.FormatInt(int64(i), 10) + ")"
}
