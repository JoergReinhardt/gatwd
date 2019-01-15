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
	"strconv"

	d "github.com/JoergReinhardt/godeep/data"
	l "github.com/JoergReinhardt/godeep/lang"
)

///////// MONO- / POLY-MORPHISM ///////////

type (
	pattern   func() (id int, args []Token, ret Token)
	monoid    func() (pat pattern, fnc Functor)
	polymorph func() (id int, name string, mon []monoid)
)

// patterns are slices of tokens that can be compared with one another
func (s pattern) String() string {
	return strconv.Itoa(s.Id()) + " " + tokens(s.Args()).String()
}
func (s pattern) Flag() d.BitFlag { return Internal.Flag() }
func (s pattern) Id() int         { id, _, _ := s(); return id }
func (s pattern) RetVal() Token   { _, _, ret := s(); return ret }
func (s pattern) Args() []Token {
	_, args, _ := s()
	return append(join(newToken(Syntax_Token, l.RightArrow), args))
}

// signature pattern of a literal function that takes a particular set of input
// parameters and returns a particular set of return values (this get's called)
func (i monoid) Id() int             { pat, _ := i(); return pat.Id() }
func (s monoid) Flag() d.BitFlag     { return Internal.Flag() }
func (i monoid) Args() []Token       { pat, _ := i(); return pat.Args() }
func (i monoid) RetVal() Token       { pat, _ := i(); return pat.RetVal() }
func (i monoid) Fnc() Functor        { _, fnc := i(); return fnc }
func (i monoid) Call(d ...Data) Data { _, fnc := i(); return fnc.Call(d...) }
func (s monoid) String() string {
	return strconv.Itoa(s.Id()) + " " + tokens(s.Args()).String()
}

// slice of signatures and associated isomorphic implementations
// polymorph defined with a name
func (n polymorph) Id() int         { id, _, _ := n(); return id }
func (s polymorph) Flag() d.BitFlag { return Internal.Flag() }
func (n polymorph) Name() string    { _, name, _ := n(); return name }
func (n polymorph) Monom() []monoid { _, _, m := n(); return m }
func (s polymorph) String() string {
	var str string
	for _, mon := range s.Monom() {
		str = str + tokens(mon.Args()).String() + "\n"
	}
	return strconv.Itoa(s.Id()) + " " + str
}

// generation of new types starts with the generation of a pattern, which in
// turn retrieves, or generates an id depending on preexistence of the
// particular pattern. One way of passing the pattern in, is as a slice of
// mixed syntax, type-flag & string-data tokens.
func newPattern(id int, retVal Token, args ...Token) (p pattern) {
	return p
}
func newMonoid(pat pattern, fnc Functor) monoid {
	return func() (pattern, Functor) { return pat, fnc }
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
