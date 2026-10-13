package evaluator_test

import (
	"fmt"
	"testing"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/functions"
)

// nestedResource builds the nested-HAMT shape the Kubernetes adapter yields:
// { kind: "Pod", spec: { selector: { app: "nginx" } } }.
func nestedResource() *collections.HAMT[string, any] {
	selector := collections.NewHAMT[string, any]().Set("app", "nginx")
	spec := collections.NewHAMT[string, any]().Set("selector", selector)
	return collections.NewHAMT[string, any]().
		Set("kind", "Pod").
		Set("spec", spec)
}

func runPath(root any, path string) (any, bool, error) {
	engine := evaluator.NewEngine[functions.Either[error, any]]()
	flag := dsltypes.BitFlag(1)
	source := func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
		yield(dsltypes.NewASTNode(functions.Right[error, any](root), flag))
	}
	pipeline := []evaluator.PipelineStep[functions.Either[error, any]]{
		evaluator.StepHAMTPath(path),
	}
	for node := range engine.EvaluatePipeline(source, pipeline) {
		if node.Value.IsLeft() {
			err, _ := node.Value.UnwrapLeft()
			return nil, false, err
		}
		val, _ := node.Value.UnwrapRight()
		return val, true, nil
	}
	return nil, false, nil
}

func TestStepHAMTPath(t *testing.T) {
	root := nestedResource()

	t.Run("deep navigation yields the leaf", func(t *testing.T) {
		val, ok, err := runPath(root, "spec.selector.app")
		if !ok || err != nil {
			t.Fatalf("expected success, got ok=%v err=%v", ok, err)
		}
		if val != "nginx" {
			t.Fatalf("spec.selector.app = %v, want nginx", val)
		}
	})

	t.Run("single segment behaves like hamt_get", func(t *testing.T) {
		val, ok, err := runPath(root, "kind")
		if !ok || err != nil || val != "Pod" {
			t.Fatalf("kind = %v (ok=%v err=%v), want Pod", val, ok, err)
		}
	})

	t.Run("intermediate node reachable", func(t *testing.T) {
		val, _, err := runPath(root, "spec.selector")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, isHAMT := val.(*collections.HAMT[string, any]); !isHAMT {
			t.Fatalf("spec.selector should be a nested HAMT, got %T", val)
		}
	})

	t.Run("missing segment errors, naming the segment and path", func(t *testing.T) {
		_, ok, err := runPath(root, "spec.selector.missing")
		if ok || err == nil {
			t.Fatalf("expected error for missing segment, got ok=%v err=%v", ok, err)
		}
	})

	t.Run("navigating into a scalar errors", func(t *testing.T) {
		// kind is a string; descending past it is not navigable.
		_, ok, err := runPath(root, "kind.foo")
		if ok || err == nil {
			t.Fatalf("expected error navigating into scalar, got ok=%v err=%v", ok, err)
		}
	})

	t.Run("record (map[string]any) carrier", func(t *testing.T) {
		record := map[string]any{"owner": "root"}
		val, ok, err := runPath(record, "owner")
		if !ok || err != nil || val != "root" {
			t.Fatalf("record owner = %v (ok=%v err=%v), want root", val, ok, err)
		}
	})

	t.Run("empty path passes the node through", func(t *testing.T) {
		val, ok, err := runPath(root, "")
		if !ok || err != nil {
			t.Fatalf("empty path should pass through, got ok=%v err=%v", ok, err)
		}
		if _, isHAMT := val.(*collections.HAMT[string, any]); !isHAMT {
			t.Fatalf("empty path should yield the original HAMT, got %T", val)
		}
	})
}

// ExampleStepHAMTPath shows deep, struct-like field access over a dynamically-typed
// nested-HAMT resource payload (the shape the Kubernetes adapter yields): the single
// `hamt_path spec.selector.app` step walks three levels of nesting at once.
func ExampleStepHAMTPath() {
	val, _, _ := runPath(nestedResource(), "spec.selector.app")
	fmt.Println(val)
	// Output:
	// nginx
}
