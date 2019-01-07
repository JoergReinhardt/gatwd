// Code generated by "stringer -type=TokType"; DO NOT EDIT.

package lang

import "strconv"

const _TokType_name = "NoneBlankUnderscoreAsteriskDotCommaColonSemicolonMinusPlusOrXorAndEqualLesserGreaterLesseqGreaterqLeftParRightParLeftBraRightBraLeftCurRightCurSlashNotDecIncDoubEqualTripEqualRightArrowLeftArrowFatLArrowFatRArrowDoubColSing_quoteDoub_quoteBckSlaLambdaNumberLetterCapitalGenTypeHeadWordTailWordInWordConWordLetWordMutableWordWhereWordOtherwiseWordIfWordThenWordElseWordCaseWordOfWordDataWordTypeWordTypeIdentFuncIdent"

var _TokType_map = map[TokType]string{
	1:                  _TokType_name[0:4],
	2:                  _TokType_name[4:9],
	4:                  _TokType_name[9:19],
	8:                  _TokType_name[19:27],
	16:                 _TokType_name[27:30],
	32:                 _TokType_name[30:35],
	64:                 _TokType_name[35:40],
	128:                _TokType_name[40:49],
	256:                _TokType_name[49:54],
	512:                _TokType_name[54:58],
	1024:               _TokType_name[58:60],
	2048:               _TokType_name[60:63],
	4096:               _TokType_name[63:66],
	8192:               _TokType_name[66:71],
	16384:              _TokType_name[71:77],
	32768:              _TokType_name[77:84],
	65536:              _TokType_name[84:90],
	131072:             _TokType_name[90:98],
	262144:             _TokType_name[98:105],
	524288:             _TokType_name[105:113],
	1048576:            _TokType_name[113:120],
	2097152:            _TokType_name[120:128],
	4194304:            _TokType_name[128:135],
	8388608:            _TokType_name[135:143],
	16777216:           _TokType_name[143:148],
	33554432:           _TokType_name[148:151],
	67108864:           _TokType_name[151:154],
	134217728:          _TokType_name[154:157],
	268435456:          _TokType_name[157:166],
	536870912:          _TokType_name[166:175],
	1073741824:         _TokType_name[175:185],
	2147483648:         _TokType_name[185:194],
	4294967296:         _TokType_name[194:203],
	8589934592:         _TokType_name[203:212],
	17179869184:        _TokType_name[212:219],
	34359738368:        _TokType_name[219:229],
	68719476736:        _TokType_name[229:239],
	137438953472:       _TokType_name[239:245],
	274877906944:       _TokType_name[245:251],
	549755813888:       _TokType_name[251:257],
	1099511627776:      _TokType_name[257:263],
	2199023255552:      _TokType_name[263:270],
	4398046511104:      _TokType_name[270:277],
	8796093022208:      _TokType_name[277:285],
	17592186044416:     _TokType_name[285:293],
	35184372088832:     _TokType_name[293:299],
	70368744177664:     _TokType_name[299:306],
	140737488355328:    _TokType_name[306:313],
	281474976710656:    _TokType_name[313:324],
	562949953421312:    _TokType_name[324:333],
	1125899906842624:   _TokType_name[333:346],
	2251799813685248:   _TokType_name[346:352],
	4503599627370496:   _TokType_name[352:360],
	9007199254740992:   _TokType_name[360:368],
	18014398509481984:  _TokType_name[368:376],
	36028797018963968:  _TokType_name[376:382],
	72057594037927936:  _TokType_name[382:390],
	144115188075855872: _TokType_name[390:398],
	288230376151711744: _TokType_name[398:407],
	576460752303423488: _TokType_name[407:416],
}

func (i TokType) String() string {
	if str, ok := _TokType_map[i]; ok {
		return str
	}
	return "TokType(" + strconv.FormatInt(int64(i), 10) + ")"
}
