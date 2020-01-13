// Code generated by "stringer -type=TyFnc"; DO NOT EDIT.

package functions

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Type-1]
	_ = x[Data-2]
	_ = x[Value-4]
	_ = x[Partial-8]
	_ = x[Constant-16]
	_ = x[Generator-32]
	_ = x[Accumulator-64]
	_ = x[Constructor-128]
	_ = x[Parameter-256]
	_ = x[Property-512]
	_ = x[Lexical-1024]
	_ = x[Symbol-2048]
	_ = x[Index-4096]
	_ = x[Key-8192]
	_ = x[True-16384]
	_ = x[False-32768]
	_ = x[Undecided-65536]
	_ = x[Equal-131072]
	_ = x[Lesser-262144]
	_ = x[Greater-524288]
	_ = x[Min-1048576]
	_ = x[Max-2097152]
	_ = x[Polymorph-4194304]
	_ = x[Switch-8388608]
	_ = x[Case-16777216]
	_ = x[Just-33554432]
	_ = x[None-67108864]
	_ = x[Either-134217728]
	_ = x[Or-268435456]
	_ = x[Natural-536870912]
	_ = x[Integer-1073741824]
	_ = x[Real-2147483648]
	_ = x[Ratio-4294967296]
	_ = x[Letter-8589934592]
	_ = x[String-17179869184]
	_ = x[Byte-34359738368]
	_ = x[Vector-68719476736]
	_ = x[List-137438953472]
	_ = x[Enum-274877906944]
	_ = x[Set-549755813888]
	_ = x[Pair-1099511627776]
	_ = x[Tuple-2199023255552]
	_ = x[Record-4398046511104]
	_ = x[HashMap-8796093022208]
	_ = x[Group-17592186044416]
	_ = x[Functor-35184372088832]
	_ = x[Applicative-70368744177664]
	_ = x[Monad-140737488355328]
	_ = x[State-281474976710656]
	_ = x[IO-562949953421312]
	_ = x[T-18446744073709551615]
}

const _TyFnc_name = "TypeDataValuePartialConstantGeneratorAccumulatorConstructorParameterPropertyLexicalSymbolIndexKeyTrueFalseUndecidedEqualLesserGreaterMinMaxPolymorphSwitchCaseJustNoneEitherOrNaturalIntegerRealRatioLetterStringByteVectorListEnumSetPairTupleRecordHashMapGroupFunctorApplicativeMonadStateIOT"

var _TyFnc_map = map[TyFnc]string{
	1:                    _TyFnc_name[0:4],
	2:                    _TyFnc_name[4:8],
	4:                    _TyFnc_name[8:13],
	8:                    _TyFnc_name[13:20],
	16:                   _TyFnc_name[20:28],
	32:                   _TyFnc_name[28:37],
	64:                   _TyFnc_name[37:48],
	128:                  _TyFnc_name[48:59],
	256:                  _TyFnc_name[59:68],
	512:                  _TyFnc_name[68:76],
	1024:                 _TyFnc_name[76:83],
	2048:                 _TyFnc_name[83:89],
	4096:                 _TyFnc_name[89:94],
	8192:                 _TyFnc_name[94:97],
	16384:                _TyFnc_name[97:101],
	32768:                _TyFnc_name[101:106],
	65536:                _TyFnc_name[106:115],
	131072:               _TyFnc_name[115:120],
	262144:               _TyFnc_name[120:126],
	524288:               _TyFnc_name[126:133],
	1048576:              _TyFnc_name[133:136],
	2097152:              _TyFnc_name[136:139],
	4194304:              _TyFnc_name[139:148],
	8388608:              _TyFnc_name[148:154],
	16777216:             _TyFnc_name[154:158],
	33554432:             _TyFnc_name[158:162],
	67108864:             _TyFnc_name[162:166],
	134217728:            _TyFnc_name[166:172],
	268435456:            _TyFnc_name[172:174],
	536870912:            _TyFnc_name[174:181],
	1073741824:           _TyFnc_name[181:188],
	2147483648:           _TyFnc_name[188:192],
	4294967296:           _TyFnc_name[192:197],
	8589934592:           _TyFnc_name[197:203],
	17179869184:          _TyFnc_name[203:209],
	34359738368:          _TyFnc_name[209:213],
	68719476736:          _TyFnc_name[213:219],
	137438953472:         _TyFnc_name[219:223],
	274877906944:         _TyFnc_name[223:227],
	549755813888:         _TyFnc_name[227:230],
	1099511627776:        _TyFnc_name[230:234],
	2199023255552:        _TyFnc_name[234:239],
	4398046511104:        _TyFnc_name[239:245],
	8796093022208:        _TyFnc_name[245:252],
	17592186044416:       _TyFnc_name[252:257],
	35184372088832:       _TyFnc_name[257:264],
	70368744177664:       _TyFnc_name[264:275],
	140737488355328:      _TyFnc_name[275:280],
	281474976710656:      _TyFnc_name[280:285],
	562949953421312:      _TyFnc_name[285:287],
	18446744073709551615: _TyFnc_name[287:288],
}

func (i TyFnc) String() string {
	if str, ok := _TyFnc_map[i]; ok {
		return str
	}
	return "TyFnc(" + strconv.FormatInt(int64(i), 10) + ")"
}
