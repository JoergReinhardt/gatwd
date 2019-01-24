package parse

import (
	"strconv"

	d "github.com/JoergReinhardt/godeep/data"
	l "github.com/JoergReinhardt/godeep/lex"
)

//// TOKENS
func (t tokens) String() string {
	var str string
	for _, tok := range t {
		str = str + " " + tok.String()
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
	case Syntax_Token:
		str = t.Typed.(l.SyntaxItemFlag).Syntax()
	case Data_Type_Token:
		str = t.Typed.(d.Type).String()
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

///// PATTERNS MONOID
func (s Pattern) String() string {
	var str string
	for i, tok := range s.ArgToks() {
		str = str + tok.String()
		if i < len(s.ArgToks())-1 {
			str = str + " "
		}
	}
	return strconv.Itoa(s.Id()) + str
}

func (s Monoid) String() string {
	return strconv.Itoa(s.Id()) + " " + tokens(s.Tokens()).String()
}
func (s Polymorph) String() string {
	var str string
	for _, mon := range s.Monoids() {
		str = str + tokens(mon.Tokens()).String() + "\n"
	}
	return strconv.Itoa(s.Id()) + " " + str
}
