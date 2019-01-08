package functions

import "sort"

///////// POLYMORPHISM ///////////
type patterns []pattern

func (s patterns) Len() int           { return len(s) }
func (s patterns) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s patterns) Less(i, j int) bool { return s[i].Id() < s[j].Id() }
func (s patterns) hasId(id int) bool  { return s.getById(id).Id() == id }
func (s patterns) getById(id int) pattern {
	var sig = s[sort.Search(len(s),
		func(i int) bool {
			return s[i].Id() >= id
		})]
	if sig.Id() == id {
		return sig
	}
	return sig
}
func sortPatterns(s patterns) patterns { sort.Sort(s); return s }

type isomorphs []isomorph

func (m isomorphs) Len() int           { return len(m) }
func (m isomorphs) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m isomorphs) Less(i, j int) bool { return m[i].Id() < m[j].Id() }
func (m isomorphs) hasId(id int) bool  { return m.getById(id).Id() == id }
func (m isomorphs) getById(id int) isomorph {
	var iso = m[sort.Search(len(m),
		func(i int) bool {
			return m[i].Id() >= id
		})]
	if iso.Id() == id {
		return iso
	}
	return iso
}
func sortIsomorphs(m isomorphs) isomorphs { sort.Sort(m); return m }

type polymorphs []polymorph

func (p polymorphs) Len() int           { return len(p) }
func (p polymorphs) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p polymorphs) Less(i, j int) bool { return p[i].Id() < p[j].Id() }
func (m polymorphs) hasId(id int) bool  { return m.getById(id).Id() == id }
func (m polymorphs) getById(id int) polymorph {
	var poly = m[sort.Search(len(m),
		func(i int) bool {
			return m[i].Id() >= id
		})]
	if poly.Id() == id {
		return poly
	}
	return poly
}
func sortPolymorphs(p polymorphs) polymorphs { sort.Sort(p); return p }

type (
	pattern   func() (id int, tok tokens)
	derivate  func() (id int, from int, tok tokens)
	isomorph  func() (id int, tok tokens, fnc Function)
	polymorph func() (id int, tok tokens, iso isomorphs)
	namedPoly func() (id int, name string, sig tokens, iso isomorphs)
)

// patterns are slices of tokens that can be compared with one another
func (s pattern) Id() int        { id, _ := s(); return id }
func (s pattern) Tokens() tokens { _, tok := s(); return tok }

// parametric types construct derived patterns for derived types
func (d derivate) Id() int          { id, _, _ := d(); return id }
func (d derivate) DerivedFrom() int { _, id, _ := d(); return id }
func (d derivate) Tokens() tokens   { _, _, tok := d(); return tok }

// signature pattern of a literal function that takes a particular set of input
// parameters and returns a particular set of return values (this get's called)
func (i isomorph) Id() int        { id, _, _ := i(); return id }
func (i isomorph) Tokens() tokens { _, tok, _ := i(); return tok }

// slice of signatures and associated isomorphic implementations
func (p polymorph) Id() int        { id, _, _ := p(); return id }
func (p polymorph) Tokens() tokens { _, tok, _ := p(); return tok }

// polymorph defined with a name
func (n namedPoly) Id() int        { id, _, _, _ := n(); return id }
func (n namedPoly) Name() string   { _, name, _, _ := n(); return name }
func (n namedPoly) Tokens() tokens { _, _, tok, _ := n(); return tok }

// isomorphic functions implement the function interface by forwarding passed
// parameters to the embedded functions eval method. TODO: handle arguments and returns
func (i isomorph) Call(d ...data) data { _, _, fn := i(); return fn.Call(d...) }

func conPattern(tok ...Token) pattern {
	i := conUID()
	s := tok
	return func() (id int, sig tokens) {
		return i, s
	}
}
func conDerivate(deri int, tok ...Token) derivate {
	d := deri
	i := conUID()
	s := tok
	return func() (id int, deri int, sig tokens) {
		return i, d, s
	}
}
func conIsomorph(sig pattern, fnc Function) isomorph {
	s := sig
	f := fnc
	return func() (
		id int,
		tok tokens,
		fn Function,
	) {
		id, tok = s()
		return id, tok, f
	}
}
func conPolymorph(sig pattern, iso ...isomorph) polymorph {
	s := sig
	return func() (
		id int,
		tok tokens,
		iso isomorphs,
	) {
		id, tok = s()
		return id, tok, iso
	}
}
func conNamedDef(name string, pol polymorph) namedPoly {
	p := pol
	return func() (
		id int,
		name string,
		tok tokens,
		iso isomorphs,
	) {
		id, tok, iso = p()
		return id, name, tok, iso
	}
}
