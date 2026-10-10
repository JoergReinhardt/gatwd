# gatwd DSL Guide

`gatwd` is a purely functional Domain Specific Language (DSL) for declarative infrastructure crawling, evaluation, and mutation.

## 1. Algebraic Data Types (ADTs)

You must define your types before spawning them.
```haskell
-- Sum Types (Dynamic Enums)
type OS = Linux | Windows | Darwin

-- Record Types (Structured Data)
type File = { Name: String, Size: Int }
```

## 2. Spawning & Pipelines

Use `spawn` to create the initial root node. Use `|>` for lazy sequential evaluation, or `||>` (ParBind) to fan-out execution across a goroutine worker pool.
```haskell
spawn OS: Linux |> uname -a
spawn OS: Linux ||> sleep 5
```

## 3. Transducers & Streams

Because `gatwd` uses lazy `iter.Seq` trampolines, graphs can be infinite.
* `take N`: Takes exactly N nodes, then short-circuits the pipeline.
* `drop N`: Ignores the first N nodes.
* `distinct`: Uses an immutable `Set` to drop duplicate nodes (critical for cyclic graphs).
* `fold count`: Collapses the entire stream into a single scalar value.

```haskell
spawn Node: dummy |> ls . . |> distinct |> fold count
```

## 4. Compliance Auditing (Transformers)

* `expect "keyword" "ErrorMsg"`: Uses the `EitherT` Monad Transformer. If the keyword is missing, yields an explicit `Either.Left` Error Node.
* `filter "keyword"`: Acts like the `MaybeT` Transformer. If the keyword is missing, it silently drops the node without panicking.

```haskell
spawn OS: Linux |> uname -a |> expect "Linux" "CRITICAL: Not Linux"
```

## 5. Arrows & Zipping (Routing)

* `(A &&& B)` (Kleisli FanOut): Routes a node into TWO independent adapters concurrently, yielding a Tuple `(Res1, Res2)`.
* `[A *** B]` (Applicative Split): Takes a Tuple, and routes the Left side to Adapter A, and Right side to Adapter B concurrently.
* `format_tuple`: Unpacks a Tuple into a string.

```haskell
spawn OS: Linux |> (date &&& whoami) |> [echo L_%s *** echo R_%s] |> format_tuple "(%s, %s)"
```

## 6. Optics & Records (Lenses)

Convert unstructured OS output into rich Records, and use functional Lenses to immutably manipulate them.
```haskell
spawn Folder: /etc |> ls -l |> parse_record Perms Links Owner Group Size Month Day Time Name |> update Owner "root" |> focus Name
```

## 7. Pattern Matching & Memoization

* `match`: Destructures Sum Types and routes them dynamically.
* `memoize`: Uses functional caching to prevent redundant adapter execution.

```haskell
spawn OS: Linux |> match { Linux -> runBash | Windows -> runPowershell }
```
