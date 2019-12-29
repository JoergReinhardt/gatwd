package functions

import (
	"fmt"
	"math/rand"
	"strings"
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

	mapAddInt = Define(Dat(func(args ...d.Native) d.Native {
		var a, b = args[0].(d.IntVal), args[1].(d.IntVal)
		return a + b
	}),
		DefSym("+"),
		Dat(0).Type(),
		Def(
			Dat(0).Type(),
			Dat(0).Type(),
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
func randInt() DatConst {
	return Dat(rand.Intn(100)).(DatConst)
}
func randInts(n int) []Expression {
	var slice = make([]Expression, 0, n)
	for i := 0; i < n; i++ {
		slice = append(slice, randInt())
	}
	return slice
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

func TestSortVector(t *testing.T) {
	var (
		v    = NewVector(randInts(10)...)
		sort = func(a, b Expression) bool {
			return a.(DatConst)().(d.IntVal) < b.(DatConst)().(d.IntVal)
		}
	)
	fmt.Printf("random: %s\n", v)
	fmt.Printf("sorted: %s\n", v.Sort(sort))
	var tmp Expression = Dat(0)
	for _, elem := range v() {
		if elem.(DatConst)().(d.IntVal) < tmp.(DatConst)().(d.IntVal) {
			t.Fail()
		}
		tmp = elem
	}
}
func TestSearchVector(t *testing.T) {
	var elem = abc.Search(Dat("k"), func(a, b Expression) int {
		return strings.Compare(a.String(), b.String())
	})
	fmt.Println(elem)
	if elem.String() != "k" {
		t.Fail()
	}
}

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
	if seq.Head().(DatConst).Eval().(d.Numeral).Int() != 9 {
		t.Fail()
	}

	seq = seq.Cons(Dat(8)).(SeqVal)
	fmt.Printf("equence with two elements (8, 9):\n%s\n", seq)
	if seq.Head().(DatConst).Eval().(d.Numeral).Int() != 8 {
		t.Fail()
	}

	seq = seq.Cons(Dat(5), Dat(6), Dat(7)).(SeqVal)
	fmt.Printf("equence with five elements (5, 6, 7, 8, 9):\n%s\n", seq)
	if seq.Head().(DatConst).Eval().(d.Numeral).Int() != 5 {
		t.Fail()
	}

	seq = seq.Append(Dat(10), Dat(11)).(SeqVal)
	fmt.Printf("equence with two elements appended (5, 6, 7, 8, 9, 10, 11):\n%s\n", seq)
	if seq.Head().(DatConst).Eval().(d.Numeral).Int() != 5 {
		t.Fail()
	}

	seq = seq.Cons(Dat(0), Dat(1), Dat(2), Dat(3), Dat(4)).(SeqVal)
	fmt.Printf("equence with five elements (0, 1, 2, 3, 4, 5, 6, 7, 8, 9):\n%s\n", seq)
	if seq.Head().(DatConst).Eval().(d.Numeral).Int() != 0 {
		t.Fail()
	}
}

func TestVectorConsAppend(t *testing.T) {
	var vec = NewVector()
	fmt.Printf("empty vecuence: %s\n", vec)

	vec = vec.Cons(Dat(8)).(VecVal)
	fmt.Printf("vector with one element [8]:\n%s\n", vec)
	if vec.Head().(DatConst).Eval().(d.Numeral).Int() != 8 {
		t.Fail()
	}

	vec = vec.Cons(Dat(9)).(VecVal)
	fmt.Printf("vector with two elements [8, 9]:\n%s\n", vec)
	if vec.Head().(DatConst).Eval().(d.Numeral).Int() != 9 {
		t.Fail()
	}

	vec = vec.Cons(Dat(10), Dat(11), Dat(12)).(VecVal)
	fmt.Printf("vector with five elements [8, 9, 10, 11, 12]:\n%s\n", vec)
	if vec.Head().(DatConst).Eval().(d.Numeral).Int() != 12 {
		t.Fail()
	}

	vec = vec.Push(Dat(6), Dat(7)).(VecVal)
	fmt.Printf("vector with two elements pushed [8, 9, 10, 11, 12, 6, 7]:\n%s\n", vec)
	if vec.Head().(DatConst).Eval().(d.Numeral).Int() != 7 {
		t.Fail()
	}

	vec = vec.Push(Dat(0), Dat(1), Dat(2), Dat(3), Dat(4), Dat(5)).(VecVal)
	fmt.Printf("vector with five elements [8, 9, 10, 11, 12, 6, 7, 0, 1, 2, 3, 4, 5]:\n%s\n", vec)
	if vec.Head().(DatConst).Eval().(d.Numeral).Int() != 5 {
		t.Fail()
	}
}

func TestStackSequence(t *testing.T) {
	var (
		head  Expression
		tail  Continuation
		list  Sequential = NewSequence()
		stack Stack      = NewSequence(intsA()...)
	)
	fmt.Printf("stack: %s\n", stack)
	for i := 0; i < 5; i++ {
		head, stack = stack.Pop()
		list = list.Cons(head)
	}
	fmt.Printf("head after 5 pops: %s\n", head)
	if head.(DatConst)().(d.IntVal) != 4 {
		t.Fail()
	}
	for i := 0; i < 5; i++ {
		head, tail = list.Continue()
		list = tail.(Sequential)
		stack = stack.Push(head)
	}
	fmt.Printf("stack after pushing 5 popped elements back on again: %s\n", stack)
	fmt.Printf("head after pushing 5 popped elements back on again: %s\n", head)
	if head.(DatConst)().(d.IntVal) != 0 {
		t.Fail()
	}
}

func TestStackVector(t *testing.T) {
	var (
		head  Expression
		tail  Continuation
		list  Sequential = NewSequence()
		stack Stack      = NewVector(intsA()...)
	)
	fmt.Printf("stack: %s\n", stack)
	for i := 0; i < 5; i++ {
		head, stack = stack.Pop()
		list = list.Cons(head)
	}
	fmt.Printf("head after 5 pops: %s\n", head)
	if head.(DatConst)().(d.IntVal) != 5 {
		t.Fail()
	}
	for i := 0; i < 5; i++ {
		head, tail = list.Continue()
		fmt.Printf("head from within push loop: %s\n", head)
		list = tail.(Sequential)
		stack = stack.Push(head)
	}
	fmt.Printf("stack after pushing 5 popped elements back on again: %s\n", stack)
	fmt.Printf("head after pushing 5 popped elements back on again: %s\n", head)
	if head.(DatConst)().(d.IntVal) != 9 {
		t.Fail()
	}
}

func TestQueueSequence(t *testing.T) {
	var (
		head  Expression
		tail  Continuation
		list  Sequential = NewVector()
		queue Queue      = NewSequence(intsA()...)
	)
	fmt.Printf("queue: %s\n", queue)
	for i := 0; i < 5; i++ {
		head, queue = queue.Pull()
		list = list.Cons(head)
	}
	fmt.Printf("head after 5 pulls: %s\n", head)
	if head.(DatConst)().(d.IntVal) != 5 {
		t.Fail()
	}
	for i := 0; i < 5; i++ {
		head, tail = list.Continue()
		list = tail.(Sequential)
		queue = queue.Append(head)
	}
	fmt.Printf("stack after appending 5 popped elements back on again: %s\n", queue)
	fmt.Printf("head after appending 5 popped elements back on again: %s\n", head)
	if head.(DatConst)().(d.IntVal) != 9 {
		t.Fail()
	}
}

func TestQueueVector(t *testing.T) {
	var (
		head  Expression
		tail  Continuation
		list  Sequential = NewVector()
		queue Queue      = NewVector(intsA()...)
	)
	fmt.Printf("queue: %s\n", queue)
	for i := 0; i < 5; i++ {
		head, queue = queue.Pull()
		list = list.Cons(head)
	}
	fmt.Printf("head after 5 pulls: %s\n", head)
	if head.(DatConst)().(d.IntVal) != 5 {
		t.Fail()
	}
	for i := 0; i < 5; i++ {
		head, tail = list.Continue()
		list = tail.(Sequential)
		queue = queue.Append(head)
	}
	fmt.Printf("stack after appending 5 popped elements back on again: %s\n", queue)
	fmt.Printf("head after appending 5 popped elements back on again: %s\n", head)
	if head.(DatConst)().(d.IntVal) != 9 {
		t.Fail()
	}
}

func TestMapSequence(t *testing.T) {

	var m = Map(intsA, func(arg Expression) Expression {
		return addInts(arg)
	})

	fmt.Printf("list-a mapped to add-ints: %s\n", m)

	var head, tail = m.Continue()
	for !tail.Empty() {
		fmt.Printf("expression called on 10: %s\n", head.Call(Dat(10)))
		head, tail = tail.Continue()
	}

	fmt.Printf("list-a mutated?: %s\n", m)

	m = Map(m, func(arg Expression) Expression {
		var result = arg.Call(Dat(10))
		return result
	})

	fmt.Printf("mapped list-a mapped to add-10: %s\n", m)
}

func TestApplySequence(t *testing.T) {

	var m = Apply(intsA, func(head Expression, args ...Expression) Expression {
		return addInts(append([]Expression{head}, args...)...)
	})

	fmt.Printf("add-ints applyed to list-a: %s\n", m)

	if m.Call().(Paired).Left().Call(Dat(13)).(DatConst)().(d.IntVal) != 22 {
		t.Fail()
	}

	var (
		head Expression
		pair Paired
		tail = m
	)
	for !tail.Empty() {
		pair = tail.Call().(Paired)
		head, tail = pair.Left().Call(Dat(13)), pair.Right().(SeqVal)
		fmt.Printf("list called with 13: %s\n", head)
	}
}

func TestFoldSequence(t *testing.T) {
	var f = Fold(intsA, Dat(0), func(init, head Expression) Expression {
		return addInts(init, head)
	})
	fmt.Printf("folded list: %s\n", f)

	var head, tail = f.Continue()
	for i := 0; i < 8; i++ {
		head, tail = tail.Continue()
	}
	fmt.Printf("head after eight continuations: %s\n", head)
	if head.(DatConst)().(d.IntVal) != 45 {
		t.Fail()
	}
}

func TestFilterPassSequence(t *testing.T) {
	var (
		isEven = func(arg Expression) bool {
			return arg.(DatConst)().(d.IntVal)%2 == 0
		}
		odd  = Filter(intsA, isEven)
		even = Pass(intsA, isEven)
	)
	fmt.Printf("odd: %s\neven: %s\n", odd, even)
	var ohead, otail = odd.Continue()
	var ehead, etail = even.Continue()
	for i := 0; i < 3; i++ {
		ohead, otail = otail.Continue()
		ehead, etail = etail.Continue()
		fmt.Printf("odd head: %s\neven head: %s\n", ohead, ehead)
	}
	if ohead.(SeqVal).Head().(DatConst)().(d.IntVal) != 3 ||
		ehead.(SeqVal).Head().(DatConst)().(d.IntVal) != 2 {
		t.Fail()
	}
}

func TestTakeNSequence(t *testing.T) {
	var token = TakeN(intsA, 2)
	fmt.Printf("take two: %s\n", token)
	var head, tail = token.Continue()
	fmt.Printf("head: %s\n", head)
	for !tail.Empty() {
		head, tail = tail.Continue()
	}
	fmt.Printf("last element: %s\n", head)
	head, tail = head.(SeqVal).Continue()
	for !tail.Empty() {
		head, tail = tail.(SeqVal).Continue()
	}
	fmt.Printf("last elements head: %s\n", head.(VecVal).Head())
	if head.(VecVal).Head().(DatConst)().(d.IntVal) != 9 {
		t.Fail()
	}
}
