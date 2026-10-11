// Package cli implements the top-level script interpreter, REPL environment, and CLI entry point for gatwd.
//
// The Interpreter binds together:
//   - The monadic parser for compiling pipeline text into executable transducers.
//   - The evaluator engine for executing pipelines lazily over iter.Seq streams.
//   - Macro expansions (def) with runtime lazy expansion for recursive calls.
//   - Local variable bindings (let ... in ...).
//   - Top-level immutable constants (const) and native Go constant registration (RegisterConstant).
//   - Algebraic Data Type pattern matching and list destructuring ([pivot, ...rest]).
//   - Gatwd Builtins: Direct integration with transducers: chunk, checkpoint, telemetry, to_hamt, dial_tcp, scan_file.
//   - Bare Runtime (NewBareInterpreter): Minimalist execution core without standard library builtins.
//   - Gatwd Editions: Custom binary distributions extending the language via RegisterType and RegisterStep.
//   - Gatwd Appliances: Self-contained releases (Kubernetes Operators, Servers, CLIs) executing embedded scripts via LoadScriptFS.
package cli
