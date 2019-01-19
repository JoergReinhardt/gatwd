/*
TYPE IDENTITY PATTERNS

  composes type id tokenized representation of the type constructors syntax &
  patterns to identify data instances that match the given parameter type.
  results in a unique type identity pattern that can efficiently be parsed and
  checked against, as well as, dynamicly be created and extended during
  runtime. in the following steps, patterns get composed with function base
  types and flag identifyers, to declare monoids which then in turn get
  constructed to declare polymophic, possibly parametric types during prelude
  and runtime.
*/
package functions

import (
	d "github.com/JoergReinhardt/godeep/data"
	l "github.com/JoergReinhardt/godeep/lang"
)

///////// MONO- / POLY-MORPHISM ///////////

type (
	pattern   func() (id int, args []Flag, ret Flag)
	monoid    func() (pat pattern, fnc Function)
	polymorph func() (id int, name string, mon []monoid)
)

// patterns are slices of tokens that can be compared with one another
func (s pattern) Flag() d.BitFlag { return HigherOrder.Flag() }
func (s pattern) Id() int         { id, _, _ := s(); return id }
func (s pattern) Args() []Flag    { _, flags, _ := s(); return flags }
func (s pattern) RetVal() Flag    { _, _, ret := s(); return ret }
func (s pattern) Tokens() []Token {
	var toks = []Token{}
	for _, flag := range s.Args() {
		toks = append(toks, newToken(Data_Type_Token, d.Type(flag.Flag())))
	}
	return append(join(newToken(Syntax_Token, l.RightArrow), toks))
}

// signature pattern of a literal function that takes a particular set of input
// parameters and returns a particular set of return values (this get's called)
func (i monoid) Id() int             { pat, _ := i(); return pat.Id() }
func (s monoid) Flag() d.BitFlag     { return HigherOrder.Flag() }
func (i monoid) Args() []Flag        { pat, _ := i(); return pat.Args() }
func (i monoid) Tokens() []Token     { pat, _ := i(); return pat.Tokens() }
func (i monoid) RetVal() Flag        { pat, _ := i(); return pat.RetVal() }
func (i monoid) Fnc() Function       { _, fnc := i(); return fnc }
func (i monoid) Call(d ...Data) Data { _, fnc := i(); return fnc.Call(d...) }

// slice of signatures and associated isomorphic implementations
// polymorph defined with a name
func (n polymorph) Id() int         { id, _, _ := n(); return id }
func (s polymorph) Flag() d.BitFlag { return HigherOrder.Flag() }
func (n polymorph) Name() string    { _, name, _ := n(); return name }
func (n polymorph) Monom() []monoid { _, _, m := n(); return m }

// generation of new types starts with the generation of a pattern, which in
// turn retrieves, or generates an id depending on preexistence of the
// particular pattern. One way of passing the pattern in, is as a slice of
// mixed syntax, type-flag & string-data tokens.
func newPattern(id int, retVal Flag, args ...Flag) (p pattern) {
	return func() (int, []Flag, Flag) { return id, args, retVal }
}
func newMonoid(pat pattern, fnc Function) monoid {
	return func() (pattern, Function) { return pat, fnc }
}
func newPolymorph(i int, name string, mono ...monoid) polymorph {
	return func() (
		id int,
		name string,
		monom []monoid,
	) {
		return i, name, mono
	}
}

// token mangling
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
