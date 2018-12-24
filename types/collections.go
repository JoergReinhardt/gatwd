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

// return reference for every method
func (d evalMethodSet) detatchMethodSet() (
	typ func() flag,
	str func() string,
	eva, cp, re func() Evaluable,
) {
	return d.typ, d.str, d.eva, d.cp, d.re
}

//////////////////////////////////////////////////////////////////
type cellMethodSet struct {
	ms    evalMethodSet
	void  func() bool
	arity func() int
}

func (c cellMethodSet) methods() cellMethodSet { return c }
func (c cellMethodSet) Type() flag             { return c.ms.Type() }
func (c cellMethodSet) String() string         { return c.ms.String() }
func (c cellMethodSet) Eval() Evaluable        { return c.ms.Eval() }
func (c cellMethodSet) Copy() Evaluable        { return c.ms.Copy() }
func (c cellMethodSet) ref() Evaluable         { return c.ms.ref() }
func (c cellMethodSet) Empty() bool            { return c.void() }
func (c cellMethodSet) Arity() int             { return c.arity() }

func (c cellMethodSet) detatchMethodSet() (
	ms func() evalMethodSet,
	void func() bool,
	arity func() int,
) {
	return c.ms.methods, c.void, c.arity
}
func constructEmptyCell() cellMethodSet {
	ms := stripEvalMethodSet(nilVal{})
	return cellMethodSet{
		ms.methods(),
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

	c.ms = ms.methods()
	c.void = void
	c.arity = arity

	return c
}

type tupleMethodSet struct {
	len   func() int
	head  func() Cellular
	tail  func() Tupular
	decap func() (Cellular, Tupular)
	cel   cellMethodSet
}

func (t tupleMethodSet) Len() int                   { return t.len() }
func (t tupleMethodSet) Empty() bool                { return t.cel.void() }
func (t tupleMethodSet) Arity() int                 { return t.cel.arity() }
func (t tupleMethodSet) Head() Cellular             { return t.head() }
func (t tupleMethodSet) Tail() Tupular              { return t.tail() }
func (t tupleMethodSet) Decap() (Cellular, Tupular) { return t.decap() }
func (t tupleMethodSet) Type() flag                 { return t.cel.ms.typ() }
func (t tupleMethodSet) String() string             { return t.cel.ms.str() }
func (t tupleMethodSet) Eval() Evaluable            { return t.cel.ms.eva() }
func (t tupleMethodSet) Copy() Evaluable            { return t.cel.ms.cp() }
func (t tupleMethodSet) ref() Evaluable             { return t.cel.ms.re() }

func constructEmptyTuple() tupleMethodSet {

	emc := func() Evaluable { return constructEmptyCell() }

	typ := func() flag { return Tuple.Type() }
	str := func() string { return "()" }
	decap := func() (Cellular, Tupular) {
		return constructEmptyCell(),
			constructEmptyTuple()
	}
	eva := emc
	cp := emc
	re := emc
	ms := constructEvalMethodSet(typ, str, eva, cp, re)

	void := func() bool { return true }
	arity := func() int { return 0 }
	cs := cellMethodSet{ms.methods(), void, arity}
	h := func() Cellular { return constructEmptyCell() }
	t := func() Tupular { return tupleMethodSet{} }
	return tupleMethodSet{func() int { return 0 }, h, t, decap, cs}
}
func constructTuple(e ...Evaluable) tupleMethodSet {
	var tup tupleMethodSet
	var plen = len(e)
	if plen == 0 {
		return constructEmptyTuple()
	}
	var cells = make([]Cellular, 0, plen)
	for _, eva := range e {
		c := constructDataCell(eva)
		cells = append(cells, c)
	}

	clen := func() int { return len(cells) }
	arity := func() int { return plen }
	head := func() Cellular {
		if len(cells) > 0 {
			if c := cells[0]; !c.Empty() {
				return c
			}
		}
		return constructEmptyCell()
	}
	tail := func() Tupular {
		if len(cells) > 1 {
			if c := cells[1]; !c.Empty() {
				return c.(Tupular)
			}
		}
		return constructEmptyTuple()
	}
	decap := func() (Cellular, Tupular) {
		var h Cellular
		var t Tupular
		switch clen() {
		case 0:
			h = constructEmptyCell()
			t = constructEmptyTuple()
		case 1:
			h = tup.head()
			t = constructEmptyTuple()
			tup.head = func() Cellular { return constructEmptyCell() }
		case 2:
			h = tup.head()
			t = tup.tail()
			tup.head, tup.tail = func() Cellular { return t },
				func() Tupular { return constructEmptyTuple() }
		default:
			h = tup.head()
			t = tup.tail()
			lh, lt := tup.tail().Decap()
			tup.head, tup.tail = func() Cellular { return lh },
				func() Tupular { return lt }
		}
		return h, t
	}
	void := func() bool {
		if !tup.head().Empty() && !tup.tail().Empty() {
			return false
		}
		return true
	}
	typ := func() flag {
		var f flag = Tuple.Type()
		if len(cells) > 0 {
			f = fconc(f.Type(), tup.head().Type())
			if len(cells) > 1 {
				f = fconc(f.Type(), tup.tail().Type())
			}
		}
		return f
	}
	str := func() string {
		var str = "("
		if clen() > 0 {
			for i, cell := range cells {
				str = str + cell.String()
				if i < clen()-1 {
					str = str + ", "
				}
			}
		}
		return str + ")"
	}
	eva := func() Evaluable {
		return tup.Head()
	}
	cp := func() Evaluable { return tup.Copy() }
	re := func() Evaluable { return &tup }

	tup.len = clen
	tup.head = head
	tup.tail = tail
	tup.decap = decap
	tup.cel.void = void
	tup.cel.arity = arity
	tup.cel.ms.typ = typ
	tup.cel.ms.str = str
	tup.cel.ms.eva = eva
	tup.cel.ms.cp = cp
	tup.cel.ms.re = re

	return tup
}
