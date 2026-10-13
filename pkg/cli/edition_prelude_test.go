package cli

import (
	"fmt"
	"iter"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
)

// TestEditionWithPrelude verifies that an inline prelude registered via
// WithPrelude is evaluated at build time, so its definitions are available to
// pipelines run against the built interpreter.
func TestEditionWithPrelude(t *testing.T) {
	interp := NewEdition("test-prelude").
		WithPrelude("type IntVal = Int\ndef add20 = math_add 20").
		BuildInterpreter()

	vals := interp.EvaluateToValues("spawn IntVal: 5 |> add20")
	if len(vals) != 1 || fmt.Sprint(vals[0]) != "25" {
		t.Fatalf("prelude def add20 not applied: got %v, want [25]", vals)
	}
}

// TestEditionWithPreludeStacking verifies preludes accumulate and run in
// registration order: a later prelude can use definitions from an earlier one.
func TestEditionWithPreludeStacking(t *testing.T) {
	interp := NewEdition("test-stack").
		WithPrelude("type IntVal = Int\ndef add20 = math_add 20").
		WithPrelude("def add25 = add20 |> math_add 5").
		BuildInterpreter()

	vals := interp.EvaluateToValues("spawn IntVal: 0 |> add25")
	if len(vals) != 1 || fmt.Sprint(vals[0]) != "25" {
		t.Fatalf("stacked prelude (add25 using add20) failed: got %v, want [25]", vals)
	}
}

// TestEditionWithPreludeFS verifies an embedded prelude is read from the given
// fs.FS and evaluated, mirroring a //go:embed embed.FS shipped in an edition.
func TestEditionWithPreludeFS(t *testing.T) {
	fsys := fstest.MapFS{
		"lib/prelude.gatwd": &fstest.MapFile{
			Data: []byte("type IntVal = Int\ndef plus7 = math_add 7"),
		},
	}
	interp := NewEdition("test-fs").
		WithPreludeFS(fsys, "lib/prelude.gatwd").
		BuildInterpreter()

	vals := interp.EvaluateToValues("spawn IntVal: 5 |> plus7")
	if len(vals) != 1 || fmt.Sprint(vals[0]) != "12" {
		t.Fatalf("embedded prelude plus7 not applied: got %v, want [12]", vals)
	}
}

// TestAdapterStep verifies that an arbitrary Adapter[any] registered via
// AdapterStep becomes a named DSL pipeline step that fans one node out into the
// sequence of child nodes the adapter yields.
func TestAdapterStep(t *testing.T) {
	// A trivial crawl adapter: expand one node into two suffixed children.
	expand := func(node dsltypes.ASTNode[any]) iter.Seq[dsltypes.ASTNode[any]] {
		return func(yield func(dsltypes.ASTNode[any]) bool) {
			base := fmt.Sprint(node.Value)
			for _, suffix := range []string{"_a", "_b"} {
				if !yield(dsltypes.NewASTNode[any](base+suffix, node.DSLType)) {
					return
				}
			}
		}
	}

	interp := NewEdition("test-adapter").
		WithPrelude("type StrVal = String").
		WithStep(AdapterStep("expand", evaluator.Adapter[any](expand))).
		BuildInterpreter()

	vals := interp.EvaluateToValues("spawn StrVal: X |> expand")
	if len(vals) != 2 {
		t.Fatalf("AdapterStep fan-out: got %d values %v, want 2", len(vals), vals)
	}
	if fmt.Sprint(vals[0]) != "X_a" || fmt.Sprint(vals[1]) != "X_b" {
		t.Fatalf("AdapterStep children wrong: got %v, want [X_a X_b]", vals)
	}
}

// TestEditionPreludeCatShowsTheImportedFiles: `prelude list` names the entry and the .gatwd
// files beside it, and `prelude cat` prints each, so a library split over several files
// (an entry that `import`s its siblings) is readable from the binary. Files of other
// directories and other extensions are not listed.
func TestEditionPreludeCatShowsTheImportedFiles(t *testing.T) {
	fsys := fstest.MapFS{
		"lib/entry.gatwd":   {Data: []byte("import \"rel.gatwd\"\ndef one = \\u -> 1\n")},
		"lib/rel.gatwd":     {Data: []byte("def two = \\u -> 2\n")},
		"lib/notes.txt":     {Data: []byte("not a script")},
		"other/alien.gatwd": {Data: []byte("def three = \\u -> 3\n")},
	}
	tree := NewEdition("sib").WithPreludeFS(fsys, "lib/entry.gatwd").Command()
	run := func(args ...string) (int, string, string) {
		var o, e strings.Builder
		code := tree.Execute(args, strings.NewReader(""), &o, &e)
		return code, o.String(), e.String()
	}
	code, out, _ := run("prelude", "list")
	if code != 0 || out != "lib/entry.gatwd\nlib/rel.gatwd\n" {
		t.Fatalf("prelude list: exit %d, out %q; want the entry and its sibling only", code, out)
	}
	if code, out, _ := run("prelude", "cat", "lib/rel.gatwd"); code != 0 || out != "def two = \\u -> 2\n" {
		t.Fatalf("prelude cat of the imported file: exit %d, out %q", code, out)
	}
	if code, _, errOut := run("prelude", "cat", "lib/notes.txt"); code == 0 || !strings.Contains(errOut, "no prelude") {
		t.Fatalf("a non-script sibling must stay unlisted: exit %d, stderr %q", code, errOut)
	}
}
