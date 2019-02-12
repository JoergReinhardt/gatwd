/*
DATA PACKAGE

  interfaces of primary data types

  the data package implements primary data types as aliases on go's native
  types. this designg decission was choosen to provide unboxed values for code
  based on gatwd, while providing dynamic type inference, auto conversion and
  parametric algebraic types (in the functions package).

  golang itself provides a feature rich reflexion standard library, as well as
  the syntatctic constructs of the type switch, assertion and conversion, all
  of which which could be used to implemet the intendet features, but that
  would come with certain drawbacks:

  - golang reflection is deemed to be slow (no idea if that is true at all‥.
    so not a particular strong argument, which is no excuse for not eliminating
    it ;)

  - golangs reflection library is rather complex due to the rich set of
    features provided, most of which aren't needed for gatwds purpose‥.

  - type switch, type assertion & conversion, deal with type, but don't allow
    to treat the type itself, as if it where just a value (for very good resons
    though‥.).

  this is fine for every day use, especially when combined with interfaces, but
  when implementing a type system, doubly so, when it's supposed to be
  parametric, it is highly desireable to be able to handle, mutate and
  dynamicly create such internal types.


  BitFlag:

  to 'replace' golangs type switch, type aliases provide a bit flag constant
  for identification and comparison. that brings additional merrit's:

    - flag value can be sorted by precedence to later choose in which way a
      type needs to be auto converted, or as what other type it needs to be
      casted.

    - flags can be 'OR' concatenated to define & match against sets of flags,
      whithout adding data, or cpu cycles.

    - types can be treated, stored, serialized as simple uint instances.

    - each instance of a type also get's access to it's types string
      representation (aka name of the internal type).


  the types defined and bit-flag constants provided (native → alias [constant]):

    - one alias per go native type:
      folowing the naming convention int → IntVal [Int]

    - one slice type alias per go native type:
      folowing the naming convention []int → IntVec [Vector|Int]

    - aliases for types of the math/big package

      - BigIntVal → big.Int [BigInt]

      - BigFltVal → big.Float [BigFlt]

      - RatioVal  → big.Rat [Ratio]

    - aliases for types of the time package

      - TimeVal   → time.Time [Time]

      - DuraVal   → time.Duration [Duration]

    - NilVal    → struct{} [Nil]

    - ErrorVal  → struct{ error } [Error]

    - PairVal   → struct{ Primary, Primary } [Pair]

    - BitFlag   → BitFlag [Flag]

    - DataSlice → []Primary [Vector]

    - FlagSet   → []BitFlag [Vector|Flag]

    - SetString → map[StrVal]Primary [Set|String]

    - SetUint   → map[UintVal]Primary [Set|Uint]

    - SetInt    → map[IntVal]Primary [Set|Int]

    - SetFloat  → map[FloatVal]Primary [Set|Float]

    - SetFlag   → map[FlagVal]Primary [Set|Flag]


  further sets of constants are defined and named bit flag constants and
  concatenations there of concatenations, to connect the types and interfaces
  defined here semanticly to the higher order types and interfaces there of,
  defined by the functions package.
*/
package data
