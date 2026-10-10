# gatwd: Gophers All The Way Down

`gatwd` is a purely functional Category Theory engine and Turing-complete Scripting Language written in Go. 

It is designed to evaluate, crawl, and mutate highly complex infrastructure graphs (like Kubernetes or Cloud architectures) safely, concurrently, and lazily.

## Features

- **Turing-Complete DSL:** Supports `import`, `def` (macros), `let ... in` bindings, and Lambdas (`map (\x -> ...)`).
- **Algebraic Data Types:** Natively parses Sum Types (`|`), Records (`{}`), Tuples (`()`), and Collections (`Set`, `List`, `Tree`).
- **Lazy Trampolines:** Evaluates graphs dynamically using Go 1.23 `iter.Seq`, allowing for infinite data structures safely bounded by `take`.
- **Parallel Execution (`||>`):** Mechanically fans-out pipeline branches across a dynamic Goroutine worker pool.
- **Optics & Comonads:** Exposes functional Lenses (`focus`, `update`) and Zippers (`up`, `down`) directly to the scripting language for deep immutability.
- **Compliance Transformers (`EitherT`):** Uses the Error Monad Transformer to silently drop invalid nodes (`filter`) or yield explicit Error Nodes (`expect`) without crashing the pipeline.
- **Kleisli Arrows (`&&&` / `***`):** Mathematically route streams concurrently into Tuples, and apply Applicative logic over Product Types.

## The Language

`gatwd` scripts look and feel like Elm or Haskell. You define modules, build reusable functions, and lazily pipe data through transducers.

```haskell
import "kubernetes_types.gatwd"

-- 1. Define Reusable Functions
def audit_node = uname -a |> expect "Linux" "CRITICAL: Not Linux"
def count_pods = get_pods |> distinct |> fold count

-- 2. Local State Bindings
let cluster_ip = spawn Master: localhost |> focus IP in

-- 3. Execution Pipeline
spawn Worker: cluster_ip |> audit_node |> count_pods
```

## Documentation

* **[DSL Guide](docs/dsl_guide.md)**: A complete handbook on how to write `gatwd` scripts.
* **[Formal Grammar](docs/grammar.md)**: The EBNF specification for the DSL.
* **[Error Handling](docs/error_handling.md)**: How the Engine uses `EitherT`.
* **[Native Adapters](docs/native_adapters.md)**: How to write pure Go functions that bind to the Engine.

## Interactive Tour

Run the interactive tutorials to learn both the language and functional programming!

```bash
go build -o gatwd ./cmd/gatwd/...

./gatwd examples/tour/01_basics.gatwd
./gatwd examples/tour/02_transducers.gatwd
./gatwd examples/tour/06_optics_records.gatwd
./gatwd examples/tour/12_turing_complete.gatwd
```
