package types

import "fmt"

// FUNCTION CONSTRUCTORS
//
//go:generate stringer -type FixType
type FixType BitFlag

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
	Args                   = 2
)

///// functional closure over data
func (d dataClosure) Flag() BitFlag { return d().Flag() }
func chainData(x []Data, y []Data) Data {
	if len(x) > 0 {
		return sliceAppend(newSlice(x...), y...)
	}
	return x[0]
}
func encloseData(dat ...Data) dataClosure {
	return func(d ...Data) Data { return chainData(dat, d) }
}

///// lamda implementation //////////
func (l lambda) Call(in ...Data) Data {
	if len(in) == 0 {
		in = []Data{nilVal{}}
	}
	_, dat := l(in...)
	return dat
}
func (l lambda) Arity() Arity { return Arity(len(l.Args())) }
func (l lambda) Flag() BitFlag {
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
func (l lambda) Args() []BitFlag {
	a, _ := l()
	if len(a) > int(Args) {
		return a[Args:]
	}
	return []BitFlag{}
}

/// lambda constructor
func composeLambda(
	fn func(...Data) Data,
	typ Typed,
	fix FixType,
	args ...BitFlag,
) lambda {

	var f = fn

	var flags = []BitFlag{typ.Flag(), BitFlag(fix)}
	for _, at := range args {
		flags = append(flags, at.Flag())
	}

	return func(d ...Data) (args, Data) {
		if len(d) == 0 {
			return flags, nilVal{}
		}
		return args{}, f(d...)
	}
}
func applyPartial(
	lam lambda,
	args args,
	dati ...Data,
) (lambdaClosure, []Data) {
	var sig = lam.Args()
	var sigo = []BitFlag{}
	var dato = []Data{}
	var dapply = []Data{}
	var length int
	fmt.Printf("sig: %s\tdati: %s\n", sig, dati)
	if len(sig) > len(dati) {
		length = len(dati)
	} else {
		length = len(sig)
	}
	for i := 0; i < length; i++ {
		if sig[i-1].Match(dati[i-1].Flag()) {
			dapply = append(dapply, dati[i-1])
		} else {
			if len(dato) < i {
				sigo = append(sigo, sig[i-1])
				dato = append(dato, dati[i-1])
			}
		}
	}
	fno := lam
	l := func(d ...Data) Data {
		return fno.Call(append(dapply, d...)...)
	}
	return enclsoseLambda(composeLambda(
		l,
		fno.Flag(),
		fno.Fixity(),
		sigo...)), dato
}

func (lc lambdaClosure) Enclosed() lambda { return lc().(lambda) }
func (lc lambdaClosure) Flag() BitFlag    { return lc().(lambda).Flag() }
func (lc lambdaClosure) Args() []BitFlag  { return lc().(lambda).Args() }
func (lc lambdaClosure) Arity() Arity     { return lc().(lambda).Arity() }
func (lc lambdaClosure) Fixity() FixType  { return lc().(lambda).Fixity() }
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
func (f function) Flag() BitFlag       { l, _ := f(); return l.Flag() }
func (f function) Args() []BitFlag     { l, _ := f(); return l.Args() }
func (f function) Arity() Arity        { l, _ := f(); return l.Arity() }
func (f function) Fixity() FixType     { l, _ := f(); return l.Fixity() }
func (f function) Call(d ...Data) Data { l, _ := f(); return l.Call(d...) }
func composeFunction(name string, lambd lambda) function {
	n := strVal(name)
	l := lambd
	return func() (lambda, strVal) { return l, n }
}

func (fr functionClosure) Enclosed() function { return fr().(function) }
func (fr functionClosure) Name() strVal       { return fr().(function).Name() }
func (fr functionClosure) Flag() BitFlag      { return fr().(function).Flag() }
func (fr functionClosure) Args() []BitFlag    { return fr().(function).Args() }
func (fr functionClosure) Arity() Arity       { return fr().(function).Arity() }
func (fr functionClosure) Fixity() FixType    { return fr().(function).Fixity() }
func enclsoseFunction(fnc function) functionClosure {
	var f = fnc
	return func(dat ...Data) Data {
		if len(dat) == 0 {
			return f
		}
		return f.Call(dat...)
	}
}
