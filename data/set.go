package data

func NewPair(l, r Native) Paired { return PairVal{l, r} }

// implements Paired flagged Pair
func (p PairVal) Left() Native { return p.L }

func (p PairVal) Right() Native { return p.R }

func (p PairVal) Both() (Native, Native) { return p.L, p.R }

func (p PairVal) Eval(prime ...Native) Native {
	if len(prime) >= 2 {
		return PairVal{prime[0], prime[1]}
	}
	return p
}

////////////////////////////////////////////////////////////////
//// GENERIC ACCESSOR TYPED SET
///
//
func NewValSet(acc ...Paired) Mapped {
	var m = make(map[Native]Native)
	for _, pair := range acc {
		m[pair.Left().(StrVal)] = pair.Right()
	}
	return SetVal(m)
}

func (s SetVal) Eval(p ...Native) Native {
	if len(p) > 0 {
		for _, prime := range p {
			if prime.TypeNat().Flag().Match(Pair) {
				var pair = prime.(PairVal)
				s.Set(pair.Left(), pair.Right())
			}
		}
	}
	return s
}

func (s SetVal) TypeNat() TyNative { return Set.TypeNat() }

func (s SetVal) Len() int { return len(s) }

func (s SetVal) Keys() []Native {
	var keys = []Native{}
	for k, _ := range s {
		keys = append(keys, k)
	}
	return keys
}

func (s SetVal) Data() []Native {
	var dat = []Native{}
	for _, d := range s {
		dat = append(dat, d)
	}
	return dat
}

func (s SetVal) Fields() []Paired {
	var pairs = []Paired{}
	for k, d := range s {
		pairs = append(pairs, PairVal(PairVal{k, d}))
	}
	return pairs
}

func (s SetVal) Get(acc Native) (Native, bool) {
	if dat, ok := s[acc.(StrVal)]; ok {
		return dat, ok
	}
	return nil, false
}

func (s SetVal) Set(acc Native, dat Native) Mapped { s[acc.(StrVal)] = acc.(StrVal); return s }

func (s SetVal) Delete(acc Native) bool {
	if _, ok := s[acc.(StrVal)]; ok {
		delete(s, s[acc.(StrVal)])
		return ok
	}
	return false
}

// implements Mapped flagged Set

func NewStringSet(acc ...Paired) Mapped {
	var m = make(map[StrVal]Native)
	for _, pair := range acc {
		m[pair.Left().(StrVal)] = pair.Right()
	}
	return SetString(m)
}

func (s SetString) Eval(p ...Native) Native {
	if len(p) > 0 {
		for _, prime := range p {
			if prime.TypeNat().Flag().Match(Pair) {
				var pair = prime.(PairVal)
				if pair.Left().TypeNat().Flag().Match(String) {
					s.Set(pair.Left(), pair.Right())
				}
			}
		}
	}
	return s
}

func (s SetString) TypeNat() TyNative { return Set.TypeNat() }

func (s SetString) Len() int { return len(s) }

func (s SetString) Keys() []Native {
	var keys = []Native{}
	for k, _ := range s {
		keys = append(keys, k)
	}
	return keys
}

func (s SetString) Data() []Native {
	var dat = []Native{}
	for _, d := range s {
		dat = append(dat, d)
	}
	return dat
}

func (s SetString) Fields() []Paired {
	var pairs = []Paired{}
	for k, d := range s {
		pairs = append(pairs, PairVal(PairVal{k, d}))
	}
	return pairs
}

func (s SetString) Get(acc Native) (Native, bool) {
	if dat, ok := s[acc.(StrVal)]; ok {
		return dat, ok
	}
	return nil, false
}

func (s SetString) Delete(acc Native) bool {
	if _, ok := s[acc.(StrVal)]; ok {
		delete(s, acc.(StrVal))
		return ok
	}
	return false
}

func (s SetString) Set(acc Native, dat Native) Mapped { s[acc.(StrVal)] = acc.(StrVal); return s }

//////////////////////////////////////////////////////////////

func NewIntSet(acc ...Paired) Mapped {
	var m = make(map[IntVal]Native)
	for _, pair := range acc {
		m[pair.Left().(IntVal)] = pair.Right()
	}
	return SetInt(m)
}

func (s SetInt) TypeNat() TyNative { return Set.TypeNat() }

func (s SetInt) Len() int { return len(s) }

func (s SetInt) Eval(p ...Native) Native {
	if len(p) > 0 {
		for _, prime := range p {
			if prime.TypeNat().Flag().Match(Pair) {
				var pair = prime.(PairVal)
				if pair.Left().TypeNat().Flag().Match(Integers) {
					s.Set(IntVal(pair.Left().(Integer).Int()), pair.Right())
				}
			}
		}
	}
	return s
}

func (s SetInt) Keys() []Native {
	var keys = []Native{}
	for k, _ := range s {
		keys = append(keys, k)
	}
	return keys
}

func (s SetInt) Data() []Native {
	var dat = []Native{}
	for _, d := range s {
		dat = append(dat, d)
	}
	return dat
}

func (s SetInt) Fields() []Paired {
	var pairs = []Paired{}
	for k, d := range s {
		pairs = append(pairs, PairVal(PairVal{k, d}))
	}
	return pairs
}

func (s SetInt) Get(acc Native) (Native, bool) {
	if dat, ok := s[acc.(IntVal)]; ok {
		return dat, ok
	}
	return nil, false
}

func (s SetInt) Delete(acc Native) bool {
	if _, ok := s[acc.(IntVal)]; ok {
		delete(s, acc.(IntVal))
		return ok
	}
	return false
}

func (s SetInt) Set(acc Native, dat Native) Mapped { s[acc.(IntVal)] = acc.(IntVal); return s }

//////////////////////////////////////////////////////////////

func NewUintSet(acc ...Paired) Mapped {
	var m = make(map[UintVal]Native)
	for _, pair := range acc {
		m[pair.Left().(UintVal)] = pair.Right()
	}
	return SetUint(m)
}

func (s SetUint) TypeNat() TyNative { return Set.TypeNat() }

func (s SetUint) Eval(p ...Native) Native {
	if len(p) > 0 {
		for _, prime := range p {
			if prime.TypeNat().Flag().Match(Pair) {
				var pair = prime.(PairVal)
				if pair.Left().TypeNat().Flag().Match(Naturals) {
					s.Set(UintVal(pair.Left().(Natural).Uint()), pair.Right())
				}
			}
		}
	}
	return s
}

func (s SetUint) Len() int { return len(s) }

func (s SetUint) Keys() []Native {
	var keys = []Native{}
	for k, _ := range s {
		keys = append(keys, k)
	}
	return keys
}

func (s SetUint) Data() []Native {
	var dat = []Native{}
	for _, d := range s {
		dat = append(dat, d)
	}
	return dat
}

func (s SetUint) Fields() []Paired {
	var pairs = []Paired{}
	for k, d := range s {
		pairs = append(pairs, PairVal(PairVal{k, d}))
	}
	return pairs
}

func (s SetUint) Get(acc Native) (Native, bool) {
	if dat, ok := s[acc.(UintVal)]; ok {
		return dat, ok
	}
	return nil, false
}

func (s SetUint) Delete(acc Native) bool {
	if _, ok := s[acc.(UintVal)]; ok {
		delete(s, acc.(UintVal))
		return ok
	}
	return false
}

func (s SetUint) Set(acc Native, dat Native) Mapped { s[acc.(UintVal)] = acc.(UintVal); return s }

//////////////////////////////////////////////////////////////

func NewFloatSet(acc ...Paired) Mapped {
	var m = make(map[FltVal]Native)
	for _, pair := range acc {
		m[pair.Left().(FltVal)] = pair.Right()
	}
	return SetFloat(m)
}

func (s SetFloat) Len() int { return len(s) }

func (s SetFloat) TypeNat() TyNative { return Set.TypeNat() }

func (s SetFloat) Eval(p ...Native) Native {
	if len(p) > 0 {
		for _, prime := range p {
			if prime.TypeNat().Flag().Match(Pair) {
				var pair = prime.(PairVal)
				if pair.Left().TypeNat().Flag().Match(Reals) {
					s.Set(FltVal(pair.Left().(Real).Float()), pair.Right())
				}
			}
		}
	}
	return s
}

func (s SetFloat) Keys() []Native {
	var keys = []Native{}
	for k, _ := range s {
		keys = append(keys, k)
	}
	return keys
}

func (s SetFloat) Data() []Native {
	var dat = []Native{}
	for _, d := range s {
		dat = append(dat, d)
	}
	return dat
}

func (s SetFloat) Fields() []Paired {
	var pairs = []Paired{}
	for k, d := range s {
		pairs = append(pairs, PairVal(PairVal{k, d}))
	}
	return pairs
}

func (s SetFloat) Get(acc Native) (Native, bool) {
	if dat, ok := s[acc.(FltVal)]; ok {
		return dat, ok
	}
	return nil, false
}

func (s SetFloat) Delete(acc Native) bool {
	if _, ok := s[acc.(FltVal)]; ok {
		delete(s, acc.(FltVal))
		return ok
	}
	return false
}

func (s SetFloat) Set(acc Native, dat Native) Mapped { s[acc.(FltVal)] = acc.(FltVal); return s }

//////////////////////////////////////////////////////////////

func NewBitFlagSet(acc ...Paired) Mapped {
	var m = make(map[BitFlag]Native)
	for _, pair := range acc {
		m[pair.Left().(BitFlag)] = pair.Right()
	}
	return SetFlag(m)
}

func (s SetFlag) TypeNat() TyNative { return Set.TypeNat() }

func (s SetFlag) Eval(p ...Native) Native {
	if len(p) > 0 {
		for _, prime := range p {
			if prime.TypeNat().Flag().Match(Pair) {
				var pair = prime.(PairVal)
				if pair.Left().TypeNat().Flag().Match(Flag) {
					s.Set(UintVal(pair.Left().(Natural).Uint()), pair.Right())
				}
			}
		}
	}
	return s
}

func (s SetFlag) Len() int { return len(s) }

func (s SetFlag) Keys() []Native {
	var keys = []Native{}
	for k, _ := range s {
		keys = append(keys, k)
	}
	return keys
}

func (s SetFlag) Data() []Native {
	var dat = []Native{}
	for _, d := range s {
		dat = append(dat, d)
	}
	return dat
}

func (s SetFlag) Fields() []Paired {
	var parms = []Paired{}
	for k, d := range s {
		parms = append(parms, PairVal{k, d})
	}
	return parms
}

func (s SetFlag) Get(acc Native) (Native, bool) {
	if dat, ok := s[acc.(BitFlag)]; ok {
		return dat, ok
	}
	return nil, false
}

func (s SetFlag) Delete(acc Native) bool {
	if _, ok := s[acc.(BitFlag)]; ok {
		delete(s, acc.(BitFlag))
		return ok
	}
	return false
}

func (s SetFlag) Set(acc Native, dat Native) Mapped { s[acc.(BitFlag)] = acc.(BitFlag); return s }
