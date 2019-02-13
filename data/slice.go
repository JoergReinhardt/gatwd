package data

import (
	"sort"
	"strings"
)

func NewSlice(val ...Primary) DataSlice {
	l := make([]Primary, 0, len(val))
	l = append(l, val...)
	return l
}
func SliceContainedTypes(c []Primary) BitFlag {
	var flag = BitFlag(0)
	for _, d := range c {
		if FlagMatch(d.TypePrime().Flag(), Vector.TypePrime().Flag()) {
			SliceContainedTypes(d.(DataSlice))
			continue
		}
		flag = flag | d.TypePrime().Flag()
	}
	return flag
}
func (c DataSlice) TypePrime() TyPrime      { return Vector.TypePrime() }
func (c DataSlice) ContainedTypes() BitFlag { return SliceContainedTypes(c.Slice()) }
func (c DataSlice) Eval(p ...Primary) Primary {
	if len(p) > 0 {
		if len(c) > 0 {
			return SliceAppend(c, p...)
		}
		NewSlice(p...)
	}
	return c
}
func (c DataSlice) Null() DataSlice { return []Primary{} }
func (c DataSlice) Copy() Primary {
	var ns = DataSlice{}
	for _, dat := range c {
		ns = append(ns, dat.(Reproduceable).Copy())
	}
	return ns
}

// SLICE ->
func (v DataSlice) Slice() []Primary         { return v }
func (v DataSlice) GetInt(i int) Primary     { return v[i] }
func (v DataSlice) Get(i Primary) Primary    { return v[i.(IntVal).Int()] }
func (v DataSlice) SetInt(i int, d Primary)  { v[i] = d }
func (v DataSlice) Set(i Primary, d Primary) { v[i.(IntVal)] = d }
func (v DataSlice) Len() int                 { return len([]Primary(v)) }

// COLLECTION
func (s DataSlice) Empty() bool            { return SliceEmpty(s) }
func (s DataSlice) Head() (h Primary)      { return s[0] }
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
func ElemEmpty(d Primary) bool {
	// not flagged nil, not a composition either...
	if !FlagMatch(d.TypePrime().Flag(), (Nil.TypePrime().Flag() | Vector.TypePrime().Flag())) {
		if d != nil { // not a nil pointer...
			// --> not empty
			return false
		}
	}
	// since it's a composition, inspect...
	if FlagMatch(d.TypePrime().Flag(), Vector.TypePrime().Flag()) {
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
	f := SliceGet(c, 0).TypePrime().Flag()
	if SliceAll(c, func(i int, c Primary) bool {
		return FlagMatch(f, c.TypePrime().Flag())
	}) {
		return ConNativeSlice(f, c.Slice()...)
	}
	return c
}
func (c DataSlice) NativeSlice() []interface{} {
	var s = make([]interface{}, 0, c.Len())
	for _, d := range c.Slice() {
		s = append(s, d.(Identity).Ident())
	}
	return s
}

//// LIST OPERATIONS ///////
func SliceFoldL(
	c DataSlice, fn func(i int, data Primary, accu Primary) Primary, init Primary) Primary {
	var accu = init
	for i, d := range c.Slice() {
		accu = fn(i, d, accu)
	}
	return accu
}
func SliceMap(c DataSlice, fn func(i int, d Primary) Primary) DataSlice {
	var ch = make([]Primary, 0, c.Len())
	for i, d := range c.Slice() {
		ch = append(ch, fn(i, d))
	}
	return ch
}
func SliceFilter(c DataSlice, fn func(i int, d Primary) bool) DataSlice {
	var ch = []Primary{}
	for i, d := range c.Slice() {
		if fn(i, d) {
			ch = append(ch, d)
		}
	}
	return ch
}
func SliceAny(c DataSlice, fn func(i int, d Primary) bool) bool {
	var answ = false
	for i, d := range c.Slice() {
		if fn(i, d) {
			return true
		}
	}
	return answ
}
func SliceAll(c DataSlice, fn func(i int, d Primary) bool) bool {
	var answ = true
	for i, d := range c.Slice() {
		if !fn(i, d) {
			return false
		}
	}
	return answ
}
func SliceReverse(c DataSlice) DataSlice {
	var ch = make([]Primary, 0, c.Len())
	for i := c.Len() - 1; i > 0; i-- {
		ch = append(ch, SliceGet(c, i))
	}
	return ch
}

// ACCESSABLE SLICE
func SliceGet(s DataSlice, i int) Primary { return s[i] }

// MUTABLE SLICE
func SliceSet(s DataSlice, i int, v Primary) DataSlice { s[i] = v; return s }

// reversed index to access stacks and tuples, since their order is reversed
// for improved performance
func (c DataSlice) IdxRev(i int) int { return c.Len() - 1 - i }

// reversed Get method to access elements on stacks and tuples, since their
// order is reversed for improved performance
func SliceGetRev(s DataSlice, i int) Primary { return s[s.IdxRev(i)] }

// reversed Get method to mutate elements on stacks and tuples, since their
// order is reversed for improved performance
func SliceSetRev(s DataSlice, i int, v Primary) DataSlice { s[s.IdxRev(i)] = v; return s }

// ITERATOR
func SliceNext(s DataSlice) (v Primary, i DataSlice) {
	if len(s) > 0 {
		if len(s) > 1 {
			return s[0], s[1:]
		}
		return s[0], DataSlice([]Primary{NilVal{}})
	}
	return NilVal{}, DataSlice([]Primary{NilVal{}})
}

type Iter func() (Primary, Iter)

func ConIter(c DataSlice) Iter {
	data, chain := SliceNext(c)
	return func() (Primary, Iter) {
		return data, ConIter(chain)
	}
}

// BOOTOM & TOP
func SliceFirst(s DataSlice) Primary {
	if s.Len() > 0 {
		return s[0]
	}
	return nil
}
func SliceLast(s DataSlice) Primary {
	if s.Len() > 0 {
		return s[s.Len()-1]
	}
	return nil
}

// LIFO QUEUE
func SlicePut(s DataSlice, v Primary) DataSlice {
	return append(s, v)
}
func SliceAppend(s DataSlice, v ...Primary) DataSlice {
	return append(s, v...)
}
func SlicePull(s DataSlice) (Primary, DataSlice) {
	if len(s) > 0 {
		return s[0], s[1:]
	}
	return nil, nil
}

// FIFO STACK
func SliceAdd(s DataSlice, v ...Primary) DataSlice {
	return append(v, s...)
}
func SlicePush(s DataSlice, v Primary) DataSlice {
	//return append([]Data{v}, s...)
	return SlicePut(s, v)
}
func SlicePop(s DataSlice) (Primary, DataSlice) {
	if SliceLen(s) > 0 {
		//	return s[0], s[1:]
		return s[SliceLen(s)-1], s[:SliceLen(s)-1]
	}
	return nil, nil
}

// TUPLE
func SliceHead(s DataSlice) (h Primary)         { return s[0] }
func SliceTail(s DataSlice) (c []Primary)       { return s[:1] }
func SliceCon(s DataSlice, v Primary) DataSlice { return SlicePush(s, v) }
func SliceDeCap(s DataSlice) (h Primary, t DataSlice) {
	if !SliceEmpty(s) {
		return SlicePop(s)
	}
	return nil, nil
}

// SLICE
func SliceSlice(s DataSlice) []Primary { return []Primary(s) }
func SliceLen(s DataSlice) int         { return len(s) }
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
func SliceInsert(s DataSlice, i int, v Primary) DataSlice {
	s = append(s, NilVal{})
	copy(s[i+1:], s[i:])
	s[i] = v
	return s
}
func SliceInsertVector(s DataSlice, i int, v ...Primary) DataSlice {
	return append(s[:i], append(v, s[i:]...)...)
}
func SliceAttrType(s DataSlice) BitFlag { return Int.TypePrime().Flag() }

func (c DataSlice) Swap(i, j int) { c = SliceSwap(c, i, j) }
func SliceSwap(c DataSlice, i, j int) DataSlice {
	c[i], c[j] = c[j], c[i]
	return c
}
func newSliceLess(c DataSlice, compT TyPrime) func(i, j int) bool {
	chain := c
	var fn func(i, j int) bool
	f := compT.TypePrime().Flag()
	switch {
	case FlagMatch(f, Symbolic.TypePrime().Flag()):
		fn = func(i, j int) bool {
			if strings.Compare(
				string(chain[i].String()),
				string(chain[j].String()),
			) <= 0 {
				return true
			}
			return false
		}
	case FlagMatch(f, Flag.TypePrime().Flag()):
		fn = func(i, j int) bool {
			if chain[i].(TyPrime).TypePrime() <
				chain[j].(TyPrime).TypePrime() {
				return true
			}
			return false
		}
	case FlagMatch(f, Natural.TypePrime().Flag()):
		fn = func(i, j int) bool {
			if uint(chain[i].(NaturalVal).Uint()) <
				uint(chain[j].(NaturalVal).Uint()) {
				return true
			}
			return false
		}
	case FlagMatch(f, Integer.TypePrime().Flag()):
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
func SliceSort(c DataSlice, compT TyPrime) DataSlice {
	sort.Slice(c, newSliceLess(c, compT))
	return c
}
func (c DataSlice) Sort(compT TyPrime) {
	c = SliceSort(c, compT)
}

func newSliceSearchFnc(c DataSlice, comp Primary) func(i int) bool {
	var fn func(i int) bool
	f := comp.TypePrime().Flag()
	switch {
	case FlagMatch(f, Symbolic.TypePrime().Flag()):
		fn = func(i int) bool {
			return strings.Compare(c[i].String(),
				comp.String()) >= 0
		}
	case FlagMatch(f, Flag.TypePrime().Flag()):
		fn = func(i int) bool {
			return c[i].TypePrime() >=
				comp.TypePrime()
		}
	case FlagMatch(f, Natural.TypePrime().Flag()):
		fn = func(i int) bool {
			return uint(c[i].(NaturalVal).Uint()) >=
				uint(comp.(NaturalVal).Uint())
		}
	case FlagMatch(f, Integer.TypePrime().Flag()):
		fn = func(i int) bool {
			return int(c[i].(IntegerVal).Int()) >=
				int(comp.(IntegerVal).Int())
		}
	}
	return fn
}
func SliceSearch(c DataSlice, comp Primary) Primary {
	idx := sort.Search(c.Len(), newSliceSearchFnc(c, comp))
	var dat = SliceGet(c, idx)
	return dat
}
func SliceSearchRange(c DataSlice, comp Primary) []Primary {
	var idx = sort.Search(c.Len(), newSliceSearchFnc(c, comp))
	var dat = []Primary{}
	for SliceGet(c, idx).TypePrime().Flag().Match(comp.TypePrime().Flag()) {
		dat = append(dat, SliceGet(c, idx))
	}
	return dat
}
func (c DataSlice) Search(comp Primary) Primary { return SliceSearch(c, comp) }
