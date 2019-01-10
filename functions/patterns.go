package functions

import "sort"

///////// POLYMORPHISM ///////////

type (
	pattern   func() (id int, pat tokens)
	monomorph func() (pat pattern, fnc Functor)
	polymorph func() (id int, name string, mon monomorphs)
)

// patterns are slices of tokens that can be compared with one another
func (s pattern) Id() int     { id, _ := s(); return id }
func (s pattern) Pat() tokens { _, tok := s(); return tok }

// signature pattern of a literal function that takes a particular set of input
// parameters and returns a particular set of return values (this get's called)
func (i monomorph) Id() int             { pat, _ := i(); return pat.Id() }
func (i monomorph) Pat() tokens         { pat, _ := i(); return pat.Pat() }
func (i monomorph) Fnc() Functor        { _, fnc := i(); return fnc }
func (i monomorph) Call(d ...Data) Data { _, fnc := i(); return fnc.Call(d...) }

// slice of signatures and associated isomorphic implementations
// polymorph defined with a name
func (n polymorph) Id() int           { id, _, _ := n(); return id }
func (n polymorph) Pat() monomorphs   { _, _, m := n(); return m }
func (n polymorph) Name() string      { _, name, _ := n(); return name }
func (n polymorph) Monom() monomorphs { _, _, m := n(); return m }

// generation of new types starts with the generation of a pattern, which in
// turn retrieves, or generates an id depending on preexistence of the
// particular pattern. One way of passing the pattern in, is as a slice of
// mixed syntax, type-flag & string-data tokens.
func newPattern(ts *typeState, tok ...Token) pattern {
	return getOrCreatePattern(ts, tok...)
}
func newMonomorph(pat pattern, fnc Functor) monomorph {
	return func() (pattern, Functor) { return pat, fnc }
}
func newPolymorph(i int, name string, mono ...monomorph) polymorph {
	return func() (
		id int,
		name string,
		monom monomorphs,
	) {
		return i, name, mono
	}
}
func newNamedMorph(name string, poly polymorph) polymorph {
	return func() (
		i int,
		n string,
		monom monomorphs,
	) {
		i, name, mon := poly()
		return i, name, mon
	}
}

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

type monomorphs []monomorph

func (m monomorphs) Len() int           { return len(m) }
func (m monomorphs) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m monomorphs) Less(i, j int) bool { return m[i].Id() < m[j].Id() }
func (m monomorphs) hasId(id int) bool  { return m.getById(id).Id() == id }
func (m monomorphs) getById(id int) monomorph {
	var iso = m[sort.Search(len(m),
		func(i int) bool {
			return m[i].Id() >= id
		})]
	if iso.Id() == id {
		return iso
	}
	return iso
}
func sortIsomorphs(m monomorphs) monomorphs { sort.Sort(m); return m }

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
