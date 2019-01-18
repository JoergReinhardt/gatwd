// Code generated by "stringer -type=Type"; DO NOT EDIT.

package data

import "strconv"

const _Type_name = "NilBoolIntInt8Int16Int32BigIntUintUint8Uint16Uint32FloatFlt32BigFltRatioImagImag64ByteRuneBytesStringTimeDurationErrorTupleRecordVectorListFunctionArgumentParameterConstantFlagMAX_INT"

var _Type_map = map[Type]string{
	1:                    _Type_name[0:3],
	2:                    _Type_name[3:7],
	4:                    _Type_name[7:10],
	8:                    _Type_name[10:14],
	16:                   _Type_name[14:19],
	32:                   _Type_name[19:24],
	64:                   _Type_name[24:30],
	128:                  _Type_name[30:34],
	256:                  _Type_name[34:39],
	512:                  _Type_name[39:45],
	1024:                 _Type_name[45:51],
	2048:                 _Type_name[51:56],
	4096:                 _Type_name[56:61],
	8192:                 _Type_name[61:67],
	16384:                _Type_name[67:72],
	32768:                _Type_name[72:76],
	65536:                _Type_name[76:82],
	131072:               _Type_name[82:86],
	262144:               _Type_name[86:90],
	524288:               _Type_name[90:95],
	1048576:              _Type_name[95:101],
	2097152:              _Type_name[101:105],
	4194304:              _Type_name[105:113],
	8388608:              _Type_name[113:118],
	16777216:             _Type_name[118:123],
	33554432:             _Type_name[123:129],
	67108864:             _Type_name[129:135],
	134217728:            _Type_name[135:139],
	268435456:            _Type_name[139:147],
	536870912:            _Type_name[147:155],
	1073741824:           _Type_name[155:164],
	2147483648:           _Type_name[164:172],
	4294967296:           _Type_name[172:176],
	18446744073709551615: _Type_name[176:183],
}

func (i Type) String() string {
	if str, ok := _Type_map[i]; ok {
		return str
	}
	return "Type(" + strconv.FormatInt(int64(i), 10) + ")"
}
