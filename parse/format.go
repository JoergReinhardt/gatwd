package parse

import (
	d "github.com/JoergReinhardt/gatwd/data"
	f "github.com/JoergReinhardt/gatwd/functions"
	l "github.com/JoergReinhardt/gatwd/lex"
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
		str = t.Native.String()
	case Argument_Token:
		str = t.Native.String()
	case Parameter_Token:
		str = t.Native.String()
	case Pair_Token:
		str = t.Native.String()
	case Token_Collection:
		str = t.Native.String()
	}
	return str
}
