package data

import (
	"sort"
	"strings"
)

func NewSlice(val ...Data) DataSlice {
	l := make([]Data, 0, len(val))
	l = append(l, val...)
	return l
}
func SliceContainedTypes(c []Data) BitFlag {
	var flag = BitFlag(0)
	for _, d := range c {
		if FlagMatch(d.Flag(), Vector.Flag()) {
			SliceContainedTypes(d.(DataSlice))
			continue
		}
		flag = flag | d.Flag()
	}
	return flag
}
func (c DataSlice) Flag() BitFlag           { return Vector.Flag() }
func (c DataSlice) ContainedTypes() BitFlag { return SliceContainedTypes(c.Slice()) }
func (c DataSlice) Eval() Data              { return c }
func (c DataSlice) Null() DataSlice         { return []Data{} }
func (c DataSlice) Copy() Data {
	var ns = DataSlice{}
	for _, dat := range c {
		ns = append(ns, dat.(Reproduceable).Copy())
	}
	return ns
}

// SLICE ->
func (v DataSlice) Slice() []Data        { return v }
func (v DataSlice) GetInt(i int) Data    { return v[i] }
func (v DataSlice) Get(i Data) Data      { return v[i.(IntVal).Int()] }
func (v DataSlice) SetInt(i int, d Data) { v[i] = d }
func (v DataSlice) Set(i Data, d Data)   { v[i.(IntVal)] = d }
func (v DataSlice) Len() int             { return len([]Data(v)) }

// COLLECTION
func (s DataSlice) Empty() bool            { return SliceEmpty(s) }
func (s DataSlice) Head() (h Data)         { return s[0] }
func (s DataSlice) Tail() (c Consumeable)  { return s[:1] }
func (s DataSlice) Shift() (c Consumeable) { return s[:1] }

func SliceClear(s DataSlice) {
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
		if sl, ok := d.(DataSlice); ok {
			return SliceEmpty(sl)
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
func SliceEmpty(s DataSlice) bool {
	if len(s) == 0 { // empty, as in no element...
		return true
	}
	if len(s) > 0 { // empty as in contains empty elements exclusively...
		for _, elem := range SliceSlice(s) { // return at first non empty
			if !ElemEmpty(elem) {
				return false
			}
		}
	} // --> all contained elements are empty
	return true
}

///// CONVERT TO SLICE OF NATIVES ////////
func SliceToNatives(c DataSlice) Sliceable {
	f := SliceGet(c, 0).Flag()
	if SliceAll(c, func(i int, c Data) bool {
		return FlagMatch(f, c.Flag())
	}) {
		return ConNativeSlice(f, c.Slice()...)
	}
	return c
}
func (c DataSlice) NativeSlice() []interface{} {
	var s = make([]interface{}, 0, c.Len())
	for _, d := range c.Slice() {
		s = append(s, d.(Ident).Ident())
	}
	return s
}

//// LIST OPERATIONS ///////
func SliceFoldL(
	c DataSlice, fn func(i int, data Data, accu Data) Data,
	init Data,
) Data {
	var accu = init
	for i, d := range c.Slice() {
		accu = fn(i, d, accu)
	}
	return accu
}
func SliceMap(c DataSlice, fn func(i int, d Data) Data) DataSlice {
	var ch = make([]Data, 0, c.Len())
	for i, d := range c.Slice() {
		ch = append(ch, fn(i, d))
	}
	return ch
}
func SliceFilter(c DataSlice, fn func(i int, d Data) bool) DataSlice {
	var ch = []Data{}
	for i, d := range c.Slice() {
		if fn(i, d) {
			ch = append(ch, d)
		}
	}
	return ch
}
func SliceAny(c DataSlice, fn func(i int, d Data) bool) bool {
	var answ = false
	for i, d := range c.Slice() {
		if fn(i, d) {
			return true
		}
	}
	return answ
}
func SliceAll(c DataSlice, fn func(i int, d Data) bool) bool {
	var answ = true
	for i, d := range c.Slice() {
		if !fn(i, d) {
			return false
		}
	}
	return answ
}
func SliceReverse(c DataSlice) DataSlice {
	var ch = make([]Data, 0, c.Len())
	for i := c.Len() - 1; i > 0; i-- {
		ch = append(ch, SliceGet(c, i))
	}
	return ch
}

// ACCESSABLE SLICE
func SliceGet(s DataSlice, i int) Data { return s[i] }

// MUTABLE SLICE
func SliceSet(s DataSlice, i int, v Data) DataSlice { s[i] = v; return s }

// reversed index to access stacks and tuples, since their order is reversed
// for improved performance
func (c DataSlice) IdxRev(i int) int { return c.Len() - 1 - i }

// reversed Get method to access elements on stacks and tuples, since their
// order is reversed for improved performance
func SliceGetRev(s DataSlice, i int) Data { return s[s.IdxRev(i)] }

// reversed Get method to mutate elements on stacks and tuples, since their
// order is reversed for improved performance
func SliceSetRev(s DataSlice, i int, v Data) DataSlice { s[s.IdxRev(i)] = v; return s }

// ITERATOR
func SliceNext(s DataSlice) (v Data, i DataSlice) {
	if len(s) > 0 {
		if len(s) > 1 {
			return s[0], s[1:]
		}
		return s[0], DataSlice([]Data{NilVal{}})
	}
	return NilVal{}, DataSlice([]Data{NilVal{}})
}

type Iter func() (Data, Iter)

func ConIter(c DataSlice) Iter {
	data, chain := SliceNext(c)
	return func() (Data, Iter) {
		return data, ConIter(chain)
	}
}

// BOOTOM & TOP
func SliceFirst(s DataSlice) Data {
	if s.Len() > 0 {
		return s[0]
	}
	return nil
}
func SliceLast(s DataSlice) Data {
	if s.Len() > 0 {
		return s[s.Len()-1]
	}
	return nil
}

// LIFO QUEUE
func SlicePut(s DataSlice, v Data) DataSlice {
	return append(s, v)
}
func SliceAppend(s DataSlice, v ...Data) DataSlice {
	return append(s, v...)
}
func SlicePull(s DataSlice) (Data, DataSlice) {
	if len(s) > 0 {
		return s[0], s[1:]
	}
	return nil, nil
}

// FIFO STACK
func SliceAdd(s DataSlice, v ...Data) DataSlice {
	return append(v, s...)
}
func SlicePush(s DataSlice, v Data) DataSlice {
	//return append([]Data{v}, s...)
	return SlicePut(s, v)
}
func SlicePop(s DataSlice) (Data, DataSlice) {
	if SliceLen(s) > 0 {
		//	return s[0], s[1:]
		return s[SliceLen(s)-1], s[:SliceLen(s)-1]
	}
	return nil, nil
}

// TUPLE
func SliceHead(s DataSlice) (h Data)         { return s[0] }
func SliceTail(s DataSlice) (c []Data)       { return s[:1] }
func SliceCon(s DataSlice, v Data) DataSlice { return SlicePush(s, v) }
func SliceDeCap(s DataSlice) (h Data, t DataSlice) {
	if !SliceEmpty(s) {
		return SlicePop(s)
	}
	return nil, nil
}

// SLICE
func SliceSlice(s DataSlice) []Data { return []Data(s) }
func SliceLen(s DataSlice) int      { return len(s) }
func SliceSplit(s DataSlice, i int) (DataSlice, DataSlice) {
	h, t := s[:i], s[i:]
	return h, t
}
func SliceCut(s DataSlice, i, j int) DataSlice {
	copy(s[i:], s[j:])
	// to prevent a possib. mem leak
	for k, n := len(s)-j+i, len(s); k < n; k++ {
		s[k] = nil
	}
	return s[:len(s)-j+i]
}
func SliceDelete(s DataSlice, i int) DataSlice {
	copy(s[i:], s[i+1:])
	s[len(s)-1] = nil
	return s[:len(s)-1]
}
func SliceInsert(s DataSlice, i int, v Data) DataSlice {
	s = append(s, NilVal{})
	copy(s[i+1:], s[i:])
	s[i] = v
	return s
}
func SliceInsertVector(s DataSlice, i int, v ...Data) DataSlice {
	return append(s[:i], append(v, s[i:]...)...)
}
func SliceAttrType(s DataSlice) BitFlag { return Int.Flag() }

func (c DataSlice) Swap(i, j int) { c = SliceSwap(c, i, j) }
func SliceSwap(c DataSlice, i, j int) DataSlice {
	c[i], c[j] = c[j], c[i]
	return c
}
func newSliceLess(c DataSlice, compT Type) func(i, j int) bool {
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
func SliceSort(c DataSlice, compT Type) DataSlice {
	sort.Slice(c, newSliceLess(c, compT))
	return c
}
func (c DataSlice) Sort(compT Type) {
	c = SliceSort(c, compT)
}

func newSliceSearchFnc(c DataSlice, comp Data) func(i int) bool {
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
func SliceSearch(c DataSlice, comp Data) Data {
	idx := sort.Search(c.Len(), newSliceSearchFnc(c, comp))
	var dat = SliceGet(c, idx)
	return dat
}
func SliceSearchRange(c DataSlice, comp Data) []Data {
	var idx = sort.Search(c.Len(), newSliceSearchFnc(c, comp))
	var dat = []Data{}
	for SliceGet(c, idx).Flag().Match(comp.Flag()) {
		dat = append(dat, SliceGet(c, idx))
	}
	return dat
}
func (c DataSlice) Search(comp Data) Data { return SliceSearch(c, comp) }
