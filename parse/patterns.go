/*
TYPE IDENTITY PATTERNS

  patterns.go provides functions to deal with tokenized representation of
  godeep constructs, by implementing the token types and helper functions that
  get used internaly to split, join and shuffle sequences in assisting
  signature generation, parsing and the like.
*/
package parse

import (
	"bytes"

	d "github.com/JoergReinhardt/godeep/data"
	f "github.com/JoergReinhardt/godeep/functions"
	l "github.com/JoergReinhardt/godeep/lex"
)

type UID d.BitFlag

func (u UID) Flag() d.BitFlag { return d.BitFlag(u) }
func (u UID) Uint() uint      { return d.BitFlag(u).Uint() }
func (u UID) uint(uid uint)   { u = UID(uid) }

///////// MONO- / POLY-MORPHISM ///////////
type (
	Types func() d.SetString
	// Pattern   func() (id int, name string, args []d.BitFlag, ret d.BitFlag)
	//
	// provides a mapping of unique id pointing to monoid implementation
	// with to it's name, list of expected argument types and expected
	// return type
	//
	// parameter:
	// - name:  name of the polymorphic type, this is a part of
	// - args:  types, order and optinally keys of the argument set
	// - ret:   the return values type
	Pattern func() (name string, args []Token)

	// Isomorph    func() (pat Pattern, fnc f.Function)
	//
	// a monoid is the least common denominator of a function definition
	// within the godeeps internal type system. it maps a pattern of id,
	// name, list of expected argument-/ and return-types to a golang
	// function which signature it describes, to enable typesafe
	// application of generic function arguments during runtime.
	//
	// parameter:
	// - uid:     the unique id
	// - pattern: types of return value and argument set
	// - fnc:     instance providing a Call(...Data) method
	Isomorph func() (prop Property, pat Pattern, fnc f.Function)

	// Polymorph func() (pat Pattern, mon []Monoid)
	//
	// polymorphism provides different implementations for functions of the
	// same name, depending on the particular argument type applye during
	// runtime. the polymorph datatype maps the set of all monoids
	// implementing a function type, to it's pattern. During pattern
	// matching, that list will be matched against the instance encountered
	// and it will be applyed to the first function that matches its type
	Polymorph func() (name string, prop Property, toks []Token, mon []Isomorph)
)

// patterns are slices of tokens that can be compared with one another
func NewPattern(name string, args ...Token) (p Pattern) {
	return func() (string, []Token) { return name, args }
}
func (s Pattern) Name() string        { name, _ := s(); return name }
func (s Pattern) Args() []Token       { _, toks := s(); return toks }
func (s Pattern) TokenizeName() Token { return NewDataValueToken(d.StrVal(s.Name())) }
func (s Pattern) TokenizeSignature() []Token {
	var toks = []Token{
		NewDataValueToken(d.StrVal(s.Name())),
		NewSyntaxToken(l.Blank),
		NewSyntaxToken(l.DoubCol),
		NewSyntaxToken(l.Blank),
	}
	var ts = s.Args()
	var al = len(ts)
	for i, tok := range ts {
		toks = append(toks, NewTypeToken(tok))
		if i < al-1 {
			toks = append(
				append(
					toks,
					NewSyntaxToken(l.Blank),
				),
				NewSyntaxToken(l.RightArrow))
		}
	}
	return toks
}
func (p Polymorph) FullName() string {
	return l.Function.String() + l.Blank.String() + p.Signature()
}
func (s Pattern) Signature() string {
	var buf = bytes.NewBuffer([]byte{})
	for _, tok := range s.TokenizeSignature() {
		buf.WriteString(tok.String())
	}
	return buf.String()
}
func (s Pattern) String() string { return s.Signature() }

func NewMonoid(prop Property, pat Pattern, fnc f.Function) Isomorph {
	return func() (Property, Pattern, f.Function) { return prop, pat, fnc }
}
func (s Isomorph) String() string {
	return l.Function.String() +
		l.Blank.String() +
		s.Pattern().String()
}
func (m Isomorph) Flag() d.BitFlag     { return m.Fnc().Flag() }
func (m Isomorph) Kind() d.BitFlag     { return m.Fnc().Kind() }
func (m Isomorph) TokenizeFlag() Token { return NewDataTypeToken(m.Flag()) }
func (m Isomorph) TokenizeKind() Token { return NewKindToken(m.Flag()) }
func (m Isomorph) Propertys() Property { prop, _, _ := m(); return prop }
func (m Isomorph) Pattern() Pattern    { _, pat, _ := m(); return pat }
func (m Isomorph) Fnc() f.Function     { _, _, fnc := m(); return fnc }

//TODO: type checker action needs to be happening right here
func (m Isomorph) Call(d ...d.Data) d.Data { return m.Fnc().Call(d...) }

func NewPolymorph(
	uid int,
	prop Property,
	toks []Token,
	monom ...Isomorph) Polymorph {
	return func() (
		name string,
		prop Property,
		toks []Token,
		monom []Isomorph) {
		return name, prop, toks, monom
	}
}
func (p Polymorph) MonoidSigs() string {
	var str string
	if mons := p.Monoids(); len(mons) > 0 {

	}
	return str
}
func (p Polymorph) String() string      { return p.Signature() }
func (p Polymorph) Flag() d.BitFlag     { return d.Machinery.Flag() }
func (p Polymorph) Kind() d.BitFlag     { return f.Polymorph.Flag() }
func (p Polymorph) Name() string        { name, _, _, _ := p(); return name }
func (p Polymorph) Propertys() Property { _, prop, _, _ := p(); return prop }
func (p Polymorph) TypeCon() []Token    { _, _, toks, _ := p(); return toks }
func (p Polymorph) Monoids() []Isomorph { _, _, _, mons := p(); return mons }
func (p Polymorph) Append(mon Isomorph) Polymorph {
	name, prop, toks, mons := p()
	mons = append(mons, mon)
	return func() (string, Property, []Token, []Isomorph) {
		return name, prop, toks, mons
	}
}
func (p Polymorph) Signature() string {
	var str = bytes.NewBuffer([]byte{})
	var toks = p.TypeCon()
	var lt = len(toks)
	str.WriteString(p.Name())
	str.WriteString(l.Blank.String())
	str.WriteString(l.Equal.String())
	for i, tok := range toks {
		str.WriteString(l.Blank.String())
		str.WriteString(tok.String())
		if i < lt-1 {

			str.WriteString(l.Blank.String())
		}
	}
	return str.String()
}

// TYPE SYSTEM IMPLEMENTATION
//
// maps strings to polymorphic type definitions
func (t Types) names() d.SetString { return t() }
func (t Types) Lookup(name string) Polymorph {
	return t.names().Get(d.StrVal(name)).(Polymorph)
}
func (t Types) DefinePoly(name string, pol Polymorph) {
	t.names().Set(d.StrVal(name), pol)
}
func (t Types) AppendMonoid(name string, mon Isomorph) {
	pol := t.names().Get(d.StrVal(name)).(Polymorph)
	pol = pol.Append(mon)
	t.DefinePoly(name, pol)
}
func (t Types) Define(
	name string,
	prop Property,
	fnc f.Function,
	args ...Token,
) {
}
func InitTypes() TypeSystem {
	var names = d.SetString{}
	return Types(func() d.SetString {
		return names
	})
}
