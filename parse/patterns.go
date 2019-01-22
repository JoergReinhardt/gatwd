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
	f "github.com/JoergReinhardt/godeep/functions"
	l "github.com/JoergReinhardt/godeep/lex"
)

///////// MONO- / POLY-MORPHISM ///////////

type (
	Pattern   func() (id int, name string, args []d.BitFlag, ret d.BitFlag)
	Monoid    func() (pat Pattern, fnc f.Function)
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
