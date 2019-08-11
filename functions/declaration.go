package functions

import (
	"strings"

	d "github.com/joergreinhardt/gatwd/data"
)

type (
	// TUPLE TYPE
	TupleVal  []Expression
	TupleType func(...Expression) TupleVal
)

//// TUPLE VALUE CONSTRUCTOR
///
// returns a tuple type constructor, defined by tuples field types. the first
// typed argument passed may be the types symbolic name.
//
// tuple type constructor either expects a sequence of flat args, or pairs with
// an index assigned to the left fields corresponding to fields position, for
// cases where the constructor is wrapped as declared expression and arguments
// are not intendet to be passed in the correct order. non-pair arguments are
// assigned in the order they are passed in.  field types will be replaced with
// corresponding argument if it matches the fields type and an array with the
// tuple types name as first and its fields as succeeding elements will be
// returned.
//
// if partial application is intendet, constructor needs to be declared as
// expression by calling the declaration method.
func ConstructTupleType(types ...d.Typed) TupleType {
	var (
		symbol TySymbol
		ftypes []Expression
	)
	// return if no types where passed
	if len(types) == 0 {
		return nil
	}
	if len(types) > 0 {
		// set tuple typed name by first element, if its a symbol
		if Flag_Symbol.Match(types[0].FlagType()) {
			// return early if nothing but the name has been passed
			if len(types) == 1 {
				return nil
			}
			symbol = types[0].(TySymbol)
			// shift fields
			types = types[1:]
		} else {
			// generate tuple types name from type pattern
			symbol = DefSym(Tuple.String())
		}
	}
	// allocate field set with an extra element to hold the name
	ftypes = make([]Expression, 0, len(types))
	ftypes = append(ftypes, symbol)
	// expect all remaining type arguments, to be tuples field types
	for _, typ := range types {
		if Flag_Pattern.Match(typ.FlagType()) {
			ftypes = append(ftypes, typ.(TyPattern))
		}
	}
	return func(args ...Expression) TupleVal {
		var fields []Expression
		if fields == nil {
			fields = make([]Expression, 0, len(types))
			fields = append(fields, symbol)
		}
		if len(args) == len(ftypes)-1 {
			for n, arg := range args {
				if ftypes[n+1].Type().Match(arg.Type()) {
					fields = append(fields, arg)
				}
			}
			return fields
		}
		return ftypes
	}
}
func (t TupleType) TypeFnc() TyFnc                       { return Type | Tuple }
func (t TupleType) Call(args ...Expression) Expression   { return t(args...) }
func (t TupleType) Allocate(args ...Expression) TupleVal { return t(args...) }
func (t TupleType) Symbol() TySymbol                     { return t()[0].(TySymbol) }
func (t TupleType) Fields() []Expression                 { return t()[1:] }
func (t TupleType) Declare() ExprType {
	var (
		fields = t.Fields()
		types  = make([]d.Typed, 0, len(fields))
	)
	for _, field := range fields {
		types = append(types, field.Type())
	}
	return ConstructExpressionType(t, Def(types...),
		t()[0].(TySymbol), Def(types...))
}
func (t TupleType) Type() TyPattern {
	var (
		fields = t.Fields()
		types  = make([]d.Typed, 0, len(fields))
	)
	for _, field := range fields {
		types = append(types, field.Type())
	}
	return Def(t.Symbol(), Def(types...))
}
func (t TupleType) String() string {
	var (
		fields = t()
		strs   = make([]string, 0, len(fields))
	)
	for _, field := range fields {
		strs = append(strs, field.String())
	}
	return strings.Join(strs, ", ")
}

//// TUPLE VALUE
///
//
func (t TupleVal) TypeFnc() TyFnc                { return Tuple }
func (t TupleVal) Symbol() TySymbol              { return t[0].(TySymbol) }
func (t TupleVal) Name() string                  { return t.Symbol().String() }
func (t TupleVal) Fields() []Expression          { return t[1:] }
func (t TupleVal) Len() int                      { return len(t.Fields()) }
func (t TupleVal) Call(...Expression) Expression { return TupleVal(t.Fields()) }
func (t TupleVal) Type() TyPattern {
	var types = make([]d.Typed, 0, len(t))
	for _, field := range t.Fields() {
		types = append(types, field.Type())
	}
	return Def(t.Symbol(), Def(types...))
}
func (t TupleVal) String() string {
	var strs = make([]string, 0, len(t.Fields()))
	for _, field := range t.Fields() {
		strs = append(strs, field.String())
	}
	return "[" + strings.Join(strs, " ") + "]"
}
