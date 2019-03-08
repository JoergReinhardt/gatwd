// Code generated by "stringer -type=SyntaxItemFlag"; DO NOT EDIT.

package lex

import "strconv"

const _SyntaxItemFlag_name = "NoneBlankTabNewLineUnderscoreSquareRootAsteriskFullstopEllipsisCommaColonSemicolonSubstractionAdditionDotTimesDotProductCrossProductDivisionInfiniteOrXorAndEqualLesserGreaterLesserEqGreaterEqLeftParRightParLeftBraRightBraLeftCurRightCurSlashPipeNotUnequalDecrementIncrementDoubleEqualTripEqualRightArrowLeftArrowFatLArrowFatRArrowDoubColSing_quoteDoub_quoteBackSlashLambdaFunctionPolymorphMonadParameterSequenceSequenceRevIntegralIsMemberEmptySetNumberTextEtaEpsilon"

var _SyntaxItemFlag_map = map[SyntaxItemFlag]string{
	0:                   _SyntaxItemFlag_name[0:4],
	1:                   _SyntaxItemFlag_name[4:9],
	4:                   _SyntaxItemFlag_name[9:12],
	8:                   _SyntaxItemFlag_name[12:19],
	16:                  _SyntaxItemFlag_name[19:29],
	32:                  _SyntaxItemFlag_name[29:39],
	64:                  _SyntaxItemFlag_name[39:47],
	128:                 _SyntaxItemFlag_name[47:55],
	256:                 _SyntaxItemFlag_name[55:63],
	512:                 _SyntaxItemFlag_name[63:68],
	1024:                _SyntaxItemFlag_name[68:73],
	2048:                _SyntaxItemFlag_name[73:82],
	4096:                _SyntaxItemFlag_name[82:94],
	8192:                _SyntaxItemFlag_name[94:102],
	16384:               _SyntaxItemFlag_name[102:105],
	32768:               _SyntaxItemFlag_name[105:110],
	65536:               _SyntaxItemFlag_name[110:120],
	131072:              _SyntaxItemFlag_name[120:132],
	262144:              _SyntaxItemFlag_name[132:140],
	524288:              _SyntaxItemFlag_name[140:148],
	1048576:             _SyntaxItemFlag_name[148:150],
	2097152:             _SyntaxItemFlag_name[150:153],
	4194304:             _SyntaxItemFlag_name[153:156],
	8388608:             _SyntaxItemFlag_name[156:161],
	16777216:            _SyntaxItemFlag_name[161:167],
	33554432:            _SyntaxItemFlag_name[167:174],
	67108864:            _SyntaxItemFlag_name[174:182],
	134217728:           _SyntaxItemFlag_name[182:191],
	268435456:           _SyntaxItemFlag_name[191:198],
	536870912:           _SyntaxItemFlag_name[198:206],
	1073741824:          _SyntaxItemFlag_name[206:213],
	2147483648:          _SyntaxItemFlag_name[213:221],
	4294967296:          _SyntaxItemFlag_name[221:228],
	8589934592:          _SyntaxItemFlag_name[228:236],
	17179869184:         _SyntaxItemFlag_name[236:241],
	34359738368:         _SyntaxItemFlag_name[241:245],
	68719476736:         _SyntaxItemFlag_name[245:248],
	137438953472:        _SyntaxItemFlag_name[248:255],
	274877906944:        _SyntaxItemFlag_name[255:264],
	549755813888:        _SyntaxItemFlag_name[264:273],
	1099511627776:       _SyntaxItemFlag_name[273:284],
	2199023255552:       _SyntaxItemFlag_name[284:293],
	4398046511104:       _SyntaxItemFlag_name[293:303],
	8796093022208:       _SyntaxItemFlag_name[303:312],
	17592186044416:      _SyntaxItemFlag_name[312:321],
	35184372088832:      _SyntaxItemFlag_name[321:330],
	70368744177664:      _SyntaxItemFlag_name[330:337],
	140737488355328:     _SyntaxItemFlag_name[337:347],
	281474976710656:     _SyntaxItemFlag_name[347:357],
	562949953421312:     _SyntaxItemFlag_name[357:366],
	1125899906842624:    _SyntaxItemFlag_name[366:372],
	2251799813685248:    _SyntaxItemFlag_name[372:380],
	4503599627370496:    _SyntaxItemFlag_name[380:389],
	9007199254740992:    _SyntaxItemFlag_name[389:394],
	18014398509481984:   _SyntaxItemFlag_name[394:403],
	36028797018963968:   _SyntaxItemFlag_name[403:411],
	72057594037927936:   _SyntaxItemFlag_name[411:422],
	144115188075855872:  _SyntaxItemFlag_name[422:430],
	288230376151711744:  _SyntaxItemFlag_name[430:438],
	576460752303423488:  _SyntaxItemFlag_name[438:446],
	1152921504606846976: _SyntaxItemFlag_name[446:452],
	2305843009213693952: _SyntaxItemFlag_name[452:456],
	4611686018427387904: _SyntaxItemFlag_name[456:459],
	9223372036854775808: _SyntaxItemFlag_name[459:466],
}

func (i SyntaxItemFlag) String() string {
	if str, ok := _SyntaxItemFlag_map[i]; ok {
		return str
	}
	return "SyntaxItemFlag(" + strconv.FormatInt(int64(i), 10) + ")"
}
