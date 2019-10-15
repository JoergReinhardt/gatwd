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
	_ = x[Constant-8]
	_ = x[Generator-16]
	_ = x[Accumulator-32]
	_ = x[Constructor-64]
	_ = x[Property-128]
	_ = x[Argument-256]
	_ = x[Pattern-512]
	_ = x[Element-1024]
	_ = x[Lexical-2048]
	_ = x[Symbol-4096]
	_ = x[Index-8192]
	_ = x[Key-16384]
	_ = x[True-32768]
	_ = x[False-65536]
	_ = x[Undecided-131072]
	_ = x[Equal-262144]
	_ = x[Lesser-524288]
	_ = x[Greater-1048576]
	_ = x[Min-2097152]
	_ = x[Max-4194304]
	_ = x[Switch-8388608]
	_ = x[Case-16777216]
	_ = x[Just-33554432]
	_ = x[None-67108864]
	_ = x[Option-134217728]
	_ = x[Polymorph-268435456]
	_ = x[Either-536870912]
	_ = x[Or-1073741824]
	_ = x[Natural-2147483648]
	_ = x[Integer-4294967296]
	_ = x[Real-8589934592]
	_ = x[Ratio-17179869184]
	_ = x[Letter-34359738368]
	_ = x[Text-68719476736]
	_ = x[Bytes-137438953472]
	_ = x[Vector-274877906944]
	_ = x[List-549755813888]
	_ = x[Set-1099511627776]
	_ = x[Pair-2199023255552]
	_ = x[Enum-4398046511104]
	_ = x[Tuple-8796093022208]
	_ = x[Record-17592186044416]
	_ = x[Monad-35184372088832]
	_ = x[State-70368744177664]
	_ = x[IO-140737488355328]
	_ = x[Parametric-281474976710656]
	_ = x[ALL-18446744073709551615]
}

const _TyFnc_name = "TypeDataValueConstantGeneratorAccumulatorConstructorPropertyArgumentPatternElementLexicalSymbolIndexKeyTrueFalseUndecidedEqualLesserGreaterMinMaxSwitchCaseJustNoneOptionPolymorphEitherOrNaturalIntegerRealRatioLetterTextBytesVectorListSetPairEnumTupleRecordMonadStateIOParametricALL"

var _TyFnc_map = map[TyFnc]string{
	1:                    _TyFnc_name[0:4],
	2:                    _TyFnc_name[4:8],
	4:                    _TyFnc_name[8:13],
	8:                    _TyFnc_name[13:21],
	16:                   _TyFnc_name[21:30],
	32:                   _TyFnc_name[30:41],
	64:                   _TyFnc_name[41:52],
	128:                  _TyFnc_name[52:60],
	256:                  _TyFnc_name[60:68],
	512:                  _TyFnc_name[68:75],
	1024:                 _TyFnc_name[75:82],
	2048:                 _TyFnc_name[82:89],
	4096:                 _TyFnc_name[89:95],
	8192:                 _TyFnc_name[95:100],
	16384:                _TyFnc_name[100:103],
	32768:                _TyFnc_name[103:107],
	65536:                _TyFnc_name[107:112],
	131072:               _TyFnc_name[112:121],
	262144:               _TyFnc_name[121:126],
	524288:               _TyFnc_name[126:132],
	1048576:              _TyFnc_name[132:139],
	2097152:              _TyFnc_name[139:142],
	4194304:              _TyFnc_name[142:145],
	8388608:              _TyFnc_name[145:151],
	16777216:             _TyFnc_name[151:155],
	33554432:             _TyFnc_name[155:159],
	67108864:             _TyFnc_name[159:163],
	134217728:            _TyFnc_name[163:169],
	268435456:            _TyFnc_name[169:178],
	536870912:            _TyFnc_name[178:184],
	1073741824:           _TyFnc_name[184:186],
	2147483648:           _TyFnc_name[186:193],
	4294967296:           _TyFnc_name[193:200],
	8589934592:           _TyFnc_name[200:204],
	17179869184:          _TyFnc_name[204:209],
	34359738368:          _TyFnc_name[209:215],
	68719476736:          _TyFnc_name[215:219],
	137438953472:         _TyFnc_name[219:224],
	274877906944:         _TyFnc_name[224:230],
	549755813888:         _TyFnc_name[230:234],
	1099511627776:        _TyFnc_name[234:237],
	2199023255552:        _TyFnc_name[237:241],
	4398046511104:        _TyFnc_name[241:245],
	8796093022208:        _TyFnc_name[245:250],
	17592186044416:       _TyFnc_name[250:256],
	35184372088832:       _TyFnc_name[256:261],
	70368744177664:       _TyFnc_name[261:266],
	140737488355328:      _TyFnc_name[266:268],
	281474976710656:      _TyFnc_name[268:278],
	18446744073709551615: _TyFnc_name[278:281],
}

func (i TyFnc) String() string {
	if str, ok := _TyFnc_map[i]; ok {
		return str
	}
	return "TyFnc(" + strconv.FormatInt(int64(i), 10) + ")"
}
