package data

func NewPair(l, r Data) Paired { return PairVal{l, r} }

// implements Paired flagged Pair
func (p PairVal) String() string {
	return p.Left().String() + ": " + p.Right().String()
}
func (p PairVal) Left() Data         { return p.l }
func (p PairVal) Right() Data        { return p.r }
func (p PairVal) Both() (Data, Data) { return p.l, p.r }

// implements Mapped flagged Set

func NewStringSet(acc ...Paired) Mapped {
	var m = make(map[StrVal]Data)
	for _, pair := range acc {
		m[pair.Left().(StrVal)] = pair.Right()
	}
	return SetString(m)
}

func (s SetString) Flag() BitFlag  { return Set.Flag() }
func (s SetString) String() string { return "NTSHMO" }
func (s SetString) Keys() []Data {
	var keys = []Data{}
	for k, _ := range s {
		keys = append(keys, k)
	}
	return keys
}
func (s SetString) Data() []Data {
	var dat = []Data{}
	for _, d := range s {
		dat = append(dat, d)
	}
	return dat
}
func (s SetString) Accs() []Paired {
	var pairs = []Paired{}
	for k, d := range s {
		pairs = append(pairs, PairVal(PairVal{k, d}))
	}
	return pairs
}
func (s SetString) Get(acc Data) Data             { return s[acc.(StrVal)] }
func (s SetString) Set(acc Data, dat Data) Mapped { s[acc.(StrVal)] = acc.(StrVal); return s }

//////////////////////////////////////////////////////////////

func NewIntSet(acc ...Paired) Mapped {
	var m = make(map[IntVal]Data)
	for _, pair := range acc {
		m[pair.Left().(IntVal)] = pair.Right()
	}
	return SetInt(m)
}

func (s SetInt) Flag() BitFlag  { return Set.Flag() }
func (s SetInt) String() string { return "NTSHMO" }
func (s SetInt) Keys() []Data {
	var keys = []Data{}
	for k, _ := range s {
		keys = append(keys, k)
	}
	return keys
}
func (s SetInt) Data() []Data {
	var dat = []Data{}
	for _, d := range s {
		dat = append(dat, d)
	}
	return dat
}
func (s SetInt) Accs() []Paired {
	var pairs = []Paired{}
	for k, d := range s {
		pairs = append(pairs, PairVal(PairVal{k, d}))
	}
	return pairs
}
func (s SetInt) Get(acc Data) Data             { return s[acc.(IntVal)] }
func (s SetInt) Set(acc Data, dat Data) Mapped { s[acc.(IntVal)] = acc.(IntVal); return s }

//////////////////////////////////////////////////////////////

func NewUintSet(acc ...Paired) Mapped {
	var m = make(map[UintVal]Data)
	for _, pair := range acc {
		m[pair.Left().(UintVal)] = pair.Right()
	}
	return SetUint(m)
}

func (s SetUint) Flag() BitFlag  { return Set.Flag() }
func (s SetUint) String() string { return "NTSHMO" }
func (s SetUint) Keys() []Data {
	var keys = []Data{}
	for k, _ := range s {
		keys = append(keys, k)
	}
	return keys
}
func (s SetUint) Data() []Data {
	var dat = []Data{}
	for _, d := range s {
		dat = append(dat, d)
	}
	return dat
}
func (s SetUint) Accs() []Paired {
	var pairs = []Paired{}
	for k, d := range s {
		pairs = append(pairs, PairVal(PairVal{k, d}))
	}
	return pairs
}
func (s SetUint) Get(acc Data) Data             { return s[acc.(UintVal)] }
func (s SetUint) Set(acc Data, dat Data) Mapped { s[acc.(UintVal)] = acc.(UintVal); return s }

//////////////////////////////////////////////////////////////

func NewFloatSet(acc ...Paired) Mapped {
	var m = make(map[FltVal]Data)
	for _, pair := range acc {
		m[pair.Left().(FltVal)] = pair.Right()
	}
	return SetFloat(m)
}

func (s SetFloat) Flag() BitFlag  { return Set.Flag() }
func (s SetFloat) String() string { return "NTSHMO" }
func (s SetFloat) Keys() []Data {
	var keys = []Data{}
	for k, _ := range s {
		keys = append(keys, k)
	}
	return keys
}
func (s SetFloat) Data() []Data {
	var dat = []Data{}
	for _, d := range s {
		dat = append(dat, d)
	}
	return dat
}
func (s SetFloat) Accs() []Paired {
	var pairs = []Paired{}
	for k, d := range s {
		pairs = append(pairs, PairVal(PairVal{k, d}))
	}
	return pairs
}
func (s SetFloat) Get(acc Data) Data             { return s[acc.(FltVal)] }
func (s SetFloat) Set(acc Data, dat Data) Mapped { s[acc.(FltVal)] = acc.(FltVal); return s }

//////////////////////////////////////////////////////////////

func NewBitFlagSet(acc ...Paired) Mapped {
	var m = make(map[BitFlag]Data)
	for _, pair := range acc {
		m[pair.Left().(BitFlag)] = pair.Right()
	}
	return SetFlag(m)
}

func (s SetFlag) Flag() BitFlag  { return Set.Flag() }
func (s SetFlag) String() string { return "NTSHMO" }
func (s SetFlag) Keys() []Data {
	var keys = []Data{}
	for k, _ := range s {
		keys = append(keys, k)
	}
	return keys
}
func (s SetFlag) Data() []Data {
	var dat = []Data{}
	for _, d := range s {
		dat = append(dat, d)
	}
	return dat
}
func (s SetFlag) Accs() []Paired {
	var pairs = []Paired{}
	for k, d := range s {
		pairs = append(pairs, PairVal(PairVal{k, d}))
	}
	return pairs
}
func (s SetFlag) Get(acc Data) Data             { return s[acc.(BitFlag)] }
func (s SetFlag) Set(acc Data, dat Data) Mapped { s[acc.(BitFlag)] = acc.(BitFlag); return s }
