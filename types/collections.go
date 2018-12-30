package types

//go:generate stringer -type FixType
type FixType BitFlag

//go:generate stringer -type Arity
type Arity int8 // <-- this sounds so terribly wrong, for anyone born in germany

const (
	Nullary Arity = 0 + iota
	Unary
	Binary
	Trinary
	Ternary
	Quaternary
	Quinary
	Senary
	Septenary
	Octonary
	Novenary
	Denary
	Polyadic

	PreFix FixType = 0 + iota
	InFix
	PostFix
	ConFix
)

func guardArity(a Arity, d ...Data) []Data {
	var dat = []Data{}
	switch {
	case len(d) < int(a):
		dat = append(dat, nilVal{})
	case len(d) > int(a):
		dat = d[:a-1]
	case len(d) == int(a):
		dat = d
	}
	return dat
}

////// GENERIC COLLECTION CONSTRUCTOR & INSTANCE TYPES ////////
// collection propertys:
//
// - interconnecting structure is either implicitly expressed on the callstack,
// by mutual recursion allone, or materialized as pointer, index, oder map
// referencs.
//
// - all collection types can be constructed either as to be lazy, or eagerly.
//
// defaults are lazy evaluation and preferred lambdaness for interconnections
// between elements.
type ColProps BitFlag

func (c ColProps) FLag() BitFlag { return BitFlag(c) }

//go:generate stringer -type ColProps
const (
	Eager   ColProps = 1 << (63 - iota) //  eager | lazy   -> 0 = lazy   | 1 = eager
	Lambda                              //  lamba | linked -> 0 = lambda | 1 = data
	Reverse                             // left-/ | right bound -> 0 = left | 1 = right
)

type collectionDataConstructor func(d ...Data) Consumeable
type collectionInstance func(d ...Data) (head Data, tail Consumeable)

////// PROTOTYPE &  METHOD SET RECURSIVE COLLECTION (AKA NESTED) ///////
type recursiveCollectionInstance func(d ...Data) (Data, Consumeable)

func (r recursiveCollectionInstance) Eval() Data        { return r }
func (r recursiveCollectionInstance) Flag() BitFlag     { return Recursives.Flag() }
func (r recursiveCollectionInstance) Head() Data        { head, _ := r(); return head }
func (r recursiveCollectionInstance) Tail() Consumeable { _, tail := r(); return tail }
func (r recursiveCollectionInstance) String() string    { return recursiveConsumeableString(r) }
func (r recursiveCollectionInstance) Shift() Consumeable {
	tail := r.Tail()
	return conColRec(tail)
}
func (r recursiveCollectionInstance) Empty() bool {
	if !elemEmpty(r.Head()) {
		return false
	}
	if r.Tail() != nil {
		if !r.Tail().Empty() {
			return false
		}
	}
	return true
}
func (r recursiveCollectionInstance) Len() int {
	var l = 0
	if !elemEmpty(r.Head()) {
		l = l + 1
		if fmatch(r.Head().Flag(), Composed.Flag()) {
			l = l + r.Head().(Countable).Len() - 1
		}
	}
	if !r.Tail().Empty() {
		l = l + r.Tail().(Countable).Len()
	}
	return l
}

// compose new recursive collection from n-nodes
func composeColRec(c ...Data) Consumeable {
	return recursiveCollectionInstance(func(...Data) (Data, Consumeable) {
		h, t := deVoidParamSLice(c...)
		return h, composeColRec(t...)
	})
}
func composeColRecEager(c ...Data) Consumeable {
	var ta Consumeable
	h, t := deVoidParamSLice(c...)
	if h != nil {
		if !chainEmpty(t) {
			ta = composeColRecEager(t...)
		}
	}
	return recursiveCollectionInstance(func(...Data) (Data, Consumeable) {
		head, tail := h, ta
		return head, tail
	})
}

// construct successor from shifted remnant
func conColRec(c Consumeable) Consumeable {
	return recursiveCollectionInstance(func(d ...Data) (Data, Consumeable) {
		head, tail := c.Head(), c.Tail()
		return head, tail
	})
}

////// PROTOTYPE &  METHOD SET REFERENCED COLLECTION (AKA FLAT/LINKED) ///////
type referencedCollectionInstance func(d ...Data) (Data, Splitable)

func (r referencedCollectionInstance) Head() Data      { head, _ := r(); return head }
func (r referencedCollectionInstance) Tail() Splitable { _, tail := r(); return tail }
func (r referencedCollectionInstance) Shift() Consumeable {
	tail := r.Tail().Slice()
	return conColRef(tail...)
}
func (r referencedCollectionInstance) Empty() bool {
	if !elemEmpty(r.Head()) {
		return false
	}
	if r.Tail() != nil {
		if !chainEmpty(r.Tail().(chain)) {
			return false
		}
	}
	return true
}
func (r referencedCollectionInstance) Len() int {
	var l = 0
	if !elemEmpty(r.Head()) {
		l = l + 1
		if fmatch(r.Head().Flag(), Composed.Flag()) {
			l = l + r.Head().(Countable).Len() - 1
		}
	}
	if !r.Tail().Empty() {
		l = l + r.Tail().Len()
	}
	return l
}
func conColRef(d ...Data) Consumeable {
	return nil
}

////////////// COLLECTION CONSTRUCTOR HELPERS /////////////
func deVoidParamSLice(d ...Data) (Data, chain) {
	var head Data
	var c chain
	switch len(d) {
	case 0:
		head, c = nil, conChain()
	case 1:
		head, c = d[0], conChain()
	case 2:
		head, c = d[0], conChain(d[1])
	default:
		head, c = d[0], conChain(d[1:]...)
	}
	return head, c
}

/////////////////////////////////////////////////////////////////////
/////// RECURSIVE COLLECTION ///////
type reCol func() (Data, reCol)

func conRecursiveLazy(d ...Data) reCol {
	return func() (Data, reCol) {
		var head Data
		var tail reCol
		var data = d
		switch len(data) {
		case 0:
			head = nilVal{}
			tail = conRecursiveLazy(nilVal{})
		case 1:
			head = data[0]
			tail = nil
		default:
			head = data[0]
			tail = conRecursiveLazy(data[1:]...)
		}
		return head, tail
	}
}

// construt new recursive from n data elements
func conRecursiveEager(d ...Data) reCol {
	var head Data
	var tail reCol
	switch len(d) {
	case 0:
		head = nilVal{}
		tail = conRecursiveEager(nilVal{})
	case 1:
		head = d[0]
		tail = nil
	default:
		head = d[0]
		tail = conRecursiveEager(d[1:]...)
	}
	return func() (Data, reCol) {
		h := head
		t := tail
		return h, t
	}
}

//// standalone functions to implement recursive behaviour on recursively nested closures
func recolHead(r reCol) Data  { h, _ := r(); return h }
func recolTail(r reCol) reCol { _, t := r(); return t }
func recolEmpty(rec reCol) bool {
	var h Data
	var r reCol
	if rec != nil {
		h, r = rec()
		if !elemEmpty(h) {
			return false
		}
		if r != nil {
			return recolEmpty(r)
		}
	}
	return true
}

// recursive behaviour as method to implement interfaces
func (r reCol) Eval() Data    { return r }
func (r reCol) Flag() BitFlag { return (List.Flag() | r.Head().Flag()) }
func (r reCol) Head() Data    { return recolHead(r) } // --> current head
func (r reCol) Tail() reCol   { return recolTail(r) } // --> current tail
func (r reCol) Empty() bool   { return recolEmpty(r) }

//////////// FLAT COLLECTION ///////////////////////////

type flatCol func() []Data

// construt new recursive from n data elements
func conFlatColLazy(d ...Data) flatCol {
	return func() []Data {
		var dat = d
		switch len(d) {
		case 0:
			dat[0] = nil
			dat[1] = nil
		case 1:
			dat[0] = d[0].(Data)
			dat[1] = nil
		case 2:
			dat[0] = d[1].(Data)
			dat[1] = nil
		default:
			dat = append([]Data{d[1].(Data)}, d[2:]...)

		}
		return dat
	}

}

// construt new recursive from n data elements
func conFlatColEager(d ...Data) flatCol {
	var dat []Data
	switch len(d) {
	case 0:
		dat[0] = nilVal{}
		dat[1] = nilVal{}
	case 1:
		dat[0] = d[0]
		dat[1] = nilVal{}
	default:
		dat = d

	}
	return func() []Data {
		data := dat
		return data
	}

}
func (f flatCol) Flag() BitFlag { return List.Flag() | f.Head().Flag() }
func (f flatCol) Eval() Data    { return f }
func (f flatCol) Slice() []Data { return f() }
func (f flatCol) Head() Data    { return flatColHead(f) }
func (f flatCol) Tail() flatCol { return conFlatColTailEager(f) }
func (f flatCol) Len() int      { return chainLen(f()) }
func (f flatCol) Empty() bool   { return chainEmpty(f()) }

func flatColHead(f flatCol) Data            { return f()[0] }
func conFlatColTailEager(f flatCol) flatCol { tail := conFlatColEager(f()[1:]...); return tail }
func conFlatColTailLazy(f flatCol) flatCol  { tail := conFlatColLazy(f()[1:]...); return tail }
func flatColEmpty(f flatCol) bool           { return chainEmpty(f()) }
func flatColLen(f flatCol) int              { return chainLen(f()) }

/////////// FLAT TUPLE /////////////////
type tuple flatCol

func conTuple(d ...Data) tuple {
	return tuple(conFlatColEager(d...))
}

func (t tuple) Flag() BitFlag  { return Tuple.Flag() | t.Head().Flag() }
func (t tuple) Eval() Data     { return t }
func (t tuple) Head() Data     { return t()[0] }
func (t tuple) Tail() []Data   { return t()[1:] }
func (t tuple) String() string { return flatCol(t).String() }
func (t tuple) Empty() bool    { return chainEmpty(t()) }

func (t tuple) Signature() []BitFlag { return tupleSignature(t) }
func (t tuple) Arity() Arity         { return Arity(chainLen(t())) }
func tupleSignature(t tuple) []BitFlag {
	var flags = []BitFlag{}
	data := t()
	for _, dat := range data {
		flags = append(flags, dat.Flag())
	}
	return flags
}

//////// NESTED COLLECTION ///////////
type nestCol []Data

func (n nestCol) Eval() Data     { return n }
func (n nestCol) Head() Data     { h := n[0]; return h }
func (n nestCol) Tail() Data     { t := n[1:]; return t }
func (n nestCol) Flag() BitFlag  { return Recursives.Flag() }
func (n nestCol) String() string { return chain(n).String() }
func (n nestCol) Empty() bool {
	if n.Head() != nil {
		return false
	}
	return true
}
func conNestColEager(d ...Data) nestCol {
	con := func(d ...Data) nestCol { return conNestColEager(d...) }
	var nest nestCol
	switch len(d) {
	case 0:
		return nil
	case 1:
		return []Data{d[0], nil}
	case 2:
		return []Data{d[0], d[1]}
	default:
		return []Data{con(d[1:]...)}
	}
	return nest
}
func conNestColLazy(d ...Data) func(...Data) nestCol {
	var con func(...Data) nestCol
	con = func(d ...Data) nestCol {
		var data = d
		var nest nestCol
		switch len(d) {
		case 0:
			return nil
		case 1:
			nest = []Data{data[0]}
		default:
			nest = []Data{con(data[1:]...)}
		}
		return nest
	}
	return con
}
