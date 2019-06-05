package functions

import (
	"fmt"
	d "github.com/joergreinhardt/gatwd/data"
	"testing"
)

var intslice = []Callable{
	New(0), New(1), New(2), New(3), New(4), New(5), New(6), New(7), New(8),
	New(9), New(0), New(134), New(8566735), New(4534), New(3445),
	New(76575), New(2234), New(45), New(7646), New(64), New(3), New(314),
}

var intkeys = []Callable{New("zero"), New("one"), New("two"), New("three"),
	New("four"), New("five"), New("six"), New("seven"), New("eight"), New("nine"),
	New("ten"), New("eleven"), New("twelve"), New("thirteen"), New("fourteen"),
	New("fifteen"), New("sixteen"), New("seventeen"), New("eighteen"),
	New("nineteen"), New("twenty"), New("twentyone"),
}

func TestTupleConstruction(t *testing.T) {

	var tupleType = NewTupleType(
		AtomVal(d.StrVal("").Eval),
		AtomVal(d.StrVal("").Eval),
		AtomVal(d.IntVal(0).Eval),
		AtomVal(d.FltVal(0.0).Eval),
	)
	fmt.Printf("tuple type constructor: %s\n", tupleType)

	var tupleVal = tupleType(
		AtomVal(d.StrVal("field one altered").Eval),
		AtomVal(d.StrVal("field two altered").Eval),
		AtomVal(d.IntVal(23).Eval),
		AtomVal(d.FltVal(42.23).Eval),
	)
	fmt.Printf("altered tuple type fields: %s\n", tupleVal())
	fmt.Printf("fields still altered?: %s\n", tupleVal())

	tupleVal = tupleType(
		AtomVal(d.StrVal("field one altered").Eval),
		AtomVal(d.StrVal("field two altered").Eval),
		AtomVal(d.FltVal(23.42).Eval),
		AtomVal(d.FltVal(42.23).Eval),
	)

	fmt.Printf("try to set the last field to a value of the wrong type"+
		"(should be reset to default): %s\n", tupleVal())

	tupleVal = tupleType(
		AtomVal(d.StrVal("field one").Eval),
		AtomVal(d.StrVal("field two").Eval),
		AtomVal(d.IntVal(23).Eval),
		AtomVal(d.FltVal(42.23).Eval),
	)
	var elems = tupleVal(NewIndexPair(
		2,
		UnaryExpr(func(arg Callable) Callable {
			return NewAtom(arg.Eval().(d.IntVal) + d.IntVal(23))
		}),
	))
	fmt.Printf("apply addition to value of field 2: %s\n", elems)
}

func TestRecordTypeConstruction(t *testing.T) {

	var recordType = NewRecordType(
		NewPair(AtomVal(d.StrVal("key one").Eval),
			AtomVal(d.StrVal("").Eval)),
		NewPair(AtomVal(d.StrVal("key two").Eval),
			AtomVal(d.IntVal(0).Eval)),
	)
	fmt.Printf("record type: %s\n", recordType())

	var recordVal = recordType(
		NewPair(AtomVal(d.StrVal("key one").Eval),
			AtomVal(d.StrVal("altered data one").Eval)),
		NewPair(AtomVal(d.StrVal("key two").Eval),
			AtomVal(d.IntVal(23).Eval)),
	)
	fmt.Printf("altered record type: %s\n", recordVal())

	var elems = recordVal(NewKeyPair(
		"key two",
		UnaryExpr(func(arg Callable) Callable {
			if number, ok := arg.Eval().(d.IntVal); ok {
				return NewAtom(number + d.IntVal(23))
			}
			return arg
		}),
	))
	fmt.Printf("applyed addition to field two %s\n", elems)

	elems = recordVal(NewPair(
		New("key two"),
		UnaryExpr(func(arg Callable) Callable {
			if number, ok := arg.Eval().(d.IntVal); ok {
				return NewAtom(number + d.IntVal(23))
			}
			return arg
		}),
	))
	fmt.Printf("applyed addition again, using ordinary pair: %s\n", elems)

	var args = []Callable{}
	for _, elem := range elems {
		args = append(args, elem)
	}
	recordVal = recordType(args...)
	fmt.Printf("new record value: %s\n", recordVal)
	elems = recordVal(NewRecordField(
		"key two",
		UnaryExpr(func(arg Callable) Callable {
			if number, ok := arg.Eval().(d.IntVal); ok {
				return NewAtom(number + d.IntVal(23))
			}
			return arg
		}),
	))
	fmt.Printf("applyed addition yet again, using record field: %s\n", elems)
}
