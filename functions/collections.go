package functions

import d "github.com/joergreinhardt/gatwd/data"

type (
	//// COLLECTION
	ListVal func(...Callable) (Callable, ListVal)
	VecVal  func(...Callable) []Callable

	PairVal   func(...Callable) (Callable, Callable)
	KeyPair   func(...Callable) (Callable, string)
	IndexPair func(...Callable) (Callable, int)

	PairList func(...Paired) (Paired, PairList)
	PairVec  func(...Paired) []Paired
	SetVal   func(...Paired) d.Mapped
)

///////////////////////////////////////////////////////////////////////////////
//// RECURSIVE LIST OF VALUES
///
// base implementation of recursively linked lists
func ConsList(list ListVal, elems ...Callable) ListVal {
	return list.Cons(elems...)
}

func ConcatLists(a, b ListVal) ListVal {
	return ListVal(func(args ...Callable) (Callable, ListVal) {
		if len(args) > 0 {
			b = b.Cons(args...)
		}
		var head Callable
		if head, a = a(); head != nil {
			return head, ConcatLists(a, b)
		}
		return b()
	})
}

func NewList(elems ...Callable) ListVal {
	return func(args ...Callable) (Callable, ListVal) {
		if len(args) > 0 {
			elems = append(elems, args...)
		}
		if len(elems) > 0 {
			var head = elems[0]
			if len(elems) > 1 {
				return head, NewList(
					elems[1:]...,
				)
			}
			return head, NewList()
		}
		return nil, NewList()
	}
}

func (l ListVal) Cons(elems ...Callable) ListVal {
	return ListVal(func(args ...Callable) (Callable, ListVal) {
		return l(append(elems, args...)...)
	})
}

func (l ListVal) Push(elems ...Callable) ListVal {
	return ConcatLists(NewList(elems...), l)
}

func (l ListVal) Call(args ...Callable) Callable {
	if len(args) > 0 {
		return l.Cons(args...)
	}
	return l.Head()
}

// eval applys current heads eval method to passed arguments, or calle it empty
func (l ListVal) Eval(args ...d.Native) d.Native {
	if len(args) > 0 {
		return l.Cons(natToFnc(args...)...)
	}
	if head := l.Head(); head != nil {
		return head.Eval()
	}
	return d.NilVal{}
}

func (l ListVal) Empty() bool {
	if l.Head() != nil {
		if !None.Flag().Match(l.Head().TypeFnc()) ||
			!d.Nil.Flag().Match(l.Head().TypeNat()) {
			return false
		}
	}

	return true
}

// to determine the length of a recursive function, it has to be fully unwound,
// so use with care! (and ask yourself, what went wrong to make the length of a
// list be of importance)
func (l ListVal) Len() int {
	var length int
	var head, tail = l()
	if head != nil {
		length += 1 + tail.Len()
	}
	return length
}

func (l ListVal) Ident() Callable                  { return l }
func (l ListVal) Null() ListVal                    { return NewList() }
func (l ListVal) Tail() Consumeable                { _, t := l(); return t }
func (l ListVal) Head() Callable                   { h, _ := l(); return h }
func (l ListVal) Consume() (Callable, Consumeable) { return l() }
func (l ListVal) TypeFnc() TyFnc                   { return List }
func (l ListVal) TypeNat() d.TyNat                 { return l.Head().TypeNat() }

//// PAIRS OF VALUES
///
// pairs can be created empty, key & value may be constructed later
func NewEmptyPair() PairVal {
	return func(args ...Callable) (a, b Callable) {
		if len(args) > 0 {
			if len(args) > 1 {
				return args[0], args[1]
			}
			return args[0], NewNone()
		}
		return NewNone(), NewNone()
	}
}

// new pair from two callable instances
func NewPair(l, r Callable) PairVal {
	return func(args ...Callable) (Callable, Callable) {
		if len(args) > 0 {
			if len(args) > 1 {
				return args[0], args[1]
			}
			return args[0], r
		}
		return l, r
	}
}

// new pair from two native instances
func NewPairFromData(l, r d.Native) PairVal {
	return func(args ...Callable) (Callable, Callable) {
		if len(args) > 0 {
			if len(args) > 1 {
				// return pointers to natives eval functions
				return AtomVal(args[0].Eval), AtomVal(args[1].Eval)
			}

			return AtomVal(args[0].Eval), NewNone()
		}

		return AtomVal(l.Eval), AtomVal(r.Eval)
	}
}

// create a pair from literals to create instances of type DataVal, when
// key & value are later returned
func NewPairFromLiteral(l, r interface{}) PairVal {
	return func(args ...Callable) (Callable, Callable) {
		if len(args) > 0 {
			if len(args) > 1 {
				// return values eval methods as continuations
				return AtomVal(
						d.New(args[0]).Eval,
					),
					AtomVal(
						d.New(args[1]).Eval,
					)
			}
			return AtomVal(d.New(args[0]).Eval), NewNone()
		}
		return AtomVal(d.New(l).Eval), AtomVal(d.New(r).Eval)
	}
}

// pairs identity is a pair
func (p PairVal) Ident() Callable { return p }

// pair implements associative collection
func (p PairVal) Pair() Paired { return p }

// pairs implement the consumeable interface‥. construct value pairs from any
// consumeable assuming a slice where keys and values alternate
func ConsPair(list Consumeable) (PairVal, Consumeable) {
	var first, tail = list.Consume()
	if first != nil {
		var second Callable
		second, tail = tail.Consume()
		if second != nil {
			if tail != nil {
				return NewPair(first, second), tail
			}
			return NewPair(first, second), NewNone()
		}
		return NewPair(first, NewNone()), NewNone()
	}
	return NewEmptyPair(), NewNone()
}

// head returns left value to implement consumeable
func (p PairVal) Head() Callable { return p.Left() }

// tail returns right value, which either implements consumeable allready, or
// gets wrapped as a new pair, with a none instance as it's right value.
func (p PairVal) Tail() Consumeable {
	var r = p.Right()
	if r.TypeFnc().Match(Consumeables) {
		if cons, ok := r.(Consumeable); ok {
			return cons
		}
	}
	return NewPair(r, NewNone())
}

// consume returns callable head & consumeable tail values
func (p PairVal) Consume() (Callable, Consumeable) {
	l, r := p.Head(), p.Tail()
	return l, r
}

// consume pair returns either left value, case its implementing paired, and
// tail, or a new pair instance created from the first two callables and the
// tail left once those are consumed.
func (p PairVal) ConsumePair() (Paired, Consumeable) {
	// allocate left and right value
	var l, r Callable
	// assign left value to head
	l = p.Left()
	// call tail function to assign initial tail from right value
	var tail = p.Tail()
	// if left values function type matches pair flag‥.
	if l.TypeFnc().Match(Pair) {
		//‥.and left value casts paired successfully‥.
		if pair, ok := l.(Paired); ok {
			//‥.return casted pair and tail
			return pair, tail
		}
	}
	//‥.otherwise consume right value and reassign returned tail
	r, tail = tail.Consume()
	//‥.return new pair from left & right value as well as tail
	return NewPair(l, r), tail
}

// implement swappable
func (p PairVal) Swap() (Callable, Callable) { l, r := p(); return r, l }
func (p PairVal) SwappedPair() PairVal       { return NewPair(p.Right(), p.Left()) }

// implement associated
func (p PairVal) Left() Callable             { l, _ := p(); return l }
func (p PairVal) Right() Callable            { _, r := p(); return r }
func (p PairVal) Both() (Callable, Callable) { return p() }

// implement sliced
func (p PairVal) Slice() []Callable { return []Callable{p.Left(), p.Right()} }

// associative implementing element access
func (p PairVal) Key() Callable   { return p.Left() }
func (p PairVal) Value() Callable { return p.Right() }

// key and values native and functional types
func (p PairVal) KeyType() TyFnc        { return p.Left().TypeFnc() }
func (p PairVal) KeyNatType() d.TyNat   { return p.Left().TypeNat() }
func (p PairVal) ValueType() TyFnc      { return p.Right().TypeFnc() }
func (p PairVal) ValueNatType() d.TyNat { return p.Right().TypeNat() }

// composed functional type of a value pair
func (p PairVal) TypeFnc() TyFnc { return Pair }

// composed native type of a value pair
func (p PairVal) TypeNat() d.TyNat {
	return d.Pair
}

// implements compose
func (p PairVal) Empty() bool {
	if (p.Left() == nil ||
		(!p.Left().TypeFnc().Flag().Match(None) ||
			!p.Left().TypeNat().Flag().Match(d.Nil))) ||
		(p.Right() == nil ||
			(!p.Right().TypeFnc().Flag().Match(None) ||
				!p.Right().TypeNat().Flag().Match(d.Nil))) {
		return true
	}
	return false
}

// call calls the value, arguments are forwarded when calling right element
func (p PairVal) Call(args ...Callable) Callable {
	return NewPair(p.Left().Call(args...), p.Right().Call(args...))
}

// eval evaluates the value, arguments are forwarded when evaluating right element
func (p PairVal) Eval(args ...d.Native) d.Native {
	return d.NewPair(p.Left().Eval(args...), p.Right().Eval(args...))
}

//// ASSOCIATIVE PAIRS
///
// pair composed of a string key and a functional value
func NewKeyPair(key string, val Callable) KeyPair {
	return func(...Callable) (Callable, string) { return val, key }
}

func (a KeyPair) KeyStr() string {
	var _, key = a()
	return key
}

func (a KeyPair) Ident() Callable { return a }

func (a KeyPair) Pair() Paired    { return NewPair(a.Both()) }
func (a KeyPair) Pairs() []Paired { return []Paired{NewPair(a.Both())} }

func (a KeyPair) Empty() bool {
	if (a.Left() == nil ||
		(!a.Left().TypeFnc().Flag().Match(None) ||
			!a.Left().TypeNat().Flag().Match(d.Nil))) ||
		(a.Right() == nil ||
			(!a.Right().TypeFnc().Flag().Match(None) ||
				!a.Right().TypeNat().Flag().Match(d.Nil))) {
		return true
	}
	return false
}

func (a KeyPair) Both() (Callable, Callable) {
	var val, key = a()
	return NewAtom(d.StrVal(key)), val
}

// key pair implements associative interface
func (a KeyPair) GetVal(Callable) (Callable, bool) {
	var val = a.Right()
	if val != nil {
		return val, true
	}
	return NewNone(), false
}
func (a KeyPair) SetVal(key, val Callable) (Associative, bool) {
	return NewKeyPair(a.Left().String(), a.Right()), true
}
func (a KeyPair) Left() Callable {
	_, key := a()
	return NewAtom(d.StrVal(key))
}

func (a KeyPair) Right() Callable {
	val, _ := a()
	return val
}
func (a KeyPair) Key() Callable                  { return a.Left() }
func (a KeyPair) Value() Callable                { return a.Right() }
func (a KeyPair) Call(args ...Callable) Callable { return a.Right().Call(args...) }
func (a KeyPair) Eval(args ...d.Native) d.Native { return a.Right().Eval(args...) }

func (a KeyPair) KeyType() TyFnc        { return Pair }
func (a KeyPair) KeyNatType() d.TyNat   { return d.String }
func (a KeyPair) ValFncType() TyFnc     { return a.Right().TypeFnc() }
func (a KeyPair) ValNatType() d.TyNat   { return a.Right().TypeNat() }
func (a KeyPair) KeyFncType() TyFnc     { return a.Left().TypeFnc() }
func (a KeyPair) ValueType() TyFnc      { return a.Right().TypeFnc() }
func (a KeyPair) ValueNatType() d.TyNat { return a.Right().TypeNat() }

func (a KeyPair) TypeFnc() TyFnc   { return Pair | Key }
func (a KeyPair) TypeNat() d.TyNat { return d.Pair | d.String }

func ConsKeyPair(list Consumeable) (KeyPair, Consumeable) {
	var first, tail = list.Consume()
	if first != nil {
		if keyval, ok := first.Eval().(d.StrVal); ok {
			var key = string(keyval)
			var second Callable
			second, tail = tail.Consume()
			if second != nil {
				if tail != nil {
					return NewKeyPair(key, second), tail
				}
				return NewKeyPair(key, second), NewNone()
			}
			return NewKeyPair(key, NewNone()), NewNone()
		}
	}
	return NewKeyPair("", NewNone()), NewList()
}

// implement consumeable
func (p KeyPair) Consume() (Callable, Consumeable) {
	l, r := p()
	return NewAtom(d.StrVal(r)), NewList(l)
}
func (p KeyPair) Head() Callable    { l, _ := p(); return l }
func (p KeyPair) Tail() Consumeable { _, r := p(); return NewPair(NewAtom(d.StrVal(r)), NewNone()) }

// implement swappable
func (p KeyPair) Swap() (Callable, Callable) { l, r := p(); return NewAtom(d.StrVal(r)), l }
func (p KeyPair) SwappedPair() Paired        { return NewPair(p.Right(), p.Left()) }

/// pair composed of an integer and a functional value
func NewIndexPair(idx int, val Callable) IndexPair {
	return func(...Callable) (Callable, int) { return val, idx }
}

func (a IndexPair) Index() int {
	_, idx := a()
	return idx
}

func (a IndexPair) Ident() Callable { return a }

func (a IndexPair) Empty() bool {
	if (a.Left() == nil ||
		(!a.Left().TypeFnc().Flag().Match(None) ||
			!a.Left().TypeNat().Flag().Match(d.Nil))) ||
		(a.Right() == nil ||
			(!a.Right().TypeFnc().Flag().Match(None) ||
				!a.Right().TypeNat().Flag().Match(d.Nil))) {
		return true
	}
	return false
}
func (a IndexPair) Both() (Callable, Callable) {
	var val, idx = a()
	return NewAtom(d.IntVal(idx)), val
}

func (a IndexPair) Pair() Paired { return a }

func (a IndexPair) Left() Callable {
	_, idx := a()
	return NewAtom(d.IntVal(idx))
}

func (a IndexPair) Right() Callable {
	val, _ := a()
	return val
}

func (a IndexPair) Key() Callable   { return a.Left() }
func (a IndexPair) Value() Callable { return a.Right() }

func (a IndexPair) Call(args ...Callable) Callable { return a.Right().Call(args...) }
func (a IndexPair) Eval(args ...d.Native) d.Native { return a.Right().Eval(args...) }

func (a IndexPair) KeyType() TyFnc        { return Pair }
func (a IndexPair) KeyNatType() d.TyNat   { return d.Int }
func (a IndexPair) ValueType() TyFnc      { return a.Right().TypeFnc() }
func (a IndexPair) ValueNatType() d.TyNat { return a.Right().TypeNat() }

func (a IndexPair) TypeFnc() TyFnc   { return Pair | Index }
func (a IndexPair) TypeNat() d.TyNat { return d.Pair | d.Int }

func ConsIndexPair(list Consumeable) (IndexPair, Consumeable) {
	var first, tail = list.Consume()
	if first != nil {
		if idxval, ok := first.Eval().(d.IntVal); ok {
			var index = int(idxval)
			var second Callable
			second, tail = tail.Consume()
			if second != nil {
				if tail != nil {
					return NewIndexPair(index, second), tail
				}
				return NewIndexPair(index, second), NewNone()
			}
			return NewIndexPair(index, NewNone()), NewNone()
		}
	}
	return NewIndexPair(0, NewNone()), NewList()
}

// implement consumeable
func (p IndexPair) Consume() (Callable, Consumeable) {
	l, r := p()
	return NewAtom(d.StrVal(r)), NewList(l)
}
func (p IndexPair) Head() Callable    { _, r := p(); return NewAtom(d.StrVal(r)) }
func (p IndexPair) Tail() Consumeable { l, _ := p(); return NewPair(l, NewNone()) }

// implement swappable
func (p IndexPair) Swap() (Callable, Callable) { l, r := p(); return NewAtom(d.StrVal(r)), l }
func (p IndexPair) SwappedPair() Paired        { return NewPair(p.Right(), p.Left()) }

//// LIST OF PAIRS
func ConsPairList(list PairList, pairs ...Paired) PairList {
	return list.Cons(pairs...)
}
func ConcatPairLists(a, b PairList) PairList {
	return PairList(func(args ...Paired) (Paired, PairList) {
		if len(args) > 0 {
			b = b.Cons(args...)
		}
		var pair Paired
		if pair, a = a(); pair != nil {
			return pair, ConcatPairLists(a, b)
		}
		return b()
	})
}
func NewPairList(elems ...Paired) PairList {
	return func(pairs ...Paired) (Paired, PairList) {
		if len(pairs) > 0 {
			elems = append(elems, pairs...)
		}
		if len(elems) > 0 {
			var pair = elems[0]
			if len(elems) > 1 {
				return pair, NewPairList(
					elems[1:]...,
				)
			}
			return pair, NewPairList()
		}
		return nil, NewPairList()
	}
}
func (l PairList) Cons(elems ...Paired) PairList {
	return PairList(func(args ...Paired) (Paired, PairList) {
		return l(append(elems, args...)...)
	})
}
func (l PairList) Push(elems ...Paired) PairList {
	return ConcatPairLists(NewPairList(elems...), l)
}
func (l PairList) Call(args ...Callable) Callable {
	var pairs = []Paired{}
	if len(args) > 0 {
		pairs = append(pairs, argsToPaired(args...)...)
	}
	var head Callable
	head, l = l(pairs...)
	return head
}

// eval applys current heads eval method to passed arguments, or calle it empty
func (l PairList) Eval(args ...d.Native) d.Native {
	if head := l.Head(); head != nil {
		return head.Eval(args...)
	}
	return d.NilVal{}
}

func (l PairList) Empty() bool {
	if pair := l.HeadPair(); pair != nil {
		return pair.Empty()
	}
	return true
}

// to determine the length of a recursive function, it has to be fully unwound,
// so use with care! (and ask yourself, what went wrong to make the length of a
// list be of importance)
func (l PairList) Len() int {
	var length int
	var head, tail = l()
	if head != nil {
		length += 1 + tail.Len()
	}
	return length
}

func (l PairList) Ident() Callable                         { return l }
func (l PairList) Null() PairList                          { return NewPairList() }
func (l PairList) Tail() Consumeable                       { _, t := l(); return t }
func (l PairList) TailPairs() ConsumeablePairs             { _, t := l(); return t }
func (l PairList) Head() Callable                          { h, _ := l(); return h }
func (l PairList) HeadPair() Paired                        { p, _ := l(); return p }
func (l PairList) Consume() (Callable, Consumeable)        { return l() }
func (l PairList) ConsumePair() (Paired, ConsumeablePairs) { return l() }
func (l PairList) TypeFnc() TyFnc                          { return List }
func (l PairList) TypeNat() d.TyNat                        { return l.Head().TypeNat() }

//////////////////////////////////////////////////////////////////////////////////////////
//// VECTORS (SLICES) OF VALUES
func NewEmptyVector(init ...Callable) VecVal { return NewVector() }

func NewVector(init ...Callable) VecVal {
	var vector = init
	return func(args ...Callable) []Callable {
		if len(args) > 0 {
			vector = append(
				vector,
				args...,
			)
		}
		return vector
	}
}

func ConsVector(vec Vectorized, args ...Callable) VecVal {
	return NewVector(append(vec.Slice(), args...)...)
}

func AppendVectors(vec Vectorized, args ...Callable) VecVal {
	return NewVector(append(vec.Slice(), args...)...)
}

func AppendToVector(init ...Callable) VecVal {
	return func(args ...Callable) []Callable {
		return append(init, args...)
	}
}

func (v VecVal) Append(args ...Callable) VecVal {
	return NewVector(append(v(), args...)...)
}

func (v VecVal) Cons(args ...Callable) VecVal {
	return ConsVector(v, args...)
}

func (v VecVal) Ident() Callable { return v }

func (v VecVal) Call(d ...Callable) Callable { return NewVector(v(d...)...) }

func (v VecVal) Eval(args ...d.Native) d.Native {

	var result = []d.Native{}

	for _, arg := range args {
		result = append(result, arg)
	}

	return d.DataSlice(result)
}

func (v VecVal) TypeFnc() TyFnc {
	if len(v()) > 0 {
		return Vector | v.Head().TypeFnc()
	}
	return Vector | None
}

func (v VecVal) TypeNat() d.TyNat {
	if len(v()) > 0 {
		return d.Slice.TypeNat() | v.Head().TypeNat()
	}
	return d.Slice.TypeNat() | d.Nil
}

func (v VecVal) Head() Callable {
	if v.Len() > 0 {
		return v.Slice()[0]
	}
	return nil
}

func (v VecVal) Tail() Consumeable {
	if v.Len() > 1 {
		return NewVector(v.Slice()[1:]...)
	}
	return NewEmptyVector()
}

func (v VecVal) Consume() (Callable, Consumeable) {
	return v.Head(), v.Tail()
}

func (v VecVal) Empty() bool {
	if len(v()) > 0 {
		for _, val := range v() {
			if !val.TypeNat().Flag().Match(d.Nil) &&
				!val.TypeFnc().Flag().Match(None) {
				return false
			}
		}
	}
	return true
}

func (v VecVal) Len() int          { return len(v()) }
func (v VecVal) Vector() VecVal    { return v }
func (v VecVal) Slice() []Callable { return v() }

func (v VecVal) Get(i int) (Callable, bool) {
	if i < v.Len() {
		return v()[i], true
	}
	return NewNone(), false
}

func (v VecVal) Set(i int, val Callable) (Vectorized, bool) {
	if i < v.Len() {
		var slice = v()
		slice[i] = val
		return VecVal(
			func(elems ...Callable) []Callable {
				return slice
			}), true

	}
	return v, false
}

func (v VecVal) Sort(flag d.TyNat) {
	var ps = newDataSorter(v()...)
	ps.Sort(flag)
	v = NewVector(ps...)
}

func (v VecVal) Search(praed Callable) int {
	return newDataSorter(v()...).Search(praed)
}

//// ASSOCIATIVE SLICE OF VALUE PAIRS
///
// list of associative pairs in sequential order associated, sorted and
// searched by left value of the pairs
func NewEmptyPairVec() PairVec {
	return PairVec(func(args ...Paired) []Paired {
		var pairs = []Paired{}
		if len(args) > 0 {
			pairs = append(pairs, args...)
		}
		return pairs
	})
}

func NewPairVectorFromPairs(pairs ...Paired) PairVec {
	return PairVec(func(args ...Paired) []Paired {
		if len(args) > 0 {
			return append(pairs, args...)
		}
		return pairs
	})
}

func ConPairVecFromArgs(rec PairVec, args ...Callable) PairVec {
	var pairs = []Paired{}
	for _, arg := range args {
		if pair, ok := arg.(Paired); ok {
			pairs = append(pairs, pair)
		}
	}
	return NewPairVectorFromPairs(append(rec(), pairs...)...)
}

func NewPairVec(args ...Paired) PairVec {
	return NewPairVectorFromPairs(args...)
}

func ConPairVecFromPairs(rec PairVec, pairs ...Paired) PairVec {
	return NewPairVectorFromPairs(append(rec(), pairs...)...)
}

func ConsPairVec(pvec PairVec, args ...Callable) PairVec {
	var pairs = pvec.Pairs()
	for _, arg := range args {
		if pair, ok := arg.(Paired); ok {
			pairs = append(pairs, pair)
		}
	}
	return PairVec(func(args ...Paired) []Paired {
		if len(args) > 0 {
			return ConPairVecFromPairs(pvec, args...)()
		}
		return append(pvec(), pairs...)
	})
}
func (v PairVec) Cons(args ...Callable) PairVec {
	return ConsPairVec(v, args...)
}
func (v PairVec) Consume() (Callable, Consumeable) {
	return v.Head(), v.Tail()
}

func (v PairVec) Empty() bool {
	if len(v()) > 0 {
		for _, pair := range v() {
			if !pair.Empty() {
				return false
			}
		}
	}
	return true
}

func (v PairVec) KeyFncType() TyFnc {
	if v.Len() > 0 {
		return v.Pairs()[0].Left().TypeFnc()
	}
	return None
}

func (v PairVec) KeyNatType() d.TyNat {
	if v.Len() > 0 {
		return v.Pairs()[0].Left().TypeNat()
	}
	return d.Nil
}

func (v PairVec) ValFncType() TyFnc {
	if v.Len() > 0 {
		return v.Pairs()[0].Right().TypeFnc()
	}
	return None
}

func (v PairVec) ValNatType() d.TyNat {
	if v.Len() > 0 {
		return v.Pairs()[0].Right().TypeNat()
	}
	return d.Nil
}

func (v PairVec) TypeFnc() TyFnc { return Vector | Pair }

func (v PairVec) TypeNat() d.TyNat {
	if len(v()) > 0 {
		return d.Slice | v.Head().TypeNat()
	}
	return d.Slice | d.Nil.TypeNat()
}

func (v PairVec) Len() int { return len(v()) }

func (v PairVec) Sort(flag d.TyNat) {
	var ps = newPairSorter(v.Pairs()...)
	ps.Sort(flag)
	v = NewPairVectorFromPairs(ps...)
}

func (v PairVec) Search(praed Callable) int {
	return newPairSorter(v.Pairs()...).Search(praed)
}

func (v PairVec) Get(idx int) (Paired, bool) {
	if idx < v.Len()-1 {
		return v()[idx], true
	}
	return NewKeyPair("None", NewNone()), false
}

func (v PairVec) GetVal(praed Callable) (Callable, bool) {
	return NewPairVectorFromPairs(newPairSorter(v.Pairs()...).Get(praed)), true
}

func (v PairVec) Range(praed Callable) []Paired {
	return newPairSorter(v.Pairs()...).Range(praed)
}

func (v PairVec) Pairs() []Paired {
	var pairs = []Paired{}
	for _, pair := range v() {
		pairs = append(pairs, pair)
	}
	return pairs
}

func (v PairVec) ConsumePair() (Paired, ConsumeablePairs) {
	var pairs = v()
	if len(pairs) > 0 {
		if len(pairs) > 1 {
			return pairs[0], NewPairVec(pairs[1:]...)
		}
		return pairs[0], NewPairVec()
	}
	return nil, NewPairVec()
}

func (v PairVec) SwitchedPairs() []Paired {
	var switched = []Paired{}
	for _, pair := range v() {
		switched = append(
			switched,
			pair,
		)
	}
	return switched
}

func (v PairVec) SetVal(key, value Callable) (Associative, bool) {
	if idx := v.Search(key); idx >= 0 {
		var pairs = v()
		pairs[idx] = NewKeyPair(key.String(), value)
		return NewPairVec(pairs...), true
	}
	return NewPairVec(append(v(), NewKeyPair(key.String(), value))...), false
}

func (v PairVec) Slice() []Callable {
	var fncs = []Callable{}
	for _, pair := range v() {
		fncs = append(fncs, NewPair(pair.Left(), pair.Right()))
	}
	return fncs
}

func (v PairVec) HeadPair() Paired {
	if v.Len() > 0 {
		return v()[0].(Paired)
	}
	return NewPair(NewNone(), NewNone())
}
func (v PairVec) Head() Callable {
	if v.Len() > 0 {
		return v.Pairs()[0]
	}
	return nil
}

func (v PairVec) TailPairs() ConsumeablePairs {
	if v.Len() > 1 {
		return NewPairVectorFromPairs(v.Pairs()[1:]...)
	}
	return NewEmptyPairVec()
}
func (v PairVec) Tail() Consumeable {
	if v.Len() > 1 {
		return NewPairVectorFromPairs(v.Pairs()[1:]...)
	}
	return NewEmptyPairVec()
}

func (v PairVec) Call(args ...Callable) Callable {
	return v.Cons(args...)
}

func (v PairVec) Eval(p ...d.Native) d.Native {
	var slice = d.DataSlice{}
	for _, pair := range v() {
		d.SliceAppend(slice, d.NewPair(pair.Left(), pair.Right()))
	}
	return slice
}

///////////////////////////////////////////////////////////////////////////////
//// ASSOCIATIVE SET (HASH MAP OF VALUES)
///
// unordered associative set of key/value pairs that can be sorted, accessed
// and searched by the left (key) value of the pair
func ConsSet(set SetVal, pairs ...Paired) SetVal {
	var knat = set.KeyNatType()
	var vnat = set.ValNatType()
	var m = set()
	for _, arg := range pairs {
		if pair, ok := arg.(Paired); ok {
			if pair.Left().TypeNat() == knat &&
				pair.Right().TypeNat() == vnat {
				m.Set(pair.Left(), pair.Right())
			}
		}
	}
	return SetVal(func(pairs ...Paired) d.Mapped { return m })
}

// new set discriminates between sets where all members have identical keys and
// such with mixed keys and chooses the appropriate native set accordingly.
func NewSet(pairs ...Paired) SetVal {
	var set d.Mapped
	var knat d.BitFlag
	if len(pairs) > 0 {
		// first passed pair determines initial key type
		knat = pairs[0].Left().TypeNat().Flag()
		// OR concat all the keys types, to see if arguments are of
		// mixed type
		for _, pair := range pairs {
			knat = knat | pair.Left().TypeNat().Flag()
		}
		// for sets with pure key type, choose the appropriate native
		// set type
		if knat.Count() == 1 {
			switch {
			case knat.Match(d.Int):
				set = d.SetInt{}
			case knat.Match(d.Uint):
				set = d.SetUint{}
			case knat.Match(d.Flag):
				set = d.SetFlag{}
			case knat.Match(d.Float):
				set = d.SetFloat{}
			case knat.Match(d.String):
				set = d.SetString{}
			}
		} else {
			// otherwise choose a set keyed by interface type to
			// keep every possible kind of value
			set = d.SetVal{}
		}
	}
	return SetVal(func(pairs ...Paired) d.Mapped { return set })
}

// splits set into two lists, one containing all keys and the other all values
func (v SetVal) Split() (VecVal, VecVal) {
	var keys, vals = []Callable{}, []Callable{}
	for _, pair := range v.Pairs() {
		keys = append(keys, pair.Left())
		vals = append(vals, pair.Right())
	}
	return NewVector(keys...), NewVector(vals...)
}

func (v SetVal) Pairs() []Paired {
	var pairs = []Paired{}
	for _, field := range v().Fields() {
		pairs = append(
			pairs,
			NewPairFromData(
				field.Left(),
				field.Right()))
	}
	return pairs
}

// return all members keys
func (v SetVal) Keys() VecVal { k, _ := v.Split(); return k }

// return all members values
func (v SetVal) Data() VecVal { _, d := v.Split(); return d }

func (v SetVal) Len() int { return v().Len() }

func (v SetVal) Empty() bool {
	for _, pair := range v.Pairs() {
		if !pair.Empty() {
			return false
		}
	}
	return true
}

func (v SetVal) GetVal(key Callable) (Callable, bool) {
	var m = v()
	if value, ok := m.Get(key); ok {
		return NewAtom(value), ok
	}
	return NewNone(), false
}

func (v SetVal) SetVal(key, value Callable) (Associative, bool) {
	var m = v()
	return SetVal(func(pairs ...Paired) d.Mapped { return m.Set(key, value) }), true
}

func (v SetVal) Slice() []Callable {
	var pairs = []Callable{}
	for _, pair := range v.Pairs() {
		pairs = append(pairs, pair)
	}
	return pairs
}

// call method performs a value lookup
func (v SetVal) Call(args ...Callable) Callable {
	var results = []Callable{}
	for _, arg := range args {
		if val, ok := v.GetVal(arg); ok {
			results = append(results, val)
		}
	}
	if len(results) > 0 {
		if len(results) > 1 {
			return NewVector(results...)
		}
		return results[0]
	}
	return NewNone()
}

// eval method performs a value lookup and returns contained value as native
// without any conversion
func (v SetVal) Eval(args ...d.Native) d.Native {
	var results = []d.Native{}
	var m = v()
	for _, arg := range args {
		if val, ok := m.Get(arg); ok {
			results = append(results, val)
		}
	}
	if len(results) > 0 {
		if len(results) > 1 {
			return d.NewSlice(results...)
		}
		return results[0]
	}
	return d.NilVal{}
}

func (v SetVal) TypeFnc() TyFnc { return Set }

func (v SetVal) TypeNat() d.TyNat { return d.Map | d.Functor }

func (v SetVal) KeyFncType() TyFnc {
	if v.Len() > 0 {
		return v.Pairs()[0].Left().TypeFnc()
	}
	return None
}

func (v SetVal) KeyNatType() d.TyNat {
	if v.Len() > 0 {
		return v.Pairs()[0].Left().TypeNat()
	}
	return d.Nil
}

func (v SetVal) ValFncType() TyFnc {
	if v.Len() > 0 {
		return v.Pairs()[0].Right().TypeFnc()
	}
	return None
}

func (v SetVal) ValNatType() d.TyNat {
	if v.Len() > 0 {
		return v.Pairs()[0].Right().TypeNat()
	}
	return d.Nil
}

func (v SetVal) Consume() (Callable, Consumeable) {
	return v.Head(), v.Tail()
}

func (v SetVal) Head() Callable {
	if v.Len() > 0 {
		var vec = NewPairVectorFromPairs(
			v.Pairs()...,
		)
		vec.Sort(v.KeyNatType())
		return vec()[0]
	}
	return nil
}

func (v SetVal) Tail() Consumeable {
	if v.Len() > 1 {
		var vec = NewPairVectorFromPairs(
			v.Pairs()...,
		)
		vec.Sort(v.KeyNatType())
		return NewPairVec(vec()[:1]...)
	}
	return nil
}
