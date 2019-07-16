package data

import (
	"sort"
	"strings"
)

// create slice from typed native instances
func NewSlice(args ...Native) DataSlice {
	return DataSlice(args)
}

// returns the OR concatenated type flags of a given slice of native instances as bit-flag
func sliceContainsTypes(c []Native) BitFlag {
	var flag BitFlag
	for _, d := range c {
		if FlagMatch(d.Type().Flag(), Slice.Type().Flag()) {
			sliceContainsTypes(d.(DataSlice))
			continue
		}
		flag = flag | d.Type().Flag()
	}
	return flag
}

// returns type flag by OR concatenating the Slice type to the concatenated
// type flags of it's members
func (c DataSlice) Type() TyNat             { return Slice }
func (c DataSlice) ElemType() Typed         { return TyNat(sliceContainsTypes(c.Slice())) }
func (c DataSlice) ContainedTypes() BitFlag { return sliceContainsTypes(c.Slice()) }
func (c DataSlice) Append(n ...Native)      { SliceAppend(c, n...) }
func (c DataSlice) Null() Native            { return NewSlice([]Native{}...) }
func (c DataSlice) Copy() Native {
	// allocate new instance of slice of natives
	var ds = DataSlice{}
	// range over slice elements
	for _, dat := range c {
		// append deep-copy of every element to the freshly allocated slice
		ds = append(ds, dat.(Reproduceable).Copy())
	}
	// return copyed slice
	return ds
}

// SLICE ->
func (v DataSlice) Slice() []Native          { return v }
func (v DataSlice) GetInt(i int) Native      { return v[i] }
func (v DataSlice) Get(i Native) Native      { return v[i.(IntVal).GoInt()] }
func (v DataSlice) Range(s, e int) Sliceable { return NewSlice(v[s:e]) }
func (v DataSlice) SetInt(i int, d Native)   { v[i] = d }
func (v DataSlice) Set(i Native, d Native)   { v[i.(IntVal)] = d }
func (v DataSlice) Len() int                 { return len([]Native(v)) }

// COLLECTION
func (s DataSlice) Empty() bool { return SliceEmpty(s) }

// yields first element
func (s DataSlice) Head() (h Native) {
	if len(s) > 0 {
		return s[0]
	}
	return NilVal{}
}

// yields last element
func (s DataSlice) Bottom() (h Native) {
	if len(s) > 0 {
		return s[len(s)-1]
	}
	return NilVal{}
}

// yields all elements except the first
func (s DataSlice) Tail() (c DataSlice) {
	if len(s) > 1 {
		return s[:1]
	}
	// return empty slice, if there is only a single element or less
	return NewSlice(NilVal{})
}

func (s DataSlice) Shift() (head Native, tail DataSlice) {
	if s.Len() > 0 {
		head = s[0]
		if s.Len() > 1 {
			tail = DataSlice(s[1:])
			return head, tail
		}
		return head, NewSlice()
	}
	return NewNil(), NewSlice()
}

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

func ElemEmpty(d Native) bool {
	// not flagged nil, not a composition either...
	if !FlagMatch(d.Type().Flag(), (Nil.Type().Flag() | Slice.Type().Flag())) {
		if d != nil { // not a nil pointer...
			// --> not empty
			return false
		}
	}
	// since it's a composition, inspect...
	if FlagMatch(d.Type().Flag(), Slice.Type().Flag()) {
		// slice --> call sliceEmpty
		if sl, ok := d.(DataSlice); ok {
			return SliceEmpty(sl)
		}
		// other sort of collection...
		if col, ok := d.(Composed); ok {
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
	// allocate nil flag
	var flag = Nil
	// replace flag with type flag of first elment, if there are elements
	if len(c) > 0 {
		flag = c[0].Type()
	}
	// check if all elements flags match the first elements flag
	if SliceAll(c, func(i int, c Native) bool {
		return c.Type().Match(flag)
	}) {
		// if all elements yield the same type, convert to unboxed
		// slice of natives
		return NewUnboxed(flag, c.Slice()...)
	}
	// return unconverted slice, since elment types are impure
	return c
}

//// LIST OPERATIONS ///////
func SliceFoldL(
	c DataSlice, fn func(
		i int,
		data Native,
		accu Native,
	) Native,
	init Native) Native {
	var accu = init
	for i, d := range c.Slice() {
		accu = fn(i, d, accu)
	}
	return accu
}

func SliceMap(c DataSlice, fn func(i int, d Native) Native) DataSlice {
	var ch = make([]Native, 0, c.Len())
	for i, d := range c.Slice() {
		ch = append(ch, fn(i, d))
	}
	return ch
}

func SliceFilter(c DataSlice, fn func(i int, d Native) bool) DataSlice {
	var ch = []Native{}
	for i, d := range c.Slice() {
		if fn(i, d) {
			ch = append(ch, d)
		}
	}
	return ch
}

func SliceAny(c DataSlice, fn func(i int, d Native) bool) bool {
	var answ = false
	for i, d := range c.Slice() {
		if fn(i, d) {
			return true
		}
	}
	return answ
}

func SliceAll(c DataSlice, fn func(i int, d Native) bool) bool {
	for i, d := range c.Slice() {
		if !fn(i, d) {
			return false
		}
	}
	return true
}

func SliceReverse(c DataSlice) DataSlice {
	var ch = make([]Native, 0, c.Len())
	for i := c.Len() - 1; i > 0; i-- {
		ch = append(ch, SliceGet(c, i))
	}
	return ch
}

// ACCESSABLE SLICE
func SliceGet(s DataSlice, i int) Native { return s[i] }

// MUTABLE SLICE
func SliceSet(s DataSlice, i int, v Native) DataSlice { s[i] = v; return s }

// reversed index to access stacks and tuples, since their order is reversed
// for improved performance
func (c DataSlice) IdxRev(i int) int { return c.Len() - 1 - i }

// reversed Get method to access elements on stacks and tuples, since their
// order is reversed for improved performance
func SliceGetRev(s DataSlice, i int) Native { return s[s.IdxRev(i)] }

// reversed Get method to mutate elements on stacks and tuples, since their
// order is reversed for improved performance
func SliceSetRev(s DataSlice, i int, v Native) DataSlice { s[s.IdxRev(i)] = v; return s }

// ITERATOR
func SliceNext(s DataSlice) (v Native, i DataSlice) {
	if len(s) > 0 {
		if len(s) > 1 {
			return s[0], s[1:]
		}
		return s[0], DataSlice([]Native{NilVal{}})
	}
	return NilVal{}, DataSlice([]Native{NilVal{}})
}

type Iter func() (Native, Iter)

func ConIter(c DataSlice) Iter {
	data, chain := SliceNext(c)
	return func() (Native, Iter) {
		return data, ConIter(chain)
	}
}

// BOOTOM & TOP
func SliceFirst(s DataSlice) Native {
	if s.Len() > 0 {
		return s[0]
	}
	return nil
}
func SliceLast(s DataSlice) Native {
	if s.Len() > 0 {
		return s[s.Len()-1]
	}
	return nil
}

// LIFO QUEUE
func SlicePut(s DataSlice, v Native) DataSlice {
	return append(s, v)
}

func SliceAppend(s DataSlice, v ...Native) DataSlice {
	return append(s, v...)
}

func SlicePull(s DataSlice) (Native, DataSlice) {
	if len(s) > 0 {
		return s[0], s[1:]
	}
	return nil, nil
}

// FIFO STACK
func SliceAdd(s DataSlice, v ...Native) DataSlice {
	return append(v, s...)
}

func SlicePush(s DataSlice, v Native) DataSlice {
	//return append([]Data{v}, s...)
	return SlicePut(s, v)
}

func SlicePop(s DataSlice) (Native, DataSlice) {
	if SliceLen(s) > 0 {
		//	return s[0], s[1:]
		return s[SliceLen(s)-1], s[:SliceLen(s)-1]
	}
	return nil, nil
}

// TUPLE
func SliceHead(s DataSlice) (h Native) { return s[0] }

func SliceTail(s DataSlice) (c []Native) { return s[:1] }

func SliceCon(s DataSlice, v Native) DataSlice { return SlicePush(s, v) }

func SliceDeCap(s DataSlice) (h Native, t DataSlice) {
	if !SliceEmpty(s) {
		return SlicePop(s)
	}
	return nil, nil
}

// SLICE
func SliceSlice(s DataSlice) []Native { return []Native(s) }

func SliceLen(s DataSlice) int { return len(s) }

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

func SliceInsert(s DataSlice, i int, v Native) DataSlice {
	s = append(s, NilVal{})
	copy(s[i+1:], s[i:])
	s[i] = v
	return s
}

func SliceInsertVector(s DataSlice, i int, v ...Native) DataSlice {
	return append(s[:i], append(v, s[i:]...)...)
}

func SliceAttrType(s DataSlice) BitFlag { return Int.Type().Flag() }

func (c DataSlice) Swap(i, j int) { c = SliceSwap(c, i, j) }

func SliceSwap(c DataSlice, i, j int) DataSlice {
	c[i], c[j] = c[j], c[i]
	return c
}

func newSliceLess(c DataSlice, compT TyNat) func(i, j int) bool {
	chain := c
	var fn func(i, j int) bool
	f := compT.Type().Flag()
	switch {
	case FlagMatch(f, Letters.Type().Flag()):
		fn = func(i, j int) bool {
			if strings.Compare(
				string(chain[i].String()),
				string(chain[j].String()),
			) <= 0 {
				return true
			}
			return false
		}
	case FlagMatch(f, Type.Type().Flag()):
		fn = func(i, j int) bool {
			if chain[i].(TyNat).Type() <
				chain[j].(TyNat).Type() {
				return true
			}
			return false
		}
	case FlagMatch(f, Naturals.Type().Flag()):
		fn = func(i, j int) bool {
			if uint(chain[i].(Natural).GoUint()) <
				uint(chain[j].(Natural).GoUint()) {
				return true
			}
			return false
		}
	case FlagMatch(f, Integers.Type().Flag()):
		fn = func(i, j int) bool {
			if int(chain[i].(Integer).Idx()) <
				int(chain[j].(Integer).Idx()) {
				return true
			}
			return false
		}
	}
	return fn
}

func SliceSort(c DataSlice, compT TyNat) DataSlice {
	sort.Slice(c, newSliceLess(c, compT))
	return c
}

func (c DataSlice) Sort(compT TyNat) {
	c = SliceSort(c, compT)
}

func newSliceSearchFnc(c DataSlice, comp Native) func(i int) bool {
	var fn func(i int) bool
	f := comp.Type().Flag()
	switch {
	case FlagMatch(f, Letters.Type().Flag()):
		fn = func(i int) bool {
			return strings.Compare(c[i].String(),
				comp.String()) >= 0
		}
	case FlagMatch(f, Type.Type().Flag()):
		fn = func(i int) bool {
			return c[i].Type() >=
				comp.Type()
		}
	case FlagMatch(f, Naturals.Type().Flag()):
		fn = func(i int) bool {
			return uint(c[i].(Natural).GoUint()) >=
				uint(comp.(Natural).GoUint())
		}
	case FlagMatch(f, Integers.Type().Flag()):
		fn = func(i int) bool {
			return int(c[i].(Integer).Idx()) >=
				int(comp.(Integer).Idx())
		}
	}
	return fn
}

func SliceSearch(c DataSlice, comp Native) Native {
	idx := sort.Search(c.Len(), newSliceSearchFnc(c, comp))
	var dat = SliceGet(c, idx)
	return dat
}

func SliceSearchRange(c DataSlice, comp Native) []Native {
	var idx = sort.Search(c.Len(), newSliceSearchFnc(c, comp))
	var dat = []Native{}
	for SliceGet(c, idx).Type().Flag().Match(comp.Type().Flag()) {
		dat = append(dat, SliceGet(c, idx))
	}
	return dat
}

func (c DataSlice) Search(comp Native) Native { return SliceSearch(c, comp) }
