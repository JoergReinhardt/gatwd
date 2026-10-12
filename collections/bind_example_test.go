package collections_test

import (
	"fmt"

	"github.com/joergreinhardt/gatwd/collections"
)

func ExampleList_Bind() {
	// Bind (monadic flatMap) applies f to each element and concatenates the
	// resulting Lists in order.
	list := collections.Cons(1, collections.Cons(2, collections.Nil[int]()))
	res := list.Bind(func(x int) *collections.List[int] {
		return collections.Cons(x, collections.Cons(x*10, collections.Nil[int]()))
	})
	for v := range res.Seq() {
		fmt.Println(v)
	}

	// Output:
	// 1
	// 10
	// 2
	// 20
}

func ExampleList_Reject() {
	// Reject keeps the elements for which the predicate is false (complement of
	// Filter) — here, dropping even numbers retains the odd ones.
	list := collections.Cons(1, collections.Cons(2,
		collections.Cons(3, collections.Cons(4, collections.Nil[int]()))))
	res := list.Reject(func(x int) bool { return x%2 == 0 })
	for v := range res.Seq() {
		fmt.Println(v)
	}

	// Output:
	// 1
	// 3
}

func ExampleChunk_Bind() {
	// Bind flatMaps each element into a Chunk and appends the results in order
	// into a single contiguous Chunk (bounded by the source chunk's capacity).
	chunk := collections.NewChunk[int](4)
	chunk.Append(1)
	chunk.Append(2)
	res := chunk.Bind(func(x int) *collections.Chunk[int] {
		sub := collections.NewChunk[int](2)
		sub.Append(x)
		sub.Append(x * 10)
		return sub
	})
	fmt.Println(res.Elements())

	// Output:
	// [1 10 2 20]
}

func ExampleSet_Bind() {
	// Bind flatMaps each element into a Set and unions the results. The Set monad
	// is idempotent, so duplicates collapse. We print deterministic facts (Len and
	// membership) rather than the raw, unordered Set.
	set := collections.NewSet(1, 2)
	res := set.Bind(func(x int) *collections.Set[int] {
		return collections.NewSet(x, x*10)
	})
	fmt.Println("len:", res.Len())
	fmt.Println("has 1:", res.Contains(1))
	fmt.Println("has 20:", res.Contains(20))

	// Output:
	// len: 4
	// has 1: true
	// has 20: true
}

func ExampleHAMT_Intersect() {
	// Intersect keeps only the keys present in both HAMTs, taking the values from
	// the receiver. Values are looked up by key to avoid map iteration order.
	a := collections.NewHAMT[string, int]().Set("x", 1).Set("y", 2)
	b := collections.NewHAMT[string, int]().Set("y", 99).Set("z", 3)
	res := a.Intersect(b)

	y, _ := res.Get("y")
	_, hasX := res.Get("x")
	fmt.Println("len:", res.Len())
	fmt.Println("y:", y)
	fmt.Println("has x:", hasX)

	// Output:
	// len: 1
	// y: 2
	// has x: false
}

func ExampleHAMT_Diff() {
	// Diff keeps the keys of the receiver that are absent from the other HAMT,
	// preserving the receiver's values.
	a := collections.NewHAMT[string, int]().Set("x", 1).Set("y", 2)
	b := collections.NewHAMT[string, int]().Set("y", 99).Set("z", 3)
	res := a.Diff(b)

	x, _ := res.Get("x")
	_, hasY := res.Get("y")
	fmt.Println("len:", res.Len())
	fmt.Println("x:", x)
	fmt.Println("has y:", hasY)

	// Output:
	// len: 1
	// x: 1
	// has y: false
}

func ExampleHAMT_Reject() {
	// Reject keeps the entries for which the predicate is false (complement of
	// Filter) — here, rejecting odd values retains the even ones.
	h := collections.NewHAMT[string, int]().Set("a", 1).Set("b", 2).Set("c", 3)
	res := h.Reject(func(k string, v int) bool { return v%2 == 1 })

	b, _ := res.Get("b")
	fmt.Println("len:", res.Len())
	fmt.Println("b:", b)

	// Output:
	// len: 1
	// b: 2
}

func ExampleTree_Filter() {
	// Filter always keeps the root and prunes child subtrees whose root value
	// fails the predicate. Keeping even values prunes the odd child (3).
	tree := collections.NewTree(1, collections.Cons(collections.Leaf(2),
		collections.Cons(collections.Leaf(3), collections.Nil[*collections.Tree[int]]())))
	res := tree.Filter(func(x int) bool { return x%2 == 0 })
	for v := range res.WalkDepthFirst() {
		fmt.Println(v)
	}

	// Output:
	// 1
	// 2
}

func ExampleTree_Reject() {
	// Reject prunes child subtrees whose root value satisfies the predicate
	// (complement of Filter). Rejecting even values prunes child 2, keeping 3.
	tree := collections.NewTree(1, collections.Cons(collections.Leaf(2),
		collections.Cons(collections.Leaf(3), collections.Nil[*collections.Tree[int]]())))
	res := tree.Reject(func(x int) bool { return x%2 == 0 })
	for v := range res.WalkDepthFirst() {
		fmt.Println(v)
	}

	// Output:
	// 1
	// 3
}
