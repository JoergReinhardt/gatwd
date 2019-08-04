package functions

import (
	"fmt"
	"testing"

	d "github.com/joergreinhardt/gatwd/data"
)

func TestArgType(t *testing.T) {
	var at = DefineArgumentSet(Def(d.Int), Def(d.Int), Def(d.Int))
	fmt.Printf("declared arguments: %s\n", at)
	if !at.Type().Match(Def(Def(d.Int), Def(d.Int), Def(d.Int))) {
		t.Fail()
	}

	var result = at.Call(DeclareNative(1))
	fmt.Printf("match pass int: %s result type: %s\n", result, result.Type())
	if !result.Type().Match(Def(d.Int)) {
		t.Fail()
	}

	result = at.Call(DeclareNative(1), DeclareNative(1), DeclareNative(1))
	fmt.Printf("match pass three ints: %s result type: %s\n", result, result.Type())
	if !result.Type().Match(Def(Vector, Def(d.Int))) {
		t.Fail()
	}

	result = at.Call(DeclareNative(1.0))
	fmt.Printf("match pass float: %s result type: %s\n", result, result.Type())
	if !result.Type().Match(d.Float) {
		t.Fail()
	}

	at = DefineArgumentSet(Def(d.Int), Def(d.Float))
	fmt.Printf("declared arguments: %s\n", at)
	if !at.MatchArgs(DeclareNative(1), DeclareNative(1.0)) {
		t.Fail()
	}
}

func TestDeclaredExpression(t *testing.T) {

	var addInt = DeclareFunction(func(args ...Expression) Expression {
		var a, b = args[0].(DataConst).Eval().(d.IntVal), args[1].(DataConst).Eval().(d.IntVal)
		return DeclareData(a + b)
	}, Def(d.Int))

	var result = addInt(DeclareNative(23), DeclareNative(42))
	fmt.Printf("result from applying ints to addInt: %s\n", result)

	var expr = DeclareExpression(addInt, Def(d.Int), Def(d.Int))
	fmt.Printf("declared expression: %s\n", expr)

	fmt.Printf("result from applying ints to addInt: %s\n",
		expr(DeclareNative(23), DeclareNative(42)))

	fmt.Printf("result from applying two floats to addInt: %s\n",
		expr(DeclareNative(23.0), DeclareNative(42.0)))

	fmt.Printf("result from applying four ints to addInt: %s\n",
		expr(DeclareNative(23), DeclareNative(42), DeclareNative(23), DeclareNative(42)))
	fmt.Printf("result type: %s\n", expr.Type())

	fmt.Printf("result from applying five ints to addInt: %s\n",
		expr(DeclareNative(23), DeclareNative(42), DeclareNative(23),
			DeclareNative(42), DeclareNative(42)))

	fmt.Printf("result from applying six ints to addInt: %s\n",
		expr(DeclareNative(23), DeclareNative(42), DeclareNative(23),
			DeclareNative(42), DeclareNative(42), DeclareNative(42)))

	result = expr(DeclareNative(23), DeclareNative(42), DeclareNative(23), DeclareNative(42),
		DeclareNative(42), DeclareNative(42), DeclareNative(42), DeclareNative(42))
	fmt.Printf("result from applying eight ints to addInt: %s\n", result)
	fmt.Printf("result from applying two more ints oversatisfyed expr: %s\n",
		result.Call(DeclareNative(42), DeclareNative(42)))

	var partial = expr.Call(DeclareNative(23))
	fmt.Printf("result from applying one int to addInt: %s, expression: %s arg type: %s len: %d\n",
		partial, partial.(ExpressionType).Unbox(), partial.(ExpressionType).ArgType(),
		partial.(ExpressionType).ArgType().Len())

	fmt.Printf("result from applying second int to partial addInt: %s\n",
		partial.Call(DeclareNative(42)))
}
