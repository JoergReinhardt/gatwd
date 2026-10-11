// Package evaluator implements the core stream processing engine and composable pipeline transducers.
//
// The evaluation model is based on Category Theory dataflow: pipelines process streams lazily
// using Go 1.23 iterators (iter.Seq). Because iter.Seq acts as a native trampoline, computations
// can be infinite, recursive, or parallelized across goroutines without risk of stack overflow.
//
// Core Concepts:
//
//   - Engine[T]: The pipeline runner that chains adapters and transducers over input streams.
//     Supports both sequential evaluation and concurrent multi-goroutine evaluation (ParEvaluate).
//
//   - PipelineStep[T]: The fundamental transducer signature: func(iter.Seq[ASTNode[T]]) iter.Seq[ASTNode[T]].
//     Steps compose orthogonally, allowing transformations, filters, and aggregations to chain cleanly.
//
//   - EitherT Error Handling: Pipelines evaluate nodes wrapped in functions.Either[error, any].
//     If a step produces a Left (error), downstream processing automatically short-circuits and propagates
//     the error to the stream terminal without throwing panics.
//
//   - Continuations & Checkpointing: StepCheckpoint provides non-intrusive stream watermarking.
//     Every N elements, the stream offset is saved to a CheckpointStore, allowing long-running
//     computations to be paused, migrated to another instance, or resumed seamlessly.
//
//   - Performance Benchmarking: BenchmarkPipeline hooks directly into Go's testing.B harness,
//     measuring throughput (ops/sec), allocations, and nanoseconds per operation.
package evaluator
