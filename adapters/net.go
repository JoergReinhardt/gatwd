package adapters

import (
	"bufio"
	"errors"
	"io"
	"iter"
	"net"
	"time"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/functions"
)

// ListenerToSeq unfolds a net.Listener into a lazy sequence of accepted net.Conn connections.
//
// This is an anamorphism (generator/source) representing server ingress:
// every accepted connection is emitted as a Right[error, any] node wrapping net.Conn.
// When the listener is closed, the stream terminates cleanly without emitting spurious errors.
func ListenerToSeq(
	ln net.Listener,
	flag dsltypes.BitFlag,
) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
	return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
		for {
			conn, err := ln.Accept()
			if err != nil {
				// Clean shutdown if the listener was closed
				if errors.Is(err, net.ErrClosed) {
					return
				}
				errNode := dsltypes.NewASTNode(
					functions.Left[error, any](err),
					flag,
				)
				yield(errNode)
				return
			}

			node := dsltypes.NewASTNode(
				functions.Right[error, any](any(conn)),
				flag,
			)
			if !yield(node) {
				_ = conn.Close()
				return
			}
		}
	}
}

// ConnToLines unfolds an accepted net.Conn into a lazy stream of lines using bufio.Scanner.
// It closes the connection when the stream completes or early termination occurs.
func ConnToLines(
	conn net.Conn,
	flag dsltypes.BitFlag,
) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
	return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
		defer conn.Close()

		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			raw := make([]byte, len(scanner.Bytes()))
			copy(raw, scanner.Bytes())

			node := dsltypes.NewASTNode(
				functions.Right[error, any](any(string(raw))),
				flag,
			)
			if !yield(node) {
				return
			}
		}
		if err := scanner.Err(); err != nil && !errors.Is(err, io.EOF) {
			errNode := dsltypes.NewASTNode(
				functions.Left[error, any](err),
				flag,
			)
			yield(errNode)
		}
	}
}

// StepDial returns a pipeline step that dials a remote server, writes incoming node payloads,
// reads responses line-by-line, and yields them downstream.
func StepDial(
	network string,
	address string,
	timeout time.Duration,
) evaluator.PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			for node := range stream {
				if node.Value.IsLeft() {
					if !yield(node) {
						return
					}
					continue
				}

				val, _ := node.Value.UnwrapRight()
				var payload []byte
				switch v := val.(type) {
				case []byte:
					payload = v
				case string:
					payload = []byte(v)
				default:
					payload = []byte(node.String())
				}

				// Dial remote endpoint
				conn, err := net.DialTimeout(network, address, timeout)
				if err != nil {
					errNode := dsltypes.NewASTNode(
						functions.Left[error, any](err),
						node.DSLType,
					)
					if !yield(errNode) {
						return
					}
					continue
				}

				// Send payload with newline delimiter
				_, err = conn.Write(append(payload, '\n'))
				if err != nil {
					_ = conn.Close()
					errNode := dsltypes.NewASTNode(
						functions.Left[error, any](err),
						node.DSLType,
					)
					if !yield(errNode) {
						return
					}
					continue
				}

				// Read response line
				scanner := bufio.NewScanner(conn)
				if scanner.Scan() {
					resp := scanner.Text()
					respNode := dsltypes.NewASTNode(
						functions.Right[error, any](any(resp)),
						node.DSLType,
					)
					_ = conn.Close()
					if !yield(respNode) {
						return
					}
				} else {
					if scanErr := scanner.Err(); scanErr != nil {
						errNode := dsltypes.NewASTNode(
							functions.Left[error, any](scanErr),
							node.DSLType,
						)
						_ = conn.Close()
						if !yield(errNode) {
							return
						}
					}
					_ = conn.Close()
				}
			}
		}
	}
}

// StepConnLines unfolds an incoming net.Conn node into a stream of lines.
func StepConnLines() evaluator.PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			for node := range stream {
				if node.Value.IsLeft() {
					if !yield(node) {
						return
					}
					continue
				}
				val, _ := node.Value.UnwrapRight()
				if conn, ok := val.(net.Conn); ok {
					for lineNode := range ConnToLines(conn, node.DSLType) {
						if !yield(lineNode) {
							return
						}
					}
				} else {
					if !yield(node) {
						return
					}
				}
			}
		}
	}
}

// StepListenTCP starts a TCP listener on address, yields accepted connections up to maxConns,
// and shuts down cleanly.
func StepListenTCP(network, address string, maxConns int) evaluator.PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]] {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			ln, err := net.Listen(network, address)
			if err != nil {
				yield(dsltypes.NewASTNode(functions.Left[error, any](err), 0))
				return
			}
			defer ln.Close()

			count := 0
			for node := range ListenerToSeq(ln, 0) {
				count++
				if !yield(node) {
					return
				}
				if maxConns > 0 && count >= maxConns {
					return
				}
			}
		}
	}
}
