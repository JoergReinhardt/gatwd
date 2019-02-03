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
	Signature func() (name string, toks []Token)
	Closure   func() (sig Signature, prop Property, fnc f.Function)
	Polymorph func() (sig Signature, clos []Closure)
)

// patterns are slices of tokens that can be compared with one another
func NewPattern(name string, toks ...Token) (p Signature) {
	return func() (string, []Token) { return name, toks }
}
func (s Signature) Name() string        { name, _ := s(); return name }
func (s Signature) Tokens() []Token     { _, toks := s(); return toks }
func (s Signature) Flag() d.BitFlag     { return d.Object.Flag() }
func (s Signature) TokenizeName() Token { return NewDataValueToken(d.StrVal(s.Name())) }

// CLOSURE
func NewClosure(pat Signature, prop Property, fnc f.Function) Closure {
	return func() (Signature, Property, f.Function) { return pat, prop, fnc }
}
func (c Closure) String() string       { return c.Function().String() }
func (c Closure) Flag() d.BitFlag      { return c.Function().Flag() }
func (c Closure) Kind() d.BitFlag      { return c.Function().Kind() }
func (c Closure) Signature() Signature { sig, _, _ := c(); return sig }
func (c Closure) Property() Property   { _, prop, _ := c(); return prop }
func (c Closure) Function() f.Function { _, _, fnc := c(); return fnc }

//TODO: type checker action needs to be happening right here
func (m Closure) Call(d ...f.Functional) f.Functional { return m.Function().Call(d...) }

func NewPolymorph(
	sig Signature,
	monom ...Closure,
) Polymorph {
	return func() (
		signature Signature,
		monom []Closure,
	) {
		return signature, monom
	}
}
func (p Polymorph) String() string       { return p.Signature().Name() }
func (p Polymorph) Flag() d.BitFlag      { return d.Object.Flag() }
func (p Polymorph) Kind() d.BitFlag      { return f.Polymorph.Flag() }
func (p Polymorph) Signature() Signature { sig, _ := p(); return sig }
func (p Polymorph) Monoids() []Closure   { _, mons := p(); return mons }
func (p Polymorph) Append(clo Closure) Polymorph {
	sig, clos := p()
	return func() (Signature, []Closure) {
		return sig, append(clos, clo)
	}
}
