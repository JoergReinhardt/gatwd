package visualizer_test

import (
	"errors"
	"testing"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/functions"
	"github.com/joergreinhardt/gatwd/visualizer"
)

func TestVisualPretty_CardGallery(t *testing.T) {
	t.Run("Curved_Border_MockUser", func(t *testing.T) {
		opts := visualizer.ParseOptions("--curved --border --color-type=MockUser:cyan")
		user := evaluator.MockUser{
			ID:     101,
			Name:   "admin_Diana",
			Role:   "admin",
			Score:  98.75,
			Active: true,
		}
		card := visualizer.Format(user, opts)
		t.Logf("\n%s", card)
	})

	t.Run("Heavy_Border_ErrorBox", func(t *testing.T) {
		opts := visualizer.ParseOptions("--heavy --border")
		err := errors.New("CIRCUIT_BREAKER_OPEN: downstream service unreachable after 3 retries")
		box := visualizer.Format(err, opts)
		t.Logf("\n%s", box)
	})

	t.Run("Tree_Hierarchy_AST", func(t *testing.T) {
		leaf1 := collections.Leaf("Identifier(x)")
		leaf2 := collections.Leaf("Literal(42)")
		binaryExpr := collections.NewTree("BinaryOp(+)", collections.Cons(leaf1, collections.Cons(leaf2, collections.Nil[*collections.Tree[string]]())))
		root := collections.NewTree("AssignStmt(let)", collections.Cons(binaryExpr, collections.Nil[*collections.Tree[string]]()))

		opts := visualizer.ParseOptions("--curved --border")
		t.Logf("\n%s", visualizer.Format(root, opts))
	})

	t.Run("Stream_Pipeline_StepVisualize", func(t *testing.T) {
		engine := evaluator.NewEngine[functions.Either[error, any]]()
		flag := dsltypes.BitFlag(1)

		stream := evaluator.GenerateMockUsers(2, flag)
		pipeline := []evaluator.PipelineStep[functions.Either[error, any]]{
			visualizer.StepVisualize("--curved --border"),
		}

		outStream := engine.EvaluatePipeline(stream, pipeline)
		for node := range outStream {
			val, _ := node.Value.UnwrapRight()
			t.Logf("Visualized ASTNode Card:%s", val.(string))
		}
	})
}
