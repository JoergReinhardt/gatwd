package data

import (
	"sort"
	"strings"
)

type Chain []Data

func NewChain(val ...Data) Chain {
	l := make([]Data, 0, len(val))
	l = append(l, val...)
	return l
}
func ChainContainedTypes(c []Data) BitFlag {
	var flag = BitFlag(0)
	for _, d := range c {
		if FlagMatch(d.Flag(), Vector.Flag()) {
			ChainContainedTypes(d.(Chain))
			continue
		}
		flag = flag | d.Flag()
	}
	return flag
}
func (c Chain) Flag() BitFlag           { return Vector.Flag() }
func (c Chain) ContainedTypes() BitFlag { return ChainContainedTypes(c.Slice()) }
func (c Chain) Eval() Data              { return c }
func (c Chain) Null() Chain             { return []Data{} }

// SLICE ->
func (v Chain) Slice() []Data { return v }
func (v Chain) Len() int      { return len([]Data(v)) }

// COLLECTION
func (s Chain) Empty() bool            { return ChainEmpty(s) }
func (s Chain) Head() (h Data)         { return s[0] }
func (s Chain) Tail() (c Consumeable)  { return s[:1] }
func (s Chain) Shift() (c Consumeable) { return s[:1] }

func ChainClear(s Chain) {
	if len(s) > 0 {
		for _, v := range s {
			if d, ok := v.(Destructable); ok {
				d.Clear()
			}
		}
	}
	s = nil
}
func ElemEmpty(d Data) bool {
	// not flagged nil, not a composition either...
	if !FlagMatch(d.Flag(), (Nil.Flag() | Vector.Flag())) {
		if d != nil { // not a nil pointer...
			// --> not empty
			return false
		}
	}
	// since it's a composition, inspect...
	if FlagMatch(d.Flag(), Vector.Flag()) {
		// slice --> call sliceEmpty
		if sl, ok := d.(Chain); ok {
			return ChainEmpty(sl)
		}
		// other sort of collection...
		if col, ok := d.(Collected); ok {
			// --> call it's empty method
			return col.Empty()
		}
	}
	// no idea, what this is, so better call it empty
	return true
}
func ChainEmpty(s Chain) bool {
	if len(s) == 0 { // empty, as in no element...
		return true
	}
	if len(s) > 0 { // empty as in contains empty elements exclusively...
		for _, elem := range ChainSlice(s) { // return at first non empty
			if !ElemEmpty(elem) {
				return false
			}
		}
	} // --> all contained elements are empty
	return true
}

///// CONVERT TO SLICE OF NATIVES ////////
func ChainToNativeSlice(c Chain) NativeVec {
	f := ChainGet(c, 0).Flag()
	if ChainAll(c, func(i int, c Data) bool {
		return FlagMatch(f, c.Flag())
	}) {
		return ConNativeSlice(f, c.Slice()...)
	}
	return c
}
func (c Chain) NativeSlice() []interface{} {
	var s = make([]interface{}, 0, c.Len())
	for _, d := range c.Slice() {
		s = append(s, d.(Evaluable).Eval())
	}
	return s
}

//// LIST OPERATIONS ///////
func ChainFoldL(
	c Chain,
	fn func(i int, data Data, accu Data) Data,
	init Data,
) Data {
	var accu = init
	for i, d := range c.Slice() {
		accu = fn(i, d, accu)
	}
	return accu
}
func ChainMap(c Chain, fn func(i int, d Data) Data) Chain {
	var ch = make([]Data, 0, c.Len())
	for i, d := range c.Slice() {
		ch = append(ch, fn(i, d))
	}
	return ch
}
func ChainFilter(c Chain, fn func(i int, d Data) bool) Chain {
	var ch = []Data{}
	for i, d := range c.Slice() {
		if fn(i, d) {
			ch = append(ch, d)
		}
	}
	return ch
}
func ChainAny(c Chain, fn func(i int, d Data) bool) bool {
	var answ = false
	for i, d := range c.Slice() {
		if fn(i, d) {
			return true
		}
	}
	return answ
}
func ChainAll(c Chain, fn func(i int, d Data) bool) bool {
	var answ = true
	for i, d := range c.Slice() {
		if !fn(i, d) {
			return false
		}
	}
	return answ
}
func ChainReverse(c Chain) Chain {
	var ch = make([]Data, 0, c.Len())
	for i := c.Len() - 1; i > 0; i-- {
		ch = append(ch, ChainGet(c, i))
	}
	return ch
}

// ACCESSABLE SLICE
func ChainGet(s Chain, i int) Data { return s[i] }

// MUTABLE SLICE
func ChainSet(s Chain, i int, v Data) Chain { s[i] = v; return s }

// reversed index to access stacks and tuples, since their order is reversed
// for improved performance
func (c Chain) IdxRev(i int) int { return c.Len() - 1 - i }

// reversed Get method to access elements on stacks and tuples, since their
// order is reversed for improved performance
func ChainGetRev(s Chain, i int) Data { return s[s.IdxRev(i)] }

// reversed Get method to mutate elements on stacks and tuples, since their
// order is reversed for improved performance
func ChainSetRev(s Chain, i int, v Data) Chain { s[s.IdxRev(i)] = v; return s }

// ITERATOR
func ChainNext(s Chain) (v Data, i Chain) {
	if len(s) > 0 {
		if len(s) > 1 {
			return s[0], s[1:]
		}
		return s[0], Chain([]Data{NilVal{}})
	}
	return NilVal{}, Chain([]Data{NilVal{}})
}

type Iter func() (Data, Iter)

func ConIter(c Chain) Iter {
	data, chain := ChainNext(c)
	return func() (Data, Iter) {
		return data, ConIter(chain)
	}
}

// BOOTOM & TOP
func ChainFirst(s Chain) Data {
	if s.Len() > 0 {
		return s[0]
	}
	return nil
}
func ChainLast(s Chain) Data {
	if s.Len() > 0 {
		return s[s.Len()-1]
	}
	return nil
}

// LIFO QUEUE
func ChainPut(s Chain, v Data) Chain {
	return append(s, v)
}
func ChainAppend(s Chain, v ...Data) Chain {
	return append(s, v...)
}
func ChainPull(s Chain) (Data, Chain) {
	if len(s) > 0 {
		return s[0], s[1:]
	}
	return nil, nil
}

// FIFO STACK
func ChainAdd(s Chain, v ...Data) Chain {
	return append(v, s...)
}
func ChainPush(s Chain, v Data) Chain {
	//return append([]Data{v}, s...)
	return ChainPut(s, v)
}
func ChainPop(s Chain) (Data, Chain) {
	if ChainLen(s) > 0 {
		//	return s[0], s[1:]
		return s[ChainLen(s)-1], s[:ChainLen(s)-1]
	}
	return nil, nil
}

// TUPLE
func ChainHead(s Chain) (h Data)     { return s[0] }
func ChainTail(s Chain) (c []Data)   { return s[:1] }
func ChainCon(s Chain, v Data) Chain { return ChainPush(s, v) }
func ChainDecap(s Chain) (h Data, t Chain) {
	if !ChainEmpty(s) {
		return ChainPop(s)
	}
	return nil, nil
}

// SLICE
func ChainSlice(s Chain) []Data { return []Data(s) }
func ChainLen(s Chain) int      { return len(s) }
func ChainSplit(s Chain, i int) (Chain, Chain) {
	h, t := s[:i], s[i:]
	return h, t
}
func ChainCut(s Chain, i, j int) Chain {
	copy(s[i:], s[j:])
	// to prevent a possib. mem leak
	for k, n := len(s)-j+i, len(s); k < n; k++ {
		s[k] = nil
	}
	return s[:len(s)-j+i]
}
func ChainDelete(s Chain, i int) Chain {
	copy(s[i:], s[i+1:])
	s[len(s)-1] = nil
	return s[:len(s)-1]
}
func ChainInsert(s Chain, i int, v Data) Chain {
	s = append(s, NilVal{})
	copy(s[i+1:], s[i:])
	s[i] = v
	return s
}
func ChainInsertVector(s Chain, i int, v ...Data) Chain {
	return append(s[:i], append(v, s[i:]...)...)
}
func ChainAttrType(s Chain) BitFlag { return Int.Flag() }

func (c Chain) Swap(i, j int) { c = ChainSwap(c, i, j) }
func ChainSwap(c Chain, i, j int) Chain {
	c[i], c[j] = c[j], c[i]
	return c
}
func newChainLessFnc(c Chain, compT Type) func(i, j int) bool {
	chain := c
	var fn func(i, j int) bool
	f := compT.Flag()
	switch {
	case FlagMatch(f, Symbolic.Flag()):
		fn = func(i, j int) bool {
			if strings.Compare(
				string(chain[i].String()),
				string(chain[j].String()),
			) <= 0 {
				return true
			}
			return false
		}
	case FlagMatch(f, Flag.Flag()):
		fn = func(i, j int) bool {
			if chain[i].(Type).Flag() <
				chain[j].(Type).Flag() {
				return true
			}
			return false
		}
	case FlagMatch(f, Unsigned.Flag()):
		fn = func(i, j int) bool {
			if uint(chain[i].(UnsignedVal).Uint()) <
				uint(chain[j].(UnsignedVal).Uint()) {
				return true
			}
			return false
		}
	case FlagMatch(f, Integer.Flag()):
		fn = func(i, j int) bool {
			if int(chain[i].(IntegerVal).Int()) <
				int(chain[j].(IntegerVal).Int()) {
				return true
			}
			return false
		}
	}
	return fn
}
func ChainSort(c Chain, compT Type) Chain {
	sort.Slice(c, newChainLessFnc(c, compT))
	return c
}
func (c Chain) Sort(compT Type) {
	c = ChainSort(c, compT)
}

func newChainSearchFnc(c Chain, comp Data) func(i int) bool {
	var fn func(i int) bool
	f := comp.Flag()
	switch {
	case FlagMatch(f, Symbolic.Flag()):
		fn = func(i int) bool {
			return strings.Compare(c[i].String(),
				comp.String()) >= 0
		}
	case FlagMatch(f, Flag.Flag()):
		fn = func(i int) bool {
			return c[i].Flag() >=
				comp.Flag()
		}
	case FlagMatch(f, Unsigned.Flag()):
		fn = func(i int) bool {
			return uint(c[i].(UnsignedVal).Uint()) >=
				uint(comp.(UnsignedVal).Uint())
		}
	case FlagMatch(f, Integer.Flag()):
		fn = func(i int) bool {
			return int(c[i].(IntegerVal).Int()) >=
				int(comp.(IntegerVal).Int())
		}
	}
	return fn
}
func ChainSearch(c Chain, comp Data) Data {
	idx := sort.Search(c.Len(), newChainSearchFnc(c, comp))
	var dat = ChainGet(c, idx)
	return dat
}
func ChainSearchRange(c Chain, comp Data) []Data {
	var idx = sort.Search(c.Len(), newChainSearchFnc(c, comp))
	var dat = []Data{}
	for ChainGet(c, idx).Flag().Match(comp.Flag()) {
		dat = append(dat, ChainGet(c, idx))
	}
	return dat
}
func (c Chain) Search(comp Data) Data { return ChainSearch(c, comp) }
