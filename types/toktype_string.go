// Code generated by "stringer -type=TokType"; DO NOT EDIT.

package types

import "strconv"

const _TokType_name = "tok_nonetok_blanktok_underscoretok_asterisktok_dottok_commatok_colontok_semicolontok_minustok_plustok_ortok_xortok_andtok_equaltok_lessertok_greatertok_leftPartok_rightPartok_leftBratok_rightBratok_leftCurtok_rightCurtok_slashtok_nottok_dectok_inctok_doubEqualtok_rightArrowtok_leftArrowtok_fatLArrowtok_fatRArrowtok_doubColtok_sing_quotetok_doub_quotetok_bckSlatok_lambdatok_numbertok_lettertok_capitaltok_genTypetok_headWordtok_tailWordtok_inWordtok_conWordtok_letWordtok_whereWordtok_otherwiseWordtok_ifWordtok_thenWordtok_elseWordtok_caseWordtok_ofWordtok_dataWordtok_typeWordtok_typeIdenttok_funcIdent"

var _TokType_map = map[TokType]string{
	1:                 _TokType_name[0:8],
	2:                 _TokType_name[8:17],
	4:                 _TokType_name[17:31],
	8:                 _TokType_name[31:43],
	16:                _TokType_name[43:50],
	32:                _TokType_name[50:59],
	64:                _TokType_name[59:68],
	128:               _TokType_name[68:81],
	256:               _TokType_name[81:90],
	512:               _TokType_name[90:98],
	1024:              _TokType_name[98:104],
	2048:              _TokType_name[104:111],
	4096:              _TokType_name[111:118],
	8192:              _TokType_name[118:127],
	16384:             _TokType_name[127:137],
	32768:             _TokType_name[137:148],
	65536:             _TokType_name[148:159],
	131072:            _TokType_name[159:171],
	262144:            _TokType_name[171:182],
	524288:            _TokType_name[182:194],
	1048576:           _TokType_name[194:205],
	2097152:           _TokType_name[205:217],
	4194304:           _TokType_name[217:226],
	8388608:           _TokType_name[226:233],
	16777216:          _TokType_name[233:240],
	33554432:          _TokType_name[240:247],
	67108864:          _TokType_name[247:260],
	134217728:         _TokType_name[260:274],
	268435456:         _TokType_name[274:287],
	536870912:         _TokType_name[287:300],
	1073741824:        _TokType_name[300:313],
	2147483648:        _TokType_name[313:324],
	4294967296:        _TokType_name[324:338],
	8589934592:        _TokType_name[338:352],
	17179869184:       _TokType_name[352:362],
	34359738368:       _TokType_name[362:372],
	68719476736:       _TokType_name[372:382],
	137438953472:      _TokType_name[382:392],
	274877906944:      _TokType_name[392:403],
	549755813888:      _TokType_name[403:414],
	1099511627776:     _TokType_name[414:426],
	2199023255552:     _TokType_name[426:438],
	4398046511104:     _TokType_name[438:448],
	8796093022208:     _TokType_name[448:459],
	17592186044416:    _TokType_name[459:470],
	35184372088832:    _TokType_name[470:483],
	70368744177664:    _TokType_name[483:500],
	140737488355328:   _TokType_name[500:510],
	281474976710656:   _TokType_name[510:522],
	562949953421312:   _TokType_name[522:534],
	1125899906842624:  _TokType_name[534:546],
	2251799813685248:  _TokType_name[546:556],
	4503599627370496:  _TokType_name[556:568],
	9007199254740992:  _TokType_name[568:580],
	18014398509481984: _TokType_name[580:593],
	36028797018963968: _TokType_name[593:606],
}

func (i TokType) String() string {
	if str, ok := _TokType_map[i]; ok {
		return str
	}
	return "TokType(" + strconv.FormatInt(int64(i), 10) + ")"
}
