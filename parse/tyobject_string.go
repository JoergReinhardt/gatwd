// Code generated by "stringer -type=TyObject"; DO NOT EDIT.

package parse

import "strconv"

const _TyObject_name = "ConstructorFunctionClosureThunkSelectorThunkPartialApplicationGenericApplicationStackApplicationIndirectionByteCodeObjectBlackHoleArrayIOByteStream"

var _TyObject_map = map[TyObject]string{
	1:    _TyObject_name[0:11],
	2:    _TyObject_name[11:26],
	4:    _TyObject_name[26:31],
	8:    _TyObject_name[31:44],
	16:   _TyObject_name[44:62],
	32:   _TyObject_name[62:80],
	64:   _TyObject_name[80:96],
	128:  _TyObject_name[96:107],
	256:  _TyObject_name[107:121],
	512:  _TyObject_name[121:130],
	1024: _TyObject_name[130:135],
	2048: _TyObject_name[135:147],
}

func (i TyObject) String() string {
	if str, ok := _TyObject_map[i]; ok {
		return str
	}
	return "TyObject(" + strconv.FormatInt(int64(i), 10) + ")"
}
