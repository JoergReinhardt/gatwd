// Code generated by "stringer -type=TyPrimitive"; DO NOT EDIT.

package data

import "strconv"

const _TyPrimitive_name = "NilBoolInt8Int16Int32IntBigIntUint8Uint16Uint32UintFlt32FloatBigFltRatioImag64ImagTimeDurationByteRuneBytesStringErrorPairTupleRecordVectorListSetFunctionFlagMAX_INT"

var _TyPrimitive_map = map[TyPrimitive]string{
	1:                    _TyPrimitive_name[0:3],
	2:                    _TyPrimitive_name[3:7],
	4:                    _TyPrimitive_name[7:11],
	8:                    _TyPrimitive_name[11:16],
	16:                   _TyPrimitive_name[16:21],
	32:                   _TyPrimitive_name[21:24],
	64:                   _TyPrimitive_name[24:30],
	128:                  _TyPrimitive_name[30:35],
	256:                  _TyPrimitive_name[35:41],
	512:                  _TyPrimitive_name[41:47],
	1024:                 _TyPrimitive_name[47:51],
	2048:                 _TyPrimitive_name[51:56],
	4096:                 _TyPrimitive_name[56:61],
	8192:                 _TyPrimitive_name[61:67],
	16384:                _TyPrimitive_name[67:72],
	32768:                _TyPrimitive_name[72:78],
	65536:                _TyPrimitive_name[78:82],
	131072:               _TyPrimitive_name[82:86],
	262144:               _TyPrimitive_name[86:94],
	524288:               _TyPrimitive_name[94:98],
	1048576:              _TyPrimitive_name[98:102],
	2097152:              _TyPrimitive_name[102:107],
	4194304:              _TyPrimitive_name[107:113],
	8388608:              _TyPrimitive_name[113:118],
	16777216:             _TyPrimitive_name[118:122],
	33554432:             _TyPrimitive_name[122:127],
	67108864:             _TyPrimitive_name[127:133],
	134217728:            _TyPrimitive_name[133:139],
	268435456:            _TyPrimitive_name[139:143],
	536870912:            _TyPrimitive_name[143:146],
	1073741824:           _TyPrimitive_name[146:154],
	2147483648:           _TyPrimitive_name[154:158],
	18446744073709551615: _TyPrimitive_name[158:165],
}

func (i TyPrimitive) String() string {
	if str, ok := _TyPrimitive_map[i]; ok {
		return str
	}
	return "TyPrimitive(" + strconv.FormatInt(int64(i), 10) + ")"
}
