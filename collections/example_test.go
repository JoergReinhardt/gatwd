package collections_test

import (
	"fmt"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/visualizer"
)

func ExampleHAMT() {
	// Create an immutable HAMT and associate keys
	hamt := collections.NewHAMT[string, int]().Set("alpha", 100)

	opts := visualizer.ParseOptions("--curved --border")
	fmt.Print(visualizer.Format(hamt, opts))

	// Output:
	// ╭───────────────────╮
	// │ ╰─ Record         │
	// │     ╰─ alpha: 100 │
	// ╰───────────────────╯
}

func ExampleList() {
	// Create an immutable list and prepend elements (Cons)
	list := collections.Cons(2, collections.Cons(1, collections.Nil[int]()))

	opts := visualizer.ParseOptions("--curved --border")
	fmt.Print(visualizer.Format(list, opts))

	// Output:
	// ╭───────────────────────────────╮
	// │ ╰─ Struct                     │
	// │     ├─ Head: 2                │
	// │     ╰─ Tail: Struct           │
	// │         ├─ Head: 1            │
	// │         ╰─ Tail: <empty List> │
	// ╰───────────────────────────────╯
}

func ExampleSet() {
	// Create an immutable set backed by HAMT
	set := collections.NewSet("reader")

	opts := visualizer.ParseOptions("--heavy --border")
	fmt.Print(visualizer.Format(set, opts))

	// Output:
	// ╭─────────────────────╮
	// │ ┗━━ List            │
	// │     ┗━━ [0]: reader │
	// ╰─────────────────────╯
}

func ExampleChunk() {
	// Create an eager, bounded contiguous array
	chunk := collections.NewChunk[int](2)
	chunk.Append(10)
	chunk.Append(20)

	opts := visualizer.ParseOptions("--curved --border")
	fmt.Print(visualizer.Format(chunk, opts))

	// Output:
	// ╭────────────────╮
	// │ ╰─ List        │
	// │     ├─ [0]: 10 │
	// │     ╰─ [1]: 20 │
	// ╰────────────────╯
}
