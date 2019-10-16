# Type System Design

Native as well as internal types implement the typed interface. data & functions therefore share identical interfaces.

```go
type Typed interface {
        Type() d.Typed
	KindOfType
	NameTyped
	Flagged
	Matched
	Stringer
}
type KindOfType interface {
	Kind() d.Uint8Val
}

type NativeTyped interface {
	TypeNat() d.TyNat
}

type FunctionTyped interface {
	TypeFnc() TyFnc
}

```

## Type Construction

the composed type TyComp is implemented by a slice of data.Typed instances, storing type identity, return-/ & argument types:

  - Def(...d.Typed) TyComp generates and returns **type signatures**, or parts there of from n passed data.Typed instance arguments.
  - composed type fields assignments: [0 Ident, 1 Return, 2 Arguments, 3 Properties/Subtypes]
  - the composed type marker TyComp can contain other instances as well as instances of different **kinds of types** recursively:
    0. TyFnc   instances of function primitives, parts of type system included
    0. TyNat   instances of the data package implementing the native interface
    0. TySym   runtime defined symbol
    0. TyExp   type pattern can contain expressions (constraints, bounds)
    0. TyProp  call propertys like fixity, lazynes, etc‥.
    0. TyLex   lexical token

### parametric types:

  - the function package features six **agnostic types**:
    0. None                 func()
    0. Constant             func() Expression                   reflects expressions type fields
    0. Test, Trinary        func(...Expression) bool            knows if test, or trinary 
    0. Compare              func(...Expression) int             knows it is a comparator
    0. Function (untyped)   func(...Expression) Expression      reflects embedded expressions type
    0. Case                 ‥.                                  defines argument-/ and return-, reflects test type
    0. Switch               ‥.                                  reflects its enclosed cases types
    0. Maybe                func(...Expression) (Just|None)     defines the 'just' sub-types particular type
      0.0 Just Sub Type       func(...Expression) Expression
      0.0 None Sub Type       func(...Expression) Expression
    0. Option               func(...Expression) (Either|Or)     defines an 'either'-/ and 'or' sub-type
      0.0 Either Sub Type
      0.0 Or Sub Type

A particular sub-type, & value constructor will be defined for every instanciated permutation (unique in terms of argument-/ & return types), of each agnostic types.

  - a data constructor per **native type** in the functions package.
    0. Data Atomic    func(...d.Native) d.Native  define the enclosed native argument-/ & return types
    0. Data Slice     ‥.                          ‥.
    0. Data Go Slice  ‥.                          ‥.
    0. Data Pair      ‥.                          ‥.
    0. Data Map       ‥.                          ‥.
    0. Data Function  ‥.                          ‥.

  - a value-/ and type constructor per **product-/ sum-type** in the functions package.
    0. Function     func(...Expression) Expression                          all fields defined at runtime
    0. Generator    ‥.                                                      ‥.
    0. Accumulator  ‥.                                                      ‥.
    0. Tuple        func(...Expression) []Expression                        ‥.
    0. Record       func(...Expression) (Expression, Record)                ‥.
    0. Vector       func(...Expression) (Expression, Vector)                ‥.
    0. List         func(...Expression) (Expression, List)                  ‥.
    0. Monad        func(...Expression) (Expression, Monad)                 ‥.
    0. Sequence     func(...Expression) (Expression, Sequence)              ‥.
    0. Applicable   func(...Expression) (Expression, Applicable)            ‥.
    0. Polymorphic  func(...Expression) (Expression, Poly-Def)              ‥.
    0. Enumerable   func(...Expression) (Expression, (Numeral,(Enum-Def)))  ‥.


A particular sub-type and value constructor will be defined for every instanciated permutation (unique in terms of argument-/ & return types), of each agnosic, data-/, sum-/ and/or product-type. 

Type constructors keep references to all instanciated sub-type value constructors. sub-types keep references to their parent types
