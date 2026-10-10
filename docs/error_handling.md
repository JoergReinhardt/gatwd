# Error Handling & Compliance Auditing

Functional compilers are legendary for their beautiful error handling. `gatwd` embraces this by providing mathematically pure error propagation across both the **DSL Frontend** (parsing) and the **Host Engine** (execution).

## 1. Elm-Style Syntax Errors (The Frontend)

When you write an invalid script, the Parser does not just crash or return an empty sequence. It uses a custom **Error Tracker** inside the Combinator loop to track the furthest (deepest) point it reached before failing.

If you write a syntax error, the CLI dynamically prints a beautiful Elm/Rust-style pointer explaining exactly what went wrong:

```text
🔥 Syntax Error:

    spawn Server |> take fifty
                         ^-- Expected Integer here.
```

## 2. The EitherT Monad Transformer (The Engine)

If the Graph Crawler encounters a runtime error (e.g., a Network Timeout when querying Kubernetes, or an RBAC permission error), throwing a panic would destroy the entire lazy traversal sequence.

Instead, the `gatwd` Engine uses the **`EitherT` Monad Transformer** (`evaluator/eithert.go`).
It upgrades the Graph Engine from evaluating `ASTNode[T]` to evaluating `ASTNode[Either[error, T]]`.

**How it works seamlessly:**
1. If an Adapter yields an Error (an `Either.Left`), the `EitherT` wrapper instantly recognizes it.
2. It **bypasses** any subsequent Adapters in the pipeline.
3. It passes the Error node straight down the stream so you (or your logging system) can see it!
4. The rest of the graph evaluating concurrently is entirely unaffected.

## 3. Compliance Auditing (The `expect` Gate)

Because of the `EitherT` mechanics, you can use `gatwd` as a native Security and Compliance Auditing tool!

If you use the standard `filter` keyword, the engine simply drops nodes that don't match. 
If you use the `expect` keyword, the engine acts as an auditing gate. If a node fails the predicate, it is transformed into an explicit `Either.Left` Error Node:

```haskell
auditCluster = 
    spawn KubernetesCluster "prod-01"
    |> getPods
    |> expect "is_non_root" "Security Violation: Pod is running as root!"
```

If any Pod in your cluster is running as root, the pipeline yields an Error node containing that exact string, while the valid pods continue processing!
