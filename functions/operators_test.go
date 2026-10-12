package functions_test

import (
	"math"
	"math/big"
	"testing"

	"github.com/joergreinhardt/gatwd/collections"
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

func TestOperators_CollectionConcat(t *testing.T) {
	// Set ++ Set (Union)
	s1 := collections.NewSet[int](1, 2)
	s2 := collections.NewSet[int](2, 3)
	resSet, err := functions.EvalBinaryOp(functions.OpConcat, s1, s2)
	if err != nil {
		t.Fatalf("Set concat error: %v", err)
	}
	unionSet, ok := resSet.(*collections.Set[int])
	if !ok || unionSet.Len() != 3 {
		t.Fatalf("expected Set of len 3, got %T: %v", resSet, resSet)
	}

	// HAMT ++ HAMT (Merge)
	h1 := collections.NewHAMT[string, int]().Set("a", 1).Set("b", 2)
	h2 := collections.NewHAMT[string, int]().Set("b", 20).Set("c", 30)
	resHamt, err := functions.EvalBinaryOp(functions.OpConcat, h1, h2)
	if err != nil {
		t.Fatalf("HAMT concat error: %v", err)
	}
	mergedHamt, ok := resHamt.(*collections.HAMT[string, int])
	if !ok || mergedHamt.Len() != 3 {
		t.Fatalf("expected HAMT of len 3, got %T", resHamt)
	}
	vB, _ := mergedHamt.Get("b")
	if vB != 20 {
		t.Errorf("expected right-biased value 20 for 'b', got %d", vB)
	}

	// List ++ List
	l1 := collections.Cons(1, collections.Cons(2, collections.Nil[int]()))
	l2 := collections.Cons(3, collections.Cons(4, collections.Nil[int]()))
	resList, err := functions.EvalBinaryOp(functions.OpConcat, l1, l2)
	if err != nil {
		t.Fatalf("List concat error: %v", err)
	}
	catList, ok := resList.(*collections.List[int])
	if !ok || catList.Len() != 4 || catList.String() != "[1 2 3 4]" {
		t.Fatalf("expected List [1 2 3 4], got %v (%T)", resList, resList)
	}

	// Chunk ++ Chunk
	c1 := collections.NewChunk[int](2)
	c1.Append(1)
	c1.Append(2)
	c2 := collections.NewChunk[int](2)
	c2.Append(3)
	c2.Append(4)
	resChunk, err := functions.EvalBinaryOp(functions.OpConcat, c1, c2)
	if err != nil {
		t.Fatalf("Chunk concat error: %v", err)
	}
	catChunk, ok := resChunk.(*collections.Chunk[int])
	if !ok || catChunk.Len() != 4 {
		t.Fatalf("expected Chunk of len 4, got %T", resChunk)
	}
}

func TestOperators_CompareAndLexicographicalOrdering(t *testing.T) {
	// Sequences comparison
	sA := []any{1, 2}
	sB := []any{1, 3}
	cmp, err := functions.Compare(sA, sB)
	if err != nil || cmp >= 0 {
		t.Errorf("expected [1, 2] < [1, 3], got cmp=%d, err=%v", cmp, err)
	}

	// Prefix comparison (shorter is smaller)
	sC := []any{1, 2, 0}
	cmp2, err2 := functions.Compare(sA, sC)
	if err2 != nil || cmp2 >= 0 {
		t.Errorf("expected [1, 2] < [1, 2, 0], got cmp=%d, err=%v", cmp2, err2)
	}

	// Relational operators on sequences
	isLt, err := functions.EvalBinaryOp(functions.OpLt, sA, sB)
	if err != nil || isLt != true {
		t.Errorf("expected [1, 2] < [1, 3] to be true, got %v", isLt)
	}

	// Pairs comparison
	pA := functions.Pair[int, string]{Fst: 1, Snd: "alpha"}
	pB := functions.Pair[int, string]{Fst: 1, Snd: "beta"}
	cmpP, errP := functions.Compare(pA, pB)
	if errP != nil || cmpP >= 0 {
		t.Errorf("expected (1, alpha) < (1, beta), got cmp=%d, err=%v", cmpP, errP)
	}
}

func TestOperators_StructuralEqualSetsAndCollections(t *testing.T) {
	// Sets are order-independent!
	set1 := collections.NewSet[int](1, 2, 3)
	set2 := collections.NewSet[int](3, 1, 2)
	if !functions.StructuralEqual(set1, set2) {
		t.Errorf("expected {1,2,3} == {3,1,2} for sets")
	}

	// Pairs
	p1 := functions.Pair[string, int]{Fst: "id", Snd: 42}
	p2 := functions.Pair[string, int]{Fst: "id", Snd: 42}
	if !functions.StructuralEqual(p1, p2) {
		t.Errorf("expected pairs to be structurally equal")
	}

	// Either
	e1 := functions.Right[error, int](100)
	e2 := functions.Right[error, int](100)
	if !e1.Equal(e2) {
		t.Errorf("expected Either rights to be equal")
	}

	// Maybe
	m1 := functions.Just(42)
	m2 := functions.Just(42)
	if !m1.Equal(m2) {
		t.Errorf("expected Maybe justs to be equal")
	}
	mFiltered := m1.Filter(func(x int) bool { return x > 50 })
	if !mFiltered.IsNothing() {
		t.Errorf("expected Nothing after filter")
	}
}
