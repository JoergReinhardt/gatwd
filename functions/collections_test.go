package functions

import (
	"fmt"
	"testing"

	d "github.com/joergreinhardt/gatwd/data"
)

// test data
var (
	listA = NewVector(Dat(0), Dat(1), Dat(2), Dat(3),
		Dat(4), Dat(5), Dat(6), Dat(7), Dat(8), Dat(9))

	listB = NewVector(Dat(10), Dat(11), Dat(12), Dat(13),
		Dat(14), Dat(15), Dat(16), Dat(17), Dat(18), Dat(19))

	mapAddInt = Define(Lambda(func(args ...Expression) Expression {
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
		return mapAddInt.Call(args[0], Dat(1))
	}))

	accumulator = NewAccumulator(Dat(0), Lambda(func(args ...Expression) Expression {
		return mapAddInt.Call(args[0], Dat(1))
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
		list.Type().TypeArguments())
	fmt.Printf("empty list ident types: %s\n",
		list.Type().TypeIdent())
	fmt.Printf("empty list return types: %s\n",
		list.Type().TypeReturn())
	fmt.Printf("empty list type name: %s\n",
		list.Type())
}
func TestList(t *testing.T) {
	var list = NewVector(listA()...)
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

	var alist = NewVector(listA()...)
	var tail Continuation
	var head Expression

	for i := 0; i < 5; i++ {
		head, tail = alist.Continue()
		fmt.Println("for loop: " + head.String())
	}

	tail = tail.(VecVal).Cons(listB()...)

	printCons(tail)
}

func TestPushList(t *testing.T) {

	var alist = NewVector(listA()...)
	var tail Continuation
	var head Expression

	for i := 0; i < 5; i++ {
		head, tail = alist.Continue()
		fmt.Println("for loop: " + head.String())
	}

	printCons(tail)
}

func TestVector(t *testing.T) {

	var vec = NewVector(listA.Slice()...)
	fmt.Printf("vector: %s\n", vec)

	vec = vec.Cons(listB.Slice()...).(VecVal)
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

func TestPairVal(t *testing.T) {
	var pair = NewPair(NewNone(), NewNone())
	fmt.Printf("name of empty pair: %s\n", pair.Type())

	pair = NewPair(Dat(12), Dat("string"))
	fmt.Printf("name of (int,string) pair: %s\n",
		pair.Type())
	fmt.Printf("name of (int,string) pair args: %s\n",
		pair.Type().TypeArguments())
	fmt.Printf("name of (int,string) pair return: %s\n",
		pair.Type().TypeReturn())
}

func TestGenerator(t *testing.T) {
	fmt.Printf("generator: %s\n", generator)
	var answ Expression
	for i := 0; i < 10; i++ {
		answ, generator = generator()
		fmt.Printf("answer: %s generator: %s\n", answ, generator)
	}
	if answ.(NatEval).Eval().(d.IntVal) != d.IntVal(9) {
		t.Fail()
	}
}

func TestAccumulator(t *testing.T) {
	fmt.Printf("accumulator: %s \n", accumulator)
	var res Expression
	for i := 0; i < 10; i++ {
		res, accumulator = accumulator(Dat(1))
		fmt.Printf("result: %s accumulator called on argument: %s\n", res, accumulator)
	}
	if res.(NatEval).Eval().(d.IntVal) != d.IntVal(10) {
		t.Fail()
	}
}

func TestSequence(t *testing.T) {
	var seq = NewSequence(listA)
	fmt.Printf("fresh sequence: %s\n", seq)
	seq = NewSequence(listA)
	fmt.Printf("sequence second print: %s\n", seq)
	seq = NewSequence(listA)
	var head, tail = seq()
	fmt.Printf("head: %s tail: %s\n", head, tail)
	for !head.Type().Match(None) {
		fmt.Printf("head iteration: %s\n", head)
		head, tail = tail()
	}
	fmt.Printf("sequence: %s\n", seq)
	fmt.Printf("seq head: %s, tail: %s type: %s\n",
		seq.Current(), seq.Next(), seq.TypeFnc())
}

func TestMapSequential(t *testing.T) {
	var (
		a, b   Expression
		la     = NewSeqCont(listA)
		lb     = NewSeqCont(listB)
		mapped = la.Map(NewLambda(func(args ...Expression) Expression {
			a, la = la()
			b, lb = lb()
			return mapAddInt(a, b)
		}))
	)
	fmt.Printf("mapped: %s\n", mapped)
}

func TestConcatSequences(t *testing.T) {
	var lc = NewSeqCont(listA)
	fmt.Printf("new sequence from continuation list a: %s\n", lc)

	var step, next = lc.Continue()
	for !step.Type().Match(None) {
		fmt.Printf("loop til next end: %s %s\n", step, next)
		step, next = next.Continue()
	}
	fmt.Printf("head & tail after loop to end: %s %s\n", step, next)
	fmt.Printf("next step, next next: %s %s\n", next.Current(), next.Next())

	lc = lc.ConcatSeq(listB)
	fmt.Printf("list b concatenated to list-a continuation: %s\n", lc)
	fmt.Printf("concated lists a-/ & b: %s\n", lc)
}

func TestMapSequentialProduct(t *testing.T) {
	var (
		ll   = NewSequence(listA, listB, listA, listB)
		mapf = NewLambda(func(args ...Expression) Expression {
			if len(args) > 0 {
				if len(args) > 1 {
					return NewSequence(args...)
				}
				return args[0]
			}
			return NewNone()
		})
		mapped = ll.Map(mapf)
	)
	fmt.Printf("mapped sequences: %s\n", mapped)

	var flattened = mapped.(SeqVal).Flatten()
	fmt.Printf("flattened sequences: %s\n", flattened)
}

func TestFoldSequential(t *testing.T) {
	var queue = NewVector()
	fmt.Printf("empty queue: %s\n", queue)
	var (
		acc  = NewVector()
		fold = func(acc, head Expression) Expression {
			if vec, ok := acc.(VecVal); ok {
				return vec.ConcatVector(head)
			}
			return acc
		}
		folded = NewSeqCont(listA).Fold(acc, fold)
	)
	fmt.Printf("list A: %s\n", listA)
	fmt.Printf("accumulator: %s\n", acc)
	fmt.Printf("folded: %s\n", folded)
	var head, tail = folded.Continue()
	for !tail.End() {
		fmt.Printf("folded head: %s\n", head)
		head, tail = tail.Continue()
	}
	fmt.Printf("folded: %s\n", folded)
}

func TestFilterSequential(t *testing.T) {
	var (
		seq  = NewSeqCont(listA)
		test = TestFunc(func(arg Expression) bool {
			if dat, ok := arg.(NatEval); ok {
				if num, ok := dat.Eval().(d.Integer); ok {
					if num.Int()%2 == 0 {
						fmt.Printf("even: %s\n", num)
						return true
					}
				}
			}
			return false
		})
		filtered = seq.Filter(test)
	)
	fmt.Printf("filtered: %s\n", filtered)
}
