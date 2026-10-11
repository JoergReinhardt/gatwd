package adapters

import (
	"io/fs"
	"iter"
	"os"
	"path"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/functions"
)

// BuildFSTree recursively builds an immutable Rose Tree representing the file system structure.
// This allows the tree to be mathematically traversed using the Comonadic TreeZipper.
func BuildFSTree(fsys fs.FS, root string) (*collections.Tree[any], error) {
	entries, err := fs.ReadDir(fsys, root)
	if err != nil {
		return nil, err
	}

	var childTrees []*collections.Tree[any]
	for _, entry := range entries {
		fullPath := path.Join(root, entry.Name())
		if entry.IsDir() {
			subTree, err := BuildFSTree(fsys, fullPath)
			if err != nil {
				return nil, err
			}
			childTrees = append(childTrees, subTree)
		} else {
			// Leaf node for a file
			childTrees = append(childTrees, collections.Leaf[any](any(fullPath)))
		}
	}
	
	list := collections.Nil[*collections.Tree[any]]()
	for i := len(childTrees) - 1; i >= 0; i-- {
		list = collections.Cons(childTrees[i], list)
	}

	return collections.NewTree[any](any(root), list), nil
}

// WalkFSAdapter ignores the incoming stream and yields a single Comonadic Tree node
// representing the file system rooted at the given path.
func WalkFSAdapter(root string) func(dsltypes.ASTNode[functions.Either[error, any]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
	return func(node dsltypes.ASTNode[functions.Either[error, any]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			fsys := os.DirFS(".")
			
			tree, err := BuildFSTree(fsys, root)
			if err != nil {
				errNode := dsltypes.NewASTNode(functions.Left[error, any](err), node.DSLType)
				yield(errNode)
				return
			}
			
			treeNode := dsltypes.NewASTNode(functions.Right[error, any](any(tree)), node.DSLType)
			yield(treeNode)
		}
	}
}
