package types

// FUNCTION CONSTRUCTORS
//
//go:generate stringer -type FixType
type FixType flag

//go:generate stringer -type ArgPosition
type ArgPosition int8

//go:generate stringer -type Arity
type Arity int8 // <-- this sounds terribly wrong, when born in germany8

const (
	Nullary Arity = 0 + iota
	Unary
	Binary
	Trinary
	Ternary
	Quaternary
	Quinary
	Senary
	Septenary
	Octonary
	Novenary
	Denary
	Polyadic

	PreFix FixType = 0 + iota
	InFix
	PostFix
	ConFix

	// argument array layout:
	ReturnType ArgPosition = 0
	Fixity                 = 1
	ArgTypes               = 2
)

func (d dataClosure) Type() flag { return d().Flag() }
func chainData(x []Data, y []Data) Data {
	if len(x) > 0 {
		return sliceAppend(newSlice(x...), y...)
	}
	return x[0]
}
func encloseData(dat ...Data) dataClosure {
	return func(d ...Data) Data { return chainData(dat, d) }
}

func (l lambda) Call(in ...Data) Data {
	if len(in) == 0 {
		in = []Data{nilVal{}}
	}
	_, dat := l(in...)
	return dat
}
func (l lambda) Arity() Arity { return Arity(len(l.ArgTypes())) }
func (l lambda) Flag() flag {
	a, _ := l()
	if len(a) > int(ReturnType) {
		return a[ReturnType].Flag()
	}
	return Nil.Flag()
}
func (l lambda) Fixity() FixType {
	a, _ := l()
	if len(a) > int(Fixity) {
		return FixType(a[Fixity])
	}
	return ConFix
}
func (l lambda) ArgTypes() []flag {
	a, _ := l()
	if len(a) > int(ArgTypes) {
		return a[ArgTypes:]
	}
	return []flag{}
}

func composeLambda(
	fn func(...Data) Data,
	typ Typed,
	fix FixType,
	argTypes ...Typed,
) lambda {

	var f = fn

	var flags = []flag{typ.Flag(), flag(fix)}
	for _, at := range argTypes {
		flags = append(flags, at.Flag())
	}

	return func(d ...Data) (args, Data) {
		if len(d) == 0 {
			return flags, nilVal{}
		}
		return args{}, f(d...)
	}
}

func (lr lambdaClosure) Enclosed() Data   { return lr().(lambda) }
func (lr lambdaClosure) Type() flag       { return lr().(lambda).Flag() }
func (lr lambdaClosure) ArgTypes() []flag { return lr().(lambda).ArgTypes() }
func (lr lambdaClosure) Arity() Arity     { return lr().(lambda).Arity() }
func (lr lambdaClosure) Fixity() FixType  { return lr().(lambda).Fixity() }
func enclsoseLambda(lmbd lambda) lambdaClosure {
	var l = lmbd
	return func(dat ...Data) Data {
		if len(dat) == 0 {
			return l
		}
		return l.Call(dat...)
	}
}

// wrapper type for named functions
func (f function) Name() strVal        { _, n := f(); return n }
func (f function) Flag() flag          { l, _ := f(); return l.Flag() }
func (f function) ArgTypes() []flag    { l, _ := f(); return l.ArgTypes() }
func (f function) Arity() Arity        { l, _ := f(); return l.Arity() }
func (f function) Fixity() FixType     { l, _ := f(); return l.Fixity() }
func (f function) Call(d ...Data) Data { l, _ := f(); return l.Call(d...) }
func composeFunction(name string, lambd lambda) function {
	n := strVal(name)
	l := lambd
	return func() (lambda, strVal) { return l, n }
}

func (fr functionClosure) Enclosed() Data   { return fr().(function) }
func (fr functionClosure) Name() strVal     { return fr().(function).Name() }
func (fr functionClosure) Type() flag       { return fr().(function).Flag() }
func (fr functionClosure) ArgTypes() []flag { return fr().(function).ArgTypes() }
func (fr functionClosure) Arity() Arity     { return fr().(function).Arity() }
func (fr functionClosure) Fixity() FixType  { return fr().(function).Fixity() }
func enclsoseFunction(fnc function) functionClosure {
	var f = fnc
	return func(dat ...Data) Data {
		if len(dat) == 0 {
			return f
		}
		return f.Call(dat...)
	}
}
