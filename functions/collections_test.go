package functions

import (
	"fmt"
	"testing"

	d "github.com/joergreinhardt/gatwd/data"
)

// test data
var (
	intsA = NewVector(Dat(0), Dat(1), Dat(2), Dat(3),
		Dat(4), Dat(5), Dat(6), Dat(7), Dat(8), Dat(9))

	intsB = NewVector(Dat(10), Dat(11), Dat(12), Dat(13),
		Dat(14), Dat(15), Dat(16), Dat(17), Dat(18), Dat(19))

	abc = NewVector(Dat("a"), Dat("b"), Dat("c"), Dat("d"), Dat("e"),
		Dat("f"), Dat("g"), Dat("h"), Dat("i"), Dat("j"), Dat("k"), Dat("l"),
		Dat("m"), Dat("n"), Dat("o"), Dat("p"), Dat("q"), Dat("r"), Dat("s"),
		Dat("t"), Dat("u"), Dat("v"), Dat("w"), Dat("x"), Dat("y"), Dat("z"))

	mapAddInt = Define(
		Lambda(func(args ...Expression) Expression {
			if args[0].Type().Match(Data) &&
				args[1].Type().Match(Data) {
				if inta, ok := args[0].(NatEval).Eval().(d.Integer); ok {
					if intb, ok := args[1].(NatEval).Eval().(d.Integer); ok {
						return Box(inta.Int() + intb.Int())
					}
				}
			}
			return NewNone()
		}),
		DefSym("+"),
		Def(Def(Data, Constant), d.Int),
		Def(
			Def(Def(Data, Constant), d.Int),
			Def(Def(Data, Constant), d.Int),
		))

	generator = NewGenerator(Dat(0), Lambda(func(args ...Expression) Expression {
		return mapAddInt.Call(args[0], Dat(10))
	}))

	accumulator = NewAccumulator(Dat(0), Lambda(func(args ...Expression) Expression {
		return mapAddInt.Call(args[0], Dat(10))
	}))
)

// helper functions
func conList(args ...Expression) Sequential {
	return NewVector(args...)
}

func printCons(cons Continuation) {
	var head, tail = cons.Continue()
	//if !head.Type().Match(None) {
	if !head.Type().Match(None) {
		fmt.Println(head)
		printCons(tail)
	}
}

func TestEmptyList(t *testing.T) {
	var list = NewVector()
	fmt.Printf("empty list pattern length: %d\n",
		list.Len())
	fmt.Printf("empty list patterns: %s\n",
		list.Type().TypeName())
	fmt.Printf("empty list arg types: %s\n",
		list.Type().TypeArgs())
	fmt.Printf("empty list ident types: %s\n",
		list.Type().TypeId())
	fmt.Printf("empty list return types: %s\n",
		list.Type().TypeRet())
	fmt.Printf("empty list type name: %s\n",
		list.Type())
}
func TestList(t *testing.T) {
	var list = NewVector(intsA()...)
	fmt.Printf("list pattern length: %d\n",
		list.Type().Len())
	fmt.Printf("list patterns: %d\n",
		list.Type().Pattern())
	fmt.Printf("list arg types: %s\n",
		list.Type().ArgumentsName())
	fmt.Printf("list ident types: %s\n",
		list.Type().IdentName())
	fmt.Printf("list return types: %s\n",
		list.Type().ReturnName())
	fmt.Printf("list type name: %s\n",
		list.Type())
	printCons(list)
}

func TestConList(t *testing.T) {

	var alist = NewVector(intsA()...)
	var tail Continuation
	var head Expression

	for i := 0; i < 5; i++ {
		head, tail = alist.Continue()
		fmt.Println("for loop: " + head.String())
	}

	tail = tail.(VecVal).Cons(intsB()...)

	printCons(tail)
}

func TestPushList(t *testing.T) {

	var alist = NewVector(intsA()...)
	var tail Continuation
	var head Expression

	for i := 0; i < 5; i++ {
		head, tail = alist.Continue()
		fmt.Println("for loop: " + head.String())
	}

	printCons(tail)
}

func TestVector(t *testing.T) {

	var vec = NewVector(intsA.Slice()...)
	fmt.Printf("vector: %s\n", vec)

	vec = vec.Cons(intsB.Slice()...).(VecVal)
	fmt.Printf("vector after cons list-B: %s\n", vec)
	fmt.Printf("vector first: %s last: %s\n", vec.First(), vec.Last())

	var head, tail = vec.Continue()
	for !head.Type().Match(None) {
		fmt.Printf("head: %s\n", head)
		head, tail = tail.Continue()
	}
}

//func TestSortVector(t *testing.T) {
//	var vec = NewVector(Dat(13), Dat(7), Dat(3), Dat(23), Dat(42))
//	fmt.Printf("unsorted list: %s\n", vec)
//
//	var sorted = vec.Sort(func(i, j int) bool {
//		if vec.Len() > i && vec.Len() > j {
//			return vec()[i].(NatEval).Eval().(d.IntVal) <
//				vec()[j].(NatEval).Eval().(d.IntVal)
//		}
//		return false
//	})
//	fmt.Printf("sorted list: %s\n", sorted)
//}
//func TestSearchVector(t *testing.T) {
//	var vec = NewVector(Dat("one"), Dat("two"), Dat("three"), Dat("four"), Dat("five"))
//	fmt.Printf("unsorted list: %s\n", vec)
//
//	var sorted = vec.Search(
//		func(i, j int) bool {
//			if vec.Len() > i && vec.Len() > j {
//				return strings.Compare(
//					vec()[i].(NatEval).Eval().String(),
//					vec()[j].(NatEval).Eval().String()) < 0
//			}
//			return false
//		},
//		func(arg Expression) bool {
//			return strings.Compare(arg.String(), "one") == 0
//		})
//	fmt.Printf("found element one: %s\n", sorted)
//}

func TestGenerator(t *testing.T) {
	fmt.Printf("generator: %s\n", generator)
	var answ Expression
	for i := 0; i < 10; i++ {
		answ, generator = generator()
		fmt.Printf("answer: %s generator: %s\n", answ, generator)
	}
	if answ.(NatEval).Eval().(d.IntVal) != d.IntVal(90) {
		t.Fail()
	}
}

func TestAccumulator(t *testing.T) {
	fmt.Printf("accumulator: %s \n", accumulator)
	var res Expression
	for i := 0; i < 10; i++ {
		res, accumulator = accumulator(Dat(10))
		fmt.Printf("result: %s accumulator called on argument: %s\n", res, accumulator)
	}
	if res.(NatEval).Eval().(d.IntVal) != d.IntVal(100) {
		t.Fail()
	}
}

func TestSequence(t *testing.T) {
	var seq = NewSequence(intsA()...)
	fmt.Printf("fresh sequence: %s\n", seq)
	fmt.Printf("sequence second print: %s\n", seq)
	var head, tail = seq()
	fmt.Printf("head: %s tail: %s\n", head, tail)
	for !tail.Empty() {
		head, tail = tail()
		fmt.Printf("head iteration: %s\n", head)
	}
	fmt.Printf("sequence: %s\n", seq)
	fmt.Printf("seq head: %s, tail: %s type: %s\n",
		seq.Head(), seq.Tail(), seq.TypeFnc())
}

func TestSequenceConsAppend(t *testing.T) {
	var seq = NewSequence()
	fmt.Printf("empty sequence: %s\n", seq)

	seq = seq.Cons(Dat(9)).(SeqVal)
	fmt.Printf("equence with one element (9):\n%s\n", seq)
	if seq.Head().(DatAtom).Eval().(d.Numeral).Int() != 9 {
		t.Fail()
	}

	seq = seq.Cons(Dat(8)).(SeqVal)
	fmt.Printf("equence with two elements (8, 9):\n%s\n", seq)
	if seq.Head().(DatAtom).Eval().(d.Numeral).Int() != 8 {
		t.Fail()
	}

	seq = seq.Cons(Dat(5), Dat(6), Dat(7)).(SeqVal)
	fmt.Printf("equence with five elements (5, 6, 7, 8, 9):\n%s\n", seq)
	if seq.Head().(DatAtom).Eval().(d.Numeral).Int() != 5 {
		t.Fail()
	}

	seq = seq.Append(Dat(10), Dat(11)).(SeqVal)
	fmt.Printf("equence with two elements appended (5, 6, 7, 8, 9, 10, 11):\n%s\n", seq)
	if seq.Head().(DatAtom).Eval().(d.Numeral).Int() != 5 {
		t.Fail()
	}

	seq = seq.Cons(Dat(0), Dat(1), Dat(2), Dat(3), Dat(4)).(SeqVal)
	fmt.Printf("equence with five elements (0, 1, 2, 3, 4, 5, 6, 7, 8, 9):\n%s\n", seq)
	if seq.Head().(DatAtom).Eval().(d.Numeral).Int() != 0 {
		t.Fail()
	}
}

func TestVectorConsAppend(t *testing.T) {
	var vec = NewVector()
	fmt.Printf("empty vecuence: %s\n", vec)

	vec = vec.Cons(Dat(8)).(VecVal)
	fmt.Printf("vector with one element [8]:\n%s\n", vec)
	if vec.Head().(DatAtom).Eval().(d.Numeral).Int() != 8 {
		t.Fail()
	}

	vec = vec.Cons(Dat(9)).(VecVal)
	fmt.Printf("vector with two elements [8, 9]:\n%s\n", vec)
	if vec.Head().(DatAtom).Eval().(d.Numeral).Int() != 8 {
		t.Fail()
	}

	vec = vec.Cons(Dat(10), Dat(11), Dat(12)).(VecVal)
	fmt.Printf("vector with five elements [8, 9, 10, 11, 12]:\n%s\n", vec)
	if vec.Head().(DatAtom).Eval().(d.Numeral).Int() != 8 {
		t.Fail()
	}

	vec = vec.Push(Dat(6), Dat(7)).(VecVal)
	fmt.Printf("vector with two elements pushed [6, 7, 8, 9, 10, 11, 12]:\n%s\n", vec)
	if vec.Head().(DatAtom).Eval().(d.Numeral).Int() != 6 {
		t.Fail()
	}

	vec = vec.Push(Dat(0), Dat(1), Dat(2), Dat(3), Dat(4), Dat(5)).(VecVal)
	fmt.Printf("vector with five elements [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12]:\n%s\n", vec)
	if vec.Head().(DatAtom).Eval().(d.Numeral).Int() != 0 {
		t.Fail()
	}
}
