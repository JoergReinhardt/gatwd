package functions

import (
	d "github.com/JoergReinhardt/godeep/data"
	l "github.com/JoergReinhardt/godeep/lang"
)

//////// RUNTIME TYPE SPECIFICATIONS ////////
///// UID & USER DEFINED TYPE REGISTRATION ////
// TODO: make that portable, serializeable, parallelizeable, modular,
// selfcontained, distributely executed, and all the good things. by wrapping it all in a state monad
type typeState struct {
	uid     idGenerator
	tree    Token
	names   map[string]polymorph
	polys   []polymorph
	datacon []monoid
	typecon []monoid
}

func (ts *typeState) NewUid() (id int) { id, (*ts).uid = ts.uid(); return id }

func newTypeState() *typeState {
	return &typeState{
		uid:     genCount(),
		tree:    dataToken{},
		polys:   []polymorph{}, // []sig & []fnc
		typecon: []monoid{},    // sig & fnc
		datacon: []monoid{},    // sig & fnc
		names:   map[string]polymorph{},
	}
}

/////////////////////////////////////////////////////////////////////
type idGenerator func() (int, idGenerator)

func genCount() idGenerator {
	return func() (int, idGenerator) {
		var id int
		var gen idGenerator
		gen = func() (int, idGenerator) {
			id = id + 1
			return id, gen
		}
		return id, gen
	}
}
func tokS(f l.TokType) Token {
	return newToken(Syntax_Token, f)
}
func toksS(f ...l.TokType) []Token {
	var t = make([]Token, 0, len(f))
	for _, flag := range f {
		t = append(t, newToken(Syntax_Token, flag))
	}
	return t
}
func tokD(f d.Type) Token {
	return newToken(Data_Type_Token, f)
}
func toksD(f ...d.Type) []Token {
	var t = make([]Token, 0, len(f))
	for _, flag := range f {
		t = append(t, newToken(Data_Type_Token, flag))
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

// concatenate typeflags with right arrows as seperators, to generate a chain
// of curryed arguments
func newArgs(f ...d.Type) []Token {
	return join(tokS(l.LeftArrow), toksD(f...))
}
func newRetVal(f Flag) Token {
	return newToken(Return_Token, f)
}

// concatenates arguments, name of the type this signature is associated with
// and the type of the value, the associated function wil return. and returns
// the resulting signature as a chain of tokens (the name get's converted to a
// data-value token)
func tokenizeTypeDef(name string, args []d.Type, retval Token) []Token {
	return append( // concat arguments, token & name
		append(
			newArgs(args...),
			newToken(Data_Value_Token, d.New(name)),
		), retval)
}
