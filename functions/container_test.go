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

func TestPartial(t *testing.T) {
	var strconc = DefinePartial("String Concat",
		NativeExpr(func(args ...Native) Native {
			var str string
			for n, arg := range args {
				str = str + arg.String()
				if n < len(args)-1 {
					str = str + " "
				}
			}
			return NewData(d.StrVal(str))
		}),
		NewData(d.NewNull(d.String)).Type(),
		NewData(d.NewNull(d.String)),
		NewData(d.NewNull(d.String)),
		NewData(d.NewNull(d.String)),
	)

	var r0 = strconc(NewData(d.StrVal("0")))

	var r1 = strconc(NewData(d.StrVal("0")), NewData(d.StrVal("1")))

	var r2 = strconc(NewData(d.StrVal("0")), NewData(d.StrVal("1")),
		NewData(d.StrVal("2")))

	var r3 = strconc(NewData(d.StrVal("0")), NewData(d.StrVal("1")),
		NewData(d.StrVal("2")), NewData(d.StrVal("3")))

	var r4 = strconc(NewData(d.StrVal("0")), NewData(d.StrVal("1")),
		NewData(d.StrVal("2")), NewData(d.StrVal("3")),
		NewData(d.StrVal("4")))

	var r5 = strconc(NewData(d.StrVal("0")), NewData(d.StrVal("1")),
		NewData(d.StrVal("2")), NewData(d.StrVal("3")),
		NewData(d.StrVal("4")), NewData(d.StrVal("5")))

	var r6 = strconc(NewData(d.StrVal("0")), NewData(d.StrVal("1")),
		NewData(d.StrVal("2")), NewData(d.StrVal("3")),
		NewData(d.StrVal("4")), NewData(d.StrVal("5")),
		NewData(d.StrVal("6")))

	var r7 = strconc(NewData(d.StrVal("0")), NewData(d.StrVal("1")),
		NewData(d.StrVal("2")), NewData(d.StrVal("3")),
		NewData(d.StrVal("4")), NewData(d.StrVal("5")),
		NewData(d.StrVal("6")), NewData(d.StrVal("7")))

	var r8 = strconc(NewData(d.StrVal("0")), NewData(d.StrVal("1")),
		NewData(d.StrVal("2")), NewData(d.StrVal("3")),
		NewData(d.StrVal("4")), NewData(d.StrVal("5")),
		NewData(d.StrVal("6")), NewData(d.StrVal("7")),
		NewData(d.StrVal("8")))

	fmt.Println(r0)
	fmt.Println(r1)
	fmt.Println(r2)
	fmt.Println(r3)
	fmt.Println(r4)
	fmt.Println(r5)
	fmt.Println(r6)
	fmt.Println(r7)
	fmt.Println(r8)

	fmt.Printf("string concat type name: %s\n", strconc.TypeName())
	fmt.Printf("result: %s typed: %s name: %s\n", r0, r0.Type(), r0.TypeName())
	fmt.Printf("result: %s typed: %s name: %s\n", r1, r1.Type(), r1.TypeName())
	fmt.Printf("result: %s typed: %s name: %s\n", r2, r2.Type(), r2.TypeName())
	fmt.Printf("result: %s typed: %s name: %s\n", r3, r3.Type(), r3.TypeName())
	fmt.Printf("result: %s typed: %s name: %s\n", r4, r4.Type(), r4.TypeName())
	fmt.Printf("result: %s typed: %s name: %s\n", r5, r5.Type(), r5.TypeName())
	fmt.Printf("result: %s typed: %s name: %s\n", r6, r6.Type(), r6.TypeName())
	fmt.Printf("result: %s typed: %s name: %s\n", r7, r7.Type(), r7.TypeName())
	fmt.Printf("result: %s typed: %s name: %s\n", r8, r8.Type(), r8.TypeName())

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
		NewData(d.NewUboxNull(d.String)),
		NewData(d.StrVal("")),
		NewData(d.IntVal(0)),
		NewData(d.UintVal(0)),
	)

	var sv0 = strvec(NewData(d.StrVal("0")))

	var sv1 = strvec(NewData(d.StrVal("0")), NewData(d.IntVal(1)))

	var sv2 = strvec(NewData(d.StrVal("0")), NewData(d.IntVal(1)),
		NewData(d.UintVal(2)))

	var sv3 = strvec(NewData(d.StrVal("0")), NewData(d.IntVal(1)),
		NewData(d.UintVal(2)), NewData(d.StrVal("3")))

	var sv4 = strvec(NewData(d.StrVal("0")), NewData(d.IntVal(1)),
		NewData(d.UintVal(2)), NewData(d.StrVal("3")),
		NewData(d.IntVal(4)))

	var sv5 = strvec(NewData(d.StrVal("0")), NewData(d.IntVal(1)),
		NewData(d.UintVal(2)), NewData(d.StrVal("3")),
		NewData(d.IntVal(4)), NewData(d.UintVal(5)))

	var sv6 = strvec(NewData(d.StrVal("0")), NewData(d.IntVal(1)),
		NewData(d.UintVal(2)), NewData(d.StrVal("3")),
		NewData(d.IntVal(4)), NewData(d.UintVal(5)),
		NewData(d.StrVal("6")))

	var sv7 = strvec(NewData(d.StrVal("0")), NewData(d.IntVal(1)),
		NewData(d.UintVal(2)), NewData(d.StrVal("3")),
		NewData(d.IntVal(4)), NewData(d.UintVal(5)),
		NewData(d.StrVal("6")), NewData(d.IntVal(7)))

	var sv8 = strvec.Call(NewData(d.StrVal("0")), NewData(d.IntVal(1)),
		NewData(d.UintVal(2)), NewData(d.StrVal("3")),
		NewData(d.IntVal(4)), NewData(d.UintVal(5)),
		NewData(d.StrVal("6")), NewData(d.IntVal(7)),
		NewData(d.UintVal(8)))

	fmt.Printf(
		"\nstring vector: %s type: %s elem name: %s elem type: %s"+
			" slice: %s element: %s\n\n",
		sv5, sv5.(VecCol)()[0].TypeNat().String(),
		sv5.(VecCol)()[0].(NativeUbox)().TypeName(),
		sv5.(VecCol)()[0].(NativeUbox)().TypeNat(),
		sv5.(VecCol)()[0].(NativeUbox),
		sv5.(VecCol)()[0].(NativeUbox),
	)

	fmt.Printf("string vector: %s type: %s type name: %s\n",
		strvec, strvec.Type(), strvec.TypeName())
	fmt.Printf("string vector: %s type: %s type name: %s\n",
		sv0.Call(), sv0.TypeFnc(), sv0.TypeName())
	fmt.Printf("string vector: %s type: %s type name: %s\n",
		sv1.Call(), sv1.TypeFnc(), sv1.TypeName())
	fmt.Printf("string vector: %s type: %s type name: %s\n",
		sv2.Call(), sv2.TypeFnc(), sv2.TypeName())
	fmt.Printf("string vector: %s type: %s type name: %s\n",
		sv3.Call(), sv3.TypeFnc(), sv3.TypeName())
	fmt.Printf("string vector: %s type: %s type name: %s\n",
		sv4.Call(), sv4.TypeFnc(), sv4.TypeName())
	fmt.Printf("string vector: %s type: %s type name: %s\n",
		sv5.Call(), sv5.TypeFnc(), sv5.TypeName())
	fmt.Printf("string vector: %s type: %s type name: %s\n",
		sv6.Call(), sv6.TypeFnc(), sv6.TypeName())
	fmt.Printf("string vector: %s type: %s type name: %s\n",
		sv7.Call(), sv7.TypeFnc(), sv7.TypeName())
	fmt.Printf("string vector: %s type: %s type name: %s\n",
		sv8.Call(), sv8.TypeFnc(), sv8.Type().Name())

	fmt.Printf("vector head: %s head type name: %s type name: %s\n",
		sv8.(VecCol).String(), sv8.(VecCol).Head().TypeName(), sv8.(VecCol).TypeName())

}

func TestTuple(t *testing.T) {
}

func TestRecordType(t *testing.T) {
}
