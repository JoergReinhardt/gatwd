package functions

import (
	d "github.com/JoergReinhardt/godeep/data"
	l "github.com/JoergReinhardt/godeep/lang"
)

func tokS(f l.TokType) Token {
	return newToken(Hacksell_Token, f.Flag())
}
func toksS(f ...l.TokType) []Token {
	var t = make([]Token, 0, len(f))
	for _, flag := range f {
		t = append(t, newToken(Hacksell_Token, flag.Flag()))
	}
	return t
}
func tokD(f d.Type) Token {
	return newToken(Data_Type_Token, f.Flag())
}
func toksD(f ...d.Type) []Token {
	var t = make([]Token, 0, len(f))
	for _, flag := range f {
		t = append(t, newToken(Data_Type_Token, flag.Flag()))
	}
	return t
}
func putAppend(last Token, tok []Token) []Token {
	return append(tok, last)
}
func putFront(first Token, tok []Token) []Token {
	return append([]Token{first}, tok...)
}
func join(sep Token, tok []Token) []Token {
	var args = tokens{}
	for i, t := range tok {
		args = append(args, t)
		if i < len(tok)-1 {
			args = append(args, sep)
		}
	}
	return args
}
func enclose(left, right Token, tok []Token) []Token {
	var args = tokens{left}
	for _, t := range tok {
		args = append(args, t)
	}
	args = append(args, right)
	return args
}
func embed(left, tok, right []Token) []Token {
	var args = left
	args = append(args, tok...)
	args = append(args, right...)
	return args
}

func newPrecArgs(f ...d.Type) []Token {
	return join(tokS(l.LeftArrow), toksD(f...))
}
func newRetVal(f Flag) Token {
	return newToken(Func_Type_Token, f)
}
