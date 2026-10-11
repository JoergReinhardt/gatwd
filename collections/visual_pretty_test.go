package collections_test

import (
	"testing"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/visualizer"
)

func TestVisualPretty_HAMT(t *testing.T) {
	hamt := collections.NewHAMT[string, string]().
		Set("env", "production").
		Set("region", "eu-central-1").
		Set("datacenter", "fra-01").
		Set("tier", "gateway")

	opts := visualizer.ParseOptions("--curved --border")
	card := visualizer.Format(hamt, opts)
	t.Logf("=== IMMUTABLE HAMT CONFIG MAP ===\n%s", card)
}

func TestVisualPretty_RoseTreeHierarchy(t *testing.T) {
	// Build an AST representation:
	// FunctionDeclaration(main)
	//   ├── ParameterList(args)
	//   └── Block
	//         ├── Stmt(init_db)
	//         └── Stmt(listen_and_serve)
	stmt1 := collections.Leaf("Stmt(init_db)")
	stmt2 := collections.Leaf("Stmt(listen_and_serve)")
	stmts := collections.Cons(stmt1, collections.Cons(stmt2, collections.Nil[*collections.Tree[string]]()))

	block := collections.NewTree("Block", stmts)
	params := collections.Leaf("ParameterList(args)")

	fnChildren := collections.Cons(params, collections.Cons(block, collections.Nil[*collections.Tree[string]]()))
	root := collections.NewTree("FunctionDeclaration(main)", fnChildren)

	opts := visualizer.ParseOptions("--curved --border")
	card := visualizer.Format(root, opts)
	t.Logf("=== PERSISTENT ROSE TREE AST HIERARCHY ===\n%s", card)
}

func TestVisualPretty_BoundedChunk(t *testing.T) {
	type CacheLine struct {
		Tag   string
		Hits  int
		Valid bool
	}

	chunk := collections.NewChunk[CacheLine](3)
	chunk.Append(CacheLine{Tag: "0x7FFF0010", Hits: 142, Valid: true})
	chunk.Append(CacheLine{Tag: "0x7FFF0020", Hits: 89, Valid: true})
	chunk.Append(CacheLine{Tag: "0x7FFF0030", Hits: 12, Valid: false})

	opts := visualizer.ParseOptions("--curved --border --box-all")
	card := visualizer.Format(chunk, opts)
	t.Logf("=== EAGER BOUNDED CHUNK (CPU CACHE OPTIMIZED) ===\n%s", card)
}

func TestVisualPretty_SetOperations(t *testing.T) {
	roles := collections.NewSet("reader", "writer", "deployer", "auditor")

	opts := visualizer.ParseOptions("--heavy --border")
	card := visualizer.Format(roles, opts)
	t.Logf("=== IMMUTABLE RBAC ROLE SET ===\n%s", card)
}

func TestVisualPretty_PersistentList(t *testing.T) {
	list := collections.Cons("alpha", collections.Cons("beta", collections.Cons("gamma", collections.Nil[string]())))

	opts := visualizer.ParseOptions("--curved --border")
	card := visualizer.Format(list, opts)
	t.Logf("=== IMMUTABLE CONS LIST ===\n%s", card)
}
