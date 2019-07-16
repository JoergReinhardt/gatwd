package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
	s "strings"
)

type (
	EnumType  []Expression
	TupleType []Expression
)

// enumerable type declaration takes variadic arguments of the expression type.
// the arguments can either represent all elements of the declared enumerable
// type, or be a combination of min, max, elipse and/or expressions, to declare
// bound and/or infinite enumerables.
func DeclareEnum(exprs ...Expression) EnumType {
	var segments, _ = disectEnum(exprs...)
	return segments
}

// disects enumerable type into segments seperated by ellipse, bounds, or
// generator expression
func disectEnum(enum ...Expression) (head, tail []Expression) {
	head, tail = []Expression{}, enum
	return head, tail
}

func (e EnumType) IsSegmented() bool {
	for _, elem := range e {
		if elem.TypeFnc().Match(
			Bound | Generator | Lexical,
		) {
			return true
		}
	}
	return false
}
func (e EnumType) TypeFnc() TyFnc         { return Enum }
func (e EnumType) Slice() []Expression    { return e }
func (e EnumType) Get(idx int) Expression { return e[idx] }
func (e EnumType) GetElem(idx int) ElemVal {
	return NewElement(e[idx], DefValue(NewNative(idx)))
}

// string concatenates the elements string representations
func (e EnumType) String() string {
	var slice = make([]string, 0, e.Len())
	for _, expr := range e {
		slice = append(slice, expr.String())
	}
	return "[" + s.Join(slice, " | ") + "]"
}
func (e EnumType) Len() int { return len(e) }

// type method ranges over all elements to determine if they are values, or
// sub-pattern like min|max bounds or elipse to indicate a partly defined,
// possibly infinite range of elements.
func (e EnumType) Type() TyPattern {
	var elems = make(TyPattern, 0, e.Len())
	for n, elem := range e {
		if elem.TypeFnc().Match(Pattern) {
			elems = append(elem.Type())
			continue
		}
		elems = append(
			elems,
			Define(
				DefValue(NewNative(n)),
				elem.Type(),
			))
	}
	return Define(Enum, Define(elems...))
}

// call tests if the first argument can be evaluated to yield an integer, if
// that's not the case, it take the count of arguments as an index, in case it
// is not greater, than the number of elements. when called empty, the enum
// type instance itself is returned
func (e EnumType) Call(args ...Expression) Expression {
	var arglen = len(args)
	if arglen > 0 {
		var arg = args[0]
		if arg.TypeFnc().Match(Data) {
			if native, ok := arg.(Native); ok {
				if native.Type().Match(d.Integers) ||
					native.Type().Match(d.Naturals) {
					return e[int(native.(d.Numeral).Int())]
				}
			}
		}
		if arglen < e.Len() {
			return e[arglen]
		}
	}
	return e
}

func DeclareTuple(exprs ...Expression) TupleType { return exprs }
