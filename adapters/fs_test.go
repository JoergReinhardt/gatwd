package adapters

import (
	"testing"
	"testing/fstest"

	"github.com/joergreinhardt/gatwd/comonads"
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
