package data

func NewPair(l, r Native) Paired { return PairVal{l, r} }

// implements Paired flagged Pair
func (p PairVal) Left() Native           { return p.L }
func (p PairVal) Right() Native          { return p.R }
func (p PairVal) Both() (Native, Native) { return p.L, p.R }
func (p PairVal) Type() TyNat            { return Pair }
func (p PairVal) TypeKey() TyNat         { return p.L.Type() }
func (p PairVal) TypeValue() TyNat       { return p.R.Type() }

////////////////////////////////////////////////////////////////
//// GENERIC ACCESSOR TYPED SET
///
//
func NewValMap(acc ...Paired) Mapped {
	var m = make(map[Native]Native)
	for _, pair := range acc {
		m[pair.Left().(StrVal)] = pair.Right()
	}
	return MapVal(m)
}

func (s MapVal) Type() TyNat { return Map }
func (s MapVal) First() Paired {
	if s.Len() > 0 {
		return s.Fields()[0]
	}
	return NewPair(NewNil(), NewNil())
}
func (s MapVal) TypeKey() Typed   { return s.First().Left().Type() }
func (s MapVal) TypeValue() Typed { return s.First().Right().Type() }

func (s MapVal) Len() int { return len(s) }

func (s MapVal) Keys() []Native {
	var keys = []Native{}
	for k, _ := range s {
		keys = append(keys, k)
	}
	return keys
}

func (s MapVal) Data() []Native {
	var dat = []Native{}
	for _, d := range s {
		dat = append(dat, d)
	}
	return dat
}

func (s MapVal) Slice() []Native {
	var native = []Native{}
	for k, d := range s {
		native = append(native, PairVal(PairVal{k, d}))
	}
	return native
}

func (s MapVal) Fields() []Paired {
	var pairs = []Paired{}
	for k, d := range s {
		pairs = append(pairs, PairVal(PairVal{k, d}))
	}
	return pairs
}

func (s MapVal) Has(acc Native) bool {
	if _, ok := s[acc]; ok {
		return ok
	}
	return false
}

func (s MapVal) Get(acc Native) (Native, bool) {
	if dat, ok := s[acc.(StrVal)]; ok {
		return dat, ok
	}
	return nil, false
}

func (s MapVal) Set(acc Native, dat Native) Mapped {
	s[acc.(StrVal)] = dat
	return s
}

func (s MapVal) Delete(acc Native) bool {
	if _, ok := s[acc.(StrVal)]; ok {
		delete(s, s[acc.(StrVal)])
		return ok
	}
	return false
}

// implements Mapped flagged Set

func NewStringMap(acc ...Paired) Mapped {
	var m = make(map[StrVal]Native)
	for _, pair := range acc {
		m[pair.Left().(StrVal)] = pair.Right()
	}
	return MapString(m)
}

func (s MapString) First() Paired {
	if s.Len() > 0 {
		return s.Fields()[0]
	}
	return NewPair(NewNil(), NewNil())
}

func (s MapString) Type() TyNat      { return Map }
func (s MapString) TypeKey() Typed   { return String.Type() }
func (s MapString) TypeValue() Typed { return s.First().Right().Type() }

func (s MapString) Len() int { return len(s) }

func (s MapString) Keys() []Native {
	var keys = []Native{}
	for k, _ := range s {
		keys = append(keys, k)
	}
	return keys
}

func (s MapString) Data() []Native {
	var dat = []Native{}
	for _, d := range s {
		dat = append(dat, d)
	}
	return dat
}

func (s MapString) Slice() []Native {
	var native = []Native{}
	for k, d := range s {
		native = append(native, PairVal(PairVal{k, d}))
	}
	return native
}

func (s MapString) Fields() []Paired {
	var pairs = []Paired{}
	for k, d := range s {
		pairs = append(pairs, PairVal(PairVal{k, d}))
	}
	return pairs
}

func (s MapString) Get(acc Native) (Native, bool) {
	if dat, ok := s[acc.(StrVal)]; ok {
		return dat, ok
	}
	return nil, false
}

func (s MapString) Has(acc Native) bool {
	if _, ok := s.Get(acc); ok {
		return ok
	}
	return false
}

func (s MapString) HasStr(key string) bool {
	if _, ok := s.GetStr(key); ok {
		return ok
	}
	return false
}

func (s MapString) Set(acc Native, dat Native) Mapped {
	s[acc.(StrVal)] = dat
	return s
}
func (s MapString) GetStr(key string) (Native, bool) {
	if dat, ok := s[StrVal(key)]; ok {
		return dat, ok
	}
	return nil, false
}
func (s MapString) SetStr(key string, dat Native) Mapped {
	s[StrVal(key)] = dat
	return s
}

func (s MapString) Delete(acc Native) bool {
	if _, ok := s[acc.(StrVal)]; ok {
		delete(s, acc.(StrVal))
		return ok
	}
	return false
}

//////////////////////////////////////////////////////////////

func NewIntMap(acc ...Paired) Mapped {
	var m = make(map[IntVal]Native)
	for _, pair := range acc {
		m[pair.Left().(IntVal)] = pair.Right()
	}
	return MapInt(m)
}

func (s MapInt) First() Paired {
	if s.Len() > 0 {
		return s.Fields()[0]
	}
	return NewPair(NewNil(), NewNil())
}
func (s MapInt) Type() TyNat { return Map }

func (s MapInt) TypeKey() Typed   { return Int.Type() }
func (s MapInt) TypeValue() Typed { return s.First().Right().Type() }

func (s MapInt) Len() int { return len(s) }

func (s MapInt) Keys() []Native {
	var keys = []Native{}
	for k, _ := range s {
		keys = append(keys, k)
	}
	return keys
}

func (s MapInt) Data() []Native {
	var dat = []Native{}
	for _, d := range s {
		dat = append(dat, d)
	}
	return dat
}

func (s MapInt) Slice() []Native {
	var native = []Native{}
	for k, d := range s {
		native = append(native, PairVal(PairVal{k, d}))
	}
	return native
}

func (s MapInt) Fields() []Paired {
	var pairs = []Paired{}
	for k, d := range s {
		pairs = append(pairs, PairVal(PairVal{k, d}))
	}
	return pairs
}

func (s MapInt) Has(acc Native) bool {
	if _, ok := s.Get(acc); ok {
		return ok
	}
	return false
}

func (s MapInt) HasInt(idx int) bool {
	if _, ok := s.GetIdx(idx); ok {
		return ok
	}
	return false
}

func (s MapInt) Get(acc Native) (Native, bool) {
	if dat, ok := s[acc.(IntVal)]; ok {
		return dat, ok
	}
	return nil, false
}

func (s MapInt) GetIdx(idx int) (Native, bool) {
	if dat, ok := s[IntVal(idx)]; ok {
		return dat, ok
	}
	return nil, false
}

func (s MapInt) Delete(acc Native) bool {
	if _, ok := s[acc.(IntVal)]; ok {
		delete(s, acc.(IntVal))
		return ok
	}
	return false
}

func (s MapInt) Set(acc Native, dat Native) Mapped {
	s[acc.(IntVal)] = dat
	return s
}

func (s MapInt) SetIdx(idx int, dat Native) Mapped {
	s[IntVal(idx)] = dat
	return s
}

//////////////////////////////////////////////////////////////

func NewUintMap(acc ...Paired) Mapped {
	var m = make(map[UintVal]Native)
	for _, pair := range acc {
		m[pair.Left().(UintVal)] = pair.Right()
	}
	return MapUint(m)
}

func (s MapUint) First() Paired {
	if s.Len() > 0 {
		return s.Fields()[0]
	}
	return NewPair(NewNil(), NewNil())
}
func (s MapUint) Type() TyNat { return Map }

func (s MapUint) TypeKey() Typed   { return Uint.Type() }
func (s MapUint) TypeValue() Typed { return s.First().Right().Type() }

func (s MapUint) Len() int { return len(s) }

func (s MapUint) Keys() []Native {
	var keys = []Native{}
	for k, _ := range s {
		keys = append(keys, k)
	}
	return keys
}

func (s MapUint) Data() []Native {
	var dat = []Native{}
	for _, d := range s {
		dat = append(dat, d)
	}
	return dat
}

func (s MapUint) Slice() []Native {
	var native = []Native{}
	for k, d := range s {
		native = append(native, PairVal(PairVal{k, d}))
	}
	return native
}

func (s MapUint) Fields() []Paired {
	var pairs = []Paired{}
	for k, d := range s {
		pairs = append(pairs, PairVal(PairVal{k, d}))
	}
	return pairs
}

func (s MapUint) HasUint(idx uint) bool {
	if _, ok := s.GetUint(idx); ok {
		return ok
	}
	return false
}

func (s MapUint) Has(acc Native) bool {
	if _, ok := s[acc.(UintVal)]; ok {
		return ok
	}
	return false
}

func (s MapUint) Get(acc Native) (Native, bool) {
	if dat, ok := s[acc.(UintVal)]; ok {
		return dat, ok
	}
	return nil, false
}

func (s MapUint) GetUint(idx uint) (Native, bool) {
	if dat, ok := s[UintVal(idx)]; ok {
		return dat, ok
	}
	return nil, false
}

func (s MapUint) Delete(acc Native) bool {
	if _, ok := s[acc.(UintVal)]; ok {
		delete(s, acc.(UintVal))
		return ok
	}
	return false
}

func (s MapUint) Set(acc Native, dat Native) Mapped {
	s[acc.(UintVal)] = dat
	return s
}

func (s MapUint) SetUint(idx uint, dat Native) Mapped {
	s[UintVal(idx)] = dat
	return s
}

//////////////////////////////////////////////////////////////

func NewFloatMap(acc ...Paired) Mapped {
	var m = make(map[FltVal]Native)
	for _, pair := range acc {
		m[pair.Left().(FltVal)] = pair.Right()
	}
	return MapFloat(m)
}

func (s MapFloat) Len() int { return len(s) }

func (s MapFloat) First() Paired {
	if s.Len() > 0 {
		return s.Fields()[0]
	}
	return NewPair(NewNil(), NewNil())
}
func (s MapFloat) Type() TyNat { return Map }

func (s MapFloat) TypeKey() Typed   { return Float.Type() }
func (s MapFloat) TypeValue() Typed { return s.First().Right().Type() }

func (s MapFloat) Keys() []Native {
	var keys = []Native{}
	for k, _ := range s {
		keys = append(keys, k)
	}
	return keys
}

func (s MapFloat) Data() []Native {
	var dat = []Native{}
	for _, d := range s {
		dat = append(dat, d)
	}
	return dat
}

func (s MapFloat) Slice() []Native {
	var native = []Native{}
	for k, d := range s {
		native = append(native, PairVal(PairVal{k, d}))
	}
	return native
}

func (s MapFloat) Fields() []Paired {
	var pairs = []Paired{}
	for k, d := range s {
		pairs = append(pairs, PairVal(PairVal{k, d}))
	}
	return pairs
}

func (s MapFloat) HasFlt(flt float64) bool {
	if _, ok := s.GetFlt(flt); ok {
		return ok
	}
	return false
}

func (s MapFloat) Has(acc Native) bool {
	if _, ok := s[acc.(FltVal)]; ok {
		return ok
	}
	return false
}

func (s MapFloat) Get(acc Native) (Native, bool) {
	if dat, ok := s[acc.(FltVal)]; ok {
		return dat, ok
	}
	return nil, false
}

func (s MapFloat) GetFlt(flt float64) (Native, bool) {
	if dat, ok := s[FltVal(flt)]; ok {
		return dat, ok
	}
	return nil, false
}

func (s MapFloat) Delete(acc Native) bool {
	if _, ok := s[acc.(FltVal)]; ok {
		delete(s, acc.(FltVal))
		return ok
	}
	return false
}

func (s MapFloat) Set(acc Native, dat Native) Mapped {
	s[acc.(FltVal)] = dat
	return s
}

func (s MapFloat) SetFlt(flt float64, dat Native) Mapped {
	s[FltVal(flt)] = dat
	return s
}

//////////////////////////////////////////////////////////////

func NewFLagMap(acc ...Paired) Mapped {
	var m = make(map[BitFlag]Native)
	for _, pair := range acc {
		m[pair.Left().(BitFlag)] = pair.Right()
	}
	return MapFlag(m)
}

func (s MapFlag) First() Paired {
	if s.Len() > 0 {
		return s.Fields()[0]
	}
	return NewPair(NewNil(), NewNil())
}
func (s MapFlag) Type() TyNat { return Map }

func (s MapFlag) TypeKey() Typed   { return Type.Type() }
func (s MapFlag) TypeValue() Typed { return s.First().Right().Type() }

func (s MapFlag) Len() int { return len(s) }

func (s MapFlag) Keys() []Native {
	var keys = []Native{}
	for k, _ := range s {
		keys = append(keys, k)
	}
	return keys
}

func (s MapFlag) Data() []Native {
	var dat = []Native{}
	for _, d := range s {
		dat = append(dat, d)
	}
	return dat
}

func (s MapFlag) Slice() []Native {
	var native = []Native{}
	for k, d := range s {
		native = append(native, PairVal(PairVal{k, d}))
	}
	return native
}

func (s MapFlag) Fields() []Paired {
	var parms = []Paired{}
	for k, d := range s {
		parms = append(parms, PairVal{k, d})
	}
	return parms
}

func (s MapFlag) Has(acc Native) bool {
	if _, ok := s[acc.(BitFlag)]; ok {
		return ok
	}
	return false
}

func (s MapFlag) Get(acc Native) (Native, bool) {
	if dat, ok := s[acc.(BitFlag)]; ok {
		return dat, ok
	}
	return nil, false
}

func (s MapFlag) Delete(acc Native) bool {
	if _, ok := s[acc.(BitFlag)]; ok {
		delete(s, acc.(BitFlag))
		return ok
	}
	return false
}

func (s MapFlag) Set(acc Native, dat Native) Mapped {
	s[acc.(BitFlag)] = dat
	return s
}
