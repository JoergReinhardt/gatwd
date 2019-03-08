package parse

import (
	d "github.com/joergreinhardt/gatwd/data"
	f "github.com/joergreinhardt/gatwd/functions"
	l "github.com/joergreinhardt/gatwd/lex"
)

//// TOKENS
func (t tokens) String() string {
	var str string
	for _, tok := range t {
		str = str + " " + tok.String() + "\n"
	}
	return str
}
func (t tokenSlice) String() string {
	var str string
	for _, s := range t {
		str = str + tokens(s).String() + "\n"
	}
	return str
}
func (t TokVal) String() string {
	var str string
	switch t.TypeTok() {
	case TypeFnc_Token:
		str = t.Native.(f.TyFnc).String() + "\n"
	case Syntax_Token:
		str = t.Native.(l.SyntaxItemFlag).Syntax() + "\n"
	case TypeNat_Token:
		str = t.Native.(d.TyNative).String() + "\n"
	default:
		str = "Don't know how to print this token"
	}
	return str
}
func (t dataTok) String() string {
	var str string
	switch t.TypeTok() {
	case Data_Value_Token:
		str = string(t.Native.(d.StrVal))
	case Pair_Token:
		str = string(t.Native.(d.StrVal))
	case Token_Collection:
		str = string(t.Native.(d.StrVal))
	}
	return str
}
