package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
)

type (
	EnumGen    func(int) Expression
	EnumType   func(...int) ElemVal
	TupleType  EnumType
	RecordType TupleType
)

func DeclareElementGenerator(gen func(int) Expression) EnumGen { return gen }

func (g EnumGen) TypeFnc() TyFnc  { return Generator }
func (g EnumGen) Type() TyPattern { return Def(Generator, Element) }
func (g EnumGen) String() string  { return g.Type().TypeName() }
func (g EnumGen) Call(args ...Expression) Expression {
	if len(args) > 0 {
		if len(args) > 1 {
			var vec = NewVector()
			for _, arg := range args {
				vec = vec.Append(g.Call(arg))
			}
			return vec
		}
		var idx int
		var arg = args[0]
		if arg.TypeFnc().Match(Data) {
			if nat, ok := arg.(Native); ok {
				if nat.Type().Match(d.Numbers) {
					if num, ok := arg.(d.Numeral); ok {
						idx = num.GoInt()
					}
				}
			}
		}
		return g(idx)
	}
	return NewNone()
}

// helpers to define bounds from native, or go literal
func DefMinNum(min d.Numeral) TyPattern {
	return Def(Min, DefVal(NewNative(min)))
}
func DefMin(min int) TyPattern {
	return DefMinNum(New(min).(d.Numeral))
}
func DefMaxNum(max d.Numeral) TyPattern {
	return Def(Max, DefVal(NewNative(max)))
}
func DefMax(max int) TyPattern {
	return DefMaxNum(New(max).(d.Numeral))
}

// enum declaration takes a name, that might be the empty string (in which case
// the typename is generated from type pattern), a generator expression
// expected to return the expression expected at that position and a slice of
// typed instances, that may be of type TyValue, to define bounds, or
// particular values, TyLex → Ellipsis to define ranges, and a type flag to
// define the member elements type.
func DeclareEnum(name string, gen EnumGen, typeds ...d.Typed) EnumType {

	// define new pattern from typed elements
	var pattern = Def(typeds...)
	// define types name, by either name that has been passed, or its
	// signature
	var symbol TySymbol
	if name != "" {
		symbol = DefSym(name)
	} else {
		symbol = DefSym(pattern.Print("[", " | ", "]"))
	}

	return func(idx ...int) ElemVal {
		if len(idx) > 0 {
			if len(idx) > 1 { // fetch multiple elements
				var elements = make([]Expression, 0, len(idx))
				for _, pos := range idx {
					elements = append(elements,
						NewElement(
							gen(pos),
							Def(
								symbol,
								DefValGo(pos),
							)))
				}
				return NewElement(
					NewVector(elements...),
					symbol,
				)
			}
			// fetch one element
			var pos = idx[0]
			NewElement(
				gen(pos),
				Def(
					symbol,
					DefValGo(pos),
				))

		}
		// return typename and pattern
		return NewElement(symbol, pattern)
	}
}

func (e EnumType) TypeFnc() TyFnc   { return Enum }
func (e EnumType) Type() TyPattern  { return e().Type() }
func (e EnumType) TypeName() string { return e.Type().TypeName() }
func (e EnumType) String() string   { return e.TypeName() }
func (e EnumType) Len() int {
	var length int
	return length
}
func (e EnumType) Get(idx int) ElemVal {
	if idx < e.Len() || idx == -1 {
		return e(idx)
	}
	return NewElement(NewNone(), None)
}
func (e EnumType) Call(args ...Expression) Expression {
	if len(args) > 0 {
		// cast n args to int and fetch n elements
		if len(args) > 1 {
			var exprs = make([]Expression, 0, len(args))
			for _, arg := range args {
				exprs = append(exprs, e.Call(arg))
			}
			return NewVector(exprs...)
		}
		//  cast arg to int and fetch single element
		var idx int
		var arg = args[0]
		if arg.TypeFnc().Match(Data) {
			if nat, ok := arg.(Native); ok {
				if nat.Type().Match(d.Numbers) {
					if num, ok := arg.(d.Numeral); ok {
						idx = num.GoInt()
					}
				}
			}
		}
		return e(idx)
	}
	return NewNone()
}

//func DeclareTuple(exprs ...Expression) TupleType { return exprs }
