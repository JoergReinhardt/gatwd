package functions

import (
	d "github.com/JoergReinhardt/godeep/data"
	l "github.com/JoergReinhardt/godeep/lang"
)

//////// RUNTIME TYPE SPECIFICATIONS ////////
///// UID & USER DEFINED TYPE REGISTRATION ////
// TODO: make that portable, serializeable, parallelizeable, modular,
// selfcontained, distributely executed, and all the good things. by wrapping it all in a state monad

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

type typeState struct {
	uid     idGenerator
	tree    tokens
	names   map[string]polymorph
	polys   polymorphs
	datacon monomorphs
	typecon monomorphs
}

func (ts *typeState) NewUid() (id int) { id, (*ts).uid = ts.uid(); return id }

func newTypeState() *typeState {
	return &typeState{
		uid:     genCount(),
		tree:    tokens([]Token{}),
		names:   map[string]polymorph{},
		polys:   polymorphs{}, // []sig & []fnc
		datacon: monomorphs{}, // sig & fnc
		typecon: monomorphs{}, // sig & fnc
	}
}

////////
func tokS(f l.TokType) Token {
	return newToken(Syntax_Token, f.Flag())
}
func toksS(f ...l.TokType) []Token {
	var t = make([]Token, 0, len(f))
	for _, flag := range f {
		t = append(t, newToken(Syntax_Token, flag.Flag()))
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

func newArgs(f ...d.Type) []Token {
	return join(tokS(l.LeftArrow), toksD(f...))
}
func newRetVal(f Flag) Token {
	return newToken(Func_Type_Token, f)
}
func tokenizeTypeDef(name string, args []d.Type, retval Token) []Token {
	return append( // concat arguments, token & name
		append(
			newArgs(args...),
			newToken(Data_Value_Token, d.New(name)),
		),
		retval,
	)
}

/////
func padMissing(pat []Token) (head Token, tail []Token) {
	if len(pat) > 0 {
		if len(pat) > 1 {
			if len(pat) > 2 {
				head, tail = pat[0], pat[1:]
			}
			head, tail = pat[0], []Token{pat[1]}
		}
		head, tail = pat[0], nil
	}
	return head, tail
}

var delimLeft = l.LeftPar.Flag() | l.LeftBra.Flag() | l.LeftCur.Flag()
var delimRight = l.RightPar.Flag() | l.RightBra.Flag() | l.RightCur.Flag()

func getOrCreatePattern(ts *typeState, tok ...Token) pattern {
	var head Token
	var tail []Token
	var pat pattern
	if len(tok) > 0 {
		head, tail = padMissing(tok)
		// first token, expect function name, type name lambda, type
		// ident to derive name from, let, or data keyword:
		//
		//Type | name | Name | data | let | \x
		switch {
		// name definition
		case head.Type() == Data_Value_Token.Flag():
			switch {
			case d.FlagMatch(head.Flag(), d.String.Flag()):
				pat = parseIdentDef(ts, tok...)
			case d.FlagMatch(head.Flag(), d.Int.Flag()):
				pat = parseIdentDef(ts, tok...)
			}
		case head.Type() == Syntax_Token.Flag():
			switch {
			case d.FlagMatch(head.Flag(), l.Lambda.Flag()):
				pat = parseLambdaDef(ts, tail...)
			case d.FlagMatch(head.Flag(), l.DataWord.Flag()):
				pat = parseDataDecl(ts, tail...)
			case d.FlagMatch(head.Flag(), l.LetWord.Flag()):
				pat = parseLocalDataDecl(ts, tail...)
			case d.FlagMatch(delimLeft, head.Flag()):
				pat = parseDelimLeft(ts, tail...)
			case d.FlagMatch(l.RightArrow.Flag(), head.Flag()):
				pat = parseArrowRight(ts, tail...)
			case d.FlagMatch(l.Or.Flag(), head.Flag()):
				pat = parseOr(ts, tail...)
			case d.FlagMatch(l.Equal.Flag(), head.Flag()):
				pat = parseEquals(ts, tail...)
			case d.FlagMatch(l.DoubCol.Flag(), head.Flag()):
				pat = parseDoubleColon(ts, tail...)
			}
		case head.Type() == Data_Type_Token.Flag():
			pat = compDataType(ts, tok...)
		}
	}
	return pat
}
func parseDoubleColon(ts *typeState, tok ...Token) (pat pattern) {
	return pat
}
func parseEquals(ts *typeState, tok ...Token) (pat pattern) {
	return pat
}
func parseOr(ts *typeState, tok ...Token) (pat pattern) {
	return pat
}
func parseDelimLeft(ts *typeState, tok ...Token) (pat pattern) {
	return pat
}
func parseArrowRight(ts *typeState, tok ...Token) (pat pattern) {
	return pat
}
func compDataType(ts *typeState, tok ...Token) (pat pattern) {
	return pat
}
func parseLambdaDef(ts *typeState, tok ...Token) (pat pattern) {
	return pat
}
func parseIdentDef(ts *typeState, tok ...Token) (pat pattern) {
	return pat
}
func parseDataDecl(ts *typeState, tok ...Token) (pat pattern) {
	return pat
}
func parseLocalDataDecl(ts *typeState, tok ...Token) (pat pattern) {
	return pat
}
