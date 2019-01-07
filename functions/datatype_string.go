// Code generated by "stringer -type=DataType"; DO NOT EDIT.

package functions

import "strconv"

const _DataType_name = "DataPairVectorConstantUnaryBinaryNnaryTupleListChainUniSetMuliSetAssocARecordLinkDLinkNodeTree"

var _DataType_map = map[DataType]string{
	1:      _DataType_name[0:4],
	2:      _DataType_name[4:8],
	4:      _DataType_name[8:14],
	8:      _DataType_name[14:22],
	16:     _DataType_name[22:27],
	32:     _DataType_name[27:33],
	64:     _DataType_name[33:38],
	128:    _DataType_name[38:43],
	256:    _DataType_name[43:47],
	512:    _DataType_name[47:52],
	1024:   _DataType_name[52:58],
	2048:   _DataType_name[58:65],
	4096:   _DataType_name[65:71],
	8192:   _DataType_name[71:77],
	16384:  _DataType_name[77:81],
	32768:  _DataType_name[81:86],
	65536:  _DataType_name[86:90],
	131072: _DataType_name[90:94],
}

func (i DataType) String() string {
	if str, ok := _DataType_map[i]; ok {
		return str
	}
	return "DataType(" + strconv.FormatInt(int64(i), 10) + ")"
}
