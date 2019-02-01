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
		str = f.Kind(t.BitFlag).String() + "\n"
	case Syntax_Token:
		str = l.SyntaxItemFlag(t.BitFlag).Syntax() + "\n"
	case Data_Type_Token:
		str = d.Type(t.BitFlag).String() + "\n"
	default:
		str = "Don't know how to print this token"
	}
	return str
}
func (t dataTok) String() string {
	var str string
	switch t.TokType() {
	case Type_Token: // NAMED PARAMETERS
		// if type token is a pair, parameters are named
		if t.TokVal.Flag().Match(d.Parameter.Flag()) {
			parm := t.Data.(f.Parametric)
			// ARGUMENT NAME
			str = str + parm.Acc().String() +
				l.Colon.String() +
				l.Blank.String() +
				// ARGUMENT VALUE
				parm.Arg().String()
		}
		str = t.Data.String()
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
