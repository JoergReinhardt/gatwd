package types

// FUNCTION CONSTRUCTORS
//
// the closure implementations, for better performance not actually returns any
// data, when no arguments got passed. instead return data get's only yielded,
// when called with arguments.

// the abscence of parameters indicates, that the argument-/ and return value
// types are the intended and expected result by the caller.
//
// the result of an actual call is returned and the parameters omitted,
// whenever the 'call' method get's called explicitly.  the passing of
// parameters indicates that the enclosed function is intended to actually be
// called and the result to be returned.
//
//go:generate stringer -type FixType
type FixType flag

//go:generate stringer -type ArgPosition
type ArgPosition flag

//go:generate stringer -type Arity
type Arity int

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
	Constant
	FixNil

	// argument array layout:
	return_type    ArgPosition = 0
	fixity                     = 1
	argument_types             = 2
)

type args []flag

type dataClosure func(...Data) Data

func (d dataClosure) Type() flag { return d().Type() }
func chainData(x []Data, y []Data) Data {
	if len(x) > 0 {
		return sliceAppend(newSlice(x...), y...)
	}
	return x[0]
}
func encloseData(dat ...Data) dataClosure {
	return func(d ...Data) Data { return chainData(dat, d) }
}

type lambda func(...Data) (args, Data)

func (l lambda) Call(in ...Data) Data {
	if len(in) == 0 {
		in = []Data{nilVal{}}
	}
	_, dat := l(in...)
	return dat
}
func (l lambda) Arity() Arity { return Arity(len(l.ArgTypes())) }
func (l lambda) Type() flag {
	a, _ := l()
	if len(a) > int(return_type) {
		return a[return_type].Type()
	}
	return Nil.Type()
}
func (l lambda) Fixity() FixType {
	a, _ := l()
	if len(a) > int(fixity) {
		return FixType(a[fixity])
	}
	return FixNil
}
func (l lambda) ArgTypes() []flag {
	a, _ := l()
	if len(a) > int(argument_types) {
		return a[argument_types:]
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

	var flags = []flag{typ.Type(), flag(fix)}
	for _, at := range argTypes {
		flags = append(flags, at.Type())
	}

	return func(d ...Data) (args, Data) {
		if len(d) == 0 {
			return flags, nilVal{}
		}
		return args{}, f(d...)
	}
}

type lambdaClosure func(...Data) Data

func (lr lambdaClosure) Enclosed() Data   { return lr().(lambda) }
func (lr lambdaClosure) Type() flag       { return lr().(lambda).Type() }
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
type function func() (lambda, strVal)

func (f function) Name() strVal        { _, n := f(); return n }
func (f function) Type() flag          { l, _ := f(); return l.Type() }
func (f function) ArgTypes() []flag    { l, _ := f(); return l.ArgTypes() }
func (f function) Arity() Arity        { l, _ := f(); return l.Arity() }
func (f function) Fixity() FixType     { l, _ := f(); return l.Fixity() }
func (f function) Call(d ...Data) Data { l, _ := f(); return l.Call(d...) }
func composeFunction(name string, lambd lambda) function {
	n := strVal(name)
	l := lambd
	return func() (lambda, strVal) { return l, n }
}

type functionClosure func(...Data) Data

func (fr functionClosure) Enclosed() Data   { return fr().(function) }
func (fr functionClosure) Name() strVal     { return fr().(function).Name() }
func (fr functionClosure) Type() flag       { return fr().(function).Type() }
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
