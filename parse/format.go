package parse

import (
	d "github.com/JoergReinhardt/godeep/data"
	f "github.com/JoergReinhardt/godeep/functions"
	l "github.com/JoergReinhardt/godeep/lex"
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
	switch t.TokType() {
	case Kind_Token:
		str = t.Data.(f.Kind).String() + "\n"
	case Syntax_Token:
		str = t.Data.(l.SyntaxItemFlag).Syntax() + "\n"
	case Data_Type_Token:
		str = t.Data.(d.Type).String() + "\n"
	default:
		str = "Don't know how to print this token"
	}
	return str
}
func (t dataTok) String() string {
	var str string
	switch t.TokType() {
	case Data_Value_Token:
		str = t.Data.String()
	case Argument_Token:
		str = t.Data.String()
	case Parameter_Token:
		str = t.Data.String()
	case Pair_Value_Token:
		str = t.Data.String()
	case Token_Collection:
		str = t.Data.String()
	}
	return str
}
