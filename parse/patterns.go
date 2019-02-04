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
	Function    func() (inf Info, alt []Alternative)
	Alternative func() (inf Info, fnc f.Function)
)

// patterns are slices of tokens that can be compared with one another
func NewPattern(name string, toks ...Token) (p Info) {
	return func() (string, []Token) { return name, toks }
}
func (s Info) Name() string        { name, _ := s(); return name }
func (s Info) Tokens() []Token     { _, toks := s(); return toks }
func (s Info) TypePrim() d.BitFlag { return d.Object.TypePrim() }

// CLOSURE
func NewClosure(pat Info, fnc f.Function) Alternative {
	return func() (Info, f.Function) { return pat, fnc }
}
func (c Alternative) String() string       { return c.Function().String() }
func (c Alternative) TypePrim() d.BitFlag  { return c.Function().TypePrim() }
func (c Alternative) TypeHO() d.BitFlag    { return c.Function().TypeHO() }
func (c Alternative) Info() Info           { info, _ := c(); return info }
func (c Alternative) Function() f.Function { _, fnc := c(); return fnc }

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
func (p Function) String() string              { return p.Info().Name() }
func (p Function) TypePrim() d.BitFlag         { return d.Object.TypePrim() }
func (p Function) TypeHO() d.BitFlag           { return f.Polymorph.TypePrim() }
func (p Function) Info() Info                  { info, _ := p(); return info }
func (p Function) Alternatives() []Alternative { _, alts := p(); return alts }
func (p Function) Append(clo Alternative) Function {
	sig, clos := p()
	return func() (Info, []Alternative) {
		return sig, append(clos, clo)
	}
}
