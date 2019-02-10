package data

func NewPair(l, r Primary) Paired { return PairVal{l, r} }

// implements Paired flagged Pair
func (p PairVal) Left() Primary            { return p.l }
func (p PairVal) Right() Primary           { return p.r }
func (p PairVal) Both() (Primary, Primary) { return p.l, p.r }
func (p PairVal) Eval(prime ...Primary) Primary {
	if len(prime) >= 2 {
		return PairVal{prime[0], prime[1]}
	}
	return p
}

// implements Mapped flagged Set

func NewStringSet(acc ...Paired) Mapped {
	var m = make(map[StrVal]Primary)
	for _, pair := range acc {
		m[pair.Left().(StrVal)] = pair.Right()
	}
	return SetString(m)
}

func (s SetString) Eval(p ...Primary) Primary {
	if len(p) > 0 {
		for _, prime := range p {
			if prime.TypePrim().Flag().Match(Pair) {
				var pair = prime.(PairVal)
				if pair.Left().TypePrim().Flag().Match(String) {
					s.Set(pair.Left(), pair.Right())
				}
			}
		}
	}
	return s
}
func (s SetString) TypePrim() TyPrimitive { return Set.TypePrim() }
func (s SetString) Keys() []Primary {
	var keys = []Primary{}
	for k, _ := range s {
		keys = append(keys, k)
	}
	return keys
}
func (s SetString) Data() []Primary {
	var dat = []Primary{}
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
func (s SetString) Get(acc Primary) (Primary, bool) {
	if dat, ok := s[acc.(StrVal)]; ok {
		return dat, ok
	}
	return nil, false
}
func (s SetString) Set(acc Primary, dat Primary) Mapped { s[acc.(StrVal)] = acc.(StrVal); return s }

//////////////////////////////////////////////////////////////

func NewIntSet(acc ...Paired) Mapped {
	var m = make(map[IntVal]Primary)
	for _, pair := range acc {
		m[pair.Left().(IntVal)] = pair.Right()
	}
	return SetInt(m)
}

func (s SetInt) TypePrim() TyPrimitive { return Set.TypePrim() }

func (s SetInt) Eval(p ...Primary) Primary {
	if len(p) > 0 {
		for _, prime := range p {
			if prime.TypePrim().Flag().Match(Pair) {
				var pair = prime.(PairVal)
				if pair.Left().TypePrim().Flag().Match(Integer) {
					s.Set(IntVal(pair.Left().(IntegerVal).Int()), pair.Right())
				}
			}
		}
	}
	return s
}
func (s SetInt) Keys() []Primary {
	var keys = []Primary{}
	for k, _ := range s {
		keys = append(keys, k)
	}
	return keys
}
func (s SetInt) Data() []Primary {
	var dat = []Primary{}
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
func (s SetInt) Get(acc Primary) (Primary, bool) {
	if dat, ok := s[acc.(IntVal)]; ok {
		return dat, ok
	}
	return nil, false
}
func (s SetInt) Set(acc Primary, dat Primary) Mapped { s[acc.(IntVal)] = acc.(IntVal); return s }

//////////////////////////////////////////////////////////////

func NewUintSet(acc ...Paired) Mapped {
	var m = make(map[UintVal]Primary)
	for _, pair := range acc {
		m[pair.Left().(UintVal)] = pair.Right()
	}
	return SetUint(m)
}

func (s SetUint) TypePrim() TyPrimitive { return Set.TypePrim() }
func (s SetUint) Eval(p ...Primary) Primary {
	if len(p) > 0 {
		for _, prime := range p {
			if prime.TypePrim().Flag().Match(Pair) {
				var pair = prime.(PairVal)
				if pair.Left().TypePrim().Flag().Match(Natural) {
					s.Set(UintVal(pair.Left().(NaturalVal).Uint()), pair.Right())
				}
			}
		}
	}
	return s
}
func (s SetUint) Keys() []Primary {
	var keys = []Primary{}
	for k, _ := range s {
		keys = append(keys, k)
	}
	return keys
}
func (s SetUint) Data() []Primary {
	var dat = []Primary{}
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
func (s SetUint) Get(acc Primary) (Primary, bool) {
	if dat, ok := s[acc.(UintVal)]; ok {
		return dat, ok
	}
	return nil, false
}
func (s SetUint) Set(acc Primary, dat Primary) Mapped { s[acc.(UintVal)] = acc.(UintVal); return s }

//////////////////////////////////////////////////////////////

func NewFloatSet(acc ...Paired) Mapped {
	var m = make(map[FltVal]Primary)
	for _, pair := range acc {
		m[pair.Left().(FltVal)] = pair.Right()
	}
	return SetFloat(m)
}

func (s SetFloat) TypePrim() TyPrimitive { return Set.TypePrim() }
func (s SetFloat) Eval(p ...Primary) Primary {
	if len(p) > 0 {
		for _, prime := range p {
			if prime.TypePrim().Flag().Match(Pair) {
				var pair = prime.(PairVal)
				if pair.Left().TypePrim().Flag().Match(Real) {
					s.Set(FltVal(pair.Left().(RealVal).Float()), pair.Right())
				}
			}
		}
	}
	return s
}
func (s SetFloat) Keys() []Primary {
	var keys = []Primary{}
	for k, _ := range s {
		keys = append(keys, k)
	}
	return keys
}
func (s SetFloat) Data() []Primary {
	var dat = []Primary{}
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
func (s SetFloat) Get(acc Primary) (Primary, bool) {
	if dat, ok := s[acc.(FltVal)]; ok {
		return dat, ok
	}
	return nil, false
}
func (s SetFloat) Set(acc Primary, dat Primary) Mapped { s[acc.(FltVal)] = acc.(FltVal); return s }

//////////////////////////////////////////////////////////////

func NewBitFlagSet(acc ...Paired) Mapped {
	var m = make(map[BitFlag]Primary)
	for _, pair := range acc {
		m[pair.Left().(BitFlag)] = pair.Right()
	}
	return SetFlag(m)
}

func (s SetFlag) TypePrim() TyPrimitive { return Set.TypePrim() }
func (s SetFlag) Eval(p ...Primary) Primary {
	if len(p) > 0 {
		for _, prime := range p {
			if prime.TypePrim().Flag().Match(Pair) {
				var pair = prime.(PairVal)
				if pair.Left().TypePrim().Flag().Match(Flag) {
					s.Set(UintVal(pair.Left().(NaturalVal).Uint()), pair.Right())
				}
			}
		}
	}
	return s
}
func (s SetFlag) Keys() []Primary {
	var keys = []Primary{}
	for k, _ := range s {
		keys = append(keys, k)
	}
	return keys
}
func (s SetFlag) Data() []Primary {
	var dat = []Primary{}
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
func (s SetFlag) Get(acc Primary) (Primary, bool) {
	if dat, ok := s[acc.(BitFlag)]; ok {
		return dat, ok
	}
	return nil, false
}
func (s SetFlag) Set(acc Primary, dat Primary) Mapped { s[acc.(BitFlag)] = acc.(BitFlag); return s }
