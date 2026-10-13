// Package functions provides foundational functional programming constructs and algebraic data types.
//
// Types and combinators provided:
//   - Either[L, R]: The error/result monad with Left, Right, Bind, and Fold operations.
//   - Maybe[T]: The optional value type (Just / Nothing) with BindMaybe, Map, and UnwrapOr.
//   - Pair[A, B]: 2-tuples for product types, zipping, and fan-out combinators.
//   - State[S, A]: The state monad S -> (A, S) with Run, Eval, Exec, Get, Put, Modify.
//   - StateContext[S, A]: Interruptible state monad integrated with Go's context.Context.
//   - Reader[R, A]: The environment/dependency monad R -> A with Ask, Asks, Local.
//   - Writer[W, A]: The telemetry/audit monad (A, W) with Tell, Listen.
//   - Cont[R, A]: The continuation monad (A -> R) -> R with CallCC (Call-With-Current-Continuation).
//   - Currying: Curry2, Curry3, and higher-order function composition Compose.
//   - The numeric tower (numeric.go, ratio.go): Int (unbounded, canonical: an int when it fits int64,
//     a *big.Int beyond, demoted again whenever a result fits), Ratio (exact fractions, a whole
//     fraction is an Int), Float and Complex, joined by EvalBinaryOp in the lattice
//     Int ⊂ Ratio ⊂ Float ⊂ Complex; exact comparison (CompareNumbers), Negate, Pow, the rounding
//     family (RoundToInt), GCD/LCM, ToRatio/ToRatioApprox, and Not (the function behind `!`).
package functions
