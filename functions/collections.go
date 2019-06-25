package functions

import d "github.com/joergreinhardt/gatwd/data"

type (
	//// COLLECTION
	ListCol func(...Callable) (Callable, ListCol)
	VecCol  func(...Callable) []Callable
	SetCol  func(...Paired) d.Mapped

	PairVal   func(...Callable) (Callable, Callable)
	KeyPair   func(...Callable) (Callable, string)
	IndexPair func(...Callable) (Callable, int)

	PairList func(...Paired) (Paired, PairList)
	PairVec  func(...Paired) []Paired
)

///////////////////////////////////////////////////////////////////////////////
//// RECURSIVE LIST OF VALUES
///
// base implementation of recursively linked lists
func ConList(list ListCol, elems ...Callable) ListCol {
	return list.Con(elems...)
}

func ConcatLists(a, b ListCol) ListCol {
	return ListCol(func(args ...Callable) (Callable, ListCol) {
		if len(args) > 0 {
			b = b.Con(args...)
		}
		var head Callable
		if head, a = a(); head != nil {
			return head, ConcatLists(a, b)
		}
		return b()
	})
}

func NewList(elems ...Callable) ListCol {
	return func(args ...Callable) (Callable, ListCol) {
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

func (l ListCol) Con(elems ...Callable) ListCol {
	return ListCol(func(args ...Callable) (Callable, ListCol) {
		return l(append(elems, args...)...)
	})
}

func (l ListCol) Push(elems ...Callable) ListCol {
	return ConcatLists(NewList(elems...), l)
}

func (l ListCol) Call(args ...Callable) Callable {
	if len(args) > 0 {
		return l.Con(args...)
	}
	return l.Head()
}

// get n'st element in list
func (l ListCol) GetIdx(n int) Callable {
	var head, list = l()
	for i := 0; i < n; i++ {
		head, list = list()
		if head == nil {
			return NewNone()
		}
	}
	return head
}

// eval applys current heads eval method to passed arguments, or calle it empty
func (l ListCol) Eval(args ...d.Native) d.Native {
	if head := l.Head(); head != nil {
		return head.Eval()
	}
	return d.NilVal{}
}

func (l ListCol) Empty() bool {
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
func (l ListCol) Len() int {
	var length int
	var head, tail = l()
	if head != nil {
		length += 1 + tail.Len()
	}
	return length
}

func (l ListCol) Ident() Callable                  { return l }
func (l ListCol) Null() ListCol                    { return NewList() }
func (l ListCol) Tail() Consumeable                { _, t := l(); return t }
func (l ListCol) Head() Callable                   { h, _ := l(); return h }
func (l ListCol) Consume() (Callable, Consumeable) { return l() }
func (l ListCol) TypeFnc() TyFnc                   { return List }
func (l ListCol) SubType() TyFnc                   { return l.Head().TypeFnc() }
func (l ListCol) TypeNat() d.TyNat                 { return l.Head().TypeNat() }
func (l ListCol) TypeName() string {
	if l.Len() > 0 {
		return "[" + l.Head().TypeName() + "]"
	}
	return "[]"
}

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

// pairs identity is a pair
func (p PairVal) Ident() Callable { return p }

// pair implements associative collection
func (p PairVal) Pair() Paired { return p }

// pairs implement the consumeable interface‥. construct value pairs from any
// consumeable assuming a slice where keys and values alternate
func ConPair(list Consumeable) (PairVal, Consumeable) {
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
	if r.TypeFnc().Match(Collection) {
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
func (p PairVal) TypeName() string {
	return "(" + p.Key().TypeName() + ", " + p.Value().TypeName() + ")"
}
func (p PairVal) KeyType() TyFnc        { return p.Left().TypeFnc() }
func (p PairVal) KeyNatType() d.TyNat   { return p.Left().TypeNat() }
func (p PairVal) ValType() TyFnc        { return p.Right().TypeFnc() }
func (p PairVal) ValueNatType() d.TyNat { return p.Right().TypeNat() }

// composed functional type of a value pair
func (p PairVal) TypeFnc() TyFnc { return Pair }
func (p PairVal) SubType() TyFnc { return p.KeyType() | p.ValType() }

// composed native type of a value pair
func (p PairVal) TypeNat() d.TyNat { return d.Pair }

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
	return d.NewPair(p.Left().Eval(), p.Right().Eval())
}

//// ASSOCIATIVE PAIRS
///
// pair composed of a string key and a functional value
func NewKeyPair(key string, val Callable) KeyPair {
	return func(...Callable) (Callable, string) { return val, key }
}

func (p KeyPair) TypeName() string {
	return "(" + p.Key().TypeName() + ", " + p.Value().TypeName() + ")"
}
func (a KeyPair) KeyStr() string                 { _, key := a(); return key }
func (a KeyPair) Ident() Callable                { return a }
func (a KeyPair) Value() Callable                { val, _ := a(); return val }
func (a KeyPair) Left() Callable                 { return a.Value() }
func (a KeyPair) Right() Callable                { return NewNative(d.StrVal(a.KeyStr())) }
func (a KeyPair) Both() (Callable, Callable)     { return a.Left(), a.Right() }
func (a KeyPair) Pair() Paired                   { return NewPair(a.Both()) }
func (a KeyPair) Pairs() []Paired                { return []Paired{NewPair(a.Both())} }
func (a KeyPair) Key() Callable                  { return a.Right() }
func (a KeyPair) Call(args ...Callable) Callable { return a.Value().Call(args...) }
func (a KeyPair) Eval(args ...d.Native) d.Native { return a.Value().Eval() }
func (a KeyPair) KeyNatType() d.TyNat            { return d.String }
func (a KeyPair) KeyFncType() TyFnc              { return Data }
func (a KeyPair) KeyType() TyFnc                 { return Data }
func (a KeyPair) ValNatType() d.TyNat            { return a.Value().TypeNat() }
func (a KeyPair) ValFncType() TyFnc              { return a.Value().TypeFnc() }
func (a KeyPair) ValType() TyFnc                 { return a.Value().TypeFnc() }
func (a KeyPair) TypeNat() d.TyNat               { return d.Pair | d.String }
func (a KeyPair) TypeFnc() TyFnc                 { return Pair }
func (a KeyPair) SubType() TyFnc                 { return Key }

// implement consumeable
func (p KeyPair) Head() Callable                   { return p.Value() }
func (p KeyPair) Tail() Consumeable                { return NewPair(NewNative(d.StrVal(p.KeyStr())), NewNone()) }
func (p KeyPair) Consume() (Callable, Consumeable) { return p.Head(), p.Tail() }

// implement swappable
func (p KeyPair) Swap() (Callable, Callable) { l, r := p(); return NewNative(d.StrVal(r)), l }
func (p KeyPair) SwappedPair() Paired        { return NewPair(p.Right(), p.Left()) }

func (a KeyPair) Empty() bool {
	if a.Key() != nil && a.Value() != nil && a.Value().TypeFnc() != None {
		return true
	}
	return false
}

// key pair implements associative interface
func (a KeyPair) GetVal(Callable) (Callable, bool) {
	var val = a.Value()
	if val != nil {
		return val, true
	}
	return NewNone(), false
}
func (a KeyPair) SetVal(key, val Callable) (Associative, bool) {
	return NewKeyPair(a.KeyStr(), a.Value()), true
}

func ConKeyPair(list Consumeable) (KeyPair, Consumeable) {
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

///////////////////////////////////////////////////////////////////////////////
/// pair composed of an integer and a functional value
func NewIndexPair(idx int, val Callable) IndexPair {
	return func(...Callable) (Callable, int) { return val, idx }
}
func (a IndexPair) Ident() Callable                { return a }
func (a IndexPair) Index() int                     { _, idx := a(); return idx }
func (a IndexPair) Value() Callable                { val, _ := a(); return val }
func (a IndexPair) Left() Callable                 { return a.Value() }
func (a IndexPair) Right() Callable                { return NewNative(d.IntVal(a.Index())) }
func (a IndexPair) Both() (Callable, Callable)     { return a.Left(), a.Right() }
func (a IndexPair) Pair() Paired                   { return a }
func (a IndexPair) Pairs() []Paired                { return []Paired{NewPair(a.Both())} }
func (a IndexPair) Key() Callable                  { return a.Right() }
func (a IndexPair) Call(args ...Callable) Callable { return a.Value().Call(args...) }
func (a IndexPair) Eval(args ...d.Native) d.Native { return a.Value().Eval() }
func (a IndexPair) ValType() TyFnc                 { return a.Value().TypeFnc() }
func (a IndexPair) ValNatType() d.TyNat            { return a.Value().TypeNat() }
func (a IndexPair) ValFncType() TyFnc              { return a.Value().TypeFnc() }
func (a IndexPair) KeyType() TyFnc                 { return Key }
func (a IndexPair) KeyFncType() TyFnc              { return Key }
func (a IndexPair) KeyNatType() d.TyNat            { return d.Int }
func (a IndexPair) TypeFnc() TyFnc                 { return Pair }
func (a IndexPair) SubType() TyFnc                 { return Index }
func (a IndexPair) TypeNat() d.TyNat               { return d.Pair | d.Int | a.ValNatType() }
func (p IndexPair) TypeName() string {
	return "(" + p.Left().TypeName() + ", " + p.Value().TypeName() + ")"
}

// implement consumeable
func (p IndexPair) Head() Callable                   { return p.Value() }
func (p IndexPair) Tail() Consumeable                { return NewPair(New(d.IntVal(p.Index())), NewNone()) }
func (p IndexPair) Consume() (Callable, Consumeable) { return p.Head(), p.Tail() }

// implement swappable
func (p IndexPair) Swap() (Callable, Callable) { l, r := p(); return NewNative(d.StrVal(r)), l }
func (p IndexPair) SwappedPair() Paired        { return NewPair(p.Right(), p.Left()) }
func (a IndexPair) Empty() bool {
	if a.Index() >= 0 && a.Value() != nil && a.Value().TypeFnc() != None {
		return true
	}
	return false
}

// key pair implements associative interface
func (a IndexPair) GetVal(Callable) (Callable, bool) {
	var val = a.Value()
	if val != nil {
		return val, true
	}
	return NewNone(), false
}
func (a IndexPair) SetVal(index, val Callable) (Associative, bool) {
	return NewIndexPair(a.Index(), a.Value()), true
}
func ConIndexPair(list Consumeable) (IndexPair, Consumeable) {
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

//// LIST OF PAIRS
func ConPairList(list PairList, pairs ...Paired) PairList {
	return list.Con(pairs...)
}
func ConcatPairLists(a, b PairList) PairList {
	return PairList(func(args ...Paired) (Paired, PairList) {
		if len(args) > 0 {
			b = b.Con(args...)
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
func (l PairList) Con(elems ...Paired) PairList {
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
		return head.Eval()
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
func (l PairList) SubType() TyFnc                          { return Pair }
func (l PairList) TypeNat() d.TyNat                        { return l.Head().TypeNat() }
func (l PairList) KeyType() TyFnc {
	return l.Head().(PairVal).KeyType()
}
func (l PairList) ValType() TyFnc {
	return l.Head().(PairVal).ValType()
}
func (l PairList) TypeName() string {
	if l.Len() > 0 {
		return "[" + l.HeadPair().TypeName() + "]"
	}
	return "[]"
}

// helper function to group arguments pair wise. assumes the arguments to
// either implement paired, or be alternating pairs of key & value. in case the
// number of passed arguments that are not pairs is uneven, last field will be
// filled up with a value of type none
func argsToPaired(args ...Callable) []Paired {
	var pairs = []Paired{}
	var alen = len(args)
	for i, arg := range args {
		if arg.TypeFnc().Match(Pair) {
			pairs = append(pairs, arg.(Paired))
		}
		if i < alen-2 {
			i = i + 1
			pairs = append(pairs, NewPair(arg, args[i]))
		}
		pairs = append(pairs, NewPair(arg, NewNone()))
	}
	return pairs
}

//////////////////////////////////////////////////////////////////////////////////////////
//// VECTORS (SLICES) OF VALUES
func NewEmptyVector(init ...Callable) VecCol { return NewVector() }

func NewVector(init ...Callable) VecCol {
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

func ConVector(vec Vectorized, args ...Callable) VecCol {
	return NewVector(append(vec.Slice(), args...)...)
}

func AppendVectors(vec Vectorized, args ...Callable) VecCol {
	return NewVector(append(vec.Slice(), args...)...)
}

func AppendArgToVector(init ...Callable) VecCol {
	return func(args ...Callable) []Callable {
		return append(init, args...)
	}
}

func (v VecCol) Append(args ...Callable) VecCol {
	return NewVector(append(v(), args...)...)
}

func (v VecCol) Con(args ...Callable) VecCol {
	return ConVector(v, args...)
}

func (v VecCol) Ident() Callable { return v }

func (v VecCol) Call(d ...Callable) Callable {
	return NewVector(v(d...)...)
}

func (v VecCol) Eval(args ...d.Native) d.Native {

	var results = []d.Native{}

	for _, arg := range v() {
		results = append(results, arg.Eval())
	}

	return d.NewSlice(results...)
}

func (v VecCol) TypeFnc() TyFnc { return Vector }
func (v VecCol) SubType() TyFnc { return v.Head().TypeFnc() }

func (v VecCol) TypeNat() d.TyNat {
	if len(v()) > 0 {
		return d.Slice.TypeNat() | v.Head().TypeNat()
	}
	return d.Slice.TypeNat() | d.Nil
}

func (v VecCol) Head() Callable {
	if v.Len() > 0 {
		return v.Slice()[0]
	}
	return nil
}

func (v VecCol) Tail() Consumeable {
	if v.Len() > 1 {
		return NewVector(v.Slice()[1:]...)
	}
	return NewEmptyVector()
}

func (v VecCol) Consume() (Callable, Consumeable) {
	return v.Head(), v.Tail()
}

func (v VecCol) Empty() bool {
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

func (v VecCol) Len() int          { return len(v()) }
func (v VecCol) Vector() VecCol    { return v }
func (v VecCol) Slice() []Callable { return v() }

func (v VecCol) Get(i int) (Callable, bool) {
	if i < v.Len() {
		return v()[i], true
	}
	return NewNone(), false
}

func (v VecCol) Set(i int, val Callable) (Vectorized, bool) {
	if i < v.Len() {
		var slice = v()
		slice[i] = val
		return VecCol(
			func(elems ...Callable) []Callable {
				return slice
			}), true

	}
	return v, false
}

func (v VecCol) Sort(flag d.TyNat) {
	var ps = SortData(v()...)
	ps.Sort(flag)
	v = NewVector(ps...)
}

func (v VecCol) Search(praed Callable) int {
	return SortData(v()...).Search(praed)
}
func (v VecCol) TypeName() string {
	if v.Len() > 0 {
		return "[" + v.SubType().TypeName() + "]"
	}
	return "[]"
}

///////////////////////////////////////////////////////////////////////////////
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

func ConPairListFromArgs(rec PairVec, args ...Callable) PairVec {
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

func ConPairVec(rec PairVec, pairs ...Paired) PairVec {
	return NewPairVectorFromPairs(append(rec(), pairs...)...)
}

func ConPairVecFromArgs(pvec PairVec, args ...Callable) PairVec {
	var pairs = pvec.Pairs()
	for _, arg := range args {
		if pair, ok := arg.(Paired); ok {
			pairs = append(pairs, pair)
		}
	}
	return PairVec(func(args ...Paired) []Paired {
		if len(args) > 0 {
			return ConPairVec(pvec, args...)()
		}
		return append(pvec(), pairs...)
	})
}
func (v PairVec) Con(args ...Callable) PairVec {
	return ConPairVecFromArgs(v, args...)
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

func (v PairVec) KeyType() TyFnc {
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

func (v PairVec) ValType() TyFnc {
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

func (v PairVec) TypeFnc() TyFnc { return Vector }
func (v PairVec) SubType() TyFnc { return Pair }

func (v PairVec) TypeNat() d.TyNat {
	if len(v()) > 0 {
		return d.Slice | v.Head().TypeNat()
	}
	return d.Slice | d.Nil.TypeNat()
}

func (v PairVec) Len() int { return len(v()) }

func (v PairVec) Sort(flag d.TyNat) {
	var ps = SortPairs(v.Pairs()...)
	ps.Sort(flag)
	v = NewPairVectorFromPairs(ps...)
}

func (v PairVec) Search(praed Callable) int {
	return SortPairs(v.Pairs()...).Search(praed)
}

func (v PairVec) Get(idx int) (Paired, bool) {
	if idx < v.Len()-1 {
		return v()[idx], true
	}
	return NewKeyPair("None", NewNone()), false
}

func (v PairVec) GetVal(praed Callable) (Callable, bool) {
	return NewPairVectorFromPairs(SortPairs(v.Pairs()...).Get(praed)), true
}

func (v PairVec) Range(praed Callable) []Paired {
	return SortPairs(v.Pairs()...).Range(praed)
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
	return v.Con(args...)
}

func (v PairVec) Eval(args ...d.Native) d.Native {
	var slice = d.DataSlice{}
	for _, pair := range v() {
		d.SliceAppend(slice, d.NewPair(pair.Left(), pair.Right()))
	}
	return slice
}
func (v PairVec) TypeName() string {
	if v.Len() > 0 {
		return "[" + v.SubType().TypeName() + "]"
	}
	return "[]"
}

///////////////////////////////////////////////////////////////////////////////
//// ASSOCIATIVE SET (HASH MAP OF VALUES)
///
// unordered associative set of key/value pairs that can be sorted, accessed
// and searched by the left (key) value of the pair
func ConSet(set SetCol, pairs ...Paired) SetCol {
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
	return SetCol(func(pairs ...Paired) d.Mapped { return m })
}

// new set discriminates between sets where all members have identical keys and
// such with mixed keys and chooses the appropriate native set accordingly.
func NewSet(pairs ...Paired) SetCol {
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
	return SetCol(func(pairs ...Paired) d.Mapped { return set })
}

// splits set into two lists, one containing all keys and the other all values
func (v SetCol) Split() (VecCol, VecCol) {
	var keys, vals = []Callable{}, []Callable{}
	for _, pair := range v.Pairs() {
		keys = append(keys, pair.Left())
		vals = append(vals, pair.Right())
	}
	return NewVector(keys...), NewVector(vals...)
}

func (v SetCol) Pairs() []Paired {
	var pairs = []Paired{}
	for _, field := range v().Fields() {
		pairs = append(
			pairs,
			NewPair(
				NewNative(field.Left()),
				NewNative(field.Right())))
	}
	return pairs
}

// return all members keys
func (v SetCol) Keys() VecCol { k, _ := v.Split(); return k }

// return all members values
func (v SetCol) Data() VecCol { _, d := v.Split(); return d }

func (v SetCol) Len() int { return v().Len() }

func (v SetCol) Empty() bool {
	for _, pair := range v.Pairs() {
		if !pair.Empty() {
			return false
		}
	}
	return true
}

func (v SetCol) GetVal(key Callable) (Callable, bool) {
	var m = v()
	if value, ok := m.Get(key); ok {
		return NewNative(value), ok
	}
	return NewNone(), false
}

func (v SetCol) SetVal(key, value Callable) (Associative, bool) {
	var m = v()
	return SetCol(func(pairs ...Paired) d.Mapped { return m.Set(key, value) }), true
}

func (v SetCol) Slice() []Callable {
	var pairs = []Callable{}
	for _, pair := range v.Pairs() {
		pairs = append(pairs, pair)
	}
	return pairs
}

// call method performs a value lookup
func (v SetCol) Call(args ...Callable) Callable {
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
func (v SetCol) Eval(args ...d.Native) d.Native {
	return d.NewNil()
}

func (v SetCol) TypeFnc() TyFnc { return Set }
func (v SetCol) SubType() TyFnc { return Pair }
func (v SetCol) TypeName() string {
	if v.Len() > 0 {
		return "{" + v.Pairs()[0].TypeName() + ":: " + v.Pairs()[0].TypeName() + "}"
	}
	return "{}"
}

func (v SetCol) TypeNat() d.TyNat { return d.Map | d.Function }

func (v SetCol) KeyType() TyFnc {
	if v.Len() > 0 {
		return v.Pairs()[0].Left().TypeFnc()
	}
	return None
}

func (v SetCol) KeyNatType() d.TyNat {
	if v.Len() > 0 {
		return v.Pairs()[0].Left().TypeNat()
	}
	return d.Nil
}

func (v SetCol) ValType() TyFnc {
	if v.Len() > 0 {
		return v.Pairs()[0].Right().TypeFnc()
	}
	return None
}

func (v SetCol) ValNatType() d.TyNat {
	if v.Len() > 0 {
		return v.Pairs()[0].Right().TypeNat()
	}
	return d.Nil
}

func (v SetCol) Consume() (Callable, Consumeable) {
	return v.Head(), v.Tail()
}

func (v SetCol) Head() Callable {
	if v.Len() > 0 {
		var vec = NewPairVectorFromPairs(
			v.Pairs()...,
		)
		vec.Sort(v.KeyNatType())
		return vec()[0]
	}
	return nil
}

func (v SetCol) Tail() Consumeable {
	if v.Len() > 1 {
		var vec = NewPairVectorFromPairs(
			v.Pairs()...,
		)
		vec.Sort(v.KeyNatType())
		return NewPairVec(vec()[:1]...)
	}
	return nil
}
