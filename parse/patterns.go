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
)

///////// MONO- / POLY-MORPHISM ///////////
type (
	Info        func() (name string, toks []Token)
	Alternative func() (sig Info, prop Property, fnc f.Function)
	Function    func() (sig Info, clos []Alternative)
)

// patterns are slices of tokens that can be compared with one another
func NewPattern(name string, toks ...Token) (p Info) {
	return func() (string, []Token) { return name, toks }
}
func (s Info) Name() string        { name, _ := s(); return name }
func (s Info) Tokens() []Token     { _, toks := s(); return toks }
func (s Info) Flag() d.BitFlag     { return d.Object.TypePrim() }
func (s Info) TokenizeName() Token { return NewDataValueToken(d.StrVal(s.Name())) }

// CLOSURE
func NewClosure(pat Info, prop Property, fnc f.Function) Alternative {
	return func() (Info, Property, f.Function) { return pat, prop, fnc }
}
func (c Alternative) String() string       { return c.Function().String() }
func (c Alternative) Flag() d.BitFlag      { return c.Function().TypePrim() }
func (c Alternative) Kind() d.BitFlag      { return c.Function().TypeHO() }
func (c Alternative) Signature() Info      { sig, _, _ := c(); return sig }
func (c Alternative) Property() Property   { _, prop, _ := c(); return prop }
func (c Alternative) Function() f.Function { _, _, fnc := c(); return fnc }

//TODO: type checker action needs to be happening right here
func (m Alternative) Call(d ...f.Functional) f.Functional { return m.Function().Call(d...) }

func NewPolymorph(
	sig Info,
	monom ...Alternative) Function {
	return func() (
		signature Info,
		monom []Alternative) {
		return signature, monom
	}
}
func (p Function) String() string         { return p.Signature().Name() }
func (p Function) Flag() d.BitFlag        { return d.Object.TypePrim() }
func (p Function) Kind() d.BitFlag        { return f.Polymorph.TypePrim() }
func (p Function) Signature() Info        { sig, _ := p(); return sig }
func (p Function) Monoids() []Alternative { _, mons := p(); return mons }
func (p Function) Append(clo Alternative) Function {
	sig, clos := p()
	return func() (Info, []Alternative) {
		return sig, append(clos, clo)
	}
}
