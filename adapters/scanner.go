package adapters

import (
	"bufio"
	"iter"
	"os"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/functions"
)

// Scanner is a generalized interface compatible with bufio.Scanner but also
// any custom scanner (SQL row scanners, tokenizers, protocol decoders, etc.).
//
// The contract is:
//   - Scan() advances to the next token and returns true if one exists.
//   - Bytes() returns the raw bytes of the current token.
//   - Err() returns any error encountered during scanning (not io.EOF).
type Scanner interface {
	Scan() bool
	Bytes() []byte
	Err() error
}

// ScannerToSeq unfolds a Scanner into a lazy iterator of ASTNodes.
//
// This is an anamorphism (unfold): it generates a potentially infinite stream
// from a seed (the scanner's internal state) by repeatedly calling Scan().
// Each scanned token is wrapped as a Right[error, any] value. If the scanner
// terminates with an error, a final Left[error, any] node is emitted.
//
// The flag parameter assigns the DSL quasi-type to every emitted node.
func ScannerToSeq(scanner Scanner, flag dsltypes.BitFlag) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
	return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
		for scanner.Scan() {
			// Copy the bytes to avoid aliasing the scanner's internal buffer.
			raw := make([]byte, len(scanner.Bytes()))
			copy(raw, scanner.Bytes())

			node := dsltypes.NewASTNode(
				functions.Right[error, any](any(raw)),
				flag,
			)
			if !yield(node) {
				return
			}
		}
		if err := scanner.Err(); err != nil {
			errNode := dsltypes.NewASTNode(
				functions.Left[error, any](err),
				flag,
			)
			yield(errNode)
		}
	}
}

// ScannerAdapter returns an evaluator.Adapter that ignores the incoming node's
// value and instead yields all tokens from the Scanner as a lazy stream.
//
// This allows a Scanner to be composed into an Engine pipeline via
// evaluator.StepAdapter. The incoming node's DSLType is used as the flag
// for all emitted nodes, preserving type context through the pipeline.
func ScannerAdapter(scanner Scanner) evaluator.Adapter[functions.Either[error, any]] {
	return func(node dsltypes.ASTNode[functions.Either[error, any]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		// Short-circuit: propagate upstream errors without touching the scanner.
		if node.Value.IsLeft() {
			return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
				yield(node)
			}
		}
		return ScannerToSeq(scanner, node.DSLType)
	}
}

// StepScanFile creates a pipeline step that opens the specified file and streams
// its tokens using the optional bufio.SplitFunc (defaults to lines if splitFunc is nil).
func StepScanFile(path string, splitFunc bufio.SplitFunc) evaluator.PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			file, err := os.Open(path)
			if err != nil {
				yield(evaluator.InjectError[any](err, 0))
				return
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			if splitFunc != nil {
				scanner.Split(splitFunc)
			}

			for node := range ScannerToSeq(scanner, 0) {
				if node.Value.IsRight() {
					val, _ := node.Value.UnwrapRight()
					if b, ok := val.([]byte); ok {
						node = dsltypes.NewASTNode(functions.Right[error, any](any(string(b))), node.DSLType)
					}
				}
				if !yield(node) {
					return
				}
			}
		}
	}
}

