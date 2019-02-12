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
	case TypeHO_Token:
		str = t.Primary.(f.TyHigherOrder).String() + "\n"
	case Syntax_Token:
		str = t.Primary.(l.SyntaxItemFlag).Syntax() + "\n"
	case TypePrim_Token:
		str = t.Primary.(d.TyPrimitive).String() + "\n"
	default:
		str = "Don't know how to print this token"
	}
	return str
}
func (t dataTok) String() string {
	var str string
	switch t.TypeTok() {
	case Data_Value_Token:
		str = t.Primary.String()
	case Argument_Token:
		str = t.Primary.String()
	case Parameter_Token:
		str = t.Primary.String()
	case Pair_Value_Token:
		str = t.Primary.String()
	case Token_Collection:
		str = t.Primary.String()
	}
	return str
}
