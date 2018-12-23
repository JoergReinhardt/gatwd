package types

func newSlice(val ...Evaluable) slice {
	l := make([]Evaluable, 0, len(val))
	l = append(l, val...)
	return l
}

type Function func(...Evaluable) Evaluable

// normalize narity
type Nullary Function
type Unary Function
type Binary Function
type Nary Function

func (Nullary) Arity() int { return 0 }
func (Unary) Arity() int   { return 1 }
func (Binary) Arity() int  { return 2 }
func (Nary) Arity() int    { return -1 }

func normalizeNullary(n func() Evaluable) Callable {
	return Nullary(func(...Evaluable) Evaluable {
		return n()
	})
}
func normalizeUnary(u func(e Evaluable) Evaluable) Callable {
	return Unary(func(e ...Evaluable) Evaluable {
		return u(e[0])
	})
}
func normalizeBinary(b func(a, b Evaluable) Evaluable) Callable {
	return Binary(func(e ...Evaluable) Evaluable {
		return b(e[0], e[1])
	})
}
func normalizeNary(n func(e ...Evaluable) Evaluable) Callable {
	return Nary(func(e ...Evaluable) Evaluable {
		return n(e...)
	})
}

//////////////////////////////////////////////////////////////////
type evalMethodSet struct {
	typ         func() flag
	str         func() string
	eva, cp, re func() Evaluable
}

func (d evalMethodSet) Type() flag      { return d.typ() }
func (d evalMethodSet) String() string  { return d.str() }
func (d evalMethodSet) Eval() Evaluable { return d.eva() }
func (d evalMethodSet) Copy() Evaluable { return d.cp() }
func (d evalMethodSet) ref() Evaluable  { return d.re() }

// return reference to method set
func (d evalMethodSet) methods() evalMethodSet { return d }

// return reference for every method
func (d evalMethodSet) detatchMethoSet() (
	typ func() flag,
	str func() string,
	eva, cp, re func() Evaluable,
) {
	return d.typ, d.str, d.eva, d.cp, d.re
}

// construct method set
func constructEvalMethodSet(
	typ func() flag,
	str func() string,
	eva, cp, re func() Evaluable,
) evalMethodSet {
	return evalMethodSet{typ, str, eva, cp, re}
}

// strip methods from evaluable
func stripEvalMethodSet(e Evaluable) evalMethodSet {
	return evalMethodSet{e.Type, e.String, e.Eval, e.Copy, e.ref}
}

//////////////////////////////////////////////////////////////////
type cellMethodSet struct {
	ms    func() evalMethodSet
	void  func() bool
	arity func() int
}

func (c cellMethodSet) methods() cellMethodSet { return c }
func (c cellMethodSet) Type() flag             { return c.ms().Type() }
func (c cellMethodSet) String() string         { return c.ms().String() }
func (c cellMethodSet) Eval() Evaluable        { return c.ms().Eval() }
func (c cellMethodSet) Copy() Evaluable        { return c.ms().Copy() }
func (c cellMethodSet) ref() Evaluable         { return c.ms().ref() }
func (c cellMethodSet) Empty() bool            { return c.void() }
func (c cellMethodSet) Arity() int             { return c.arity() }

func constructEmptyCell() cellMethodSet {
	ms := stripEvalMethodSet(nilVal{})
	return cellMethodSet{
		ms.methods,
		func() bool { return true },
		func() int { return 0 },
	}
}
func constructDataCell(e Evaluable) cellMethodSet {
	if fmatch(e.Type(), Nil.Type()) || e == nil {
		return constructEmptyCell()
	}

	var void func() bool
	var arity func() int

	if fmatch(e.Type(), Nullable.Type()) {
		void = func() bool { return e == nil }
		arity = func() int { return 1 }
	}
	if fmatch(e.Type(), Composed.Type()) {

		if vo, ok := e.(Cellular); ok {
			void = vo.Empty
		}

		if ari, ok := e.(Callable); ok {
			arity = ari.Arity
		}

		if siz, ok := e.(Size); ok {
			if siz.Len() > 0 {
				arity = func() int { return siz.Len() }
			}
		}

	}

	var c cellMethodSet

	typ := func() flag { return fconc(e.Type(), Cell.Type()) }
	str := func() string { return "[" + e.String() + "]" }
	eva := func() Evaluable { return constructDataCell(e.Eval()) }
	re := func() Evaluable { return constructDataCell(e.ref()) }
	cp := func() Evaluable { return e.Copy() }

	ms := constructEvalMethodSet(typ, str, eva, cp, re)

	c.ms = ms.methods
	c.void = void
	c.arity = arity

	return c
}
