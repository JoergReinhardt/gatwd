/*
TYPE IDENTITY PATTERNS

  patterns.go provides functions to deal with tokenized representation of
  godeep constructs, by implementing the token types and helper functions that
  get used internaly to split, join and shuffle sequences in assisting
  signature generation, parsing and the like.
*/
package parse

import (
	d "github.com/JoergReinhardt/godeep/data"
	f "github.com/JoergReinhardt/godeep/functions"
	l "github.com/JoergReinhardt/godeep/lex"
)

///////// MONO- / POLY-MORPHISM ///////////

type (
	// Pattern   func() (id int, name string, args []d.BitFlag, ret d.BitFlag)
	//
	// provides a mapping of unique id pointing to monoid implementation
	// with to it's name, list of expected argument types and expected
	// return type
	Pattern func() (id int, name string, args []d.BitFlag, ret d.BitFlag)

	// Monoid    func() (pat Pattern, fnc f.Function)
	//
	// a monoid is the least common denominator of a function definition
	// within the godeeps internal type system. it maps a pattern of id,
	// name, list of expected argument-/ and return-types to a golang
	// function which signature it describes, to enable typesafe
	// application of generic function arguments during runtime.
	Monoid func() (pat Pattern, fnc f.Function)

	// Polymorph func() (pat Pattern, mon []Monoid)
	//
	// polymorphism provides different implementations for functions of the
	// same name, depending on the particular argument type applye during
	// runtime. the polymorph datatype maps the set of all monoids
	// implementing a function type, to it's pattern. During pattern
	// matching, that list will be matched against the instance encountered
	// and it will be applyed to the first function that matches its type
	Polymorph func() (pat Pattern, mon []Monoid)
)

// patterns are slices of tokens that can be compared with one another
func NewPattern(
	id int,
	name string,
	retVal d.BitFlag,
	args ...d.BitFlag,
) (p Pattern) {
	return func() (
		int,
		string,
		[]d.BitFlag,
		d.BitFlag) {
		return id, name, args, retVal
	}
}
func (s Pattern) Flag() d.BitFlag   { return d.Flag.Flag() }
func (s Pattern) Id() int           { id, _, _, _ := s(); return id }
func (s Pattern) Name() string      { _, name, _, _ := s(); return name }
func (s Pattern) Args() []d.BitFlag { _, _, flags, _ := s(); return flags }
func (s Pattern) RetVal() d.BitFlag { _, _, _, retval := s(); return retval }
func (s Pattern) Signature() string {
	var sig string
	for _, tok := range s.SigToks() {
		sig = sig + " " + tok.String()
	}
	return sig
}
func (s Pattern) SigToks() []Token {
	return append(
		append(s.ArgToks(),
			newToken(Data_Value_Token,
				d.StrVal(s.Name()))),
		newToken(
			Data_Type_Token,
			s.RetVal()))
}
func (s Pattern) ArgToks() []Token {
	var toks = []Token{}
	for _, flag := range s.Args() {
		toks = append(toks, newToken(Data_Type_Token, d.Type(flag.Flag())))
	}
	return append(tokJoin(newToken(Syntax_Token, l.RightArrow), toks))
}

func NewMonoid(pat Pattern, fnc f.Function) Monoid {
	return func() (Pattern, f.Function) { return pat, fnc }
}
func (i Monoid) Id() int                 { pat, _ := i(); return pat.Id() }
func (s Monoid) Flag() d.BitFlag         { return d.Flag.Flag() }
func (i Monoid) Args() []d.BitFlag       { pat, _ := i(); return pat.Args() }
func (i Monoid) Tokens() []Token         { pat, _ := i(); return pat.ArgToks() }
func (i Monoid) RetVal() d.BitFlag       { pat, _ := i(); return pat.RetVal() }
func (i Monoid) Fnc() d.Data             { _, fnc := i(); return fnc }
func (i Monoid) Call(d ...d.Data) d.Data { _, fnc := i(); return fnc.Call(d...) }

func NewPolymorph(pats Pattern, mono ...Monoid) Polymorph {
	return func() (
		pat Pattern,
		monom []Monoid,
	) {
		return pat, mono
	}
}
func (s Polymorph) Flag() d.BitFlag   { return d.Flag.Flag() }
func (n Polymorph) Pat() Pattern      { pat, _ := n(); return pat }
func (n Polymorph) Id() int           { return n.Id() }
func (n Polymorph) Name() string      { return n.Name() }
func (n Polymorph) Monoids() []Monoid { _, m := n(); return m }
