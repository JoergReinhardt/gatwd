// Package adapters provides IO bridges and anamorphic sequence generators (unfolds) for gatwd pipelines.
//
// In functional programming, side-effects and IO are strictly decoupled from computation logic.
// Adapters act as boundaries that take external systems (files, network sockets, scanners, HTTP APIs)
// and translate them into pure, lazy Go iterators (iter.Seq) of ASTNodes.
//
// Key Adapters:
//
//   - ScannerAdapter & ScannerToSeq: Bridges Go's standard bufio.Scanner or any generalized
//     Scanner into an anamorphic stream of token nodes.
//
//   - ListenerToSeq: Unfolds any net.Listener (TCP, Unix socket) into a continuous stream of
//     accepted net.Conn nodes, providing clean server ingress.
//
//   - StepDial & ConnToLines: Transducer for network client requests (egress) and streaming
//     connection line readers.
//
//   - WalkFSAdapter: Traverses any Go fs.FS hierarchical filesystem, lifting directories into
//     navigable Comonadic Rose Trees.
//
//   - WriteFile: Drains stream items into persistent files using Go io.Writer.
package adapters
