package functions_test

import (
	"fmt"
	"math/big"

	"github.com/joergreinhardt/gatwd/functions"
)

// A Ratio is an exact fraction in lowest terms; a whole fraction is an Int; arithmetic is
// exact where a Float is not.
func ExampleNewRatio() {
	third, _ := functions.NewRatio(big.NewInt(2), big.NewInt(6))
	fmt.Println(third) // lowest terms, and it prints as the literal the lexer reads back

	sum, _ := functions.EvalBinaryOp(functions.OpAdd, third, third)
	fmt.Println(sum)

	whole, _ := functions.EvalBinaryOp(functions.OpMul, third, 3)
	fmt.Printf("%v %T\n", whole, whole) // 1/3 * 3 demotes to an Int

	tenth, _ := functions.ParseNumberLiteral("0.1r")
	f1, _ := functions.EvalBinaryOp(functions.OpAdd, tenth, tenth)
	exact, _ := functions.EvalBinaryOp(functions.OpEq, f1, functions.CanonRat(big.NewRat(1, 5)))
	a, b := 0.1, 0.2 // variables: Go would fold the constants exactly
	float, _ := functions.EvalBinaryOp(functions.OpEq, a+b, 0.3)
	fmt.Println(exact, float)
	// Output:
	// 1r3
	// 2r3
	// 1 int
	// true false
}

// to_ratio recovers the number you meant from a Float; to_ratio_exact keeps the binary value.
func ExampleToRatio() {
	meant, _ := functions.ToRatio(0.1)
	stored, _ := functions.ToRatioExact(0.1)
	approx, _ := functions.ToRatioApprox(3.14159265, 100)
	fmt.Println(meant)
	fmt.Println(stored)
	fmt.Println(approx)
	// Output:
	// 1r10
	// 3602879701896397r36028797018963968
	// 311r99
}

// The logical Int is unbounded and canonical: it demotes when a result fits int64 again.
func ExampleCanonInt() {
	big1, _ := new(big.Int).SetString("9223372036854775808", 10) // 2^63
	up, _ := functions.EvalBinaryOp(functions.OpAdd, big1, 0)
	down, _ := functions.EvalBinaryOp(functions.OpSub, big1, 1)
	fmt.Printf("%T %v\n", up, up)
	fmt.Printf("%T %v\n", down, down)
	// Output:
	// *big.Int 9223372036854775808
	// int 9223372036854775807
}
