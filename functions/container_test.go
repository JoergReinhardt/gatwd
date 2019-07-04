package functions

import (
	"fmt"
	"testing"

	d "github.com/joergreinhardt/gatwd/data"
)

var intslice = []Expression{
	New(0), New(1), New(2), New(3), New(4), New(5), New(6), New(7), New(8),
	New(9), New(0), New(134), New(8566735), New(4534), New(3445),
	New(76575), New(2234), New(45), New(7646), New(64), New(3), New(314),
}

var intkeys = []Expression{New("zero"), New("one"), New("two"), New("three"),
	New("four"), New("five"), New("six"), New("seven"), New("eight"), New("nine"),
	New("ten"), New("eleven"), New("twelve"), New("thirteen"), New("fourteen"),
	New("fifteen"), New("sixteen"), New("seventeen"), New("eighteen"),
	New("nineteen"), New("twenty"), New("twentyone"),
}

func TestNary(t *testing.T) {
	var strconc = DefinePartial("String Concat",
		NativeExpr(func(args ...d.Native) d.Native {
			var str string
			for n, arg := range args {
				str = str + arg.String()
				if n < len(args)-1 {
					str = str + " "
				}
			}
			return NewNative(d.StrVal(str))
		}),
		NewNative(d.NewNull(d.String)),
		NewNative(d.NewNull(d.String)),
		NewNative(d.NewNull(d.String)),
		NewNative(d.NewNull(d.String)),
	)

	var r0 = strconc(NewNative(d.StrVal("0")))

	var r1 = strconc(NewNative(d.StrVal("0")), NewNative(d.StrVal("1")))

	var r2 = strconc(NewNative(d.StrVal("0")), NewNative(d.StrVal("1")),
		NewNative(d.StrVal("2")))

	var r3 = strconc(NewNative(d.StrVal("0")), NewNative(d.StrVal("1")),
		NewNative(d.StrVal("2")), NewNative(d.StrVal("3")))

	var r4 = strconc(NewNative(d.StrVal("0")), NewNative(d.StrVal("1")),
		NewNative(d.StrVal("2")), NewNative(d.StrVal("3")),
		NewNative(d.StrVal("4")))

	var r5 = strconc(NewNative(d.StrVal("0")), NewNative(d.StrVal("1")),
		NewNative(d.StrVal("2")), NewNative(d.StrVal("3")),
		NewNative(d.StrVal("4")), NewNative(d.StrVal("5")))

	var r6 = strconc(NewNative(d.StrVal("0")), NewNative(d.StrVal("1")),
		NewNative(d.StrVal("2")), NewNative(d.StrVal("3")),
		NewNative(d.StrVal("4")), NewNative(d.StrVal("5")),
		NewNative(d.StrVal("6")))

	var r7 = strconc(NewNative(d.StrVal("0")), NewNative(d.StrVal("1")),
		NewNative(d.StrVal("2")), NewNative(d.StrVal("3")),
		NewNative(d.StrVal("4")), NewNative(d.StrVal("5")),
		NewNative(d.StrVal("6")), NewNative(d.StrVal("7")))

	var r8 = strconc(NewNative(d.StrVal("0")), NewNative(d.StrVal("1")),
		NewNative(d.StrVal("2")), NewNative(d.StrVal("3")),
		NewNative(d.StrVal("4")), NewNative(d.StrVal("5")),
		NewNative(d.StrVal("6")), NewNative(d.StrVal("7")),
		NewNative(d.StrVal("8")))

	fmt.Println(r0.Call())
	fmt.Println(r1.Call())
	fmt.Println(r2.Call())
	fmt.Println(r3.Call())
	fmt.Println(r4.Call())
	fmt.Println(r5.Call())
	fmt.Println(r6.Call())
	fmt.Println(r7.Call())
	fmt.Println(r8.Call())

	fmt.Printf("string concat type name: %s\n", strconc.TypeName())
	fmt.Printf("typed: %s name: %s\n", r0.Type(), r0.TypeName())
	fmt.Printf("typed: %s name: %s\n", r1.Type(), r1.TypeName())
	fmt.Printf("typed: %s name: %s\n", r2.Type(), r2.TypeName())
	fmt.Printf("typed: %s name: %s\n", r3.Type(), r3.TypeName())
	fmt.Printf("typed: %s name: %s\n", r4.Type(), r4.TypeName())
	fmt.Printf("typed: %s name: %s\n", r5.Type(), r5.TypeName())
	fmt.Printf("typed: %s name: %s\n", r6.Type(), r6.TypeName())
	fmt.Printf("typed: %s name: %s\n", r7.Type(), r7.TypeName())
	fmt.Printf("typed: %s name: %s\n", r8.Type(), r8.TypeName())

	var strvec = DefinePartial("String Vector",
		NativeExpr(func(args ...d.Native) d.Native {
			if len(args) > 0 {
				var strs []d.Native
				for _, arg := range args {
					strs = append(strs, d.StrVal(arg.String()))
				}
				return d.NewSlice(strs...)
			}
			return d.NewNil()
		}),
		NewNative(d.NewUboxNull(d.String)),
		NewNative(d.StrVal("")),
		NewNative(d.IntVal(0)),
		NewNative(d.UintVal(0)),
	)

	var sv0 = strvec(NewNative(d.StrVal("0")))

	var sv1 = strvec(NewNative(d.StrVal("0")), NewNative(d.IntVal(1)))

	var sv2 = strvec(NewNative(d.StrVal("0")), NewNative(d.IntVal(1)),
		NewNative(d.UintVal(2)))

	var sv3 = strvec(NewNative(d.StrVal("0")), NewNative(d.IntVal(1)),
		NewNative(d.UintVal(2)), NewNative(d.StrVal("3")))

	var sv4 = strvec(NewNative(d.StrVal("0")), NewNative(d.IntVal(1)),
		NewNative(d.UintVal(2)), NewNative(d.StrVal("3")),
		NewNative(d.IntVal(4)))

	var sv5 = strvec(NewNative(d.StrVal("0")), NewNative(d.IntVal(1)),
		NewNative(d.UintVal(2)), NewNative(d.StrVal("3")),
		NewNative(d.IntVal(4)), NewNative(d.UintVal(5)))

	var sv6 = strvec(NewNative(d.StrVal("0")), NewNative(d.IntVal(1)),
		NewNative(d.UintVal(2)), NewNative(d.StrVal("3")),
		NewNative(d.IntVal(4)), NewNative(d.UintVal(5)),
		NewNative(d.StrVal("6")))

	var sv7 = strvec(NewNative(d.StrVal("0")), NewNative(d.IntVal(1)),
		NewNative(d.UintVal(2)), NewNative(d.StrVal("3")),
		NewNative(d.IntVal(4)), NewNative(d.UintVal(5)),
		NewNative(d.StrVal("6")), NewNative(d.IntVal(7)))

	var sv8 = strvec(NewNative(d.StrVal("0")), NewNative(d.IntVal(1)),
		NewNative(d.UintVal(2)), NewNative(d.StrVal("3")),
		NewNative(d.IntVal(4)), NewNative(d.UintVal(5)),
		NewNative(d.StrVal("6")), NewNative(d.IntVal(7)),
		NewNative(d.UintVal(8)))

	fmt.Printf(
		"\nstring vector: %s type: %s elem name: %s elem type: %s"+
			" slice: %s element: %s\n\n",
		sv5, sv5.(VecCol)()[0].TypeNat().String(),
		sv5.(VecCol)()[0].(NativeUbox)().TypeName(),
		sv5.(VecCol)()[0].(NativeUbox)().TypeNat(),
		sv5.(VecCol)()[0].(NativeUbox).Slice(),
		sv5.(VecCol)()[0].(NativeUbox).Slice()[1],
	)

	fmt.Printf("string vector: %s type name: %s\n", strvec, strvec.TypeName())
	fmt.Printf("string vector: %s type name: %s\n", sv0, sv0.TypeName())
	fmt.Printf("string vector: %s type name: %s\n", sv1, sv1.TypeName())
	fmt.Printf("string vector: %s type name: %s\n", sv2, sv2.TypeName())
	fmt.Printf("string vector: %s type name: %s\n", sv3, sv3.TypeName())
	fmt.Printf("string vector: %s type name: %s\n", sv4, sv4.TypeName())
	fmt.Printf("string vector: %s type name: %s\n", sv5, sv5.TypeName())
	fmt.Printf("string vector: %s type name: %s\n", sv6, sv6.TypeName())
	fmt.Printf("string vector: %s type name: %s\n", sv7, sv7.TypeName())
	fmt.Printf("string vector: %s type name: %s\n", sv8, sv8.TypeName())

}

func TestTuple(t *testing.T) {
}

func TestRecordType(t *testing.T) {
}
