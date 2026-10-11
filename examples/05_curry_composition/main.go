package main

import (
	"fmt"

	"github.com/joergreinhardt/gatwd/functions"
	"github.com/joergreinhardt/gatwd/visualizer"
)

func main() {
	headerOpts := visualizer.ParseOptions("--heavy --border --border-color=cyan --color=white")
	fmt.Println(visualizer.Format("Operator: Curry & Compose", headerOpts))

	opts := visualizer.ParseOptions("--curved --border --border-color=green")

	// A normal uncurried Go function taking two arguments
	add := func(a, b int) int { return a + b }

	// 1. Curry it!
	curriedAdd := functions.Curry2(add)

	// 2. Partial Application
	addTen := curriedAdd(10)

	// 3. Another curried function
	multiplyByTwo := func(x int) int { return x * 2 }

	// 4. Compose them: f(g(x))
	// This means: add 10, then multiply the result by 2!
	addTenThenDouble := functions.Compose(multiplyByTwo, addTen)

	input := 5
	result := addTenThenDouble(input) // (5 + 10) * 2 = 30

	fmt.Println(visualizer.Format(map[string]any{
		"Pipeline":   "Compose(multiplyByTwo, Curry2(add)(10))",
		"Input":      input,
		"Expression": "(5 + 10) * 2",
		"Result":     result,
		"Status":     "Success",
	}, opts))
}
