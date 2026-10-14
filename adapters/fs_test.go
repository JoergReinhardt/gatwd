package adapters

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/comonads"
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/functions"
)

func TestBuildFSTree(t *testing.T) {
	// Create an in-memory test file system
	fsys := fstest.MapFS{
		"rootfile.txt":     &fstest.MapFile{Data: []byte("hello")},
		"subdir/file1.txt": &fstest.MapFile{Data: []byte("file1")},
		"subdir/file2.txt": &fstest.MapFile{Data: []byte("file2")},
	}

	// Build the tree (using "." as root to represent the base of MapFS)
	tree, err := BuildFSTree(fsys, ".")
	if err != nil {
		t.Fatalf("Failed to build FS tree: %v", err)
	}

	if tree.Value() != "." {
		t.Errorf("Expected root value to be '.', got %v", tree.Value())
	}

	// Traverse with Zipper
	z := comonads.NewZipper(tree)

	// Move down into "." children
	// MapFS order is usually deterministic (alphabetical)
	// Children should be "rootfile.txt" and "subdir"
	// Wait, fs.ReadDir on MapFS returns them sorted.
	// 1. rootfile.txt
	// 2. subdir

	zChild1 := z.Down()
	if zChild1 == nil {
		t.Fatalf("Expected children in '.'")
	}

	val1 := zChild1.Extract().(string)

	zChild2 := zChild1.Right()
	if zChild2 == nil {
		t.Fatalf("Expected a second child")
	}

	val2 := zChild2.Extract().(string)

	// Check that we found both "rootfile.txt" and "subdir"
	foundRootFile := val1 == "rootfile.txt" || val2 == "rootfile.txt"
	foundSubDir := val1 == "subdir" || val2 == "subdir"

	if !foundRootFile || !foundSubDir {
		t.Errorf("Expected 'rootfile.txt' and 'subdir' as children, got %v and %v", val1, val2)
	}

	// Navigate into subdir
	var zSubDir *comonads.TreeZipper[any]
	if val1 == "subdir" {
		zSubDir = zChild1
	} else {
		zSubDir = zChild2
	}

	zSubChild1 := zSubDir.Down()
	if zSubChild1 == nil {
		t.Fatalf("Expected children in 'subdir'")
	}

	valSub1 := zSubChild1.Extract().(string)
	if valSub1 != "subdir/file1.txt" {
		t.Errorf("Expected 'subdir/file1.txt', got %v", valSub1)
	}

	zSubChild2 := zSubChild1.Right()
	if zSubChild2 == nil {
		t.Fatalf("Expected a second child in 'subdir'")
	}

	valSub2 := zSubChild2.Extract().(string)
	if valSub2 != "subdir/file2.txt" {
		t.Errorf("Expected 'subdir/file2.txt', got %v", valSub2)
	}
}

// TestWalkFSStreamAdapterReadsRootFromUpstream pins the whole point of the stream-driven form:
// the SAME adapter value, with no reconstruction, walks whatever directory arrives on the
// stream — unlike WalkFSAdapter(root), which fixes root once at construction.
func TestWalkFSStreamAdapterReadsRootFromUpstream(t *testing.T) {
	adapter := WalkFSStreamAdapter()
	flag := dsltypes.BitFlag(1)

	for _, root := range []string{".", "testdata"} {
		in := dsltypes.NewASTNode(functions.Right[error, any](any(root)), flag)
		var got *collections.Tree[any]
		for out := range adapter(in) {
			val, ok := out.Value.UnwrapRight()
			if !ok {
				err, _ := out.Value.UnwrapLeft()
				t.Fatalf("walk_fs %q: %v", root, err)
			}
			tree, ok := val.(*collections.Tree[any])
			if !ok {
				t.Fatalf("walk_fs %q: want *collections.Tree[any], got %T", root, val)
			}
			got = tree
		}
		if got == nil {
			t.Fatalf("walk_fs %q: produced no node", root)
		}
		if got.Value() != root {
			t.Errorf("walk_fs %q: tree root value = %v, want %v", root, got.Value(), root)
		}
	}
}

// TestWalkFSStreamAdapterFailsLoudOnNonPath is the negative control: a non-string upstream value
// must surface as a named error, never an empty or silently-wrong result (Mandate 6).
func TestWalkFSStreamAdapterFailsLoudOnNonPath(t *testing.T) {
	adapter := WalkFSStreamAdapter()
	flag := dsltypes.BitFlag(1)
	in := dsltypes.NewASTNode(functions.Right[error, any](any(5)), flag)

	var gotErr error
	count := 0
	for out := range adapter(in) {
		count++
		err, isLeft := out.Value.UnwrapLeft()
		if !isLeft {
			t.Fatalf("walk_fs on a non-path value: want a Left, got a Right")
		}
		gotErr = err
	}
	if count != 1 {
		t.Fatalf("walk_fs on a non-path value: produced %d nodes, want exactly 1", count)
	}
	if gotErr == nil || !strings.Contains(gotErr.Error(), "expected a directory path") {
		t.Fatalf("walk_fs on a non-path value: error = %v, want it to name what was expected", gotErr)
	}
}
