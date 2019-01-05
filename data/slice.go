package data

import (
	"math/big"
	"time"
)

type (
	Chain []Data
	Set   map[Data]Data
)

type Slicer interface {
	Dats() []Data
}

type (
	NativeVector []interface{}
	NilVec       []struct{}
	BoolVec      []bool
	IntVec       []int
	Int8Vec      []int8
	Int16Vec     []int16
	Int32Vec     []int32
	UintVec      []uint
	Uint8Vec     []uint8
	Uint16Vec    []uint16
	Uint32Vec    []uint32
	FltVec       []float64
	Flt32Vec     []float32
	ImagVec      []complex128
	Imag64Vec    []complex64
	ByteVec      []byte
	RuneVec      []rune
	BytesVec     [][]byte
	StrVec       []string
	BigIntVec    []big.Int
	BigFltVec    []big.Float
	RatioVec     []big.Rat
	TimeVec      []time.Time
	DuraVec      []time.Duration
	ErrorVec     []struct{ v error }
	FlagSet      []BitFlag
)

func (v NativeVector) Dat(i int) interface{} { return v[i] }
func (v NilVec) Dat(i int) NilVal            { return NilVal(v[i]) }
func (v BoolVec) Dat(i int) BoolVal          { return BoolVal(v[i]) }
func (v IntVec) Dat(i int) IntVal            { return IntVal(v[i]) }
func (v Int8Vec) Dat(i int) Int8Val          { return Int8Val(v[i]) }
func (v Int16Vec) Dat(i int) Int16Val        { return Int16Val(v[i]) }
func (v Int32Vec) Dat(i int) Int32Val        { return Int32Val(v[i]) }
func (v UintVec) Dat(i int) UintVal          { return UintVal(v[i]) }
func (v Uint8Vec) Dat(i int) Uint8Val        { return Uint8Val(v[i]) }
func (v Uint16Vec) Dat(i int) Uint16Val      { return Uint16Val(v[i]) }
func (v Uint32Vec) Dat(i int) Uint32Val      { return Uint32Val(v[i]) }
func (v FltVec) Dat(i int) FltVal            { return FltVal(v[i]) }
func (v Flt32Vec) Dat(i int) Flt32Val        { return Flt32Val(v[i]) }
func (v ImagVec) Dat(i int) ImagVal          { return ImagVal(v[i]) }
func (v Imag64Vec) Dat(i int) Imag64Val      { return Imag64Val(v[i]) }
func (v ByteVec) Dat(i int) ByteVal          { return ByteVal(v[i]) }
func (v RuneVec) Dat(i int) RuneVal          { return RuneVal(v[i]) }
func (v BytesVec) Dat(i int) BytesVal        { return BytesVal(v[i]) }
func (v StrVec) Dat(i int) StrVal            { return StrVal(v[i]) }
func (v BigIntVec) Dat(i int) BigIntVal      { return BigIntVal(v[i]) }
func (v BigFltVec) Dat(i int) BigFltVal      { return BigFltVal(v[i]) }
func (v RatioVec) Dat(i int) RatioVal        { return RatioVal(v[i]) }
func (v TimeVec) Dat(i int) TimeVal          { return TimeVal(v[i]) }
func (v DuraVec) Dat(i int) DuraVal          { return DuraVal(v[i]) }
func (v ErrorVec) Dat(i int) ErrorVal        { return ErrorVal(v[i]) }

func (v NativeVector) Dats(i, j int) interface{} { return v[i] }
func (v NilVec) Dats(i, j int) NilVec            { return NilVec(v[i:j]) }
func (v BoolVec) Dats(i, j int) BoolVec          { return BoolVec(v[i:j]) }
func (v IntVec) Dats(i, j int) IntVec            { return IntVec(v[i:j]) }
func (v Int8Vec) Dats(i, j int) Int8Vec          { return Int8Vec(v[i:j]) }
func (v Int16Vec) Dats(i, j int) Int16Vec        { return Int16Vec(v[i:j]) }
func (v Int32Vec) Dats(i, j int) Int32Vec        { return Int32Vec(v[i:j]) }
func (v UintVec) Dats(i, j int) UintVec          { return UintVec(v[i:j]) }
func (v Uint8Vec) Dats(i, j int) Uint8Vec        { return Uint8Vec(v[i:j]) }
func (v Uint16Vec) Dats(i, j int) Uint16Vec      { return Uint16Vec(v[i:j]) }
func (v Uint32Vec) Dats(i, j int) Uint32Vec      { return Uint32Vec(v[i:j]) }
func (v FltVec) Dats(i, j int) FltVec            { return FltVec(v[i:j]) }
func (v Flt32Vec) Dats(i, j int) Flt32Vec        { return Flt32Vec(v[i:j]) }
func (v ImagVec) Dats(i, j int) ImagVec          { return ImagVec(v[i:j]) }
func (v Imag64Vec) Dats(i, j int) Imag64Vec      { return Imag64Vec(v[i:j]) }
func (v ByteVec) Dats(i, j int) ByteVec          { return ByteVec(v[i:j]) }
func (v RuneVec) Dats(i, j int) RuneVec          { return RuneVec(v[i:j]) }
func (v BytesVec) Dats(i, j int) BytesVec        { return BytesVec(v[i:j]) }
func (v StrVec) Dats(i, j int) StrVec            { return StrVec(v[i:j]) }
func (v BigIntVec) Dats(i, j int) BigIntVec      { return BigIntVec(v[i:j]) }
func (v BigFltVec) Dats(i, j int) BigFltVec      { return BigFltVec(v[i:j]) }
func (v RatioVec) Dats(i, j int) RatioVec        { return RatioVec(v[i:j]) }
func (v TimeVec) Dats(i, j int) TimeVec          { return TimeVec(v[i:j]) }
func (v DuraVec) Dats(i, j int) DuraVec          { return DuraVec(v[i:j]) }
func (v ErrorVec) Dats(i, j int) ErrorVec        { return ErrorVec(v[i:j]) }

func (v NativeVector) nat(i int) interface{}   { return v[i] }
func (v NilVec) Nat(i int) struct{}            { return v[i] }
func (v BoolVec) Nat(i int) bool               { return v[i] }
func (v IntVec) Nat(i int) int                 { return v[i] }
func (v Int8Vec) Nat(i int) int8               { return v[i] }
func (v Int16Vec) Nat(i int) int16             { return v[i] }
func (v Int32Vec) Nat(i int) int32             { return v[i] }
func (v UintVec) Nat(i int) uint               { return v[i] }
func (v Uint8Vec) Nat(i int) uint8             { return v[i] }
func (v Uint16Vec) Nat(i int) uint16           { return v[i] }
func (v Uint32Vec) Nat(i int) uint32           { return v[i] }
func (v FltVec) Nat(i int) float64             { return v[i] }
func (v Flt32Vec) Nat(i int) float32           { return v[i] }
func (v ImagVec) Nat(i int) complex128         { return v[i] }
func (v Imag64Vec) Nat(i int) complex64        { return v[i] }
func (v ByteVec) Nat(i int) byte               { return v[i] }
func (v RuneVec) Nat(i int) rune               { return v[i] }
func (v BytesVec) Nat(i int) []byte            { return v[i] }
func (v StrVec) Nat(i int) string              { return v[i] }
func (v BigIntVec) Nat(i int) big.Int          { return v[i] }
func (v BigFltVec) Nat(i int) big.Float        { return v[i] }
func (v RatioVec) Nat(i int) big.Rat           { return v[i] }
func (v TimeVec) Nat(i int) time.Time          { return v[i] }
func (v DuraVec) Nat(i int) time.Duration      { return v[i] }
func (v ErrorVec) Nat(i int) struct{ v error } { return v[i] }
func (v FlagSet) Nat(i int) BitFlag            { return v[i] }

func (v NilVec) intf(i int) interface{}    { return v[i] }
func (v BoolVec) intf(i int) interface{}   { return v[i] }
func (v IntVec) intf(i int) interface{}    { return v[i] }
func (v Int8Vec) intf(i int) interface{}   { return v[i] }
func (v Int16Vec) intf(i int) interface{}  { return v[i] }
func (v Int32Vec) intf(i int) interface{}  { return v[i] }
func (v UintVec) intf(i int) interface{}   { return v[i] }
func (v Uint8Vec) intf(i int) interface{}  { return v[i] }
func (v Uint16Vec) intf(i int) interface{} { return v[i] }
func (v Uint32Vec) intf(i int) interface{} { return v[i] }
func (v FltVec) intf(i int) interface{}    { return v[i] }
func (v Flt32Vec) intf(i int) interface{}  { return v[i] }
func (v ImagVec) intf(i int) interface{}   { return v[i] }
func (v Imag64Vec) intf(i int) interface{} { return v[i] }
func (v ByteVec) intf(i int) interface{}   { return v[i] }
func (v RuneVec) intf(i int) interface{}   { return v[i] }
func (v BytesVec) intf(i int) interface{}  { return v[i] }
func (v StrVec) intf(i int) interface{}    { return v[i] }
func (v BigIntVec) intf(i int) interface{} { return v[i] }
func (v BigFltVec) intf(i int) interface{} { return v[i] }
func (v RatioVec) intf(i int) interface{}  { return v[i] }
func (v TimeVec) intf(i int) interface{}   { return v[i] }
func (v DuraVec) intf(i int) interface{}   { return v[i] }
func (v ErrorVec) intf(i int) interface{}  { return v[i] }
func (v FlagSet) intf(i int) interface{}   { return v[i] }

func (v NilVec) ints(i, j int) interface{}    { return v[i:j] }
func (v BoolVec) ints(i, j int) interface{}   { return v[i:j] }
func (v IntVec) ints(i, j int) interface{}    { return v[i:j] }
func (v Int8Vec) ints(i, j int) interface{}   { return v[i:j] }
func (v Int16Vec) ints(i, j int) interface{}  { return v[i:j] }
func (v Int32Vec) ints(i, j int) interface{}  { return v[i:j] }
func (v UintVec) ints(i, j int) interface{}   { return v[i:j] }
func (v Uint8Vec) ints(i, j int) interface{}  { return v[i:j] }
func (v Uint16Vec) ints(i, j int) interface{} { return v[i:j] }
func (v Uint32Vec) ints(i, j int) interface{} { return v[i:j] }
func (v FltVec) ints(i, j int) interface{}    { return v[i:j] }
func (v Flt32Vec) ints(i, j int) interface{}  { return v[i:j] }
func (v ImagVec) ints(i, j int) interface{}   { return v[i:j] }
func (v Imag64Vec) ints(i, j int) interface{} { return v[i:j] }
func (v ByteVec) ints(i, j int) interface{}   { return v[i:j] }
func (v RuneVec) ints(i, j int) interface{}   { return v[i:j] }
func (v BytesVec) ints(i, j int) interface{}  { return v[i:j] }
func (v StrVec) ints(i, j int) interface{}    { return v[i:j] }
func (v BigIntVec) ints(i, j int) interface{} { return v[i:j] }
func (v BigFltVec) ints(i, j int) interface{} { return v[i:j] }
func (v RatioVec) ints(i, j int) interface{}  { return v[i:j] }
func (v TimeVec) ints(i, j int) interface{}   { return v[i:j] }
func (v DuraVec) ints(i, j int) interface{}   { return v[i:j] }
func (v ErrorVec) ints(i, j int) interface{}  { return v[i:j] }
func (v FlagSet) ints(i, j int) interface{}   { return v[i:j] }

func (v NilVec) Nats(i, j int) []struct{}       { return NilVec(v[i:j]) }
func (v BoolVec) Nats(i, j int) []bool          { return BoolVec(v[i:j]) }
func (v IntVec) Nats(i, j int) []int            { return IntVec(v[i:j]) }
func (v Int8Vec) Nats(i, j int) []int8          { return Int8Vec(v[i:j]) }
func (v Int16Vec) Nats(i, j int) []int16        { return Int16Vec(v[i:j]) }
func (v Int32Vec) Nats(i, j int) []int32        { return Int32Vec(v[i:j]) }
func (v UintVec) Nats(i, j int) []uint          { return UintVec(v[i:j]) }
func (v Uint8Vec) Nats(i, j int) []uint8        { return Uint8Vec(v[i:j]) }
func (v Uint16Vec) Nats(i, j int) []uint16      { return Uint16Vec(v[i:j]) }
func (v Uint32Vec) Nats(i, j int) []uint32      { return Uint32Vec(v[i:j]) }
func (v FltVec) Nats(i, j int) []float64        { return FltVec(v[i:j]) }
func (v Flt32Vec) Nats(i, j int) []float32      { return Flt32Vec(v[i:j]) }
func (v ImagVec) Nats(i, j int) []complex128    { return ImagVec(v[i:j]) }
func (v Imag64Vec) Nats(i, j int) []complex64   { return Imag64Vec(v[i:j]) }
func (v ByteVec) Nats(i, j int) []byte          { return ByteVec(v[i:j]) }
func (v RuneVec) Nats(i, j int) []rune          { return RuneVec(v[i:j]) }
func (v BytesVec) Nats(i, j int) [][]byte       { return BytesVec(v[i:j]) }
func (v StrVec) Nats(i, j int) []string         { return StrVec(v[i:j]) }
func (v BigIntVec) Nats(i, j int) []big.Int     { return BigIntVec(v[i:j]) }
func (v BigFltVec) Nats(i, j int) []big.Float   { return BigFltVec(v[i:j]) }
func (v RatioVec) Nats(i, j int) []big.Rat      { return RatioVec(v[i:j]) }
func (v TimeVec) Nats(i, j int) []time.Time     { return TimeVec(v[i:j]) }
func (v DuraVec) Nats(i, j int) []time.Duration { return DuraVec(v[i:j]) }

func (v NilVec) Flag() BitFlag    { return Slice.Flag() | Nil.Flag() }
func (v BoolVec) Flag() BitFlag   { return Slice.Flag() | Bool.Flag() }
func (v IntVec) Flag() BitFlag    { return Slice.Flag() | Int.Flag() }
func (v Int8Vec) Flag() BitFlag   { return Slice.Flag() | Int8.Flag() }
func (v Int16Vec) Flag() BitFlag  { return Slice.Flag() | Int16.Flag() }
func (v Int32Vec) Flag() BitFlag  { return Slice.Flag() | Int32.Flag() }
func (v UintVec) Flag() BitFlag   { return Slice.Flag() | Uint.Flag() }
func (v Uint8Vec) Flag() BitFlag  { return Slice.Flag() | Uint8.Flag() }
func (v Uint16Vec) Flag() BitFlag { return Slice.Flag() | Uint16.Flag() }
func (v Uint32Vec) Flag() BitFlag { return Slice.Flag() | Uint32.Flag() }
func (v FltVec) Flag() BitFlag    { return Slice.Flag() | Float.Flag() }
func (v Flt32Vec) Flag() BitFlag  { return Slice.Flag() | Flt32.Flag() }
func (v ImagVec) Flag() BitFlag   { return Slice.Flag() | Imag.Flag() }
func (v Imag64Vec) Flag() BitFlag { return Slice.Flag() | Imag64.Flag() }
func (v ByteVec) Flag() BitFlag   { return Slice.Flag() | Byte.Flag() }
func (v RuneVec) Flag() BitFlag   { return Slice.Flag() | Rune.Flag() }
func (v BytesVec) Flag() BitFlag  { return Slice.Flag() | Bytes.Flag() }
func (v StrVec) Flag() BitFlag    { return Slice.Flag() | String.Flag() }
func (v BigIntVec) Flag() BitFlag { return Slice.Flag() | BigInt.Flag() }
func (v BigFltVec) Flag() BitFlag { return Slice.Flag() | BigFlt.Flag() }
func (v RatioVec) Flag() BitFlag  { return Slice.Flag() | Ratio.Flag() }
func (v TimeVec) Flag() BitFlag   { return Slice.Flag() | Time.Flag() }
func (v DuraVec) Flag() BitFlag   { return Slice.Flag() | Duration.Flag() }

func (v NilVec) String() string    { return v.String() }
func (v BoolVec) String() string   { return v.String() }
func (v IntVec) String() string    { return v.String() }
func (v Int8Vec) String() string   { return v.String() }
func (v Int16Vec) String() string  { return v.String() }
func (v Int32Vec) String() string  { return v.String() }
func (v UintVec) String() string   { return v.String() }
func (v Uint8Vec) String() string  { return v.String() }
func (v Uint16Vec) String() string { return v.String() }
func (v Uint32Vec) String() string { return v.String() }
func (v FltVec) String() string    { return v.String() }
func (v Flt32Vec) String() string  { return v.String() }
func (v ImagVec) String() string   { return v.String() }
func (v Imag64Vec) String() string { return v.String() }
func (v ByteVec) String() string   { return v.String() }
func (v RuneVec) String() string   { return v.String() }
func (v BytesVec) String() string  { return v.String() }
func (v StrVec) String() string    { return v.String() }
func (v BigIntVec) String() string { return v.String() }
func (v BigFltVec) String() string { return v.String() }
func (v RatioVec) String() string  { return v.String() }
func (v TimeVec) String() string   { return v.String() }
func (v DuraVec) String() string   { return v.String() }

func (v NilVec) Eval() Data    { return v }
func (v BoolVec) Eval() Data   { return v }
func (v IntVec) Eval() Data    { return v }
func (v Int8Vec) Eval() Data   { return v }
func (v Int16Vec) Eval() Data  { return v }
func (v Int32Vec) Eval() Data  { return v }
func (v UintVec) Eval() Data   { return v }
func (v Uint8Vec) Eval() Data  { return v }
func (v Uint16Vec) Eval() Data { return v }
func (v Uint32Vec) Eval() Data { return v }
func (v FltVec) Eval() Data    { return v }
func (v Flt32Vec) Eval() Data  { return v }
func (v ImagVec) Eval() Data   { return v }
func (v Imag64Vec) Eval() Data { return v }
func (v ByteVec) Eval() Data   { return v }
func (v RuneVec) Eval() Data   { return v }
func (v BytesVec) Eval() Data  { return v }
func (v StrVec) Eval() Data    { return v }
func (v BigIntVec) Eval() Data { return v }
func (v BigFltVec) Eval() Data { return v }
func (v RatioVec) Eval() Data  { return v }
func (v TimeVec) Eval() Data   { return v }
func (v DuraVec) Eval() Data   { return v }

func conChain(val ...Data) Chain {
	l := make([]Data, 0, len(val))
	l = append(l, val...)
	return l
}
func (c Chain) Flag() BitFlag { return Slice.Flag() }
func (c Chain) Eval() Data    { return c }

// SLICE ->
func (v Chain) Slice() []Data { return v }
func (v Chain) Len() int      { return len(v) }

// COLLECTION
func (s Chain) Empty() bool            { return ChainEmpty(s) }
func (s Chain) Head() (h Data)         { return s[0] }
func (s Chain) Tail() (c Consumeable)  { return s[:1] }
func (s Chain) Shift() (c Consumeable) { return s[:1] }

func ChainClear(s Chain) {
	if len(s) > 0 {
		for i, v := range s {
			if !Match(v.Flag(), Nullable.Flag()) {
				if d, ok := v.(Destructable); ok {
					d.Clear()
				}
			}
			s[i] = nil
		}
	}
	s = nil
}
func elemEmpty(d Data) bool {
	// not flagged nil, not a composition either...
	if !Match(d.Flag(), (Nil.Flag() | Slice.Flag())) {
		if d != nil { // not a nil pointer...
			// --> not empty
			return false
		}
	}
	// since it's a composition, inspect...
	if Match(d.Flag(), Slice.Flag()) {
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
			if !elemEmpty(elem) {
				return false
			}
		}
	} // --> all contained elements are empty
	return true
}

// ACCESSABLE SLICE
func ChainGet(s Chain, i int) Data { return s[i] }

// MUTABLE SLICE
func ChainSet(s Chain, i int, v Data) Chain { s[i] = v; return s }

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

// BOOTOM & TOP
func ChainFirst(s Chain) Data {
	if s.Len() > 0 {
		return s[0]
	}
	return NilVal{}
}
func ChainLast(s Chain) Data {
	if s.Len() > 0 {
		return s[s.Len()-1]
	}
	return NilVal{}
}

// LIFO QUEUE
func ChainPut(s Chain, v Data) Chain {
	return append(s, v)
}
func ChainAppend(s Chain, v ...Data) Chain {
	return append(s, v...)
}
func ChainPull(s Chain) (Data, Chain) {
	if s.Len() > 0 {
		return s[s.Len()-1], s[:s.Len()-1]
	}
	return NilVal{}, s
}

// FIFO STACK
func ChainAdd(s Chain, v ...Data) Chain {
	return append(v, s...)
}
func ChainPush(s Chain, v Data) Chain {
	return append([]Data{v}, s...)
}
func ChainPop(s Chain) (Data, Chain) {
	if len(s) > 0 {
		return s[0], s[1:]
	}
	return NilVal{}, s
}

// TUPLE
func ChainHead(s Chain) (h Data)   { return s[0] }
func ChainTail(s Chain) (c []Data) { return s[:1] }
func ChainDecap(s Chain) (h Data, t Chain) {
	if !ChainEmpty(s) {
		return s[0], t[:1]
	}
	return NilVal{}, conChain(NilVal{})
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

///// TODO: perf test that‥. test sliding window, or similar sophistited shenaegans.
//func ChainAdd(s Chain, v ...Data) Chain {
//	if len(s) >= cap(s)+len(v)/2 {
//		return append(append(make([]Data, 0, len(v)+len(s)), v...), s...)
//	}
//	return append(v, s...)
//}
//func ChainPush(s Chain, v Data) Chain {
//	if len(s) >= cap(s)/2 {
//		return append(append(make([]Data, 0, (len(s))*2), v), s...)
//	}
//	return append([]Data{v}, s...)
//}
