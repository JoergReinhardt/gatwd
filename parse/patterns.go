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
package parse

import (
	d "github.com/JoergReinhardt/godeep/data"
	l "github.com/JoergReinhardt/godeep/lang"
)

///////// MONO- / POLY-MORPHISM ///////////

type (
	pattern   func() (id int, args []d.BitFlag, ret d.BitFlag)
	monoid    func() (pat pattern, fnc d.Data)
	polymorph func() (id int, name string, mon []monoid)
)

// patterns are slices of tokens that can be compared with one another
func (s pattern) Flag() d.BitFlag   { return d.Flag.Flag() }
func (s pattern) Id() int           { id, _, _ := s(); return id }
func (s pattern) Args() []d.BitFlag { _, flags, _ := s(); return flags }
func (s pattern) RetVal() d.BitFlag { _, _, ret := s(); return ret }
func (s pattern) Tokens() []Token {
	var toks = []Token{}
	for _, flag := range s.Args() {
		toks = append(toks, newToken(Data_Type_Token, d.Type(flag.Flag())))
	}
	return append(tokJoin(newToken(Syntax_Token, l.RightArrow), toks))
}

// signature pattern of a literal function that takes a particular set of input
// parameters and returns a particular set of return values (this get's called)
func (i monoid) Id() int           { pat, _ := i(); return pat.Id() }
func (s monoid) Flag() d.BitFlag   { return d.Flag.Flag() }
func (i monoid) Args() []d.BitFlag { pat, _ := i(); return pat.Args() }
func (i monoid) Tokens() []Token   { pat, _ := i(); return pat.Tokens() }
func (i monoid) RetVal() d.BitFlag { pat, _ := i(); return pat.RetVal() }
func (i monoid) Fnc() d.Data       { _, fnc := i(); return fnc }

//func (i monoid) Call(d ...d.Data) d.Data { _, fnc := i(); return fnc.Call(d...) }

// slice of signatures and associated isomorphic implementations
// polymorph defined with a name
func (n polymorph) Id() int         { id, _, _ := n(); return id }
func (s polymorph) Flag() d.BitFlag { return d.Flag.Flag() }
func (n polymorph) Name() string    { _, name, _ := n(); return name }
func (n polymorph) Monom() []monoid { _, _, m := n(); return m }

// generation of new types starts with the generation of a pattern, which in
// turn retrieves, or generates an id depending on preexistence of the
// particular pattern. One way of passing the pattern in, is as a slice of
// mixed syntax, type-flag & string-data tokens.
func newPattern(id int, retVal d.BitFlag, args ...d.BitFlag) (p pattern) {
	return func() (int, []d.BitFlag, d.BitFlag) { return id, args, retVal }
}
func newMonoid(pat pattern, fnc d.Data) monoid {
	return func() (pattern, d.Data) { return pat, fnc }
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
