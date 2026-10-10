package main

import (
	"fmt"
	"github.com/joergreinhardt/gatwd/functions"
)

func main() {
	fmt.Println("=== Operator: Curry & Compose ===")

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

	result := addTenThenDouble(5) // (5 + 10) * 2 = 30
	fmt.Printf("Curried, Partially Applied, and Composed: (5 + 10) * 2 = %d\n", result)
}
