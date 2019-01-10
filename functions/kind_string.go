// Code generated by "stringer -type=Kind"; DO NOT EDIT.

package functions

import "strconv"

const _Kind_name = "ValueParameterAttributAccessorDoubleVectorConstantUnaryBinaryNnaryTupleListChainUniSetMuliSetAssocARecordLinkDLinkNodeTreeInternal"

var _Kind_map = map[Kind]string{
	1:       _Kind_name[0:5],
	2:       _Kind_name[5:14],
	4:       _Kind_name[14:22],
	8:       _Kind_name[22:30],
	16:      _Kind_name[30:36],
	32:      _Kind_name[36:42],
	64:      _Kind_name[42:50],
	128:     _Kind_name[50:55],
	256:     _Kind_name[55:61],
	512:     _Kind_name[61:66],
	1024:    _Kind_name[66:71],
	2048:    _Kind_name[71:75],
	4096:    _Kind_name[75:80],
	8192:    _Kind_name[80:86],
	16384:   _Kind_name[86:93],
	32768:   _Kind_name[93:99],
	65536:   _Kind_name[99:105],
	131072:  _Kind_name[105:109],
	262144:  _Kind_name[109:114],
	524288:  _Kind_name[114:118],
	1048576: _Kind_name[118:122],
	2097152: _Kind_name[122:130],
}

func (i Kind) String() string {
	if str, ok := _Kind_map[i]; ok {
		return str
	}
	return "Kind(" + strconv.FormatInt(int64(i), 10) + ")"
}
