// Code generated by "stringer -type=TyFnc"; DO NOT EDIT.

package functions

import "strconv"

const _TyFnc_name = "TypeDataValueClassLambdaConstantGeneratorAccumulatorPropertyArgumentPatternElementLexicalSymbolArityIndexKeyTrueFalseUndecidedEqualLesserGreaterMinMaxSwitchCaseThenElseJustNoneEitherOrNumbersLettersBytesTextListVectorSetPairEnumTupleRecordStateIOParametric"

var _TyFnc_map = map[TyFnc]string{
	1:              _TyFnc_name[0:4],
	2:              _TyFnc_name[4:8],
	4:              _TyFnc_name[8:13],
	8:              _TyFnc_name[13:18],
	16:             _TyFnc_name[18:24],
	32:             _TyFnc_name[24:32],
	64:             _TyFnc_name[32:41],
	128:            _TyFnc_name[41:52],
	256:            _TyFnc_name[52:60],
	512:            _TyFnc_name[60:68],
	1024:           _TyFnc_name[68:75],
	2048:           _TyFnc_name[75:82],
	4096:           _TyFnc_name[82:89],
	8192:           _TyFnc_name[89:95],
	16384:          _TyFnc_name[95:100],
	32768:          _TyFnc_name[100:105],
	65536:          _TyFnc_name[105:108],
	131072:         _TyFnc_name[108:112],
	262144:         _TyFnc_name[112:117],
	524288:         _TyFnc_name[117:126],
	1048576:        _TyFnc_name[126:131],
	2097152:        _TyFnc_name[131:137],
	4194304:        _TyFnc_name[137:144],
	8388608:        _TyFnc_name[144:147],
	16777216:       _TyFnc_name[147:150],
	33554432:       _TyFnc_name[150:156],
	67108864:       _TyFnc_name[156:160],
	134217728:      _TyFnc_name[160:164],
	268435456:      _TyFnc_name[164:168],
	536870912:      _TyFnc_name[168:172],
	1073741824:     _TyFnc_name[172:176],
	2147483648:     _TyFnc_name[176:182],
	4294967296:     _TyFnc_name[182:184],
	8589934592:     _TyFnc_name[184:191],
	17179869184:    _TyFnc_name[191:198],
	34359738368:    _TyFnc_name[198:203],
	68719476736:    _TyFnc_name[203:207],
	137438953472:   _TyFnc_name[207:211],
	274877906944:   _TyFnc_name[211:217],
	549755813888:   _TyFnc_name[217:220],
	1099511627776:  _TyFnc_name[220:224],
	2199023255552:  _TyFnc_name[224:228],
	4398046511104:  _TyFnc_name[228:233],
	8796093022208:  _TyFnc_name[233:239],
	17592186044416: _TyFnc_name[239:244],
	35184372088832: _TyFnc_name[244:246],
	70368744177664: _TyFnc_name[246:256],
}

func (i TyFnc) String() string {
	if str, ok := _TyFnc_map[i]; ok {
		return str
	}
	return "TyFnc(" + strconv.FormatInt(int64(i), 10) + ")"
}
