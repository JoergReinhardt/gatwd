package transformers

import (
	"context"
	"testing"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/functions"
)

func TestStateT(t *testing.T) {
	items := func(yield func(int) bool) {
		for i := 1; i <= 3; i++ {
			if !yield(i) {
				return
			}
		}
	}

	// Accumulator state: running sum
	step := func(s int, a int) (string, int) {
		return "item", s + a
	}

	transformer := StateT(0, step)
	var pairs []functions.Pair[string, int]
	for p := range transformer(items) {
		pairs = append(pairs, p)
	}

	if len(pairs) != 3 {
		t.Fatalf("expected 3 pairs, got %d", len(pairs))
	}
	if pairs[2].Snd != 6 {
		t.Errorf("expected final sum 6, got %d", pairs[2].Snd)
	}
}

func TestStateContextT_Cancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	stream := func(yield func(dsltypes.ASTNode[functions.Either[error, int]]) bool) {
		flag := dsltypes.BitFlag(1)
		for i := 1; i <= 10; i++ {
			node := dsltypes.NewASTNode(functions.Right[error, int](i), flag)
			if !yield(node) {
				return
			}
		}
	}

	step := func(ctx context.Context, s int, a int) (int, int, error) {
		if a == 3 {
			cancel() // cancel mid-stream
		}
		return a * 10, s + a, nil
	}

	flag := dsltypes.BitFlag(1)
	_ = flag
	trans := StateContextT(ctx, 0, step)

	var count int
	var hadError bool
	for n := range trans(stream) {
		count++
		if n.Value.IsLeft() {
			hadError = true
			break
		}
	}

	if !hadError {
		t.Errorf("expected cancellation error node yielded into stream")
	}
}
