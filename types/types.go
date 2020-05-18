package types

type tflg uint32

//go:generate stringer -type tflg
const (
	None tflg = 1<<iota - 1
	Truth
	Uint
	Int
	Flt
	Img
	Time
	Span
	Byte
	Rune
	Bytes
	String
	Vector
	Tuple
	List
	Fnc // map one classes objects onto another class
	Op  // operate on (take|return) objects of one class
	Just
	Either
	Or

	// TRUTH
	False bool = false
	True  bool = true

	Num = Uint | Int | Flt | Img // numeric
	Tmp = Time | Span            // temporal
	Bin = Byte | Bytes           // binary
	Txt = Rune | String          // textual
	Col = Vector | Tuple | List  // collection
	Par = Fnc | Op               // parameterized

	Maybe = Just | None
	Alter = Either | Or

	T = Truth | Num | Tmp | Bin | Txt | Col |
		Par | Maybe | Alter

	//  unit, neutral & sign
	Neg  int = -1 // negative
	Zero int = 0  // ambivalent
	Unit int = 1  // positive
)
