// Package cli implements the top-level script interpreter, REPL environment, and CLI entry point for gatwd.
//
// The Interpreter binds together:
//   - The monadic parser for compiling pipeline text into executable transducers.
//   - The evaluator engine for executing pipelines lazily over iter.Seq streams.
//   - Macro expansions (def) with runtime lazy expansion for recursive calls.
//   - Local variable bindings (let ... in ...).
//   - Top-level immutable constants (const) and native Go constant registration (RegisterConstant).
//   - Algebraic Data Type pattern matching and list destructuring ([pivot, ...rest]).
//   - Gatwd Builtins: Direct integration with transducers: chunk, checkpoint, telemetry, to_hamt, dial_tcp, scan_file, now (an injectable-clock adapter step, Interpreter.Clock).
//   - Bare Runtime (NewBareInterpreter): Minimalist execution core without standard library builtins.
//   - Gatwd Editions (NewEdition): Fluent builder for custom distributions with domain types and steps.
//   - Gatwd Appliances (NewAppliance, NewApplianceTree): Hot-reloadable releases with embedded scripts and a ConfigMap override layered on top (OverrideLayer, or OverrideReplace); a check/run command line with --format text|json and exit statuses 0, 1 and 2.
//   - The command line (CommandTree, Edition.Command): a cobra tree — script runner/REPL, run, eval, check, task, prelude, builtins, version, completion — that maps flags to Interpreter.Run. Output goes to Interpreter.Out.
//   - Ergonomic Step Helpers: MapStep, FilterStep, ArgStep, SinkStep for boilerplate-free transducer creation.
package cli
