package collections_test

import (
	"fmt"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/visualizer"
)

func ExampleHAMT() {
	// Create an empty immutable HAMT
	hamt := collections.NewHAMT[string, int]()

	// Associate keys immutably
	hamt1 := hamt.Set("alpha", 100).Set("beta", 200)

	val1, _ := hamt1.Get("alpha")
	val2, _ := hamt1.Get("beta")

	fmt.Printf("alpha: %d, beta: %d, original len: %d\n", val1, val2, hamt.Len())
	// Output:
	// alpha: 100, beta: 200, original len: 0
}

func ExampleList() {
	// Create an empty list and prepend elements (Cons)
	list := collections.Cons(3, collections.Cons(2, collections.Cons(1, collections.Nil[int]())))

	head, _ := list.Head()
	tail, _ := list.Tail()
	tailHead, _ := tail.Head()

	fmt.Printf("head: %d, next: %d\n", head, tailHead)
	// Output:
	// head: 3, next: 2
}

func ExampleSet() {
	// Create an immutable set backed by HAMT
	set := collections.NewSet("reader", "writer")
	set2 := set.Add("admin")

	fmt.Printf("set has admin: %t, set2 has admin: %t\n", set.Contains("admin"), set2.Contains("admin"))
	// Output:
	// set has admin: false, set2 has admin: true
}

func ExampleChunk() {
	// Create an eager, bounded contiguous array
	chunk := collections.NewChunk[int](4)
	chunk.Append(10)
	chunk.Append(20)

	v0, _ := chunk.Get(0)
	v1, _ := chunk.Get(1)

	fmt.Printf("len: %d, cap: %d, [0]: %d, [1]: %d\n", chunk.Len(), chunk.Capacity(), v0, v1)
	// Output:
	// len: 2, cap: 4, [0]: 10, [1]: 20
}

func ExampleHAMT_visualized() {
	hamt := collections.NewHAMT[string, int]().Set("workers", 8)

	opts := visualizer.ParseOptions("--curved --border")
	fmt.Print(visualizer.Format(hamt, opts))

	// Output:
	// ╭───────────────────╮
	// │ ╰─ Record         │
	// │     ╰─ workers: 8 │
	// ╰───────────────────╯
}

func ExampleChunk_visualized() {
	chunk := collections.NewChunk[string](2)
	chunk.Append("cached_A")
	chunk.Append("cached_B")

	opts := visualizer.ParseOptions("--curved --border")
	fmt.Print(visualizer.Format(chunk, opts))

	// Output:
	// ╭──────────────────────╮
	// │ ╰─ List              │
	// │     ├─ [0]: cached_A │
	// │     ╰─ [1]: cached_B │
	// ╰──────────────────────╯
}
