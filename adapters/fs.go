package adapters

import (
	"fmt"
	"io/fs"
	"iter"
	"os"
	"path"
	"strings"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/functions"
)

// BuildFSTree recursively builds an immutable Rose Tree representing the file system structure.
// This allows the tree to be mathematically traversed using the Comonadic TreeZipper. root is an
// fs.FS-relative path (see fs.ValidPath): never absolute, never "..". Node and leaf values are
// root joined with each entry's name, in root's own relative form.
func BuildFSTree(fsys fs.FS, root string) (*collections.Tree[any], error) {
	return buildFSTree(fsys, root, root)
}

// buildFSTree walks fsRoot (the fs.FS-relative path actually read) while labelling nodes and
// leaves with labelRoot (the path reported upward). The two diverge only when the caller wants
// absolute-looking labels over an fs.FS that, per the io/fs contract, must be addressed with
// relative paths — see WalkFSAdapter.
func buildFSTree(fsys fs.FS, fsRoot, labelRoot string) (*collections.Tree[any], error) {
	entries, err := fs.ReadDir(fsys, fsRoot)
	if err != nil {
		return nil, err
	}

	var childTrees []*collections.Tree[any]
	for _, entry := range entries {
		fsFull := path.Join(fsRoot, entry.Name())
		labelFull := path.Join(labelRoot, entry.Name())
		if entry.IsDir() {
			subTree, err := buildFSTree(fsys, fsFull, labelFull)
			if err != nil {
				return nil, err
			}
			childTrees = append(childTrees, subTree)
		} else {
			// Leaf node for a file
			childTrees = append(childTrees, collections.Leaf[any](any(labelFull)))
		}
	}

	list := collections.Nil[*collections.Tree[any]]()
	for i := len(childTrees) - 1; i >= 0; i-- {
		list = collections.Cons(childTrees[i], list)
	}

	return collections.NewTree[any](any(labelRoot), list), nil
}

// WalkFSAdapter ignores the incoming stream and yields a single Comonadic Tree node
// representing the file system rooted at the given path. root may be relative (resolved
// against the process's current directory, as before) or absolute: fs.FS forbids absolute
// paths (fs.ValidPath), so an absolute root is read through os.DirFS("/") with its leading
// slash stripped for the fs.FS call, while every reported node/leaf keeps the original,
// absolute-looking path the caller gave — a script that does not control the process's cwd
// (e.g. an Edition invoked from an arbitrary directory) can still name a fixed tree to walk.
func WalkFSAdapter(root string) func(dsltypes.ASTNode[functions.Either[error, any]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
	return func(node dsltypes.ASTNode[functions.Either[error, any]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			fsDir, fsRoot := ".", root
			if path.IsAbs(root) {
				fsDir = "/"
				fsRoot = strings.TrimPrefix(root, "/")
				if fsRoot == "" {
					fsRoot = "."
				}
			}
			fsys := os.DirFS(fsDir)

			tree, err := buildFSTree(fsys, fsRoot, root)
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

// WalkFSStreamAdapter walks the directory named by EACH upstream value, instead of one path
// fixed at pipeline-construction time (see WalkFSAdapter). This is what lets a reusable script
// name the resource to walk as a runtime argument rather than a literal baked into the step: a
// task script wires `walk_fs` plain into its pipeline once and is pointed at a different root on
// every run via `-- <path>`, the same way `manifests` already reads its path from the stream
// instead of taking one as a step argument.
func WalkFSStreamAdapter() func(dsltypes.ASTNode[functions.Either[error, any]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
	return func(node dsltypes.ASTNode[functions.Either[error, any]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			rightVal, ok := node.Value.UnwrapRight()
			root, isStr := rightVal.(string)
			if !ok || !isStr || root == "" {
				errNode := dsltypes.NewASTNode(functions.Left[error, any](
					fmt.Errorf("walk_fs: expected a directory path, got %v", node.Value)), node.DSLType)
				yield(errNode)
				return
			}
			for n := range WalkFSAdapter(root)(node) {
				if !yield(n) {
					return
				}
			}
		}
	}
}
