package adapters

import (
	"bufio"
	"fmt"
	"io"
	"iter"
	"os"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/functions"
)

// ReadLines adapter creates an iterator that reads lines from an io.Reader
func ReadLines(reader io.Reader, flag dsltypes.BitFlag) func(dsltypes.ASTNode[functions.Either[error, any]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
	return func(node dsltypes.ASTNode[functions.Either[error, any]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			if node.Value.IsLeft() {
				yield(node)
				return
			}

			scanner := bufio.NewScanner(reader)
			for scanner.Scan() {
				line := scanner.Text()
				newNode := dsltypes.NewASTNode(functions.Right[error, any](any(line)), flag)
				if !yield(newNode) {
					return
				}
			}
			if err := scanner.Err(); err != nil {
				errNode := dsltypes.NewASTNode(functions.Left[error, any](err), flag)
				yield(errNode)
			}
		}
	}
}

// WriteFile adapter acts as a terminal sink, opening a file and writing all incoming string nodes to it.
func WriteFile(filename string, appendMode bool) func(dsltypes.ASTNode[functions.Either[error, any]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
	return func(node dsltypes.ASTNode[functions.Either[error, any]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			if node.Value.IsLeft() {
				yield(node)
				return
			}

			flags := os.O_CREATE | os.O_WRONLY
			if appendMode {
				flags |= os.O_APPEND
			} else {
				flags |= os.O_TRUNC
			}

			file, err := os.OpenFile(filename, flags, 0644)
			if err != nil {
				errNode := dsltypes.NewASTNode(functions.Left[error, any](err), node.DSLType)
				yield(errNode)
				return
			}
			defer file.Close()

			val, _ := node.Value.UnwrapRight()
			strVal := fmt.Sprintf("%v\n", val)
			
			if _, err := file.WriteString(strVal); err != nil {
				errNode := dsltypes.NewASTNode(functions.Left[error, any](err), node.DSLType)
				yield(errNode)
				return
			}

			// Pass the node downstream unaltered so it can still be processed or visualized
			yield(node)
		}
	}
}
