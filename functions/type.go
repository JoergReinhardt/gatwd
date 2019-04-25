package functions

import (
	d "github.com/joergreinhardt/gatwd/data"
	"strings"
)

//go:generate stringer -type=TyFnc
const (
	/// TYPE RELATED FLAGS ///
	Type TyFnc = 1 << iota
	TypeSum
	TypeProduct
	Constructor
	Expression
	ExprProperys
	ExprArity
	Data
	/// FUNCTORS AND MONADS ///
	Applicable
	Operator
	Functor
	Monad
	/// MONADIC SUB TYPES ///
	False
	True
	Undecided
	Lesser
	Greter
	Just
	None
	Case
	Switch
	Either
	Or
	If
	Else
	While
	Do
	/// TYPE CLASSES ///
	Equality
	Truth
	Order
	Number
	Index
	Symbol
	Error
	/// COLLECTION TYPES ///
	Pair
	Tuple
	Enum
	Set
	List
	Record
	Vector
	/// HIGHER ORDER TYPE IS THE HIGHEST FLAG ///
	HigherOrder

	Kind = Data | Expression

	ExprProps = ExprProperys | ExprArity

	Morphisms = Constructor | Operator | Functor |
		Applicable | Monad

	Options = False | True | Just | None | Case |
		Switch | Either | Or | If | Else |
		While | Do

	Collections = Pair | Tuple | Enum | Set |
		List | Vector | Record

	Classes = Truth | Equality | Order | Number |
		Symbol | Error
)

//// FIRST ORDER TYPE FUNCTION
///
// first order type yields a types higher order type name and both, native and
// functional type flags.
type (
	TypeFOFnc   func() (string, d.TyNative, TyFnc)
	ProdTypeFnc func() (string, d.TyNative, TyFnc, HOTyped)
	SumTypeFnc  func() (string, d.TyNative, TyFnc, HOTyped, []HOTyped)

	// a type constructor composes first- and/or higher order types to
	// derive further higher order types.
	TypeConstructor func(...HOTyped) (HOTyped, bool)
	// data constructor returns either an instance of it's higher order
	// type & true, or some other callable, often a NoOp, or a vector of
	// the argument set to pass on recursively), and false
	DataConstructor func(...Callable) (Callable, bool)

	// satisfyed data constructors (aka constant data) & constant
	// expressions return their type constructor as second argument.
	TypedDataValue          func() (DataVal, TypeConstructor)
	TypedConstantExpression func() (ConstantExpr, TypeConstructor)

	// expression constructors return their type constructor and additional
	// parameters, when called without arguments. called with arguments, a
	// new typed variadic-, nary-, or proper-expression is composed from
	// resulting expression & typeconstructor, when arguments are applyed
	// to initial expression & type constructor and returned.
	TypedVariadicExpression func(...Callable) Callable // TypeConstructFnc
	TypedNaryExpression     func(...Callable) Callable // TypeConstructFnc,Arity
	TypedProperExpression   func(...Callable) Callable // TypeConstructFnc,Arity,Propertys
)

////////////////////////////////////////////////////////////////////////////////////////
//// FIRST ORDER TYPES
///
// yields the types name (usually the string representation of the native and
// functional type flags of the type, unless it's an alias type)
func NewTypeRoot() TypeFOFnc {

	return TypeFOFnc(

		func() (string, d.TyNative, TyFnc) {

			return "RootType", d.Flag, Type
		})
}

func NewTypeError() HOTyped {

	return TypeFOFnc(
		func() (string, d.TyNative, TyFnc) {
			return "Type Error",
				d.Error,
				Error
		})
}

// first instance higher order type from static functional type instance
func NewFirstOrderTypeFromExpr(expr Callable) TypeFOFnc {

	return NewFirstOrderTypeFromBaseTypes(expr.TypeNat(), expr.TypeFnc())
}

// first order type constructor that derives the types name from the native-/
// and functional type flags string representaion.
func NewFirstOrderTypeFromBaseTypes(nt d.TyNative, ft TyFnc) TypeFOFnc {

	return func() (string, d.TyNative, TyFnc) {

		// generate name from flags string representation
		return nt.String() + " " + ft.String(), nt, ft
	}
}

// create a first order type and pass in the types name as string argument
func NewFirstOrderAliasFromBaseTypes(name string, nt d.TyNative, ft TyFnc) TypeFOFnc {

	return func() (string, d.TyNative, TyFnc) {

		return name, nt, ft
	}
}

// create a first order type and pass in the types name as string argument
func NewFirstOrderAliasFromExpr(name string, expr Callable) TypeFOFnc {

	return func() (string, d.TyNative, TyFnc) {

		return name, expr.TypeNat(), expr.TypeFnc()
	}
}

func (t TypeFOFnc) TypeName() string {
	var name, _, _ = t()
	return name
}

// yields static native and function type flags of the first order type
func (t TypeFOFnc) TypeBase() (d.TyNative, TyFnc) {
	var _, tnat, tfnc = t()
	return tnat, tfnc
}

func (t TypeFOFnc) TypeRoot() bool { return true }

// yields the type systems root type with flags d.Flag & Type
func (t TypeFOFnc) TypeParent() HOTyped { return NewTypeRoot() }

// yields the 'Type' flag
func (t TypeFOFnc) TypeFnc() TyFnc { return Type }

// yields the 'd.Flag' flag
func (t TypeFOFnc) TypeNat() d.TyNative { return d.Flag }

// string function of a first order type concatenates the types name to it's
// native and functional type flags string representations, if they happen to
// differ. most first order type names are directly derived from their type
// flags and have base names identical to the string representation of those
// flags, in which case the base name is omitted.
func (t TypeFOFnc) String() string {

	var nat, fnc = t.TypeBase()

	var str = nat.String() + " " + fnc.String()

	if str != t.TypeName() {

		return t.TypeName() + " " + str
	}

	return str
}

func (t TypeFOFnc) TypeMatch(args ...HOTyped) bool { return matchFOType(t, args...) }

func matchFOType(t TypeFOFnc, args ...HOTyped) bool {

	var nat, fnc = t.TypeBase()
	var mnat, mfnc = nat.Flag().Match, fnc.Flag().Match

	for _, arg := range args {

		if !mnat(arg.TypeNat()) || !mfnc(arg.TypeFnc()) {

			return false
		}
	}

	return true
}

// calling a type verifys if the native and functional types of a given list of
// arguments match the types native-/ & functional type flags.
func (t TypeFOFnc) Call(args ...Callable) Callable {

	return NewFromData(d.BoolVal(t.TypeMatch(castArgsHO(args...)...)))
}

// eval converts arguments to callables first and calls call
func (t TypeFOFnc) Eval(args ...d.Native) d.Native {
	return t.Call(functionalizeNatives(args...)...).Eval()
}

////////////////////////////////////////////////////////////////////////////////////////
//// SUM TYPE FUNCTION
///
// sum type function is the base for all higher order types, including product
// types, where sum of derivable types, as well as sum of type arguments are
// sum types. for parametric types the sum of all type signatures bound to
// equeations, is a sum type.
func NewSumTypeFnc(types []HOTyped, parent ...HOTyped) SumTypeFnc {

	var names = []string{}
	var tnat, tfnc = d.TyNative(0), TyFnc(0)

	for _, typ := range types {

		names = append(names, typ.TypeName())
		tnat = tnat | typ.TypeNat()
		tfnc = tfnc | typ.TypeFnc()
	}

	var name = "["

	name = strings.Join(names, ", ")

	name = name + "]"

	if len(parent) == 0 {
		parent = []HOTyped{NewTypeRoot()}
	}

	return func() (string, d.TyNative, TyFnc, HOTyped, []HOTyped) {
		return name, tnat, tfnc, parent[0], types
	}

}

func NewSumTypeAlias(st SumTypeFnc, name string) SumTypeFnc {

	var _, tnat, tfnc, parent, types = st()

	return func() (string, d.TyNative, TyFnc, HOTyped, []HOTyped) {

		return name, tnat, tfnc, parent, types
	}
}

func (t SumTypeFnc) TypeName() string {

	var name = "("

	name = strings.Join(
		strings.Split(
			t.String(),
			" ",
		),
		", ",
	)

	return name + ")"
}

// string representation concatenates all parents names recursively, till first
// order type is reached.
func (t SumTypeFnc) String() string {

	return t.TypeName() + " " + t.TypeParent().String()
}

// sum type is parametric and has members, so it can't be root
func (t SumTypeFnc) TypeRoot() bool { return false }

// returns the parent type as instance of HOTyped interface
func (t SumTypeFnc) TypeParent() HOTyped {

	var _, _, _, hot, _ = t()

	return hot
}

// returns the types native-/ & functional type flags
func (t SumTypeFnc) TypeBase() (d.TyNative, TyFnc) {

	var _, tnat, tfnc, _, _ = t()

	return tnat, tfnc
}

// return member types of sum type
func (t SumTypeFnc) TypeMembers() []HOTyped {

	var _, _, _, _, mem = t()

	return mem
}

// returns true, if passed arguments match sumtypes member types in number and
// type, independent from order in which they got passed
func (t SumTypeFnc) TypeMemberProductMatch(args ...Callable) (bool, int) {

	return matchTypeMemberProduct(t, args...)

}
func matchTypeMemberProduct(t SumTypeFnc, args ...Callable) (bool, int) {

	var mems = t.TypeMembers()

	if len(args) != len(mems) {
		return false, -1
	}

	// range over list of all members
	for idx, mem := range mems {

		// range over list of all arguments, per type member
		for adx, hot := range castArgsHO(args...) {

			// if current member matches current argument
			if mem.TypeMatch(hot) {

				// if it happens to be the first member type
				if idx == 0 {

					// if there are more members
					if len(mems) > 1 {

						// exclude satisfyed member
						// from list
						mems = mems[1:]
					}

					mems = []HOTyped{mems[0]}

				}

				// should the matching argument happen to be
				// the first argument, exclude it from list.
				if adx == 0 {

					if len(args) > 1 {

						args = args[1:]
					}

					args = []Callable{args[0]}

				}

				// argument matches last type in list → exclude
				// last membertype from list
				if idx == len(mems)-1 {

					mems = append(mems[:idx], mems[idx+1])
				}

				// last argument passed matches this member
				// type → exclude last argument from list
				if adx == len(args)-1 {

					args = append(args[:adx], args[adx+1])
				}

				// exclude any satisfyed member-type inbetween
				// first and last member
				if idx > 0 && idx < len(args)-2 {

					mems = append(mems[:idx], mems[idx+1:]...)
				}

				// exclude any matching argument inbetween
				// first and last from list of arguments
				if adx > 0 && adx < len(args)-2 {

					args = append(args[:adx], args[adx+1:]...)
				}
			}
		}
	}

	// all arguments are depleted‥.
	if len(args) == 0 {

		// all member types are satisfyed
		if len(mems) == 0 {

			// return true and number of member types
			return true, len(t.TypeMembers())
		}
	}

	// return false and minus two, if number of arguments matched number of
	// member types, but typechecks failed.
	return false, -2
}

// returns true and the number of contained members/passed arguments , if the
// passed set of arguments satisfys all contained member types in correct
// order.
//
// returns false and minus one, in case number of passed arguments fails to
// match number of member types. returns false and the index of the first type,
// that failed to match, in case a type check fails to match.
func (t SumTypeFnc) TypeSignatureMatch(args ...Callable) (bool, int) {

	return matchSumTypeSignature(t, args...)
}
func matchSumTypeSignature(t SumTypeFnc, args ...Callable) (bool, int) {

	var lena = len(args)

	var mems = t.TypeMembers()

	if lena != len(mems) {

		return false, -1
	}

	for idx, hot := range castArgsHO(args...) {

		if !mems[idx].TypeMatch(hot) {

			return false, idx
		}
	}

	return true, lena
}

// matches against first argument and any further rguments recursively against
// it's parent type
//
// returns true and index of first matching member, if any of the contained
// members matches the argument set. returns false and minus one, when no
// arguments where passed & false minus two, if the argument set matched none
// of the members.
func (t SumTypeFnc) TypeMatch(args ...Callable) bool {
	var result, _ = matchSumType(t, args...)
	return result
}
func matchSumType(t SumTypeFnc, args ...Callable) (bool, int) {

	// no arguments can't be true
	if len(args) == 0 {

		return false, -1
	}

	// if parent type happens to match remaining arguments, test the first
	// argument against all member types. if any type member matches the
	// argument, return true and index of matching type
	for idx, memtype := range t.TypeMembers() {

		// match membertypes against arguments
		if memtype.TypeMatch(castArgsHO(args...)...) {

			// return a match and membertype indes, if types match
			return true, idx
		}
	}

	// return false if none of the members matches first argument type
	// check failed
	return false, -2
}

// yields the 'Type' Flag
func (t SumTypeFnc) TypeFnc() TyFnc { return Type | TypeSum }

// yields the 'd.Flag' Flag
func (t SumTypeFnc) TypeNat() d.TyNative { return d.Flag }

func (t SumTypeFnc) Call(args ...Callable) Callable {
	return NewFromData(d.BoolVal(t.TypeMatch(args...)))
}

func (t SumTypeFnc) Eval(args ...d.Native) d.Native {

	return d.BoolVal(t.TypeMatch(functionalizeNatives(args...)...)).Eval()
}

////////////////////////////////////////////////////////////////////////////////////////
//// PRODUCT TYPE FUNCTION
///
// product type binds a derived type to it's parent type (that may be a product
// type itself).
func NewProductTypeFromNameBaseAndParent(
	name string, tnat d.TyNative, tfnc TyFnc, parent HOTyped,
) ProdTypeFnc {

	return func() (string, d.TyNative, TyFnc, HOTyped) {

		return name, tnat, tfnc, parent
	}
}

func NewProductTypeFromTypeAndParent(
	hot, parent HOTyped,
) ProdTypeFnc {

	return func() (string, d.TyNative, TyFnc, HOTyped) {

		var tnat, tfnc = hot.TypeBase()

		return parent.TypeName() +
				" " +
				hot.TypeName(),
			tnat,
			tfnc,
			parent
	}
}

func NewProductTypeFromNameAndExpr(
	name string, expr Callable, parent HOTyped,
) ProdTypeFnc {

	return func() (string, d.TyNative, TyFnc, HOTyped) {

		return parent.TypeName() +
				" " +
				name,
			expr.TypeNat(),
			expr.TypeFnc(),
			parent
	}

}

func NewProductTypeAlias(
	name string, thot HOTyped,
) ProdTypeFnc {

	return func() (string, d.TyNative, TyFnc, HOTyped) {

		var parent = thot.TypeParent()
		var tnat, tfnc = parent.TypeBase()

		return parent.TypeName() +
				" " +
				name,
			tnat,
			tfnc,
			parent
	}
}

func (t ProdTypeFnc) TypeName() string {

	var name, _, _, _ = t()

	return name
}

// string representation concatenates all parents names recursively, till first
// order type is reached.
func (t ProdTypeFnc) String() string {

	return t.TypeName() + " " + t.TypeParent().String()
}

// returns the parent type as instance of HOTyped interface
func (t ProdTypeFnc) TypeParent() HOTyped {

	var _, _, _, hot = t()

	return hot
}

// states that this is not a type root
func (t ProdTypeFnc) TypeRoot() bool { return false }

// returns the types native-/ & functional type flags
func (t ProdTypeFnc) TypeBase() (d.TyNative, TyFnc) {

	var _, tnat, tfnc, _ = t()

	return tnat, tfnc
}

// yields the 'Type' Flag
func (t ProdTypeFnc) TypeFnc() TyFnc { return Type | TypeProduct }

// yields the 'd.Flag' Flag
func (t ProdTypeFnc) TypeNat() d.TyNative { return d.Flag }

// derived types check first arguments base type to match, after passing any
// existing further arguments on to the types parent type recursively. in case
// all arguments are evaluated true by parent, parents parent‥., first argument
// is matched against this instances base type. and result is returned. every
// other outcome results in false as return value.
func (t ProdTypeFnc) TypeMatch(args ...Callable) bool { return matchProductType(t, args...) }

func matchProductType(t ProdTypeFnc, args ...Callable) bool {

	// no arguments can't be true
	if len(args) == 0 {

		return false
	}

	// fetch first argument
	var arg = args[0]

	// if there are arguments remaining, pass them on to parent types to
	// evaluate those first
	if len(args) > 1 {

		// reassign remaining arguments
		args = args[1:]

		// if evaluated to false
		if !t.TypeParent().TypeMatch(castArgsHO(args...)...) {

			// type check failed
			return false
		}
	}

	// all arguments except the first one have been tested against parent
	// types and checked out to match. test first argument against native-/
	// & functional type of this type instance.
	var nat, fnc = t.TypeBase()

	// assign type matching functions
	var mn, mf = nat.Flag().Match, fnc.Flag().Match

	// if parent type happens to match remaining arguments, test the first
	// argument against this instances native-/& functional types
	if mn(arg.TypeNat()) && mf(arg.TypeFnc()) {

		// return a match, if both flags match
		return true
	}

	// return false if any type check failed
	return false
}

// call forwards the types match method function
func (t ProdTypeFnc) Call(args ...Callable) Callable {
	return NewFromData(d.BoolVal(t.TypeMatch(args...)))
}

// eval functionalizes arguments to call the call method
func (t ProdTypeFnc) Eval(args ...d.Native) d.Native {

	return t.Call(functionalizeNatives(args...)...).Eval()
}

////////////////////////////////////////////////////////////////////////////////////////
//// TYPE CONSTRUCTOR
///
// type constructor function takes construction arguments of type callable,
// either empty-/, or instanciated with values. if those construction arguments
// // match, as indicated by boolean, evaluation yields a new higher order
// type.
//
// called without arguments, type constructor yields its parental higher order
// type component.

// new type constructor wraps a type constructor function together with a
// higher order type.
func NewTypeConstructor(hot HOTyped, derive TypeConstructor) TypeConstructor {

	return func(args ...HOTyped) (HOTyped, bool) {

		// return type constructors higher orde baser type when called
		// without arguments
		if len(args) == 0 {

			return derive()
		}

		if derived, ok := derive(args...); ok {

			return derived, ok
		}

		return NewTypeError(), false
	}
}

func (c TypeConstructor) TypeHigherOrder() HOTyped {

	var hot, ok = c()

	if ok {
		return hot
	}

	return NewTypeError()
}

func (c TypeConstructor) String() string {

	if hot, ok := c(); ok {

		return hot.String()
	}

	return "stringer error in type constructor"
}

func (c TypeConstructor) TypeRoot() bool {

	if hot, ok := c(); ok {

		return hot.TypeRoot()
	}

	return false
}

func (c TypeConstructor) TypeParent() HOTyped {

	if hot, ok := c(); ok {

		return hot.TypeParent()
	}

	return NewTypeError()
}

func (c TypeConstructor) TypeBase() (d.TyNative, TyFnc) {

	if hot, ok := c(); ok {

		return hot.TypeBase()
	}

	return NewTypeError().TypeBase()
}

func (c TypeConstructor) TypeName() string {

	if hot, ok := c(); ok {

		return hot.TypeName()
	}

	return "type constructor name error"
}

func (c TypeConstructor) TypeNat() d.TyNative {

	tnat, _ := c.TypeBase()

	return tnat
}

func (c TypeConstructor) TypeFnc() TyFnc {

	_, tfnc := c.TypeBase()

	return tfnc
}

func (c TypeConstructor) TypeMatch(args ...HOTyped) bool {

	var _, ok = c(args...)

	return ok
}

func (c TypeConstructor) Call(args ...Callable) Callable {

	var hot, ok = c(castArgsHO(args...)...)

	// returns pair with fields containing either new generated higher
	// order type and true, or error/nil instance and false.
	return NewPair(hot, NewFromData(d.BoolVal(ok)))
}

// eval feeds call
func (c TypeConstructor) Eval(args ...d.Native) d.Native {

	return c.Call(functionalizeNatives(args...)...).Eval()
}

////////////////////////////////////////////////////////////////////////////////////////
//// DATA CONSTRUCTOR
///
func NewDataConstructor(dataCon DataConstructor) DataConstructor {

	return func(args ...Callable) (Callable, bool) {

		return dataCon(args...)
	}
}

func (c DataConstructor) TypeHigherOrder() HOTyped {

	if tc, ok := c(); ok {

		if typeCon, ok := tc.(TypeConstructor); ok {

			return typeCon.TypeHigherOrder()
		}
	}

	return NewTypeError()
}

func (c DataConstructor) String() string {

	if hot, ok := c(); ok {

		return hot.String()
	}

	return "stringer error in type constructor"
}

func (c DataConstructor) TypeRoot() bool {

	return c.TypeHigherOrder().TypeRoot()

}

func (c DataConstructor) TypeParent() HOTyped {

	return c.TypeHigherOrder().TypeParent()
}

func (c DataConstructor) TypeBase() (d.TyNative, TyFnc) {

	return c.TypeHigherOrder().TypeBase()
}

func (c DataConstructor) TypeName() string {

	return c.TypeHigherOrder().TypeName()
}

func (c DataConstructor) TypeNat() d.TyNative {

	tnat, _ := c.TypeBase()

	return tnat
}

func (c DataConstructor) TypeFnc() TyFnc {

	_, tfnc := c.TypeBase()

	return tfnc
}

func (c DataConstructor) TypeMatch(args ...HOTyped) bool {

	return c.TypeHigherOrder().TypeMatch(args...)
}

func (c DataConstructor) Call(args ...Callable) Callable {

	var hot, ok = c(args...)

	// returns pair with fields containing either new generated higher
	// order type and true, or error/nil instance and false.
	return NewPair(hot, NewFromData(d.BoolVal(ok)))
}

// eval feeds call
func (c DataConstructor) Eval(args ...d.Native) d.Native {

	return c.Call(functionalizeNatives(args...)...).Eval()
}

func castArgsHO(args ...Callable) []HOTyped {

	var result = []HOTyped{}

	for _, arg := range args {

		if hot, ok := arg.(HOTyped); ok {

			result = append(result, hot)
		}
	}

	return result
}

////////////////////////////////////////////////////////////////////////////////////////
//// TYPED VARIADIC EXPRESSION
///
// a typed expression is a callable instance of an higher order type
// allocate new typed expression from callable and type constructor
func NewTypedVariadicExpression(
	dataCon DataConstructor,
	typeCon TypeConstructor,
) TypedVariadicExpression {

	// return the typed expresssion
	return func(args ...Callable) Callable {

		var ok bool
		var result Callable
		var hot = typeCon.TypeHigherOrder()

		// call without arguments‥.
		if len(args) == 0 {

			if hot, ok := typeCon(); ok {

				typeCon = NewTypeConstructor(
					hot,
					typeCon,
				)
			}

			if result, ok = dataCon(); !ok {

				result = NewNoOp()
			}
		}

		if len(args) > 0 {

			// call the the type constructor, cast arguments as HOTyped and
			// pass to type constructor, to yield derived higher
			// order type.
			if hot, ok = typeCon(castArgsHO(args...)...); ok {

				// create new type constructor from new derived higher
				// order type and constructor function constructing it
				typeCon = NewTypeConstructor(
					hot,
					typeCon,
				)
			}

			// compute result of applying dataconstructor to arguments. if
			// computation fails to return a valid result, as indicated by
			// ok set to 'false', return vector of arguments instead of
			// result and boolean false to indicate thet no valid result
			// has been found yet and provide the neccessary arguments for
			// further recursion.
			if result, ok = dataCon(args...); !ok {

				result = NewNoOp()
			}
		}

		// create new typed expression from result yielded by applying
		// arguments to the expression, and new type constructor as
		// constructed from parental and type derived from.it.
		return DataConstructor(

			func(args ...Callable) (Callable, bool) {

				if len(args) == 0 {

					return typeCon, true
				}

				return result.Call(args...), ok
			})
	}
}

// calls expression empty and returns the type constructor
func (e TypedVariadicExpression) TypeConstructor() TypeConstructor {

	if con, ok := e().(TypeConstructor); ok {

		return con
	}

	return func(...HOTyped) (HOTyped, bool) { return NewTypeError(), false }
}

// calls the type contructor empty to yield it's higher order type
func (e TypedVariadicExpression) TypeHigherOrder() HOTyped {

	var hot, _ = e.TypeConstructor()()

	return hot
}

// string function to implement callable
func (e TypedVariadicExpression) String() string { return e.TypeHigherOrder().String() }

func (e TypedVariadicExpression) Call(args ...Callable) Callable { return e(args...) }

func (e TypedVariadicExpression) Eval(args ...d.Native) d.Native {
	return e(functionalizeNatives(args...)...)
}

func (e TypedVariadicExpression) TypeNat() d.TyNative {
	tnat, _ := e.TypeHigherOrder().TypeBase()
	return tnat
}

func (e TypedVariadicExpression) TypeFnc() TyFnc {
	_, tfnc := e.TypeHigherOrder().TypeBase()
	return tfnc
}

////////////////////////////////////////////////////////////////////////////////////////
//// TYPED DATA EXPRESSION
///
// typed constant expression has to allways return it's type constructor
func NewTypedData(data DataVal, typecon TypeConstructor) TypedDataValue {

	return func() (DataVal, TypeConstructor) { return data, typecon }
}

// calls expression empty and returns the type constructor
func (e TypedDataValue) TypeConstructor() TypeConstructor { _, con := e(); return con }

// calls the type contructor empty to yield it's higher order type
func (e TypedDataValue) TypeHigherOrder() HOTyped {

	if hot, ok := e.TypeConstructor()(); ok {
		return hot
	}

	return NewTypeError()

}

// call returns the yielded constant
func (e TypedDataValue) Call(...Callable) Callable { val, _ := e(); return val }

// eval returns yielded constant as native type
func (e TypedDataValue) Eval(...d.Native) d.Native { return e.Call().Eval() }

// string function to implement callable
func (e TypedDataValue) String() string { return e.Call().String() }

func (e TypedDataValue) TypeNat() d.TyNative { tnat, _ := e.TypeHigherOrder().TypeBase(); return tnat }

func (e TypedDataValue) TypeFnc() TyFnc { _, tfnc := e.TypeHigherOrder().TypeBase(); return tfnc }

////////////////////////////////////////////////////////////////////////////////////////
//// TYPED CONSTANT EXPRESSION
///
// typed constant expression has to allways return it's type constructor
func NewTypedConstantExpr(constant ConstantExpr, typecon TypeConstructor) TypedConstantExpression {

	return func() (ConstantExpr, TypeConstructor) { return constant, typecon }
}

// calls expression empty and returns the type constructor
func (e TypedConstantExpression) TypeConstructor() TypeConstructor { _, con := e(); return con }

// calls the type contructor empty to yield it's higher order type
func (e TypedConstantExpression) TypeHigherOrder() HOTyped {

	if hot, ok := e.TypeConstructor()(); ok {
		return hot
	}

	return NewTypeError()

}

// call returns the yielded constant
func (e TypedConstantExpression) Call(...Callable) Callable { val, _ := e(); return val }

// eval returns yielded constant as native type
func (e TypedConstantExpression) Eval(...d.Native) d.Native { return e.Call().Eval() }

// string function to implement callable
func (e TypedConstantExpression) String() string { return e.Call().String() }

func (e TypedConstantExpression) TypeNat() d.TyNative {
	tnat, _ := e.TypeHigherOrder().TypeBase()
	return tnat
}

func (e TypedConstantExpression) TypeFnc() TyFnc {
	_, tfnc := e.TypeHigherOrder().TypeBase()
	return tfnc
}

////////////////////////////////////////////////////////////////////////////////////////
//// TYPED NARY EXPRESSION
///
// typed nary expression has a known arity and can be exactly-, under-, or
// oversatisfyed in number of passed arguments. typed nary expression returns either an nary
func NewTypedNaryExpression(expr NaryExpr, typecon TypeConstructor, ari int) TypedNaryExpression {

	var arity = Arity(int8(ari))

	return func(args ...Callable) Callable {

		var anum = len(args)

		// no arguments passed, return typeconstructor & arity
		if anum == 0 {

			return NewPair(typecon, arity)
		}

		// allocate result value
		var constructor = typecon
		var result Callable

		// arity exactly satisfyed by number of passed arguments
		//
		// if number of arguments equals arity. return a typed constant
		// as typed expression and ab arity of zero
		if anum == int(arity) {

			// return a constant expression, since all arguments have
			// been passed.
			var value = ConstantExpr(func() Callable {

				return expr(args...)
			})

			// if type constructor yields value‥.
			if con, ok := typecon(castArgsHO(args...)...); ok {

				// and value is a type constructor...
				if tcon, ok := con.(TypeConstructor); ok {

					// use it as resulting constructor
					constructor = tcon
				}
			}

			// compose resulting type expression from resulting
			// value & constructed type
			result = NewTypedConstantExpr(
				value,
				constructor,
			)

			// set arity to zero, there are no more arguments to
			// expect
			arity = Arity(0)
		}

		// arity is over-satisfyed by number of passed arguments
		//
		// return result as head, followed by abundant arguments as
		// tail of consumeable list
		if anum > int(arity) {

			var value = NewList(ConstantExpr(func() Callable {

				return expr(args...)

			})).Con(args...)

			if con, ok := typecon(castArgsHO(args[:ari]...)...); ok {

				if tcon, ok := con.(TypeConstructor); ok {

					constructor = tcon
				}
			}

			result = NewTypedNaryExpression(
				value.Call,
				constructor,
				anum-int(arity),
			)
		}

		// arity is under-satisfyed by arguments passed in call
		//
		// return a partialy applyed function with reduced arity
		if anum < int(arity) {

			// reduce arity by number of arguments
			arity = arity - Arity(anum)

			// create partial constructor passing all available arguments to type constructor
			if con, ok := typecon(castArgsHO(args...)...); ok {

				if tcon, ok := con.(TypeConstructor); ok {

					constructor = tcon
				}
			}

			// create partialy applyed expression  by creating a
			// call continuation, that expects the missing
			// arguments and has it's arity reduced by number of
			// arguments that got passed allready and a constructor
			// yielded from applying the arguments to the initially
			// passed type constructor.
			var value = DataConstructor(

				func(missing ...Callable) Callable {

					return NewTypedNaryExpression(

						NaryExpr(func(missing ...Callable) Callable {

							return expr(args...).Call(missing...)
						}),

						constructor,

						int(arity),
					)
				})

			result = NewTypedVariadicExpression(
				value,
				constructor,
			)
		}

		// return paired value containing expression resulting from
		// applying expression to arguments, and its remaining arity.
		return NewPair(result, arity)
	}
}

func (e TypedNaryExpression) Arity() Arity {
	return e().(PairVal).Right().(Arity)
}

// calls expression empty and returns the type constructor
func (e TypedNaryExpression) TypeConstructor() TypeConstructor {
	return e().(PairVal).Left().(TypeConstructor)

}

// calls the type contructor empty to yield it's higher order type
func (e TypedNaryExpression) TypeHigherOrder() HOTyped {

	if hot, ok := e.TypeConstructor()(); ok {
		return hot
	}

	return NewTypeError()

}

// call returns the yielded constant
func (e TypedNaryExpression) Call(args ...Callable) Callable { return e(args...) }

// eval returns yielded constant as native type
func (e TypedNaryExpression) Eval(args ...d.Native) d.Native { return e.Call().Eval(args...) }

// string function to implement callable
func (e TypedNaryExpression) String() string { return e.Call().String() }

func (e TypedNaryExpression) TypeNat() d.TyNative {
	tnat, _ := e.TypeHigherOrder().TypeBase()
	return tnat
}

func (e TypedNaryExpression) TypeFnc() TyFnc {
	_, tfnc := e.TypeHigherOrder().TypeBase()
	return tfnc
}

////////////////////////////////////////////////////////////////////////////////////////
//// NARY TYPED EXPRESSION WITH PROPERTYS
///
//
func NewProperTypeExpression(
	expr Callable,
	con TypeConstructor,
	arity int,
	props Propertys,
) TypedProperExpression {

	return func(args ...Callable) Callable {

		var nexpr = NewTypedNaryExpression(expr.Call, con, arity)

		if len(args) == 0 {

			return NewVector(nexpr, Arity(arity), props)
		}

		return NewProperTypeExpression(
			ConstantExpr(func() Callable { return expr.Call(args...) }),
			NewTypeConstructor(con.TypeHigherOrder(), con),
			arity,
			props,
		)
	}

}

func (e TypedProperExpression) TypeConstructor() TypeConstructor {

	if vec, ok := e().(VecVal); ok {

		if vec.Len() > 0 {

			if con, ok := vec.Slice()[0].(TypeConstructor); ok {

				return con
			}

		}
	}

	return NewTypeConstructor(
		NewTypeError(),
		func(...HOTyped) (HOTyped, bool) {
			return NewTypeError(), false
		})
}

func (e TypedProperExpression) Arity() Arity {

	if vec, ok := e().(VecVal); ok {

		if vec.Len() > 1 {

			if arity, ok := vec.Slice()[1].(Arity); ok {

				return arity
			}
		}
	}

	return Arity(0)
}

func (e TypedProperExpression) Propertys() Propertys {

	if vec, ok := e().(VecVal); ok {

		if vec.Len() > 2 {

			if props, ok := vec.Slice()[2].(Propertys); ok {

				return props
			}
		}
	}

	return Propertys(0)
}

// calls the type contructor empty to yield it's higher order type
func (e TypedProperExpression) TypeHigherOrder() HOTyped {

	if hot, ok := e.TypeConstructor()(); ok {
		return hot
	}

	return NewTypeError()

}

// call returns the yielded constant
func (e TypedProperExpression) Call(args ...Callable) Callable { return e(args...) }

// eval returns yielded constant as native type
func (e TypedProperExpression) Eval(...d.Native) d.Native { return e.Call().Eval() }

// string function to implement callable
func (e TypedProperExpression) String() string { return e.Call().String() }

func (e TypedProperExpression) TypeNat() d.TyNative {
	tnat, _ := e.TypeHigherOrder().TypeBase()
	return tnat
}

func (e TypedProperExpression) TypeFnc() TyFnc {
	_, tfnc := e.TypeHigherOrder().TypeBase()
	return tfnc
}

//////////////////////////////////////////////////////////////////////
//// SEMANTIC CALL PROPERTYS
///
type Arity d.Uint8Val

//go:generate stringer -type Arity
const (
	Nullary Arity = 0 + iota
	Unary
	Binary
	Ternary
	Quaternary
	Quinary
	Senary
	Septenary
	Octonary
	Novenary
	Denary
)

func (a Arity) Eval(...d.Native) d.Native { return d.Int8Val(a) }

func (a Arity) Call(...Callable) Callable { return NewFromData(a.Eval()) }

func (a Arity) Int() int            { return int(a) }
func (a Arity) Flag() d.BitFlag     { return d.BitFlag(a) }
func (a Arity) TypeNat() d.TyNative { return d.Flag }
func (a Arity) TypeFnc() TyFnc      { return HigherOrder }

func (a Arity) Match(arg Arity) bool { return a == arg }

// properys relevant for application
type Propertys d.Uint8Val

//go:generate stringer -type Propertys
const (
	Default Propertys = 0
	PostFix Propertys = 1
	InFix   Propertys = 1 + iota
	// ⌐: PreFix
	Atomic
	// ⌐: Thunk
	Eager
	// ⌐: Lazy
	RightBound
	// ⌐: Left_Bound
	Mutable
	// ⌐: Imutable
	SideEffect
	// ⌐: Pure
	Primitive
	// ⌐: Parametric
)

func (p Propertys) PostFix() bool    { return p.Flag().Match(PostFix.Flag()) }
func (p Propertys) InFix() bool      { return !p.Flag().Match(PostFix.Flag()) }
func (p Propertys) Atomic() bool     { return p.Flag().Match(Atomic.Flag()) }
func (p Propertys) Thunk() bool      { return !p.Flag().Match(Atomic.Flag()) }
func (p Propertys) Eager() bool      { return p.Flag().Match(Eager.Flag()) }
func (p Propertys) Lazy() bool       { return !p.Flag().Match(Eager.Flag()) }
func (p Propertys) RightBound() bool { return p.Flag().Match(RightBound.Flag()) }
func (p Propertys) LeftBound() bool  { return !p.Flag().Match(RightBound.Flag()) }
func (p Propertys) Mutable() bool    { return p.Flag().Match(Mutable.Flag()) }
func (p Propertys) Imutable() bool   { return !p.Flag().Match(Mutable.Flag()) }
func (p Propertys) SideEffect() bool { return p.Flag().Match(SideEffect.Flag()) }
func (p Propertys) Pure() bool       { return !p.Flag().Match(SideEffect.Flag()) }
func (p Propertys) Primitive() bool  { return p.Flag().Match(Primitive.Flag()) }
func (p Propertys) Parametric() bool { return !p.Flag().Match(Primitive.Flag()) }

func (p Propertys) TypeNat() d.TyNative { return d.Flag }
func (p Propertys) TypeFnc() TyFnc      { return HigherOrder }

func (p Propertys) Flag() d.BitFlag { return d.BitFlag(uint64(p)) }

func FlagToProp(flag d.BitFlag) Propertys { return Propertys(uint8(flag.Uint())) }

func (p Propertys) Eval(a ...d.Native) d.Native { return p }

func (p Propertys) Call(args ...Callable) Callable { return p }

func (p Propertys) MatchProperty(arg Propertys) bool {
	if p&arg != 0 {
		return true
	}
	return false
}

func (p Propertys) Match(flag d.BitFlag) bool { return p.Flag().Match(flag) }
func (p Propertys) Print() string {

	var flags = p.Flag().Decompose()
	var str string
	var l = len(flags)

	if l > 1 {
		for i, flag := range flags {
			str = str + FlagToProp(flag).String()
			if i < l-1 {
				str = str + " "
			}
		}
	}

	return p.String()
}

// type TyFnc d.BitFlag
// encodes the kind of functional data as bitflag
type TyFnc d.BitFlag

func (t TyFnc) FlagType() int8                 { return 2 }
func (t TyFnc) TypeFnc() TyFnc                 { return Type }
func (t TyFnc) TypeNat() d.TyNative            { return d.Flag }
func (t TyFnc) Call(args ...Callable) Callable { return t.TypeFnc() }
func (t TyFnc) Eval(args ...d.Native) d.Native { return t.TypeNat() }
func (t TyFnc) Flag() d.BitFlag                { return d.BitFlag(t) }
func (t TyFnc) Uint() uint                     { return d.BitFlag(t).Uint() }
