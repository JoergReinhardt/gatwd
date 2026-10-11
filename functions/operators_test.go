package functions_test

import (
	"math"
	"math/big"
	"testing"

	"github.com/joergreinhardt/gatwd/functions"
)

func TestOperators_DivisionFloatPromotion(t *testing.T) {
	// 7 / 2 should promote to float64 3.5
	res, err := functions.EvalBinaryOp(functions.OpDiv, 7, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	f, ok := res.(float64)
	if !ok || f != 3.5 {
		t.Fatalf("expected float64 3.5, got %v (%T)", res, res)
	}

	// 7 // 2 should remain integer 3
	resInt, err := functions.EvalBinaryOp(functions.OpIDiv, 7, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	i, ok := resInt.(int)
	if !ok || i != 3 {
		t.Fatalf("expected int 3, got %v (%T)", resInt, resInt)
	}
}

func TestOperators_OverflowPromotionToBigInt(t *testing.T) {
	// MaxInt64 + 1 should promote to *big.Int
	max := int64(math.MaxInt64)
	res, err := functions.EvalBinaryOp(functions.OpAdd, max, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	bigVal, ok := res.(*big.Int)
	if !ok {
		t.Fatalf("expected *big.Int promotion, got %T: %v", res, res)
	}
	expected := new(big.Int).Add(big.NewInt(math.MaxInt64), big.NewInt(1))
	if bigVal.Cmp(expected) != 0 {
		t.Fatalf("expected %v, got %v", expected, bigVal)
	}

	// MaxInt64 * 2 should promote to *big.Int
	resMul, err := functions.EvalBinaryOp(functions.OpMul, max, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	bigMul, ok := resMul.(*big.Int)
	if !ok {
		t.Fatalf("expected *big.Int promotion on mul, got %T", resMul)
	}
	expectedMul := new(big.Int).Mul(big.NewInt(math.MaxInt64), big.NewInt(2))
	if bigMul.Cmp(expectedMul) != 0 {
		t.Fatalf("expected %v, got %v", expectedMul, bigMul)
	}
}

func TestOperators_ComparisonsAndLogic(t *testing.T) {
	lt, err := functions.EvalBinaryOp(functions.OpLt, 10, 20)
	if err != nil || lt != true {
		t.Fatalf("expected 10 < 20 to be true, got %v, err=%v", lt, err)
	}

	eq, err := functions.EvalBinaryOp(functions.OpEq, 10, 10.0)
	if err != nil || eq != true {
		t.Fatalf("expected 10 == 10.0 to be true, got %v, err=%v", eq, err)
	}

	and, err := functions.EvalBinaryOp(functions.OpAnd, true, false)
	if err != nil || and != false {
		t.Fatalf("expected true && false to be false, got %v, err=%v", and, err)
	}

	concat, err := functions.EvalBinaryOp(functions.OpConcat, "hello ", "world")
	if err != nil || concat != "hello world" {
		t.Fatalf("expected 'hello world', got %v", concat)
	}
}
