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
func (t token) String() string {
	var str string
	switch t.typ {
	case Syntax_Token:
		str = t.flag.(l.SyntaxItemFlag).Syntax()
	case Data_Type_Token:
		str = d.StringBitFlag(t.flag.(d.Type).Flag())
	default:
		str = "Don't know how to print this token"
	}
	return str
}
func (t dataToken) String() string {
	var str string
	switch t.typ {
	case Data_Value_Token:
		str = t.d.(d.Data).String()
	case Argument_Token:
		str = "Arg: " + d.Type(t.Flag()).String()
	case Return_Token:
		str = "Ret: " + d.Type(t.Flag()).String()
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
