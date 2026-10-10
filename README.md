# gatwd: Gophers All The Way Down

`gatwd` is a purely functional Category Theory engine, a Turing-complete Scripting Language, and a bleeding-edge CLI Visualizer written in Go. 

It is explicitly designed to bridge the gap between elegant functional languages (like Haskell/Elm) and raw bare-metal Go performance. `gatwd` mathematically evaluates, crawls, and visualizes highly complex architectural graphs, data structures, and infrastructure networks.

## Features & The Kitchen Sink

- **Turing-Complete DSL:** Supports `import`, `def` (macros), `let ... in` bindings, and Lambdas (`map (\x -> ...)`).
- **Algebraic Data Types:** Natively parses Sum Types (`|`), Records (`{}`), Tuples (`()`), and generic Collections (`Set`, `List`, `Tree`, `BST`, `BTree`, `RBTree`).
- **Bare-Metal Interoperability:** Use `Interpreter.RegisterTypeFactory` to register native Go types (like `math/big.Int`, `net.IP`, or `time.Time`). The DSL dynamically wraps raw Go memory pointers and routes them mathematically through the execution pipeline at zero-overhead.
- **Extreme Visualization Engine ("Blackjack & Hookers"):** A dynamic, reflection-based UTF-8 visualizer. Supports curved `╭─╮` or heavy `┏━┓` borders, flow-graph sub-boxes, `Head/Tail` array truncation, and conditional ANSI colorization by structural type or node value. Also outputs direct DOT/Graphviz!
- **Lazy Trampolines:** Evaluates graphs dynamically using Go 1.23 `iter.Seq`, allowing for infinite data structures.
- **Optics & Comonads:** Exposes functional Lenses (`focus`, `update`) and Zippers (`up`, `down`) directly to the scripting language for deep, immutable tree navigation.
- **Compliance Transformers (`EitherT`):** Uses the Error Monad Transformer to silently drop invalid nodes (`filter`), assert states (`expect`), or elegantly recover from Type Assertion panics seamlessly.

## The Language

`gatwd` scripts execute raw Go logic wrapped in beautiful Category Theory combinators. 

```haskell
import "kubernetes_types.gatwd"

-- 1. Bind to Native Go Types 
type Balance = BigInt
type IPAddress = GoIP

-- 2. Define Reusable Transducers
def audit_node = uname -a |> expect "Linux" "CRITICAL: Not Linux"
def parse_access = parse_record Owner Perms |> focus Owner

-- 3. Execution Pipeline (Flow-Graph Visualization!)
-- The engine dynamically injects Go math/big.Int addition, wraps the Output in a 
-- Kleisli FanOut Tuple (&&&), and renders a beautiful UTF-8 Tree.
spawn Balance: 999999999999999999 |> math_add 1 |> (audit_node &&& parse_access) |> visualize --box-all --heavy --color-type=Record:cyan
```

## Documentation

* **[DSL Guide](docs/dsl_guide.md)**: A complete handbook on how to write `gatwd` scripts.
* **[Formal Grammar](docs/grammar.md)**: The EBNF specification for the DSL.
* **[Native Plugins & Go Types](docs/native_adapters.md)**: How to write pure Go functions and register third-party Types seamlessly.

## Interactive Tour

Run the 20 interactive tutorials to learn both the language and functional programming!

```bash
go build -o gatwd ./cmd/gatwd/...

./gatwd examples/tour/01_basics.gatwd
./gatwd examples/tour/12_turing_complete.gatwd
./gatwd examples/tour/17_rbtree_tomfoolery.gatwd
./gatwd examples/tour/20_visual_tomfoolery.gatwd
```
