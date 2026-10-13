# gatwd: Gophers All The Way Down

`gatwd` is a purely functional Category Theory engine, a Turing-complete Scripting Language, and a bleeding-edge CLI Visualizer written in Go. 

It is explicitly designed to bridge the gap between elegant functional languages (like Haskell/Elm) and raw bare-metal Go performance. `gatwd` mathematically evaluates, crawls, and visualizes highly complex architectural graphs, data structures, and infrastructure networks.

## Features & The Kitchen Sink

- **Turing-Complete DSL:** Supports `import`, `def`, `let ... in` bindings, and **first-class, curried closures with lexical scope** — `def add = \x y -> x + y` is applied in expression position (`add 1 2`), partially applied (`map (add 10)`, `fold add 0`), returned from other functions (`mk 5 10`), and applied in a pipeline with the incoming value last (`5 |> add 3`).
- **Gradual Typing (advisory):** optional signatures `def inc : Int -> Int = …` never change evaluation; `gatwd --check` uses them plus local inference to report wrong argument/result types, arity errors, operators the runtime rejects (`"a" - 1`), unbound names and bad `spawn` seeds — as warnings, with zero false positives on the repo's own scripts.
- **Algebraic Data Types & Constructor Matching:** Natively parses Sum Types (`|`), Records (`{}`), Tuples (`()`), and generic Collections (`Set`, `List`, `Tree`, `BST`, `BTree`, `RBTree`). `match` discriminates on sum-type constructors (`match { Linux -> … | Darwin -> … }`), with list/tuple destructuring and recursion.
- **Exhaustive Functional Operations:** `map`, `filter`/`reject`, `fold` (fixed reducers *and* custom `fold (\acc x -> …) seed`), `flatmap`/bind, `bimap`, and set algebra — implemented across the full type matrix (`List`, `Chunk`, `Set`, `HAMT`, `Tree`, `BST`, `Graph`, `Pair`, `Either`, `Maybe`, `Variant`), each both as a pipeline *step* and as a curried, type-preserving **function** (`hamt_set (hamt_get "k" r) n m`, `set_union a b`, `variant "Circle" 3`) usable inside lambdas and comprehensions. See [`STYLEGUIDE.md` §7](STYLEGUIDE.md).
- **Typed, Weighted Graphs:** a generic persistent directed multigraph (`Graph(T)`) with colored nodes/edges, inferred edges, BFS/DFS, Dijkstra, SCC, topological order, comonadic cursors and lens-based payload updates — the full `graph_*` step family plus expression functions (`graph_scc g`, `graph_reachable id g`) callable from comprehensions.
- **Stream Merge (`<+>`):** `(a <+> b <+> c)` feeds one incoming node through several sub-pipelines and concatenates the results — the ArrowPlus/MonadPlus combinator, used to merge many crawl streams into one.
- **Fail-Loud Tooling:** Unknown steps and failed calls surface `[ERROR] …` instead of vanishing silently, and `gatwd --check <file>` statically validates scripts (unbalanced delimiters, dangling `|>`, non-exhaustive-`match`, typing warnings). The test gate itself fails on silent statements.
- **Gatwd Builtins & Bare Runtime:** A minimal, unopinionated core engine (`cli.NewBareInterpreter()`) paired with exhaustive Go standard library builtins (`cli.NewInterpreter()`) covering collections, IO scanners, network streams, and state checkpointing.
- **Gatwd Editions & Appliances:** Build custom binary distributions (Editions) importing third-party Go libraries, or package self-contained microservices, Kubernetes Operators, and CLI tools (Appliances) shipping with embedded `.gatwd` code (`//go:embed`) and runtime ConfigMap overrides. Reference: **`gatwd-k8s`** (Edition) and **`gatwd-k8s-operator`** (Appliance) — a Kubernetes resource-graph tool whose relations (`Owns`, `Selects`, `RoutesTo`, `Configures`, `Mounts`, `Binds`, inferred `Calls`) and compliance rules are written in the DSL, over live clusters or offline manifests/kustomize output.
- **Bare-Metal Interoperability:** Use `Interpreter.RegisterType` to register native Go types (like `math/big.Int`, `net.IP`, or `time.Time`). The DSL dynamically wraps raw Go memory pointers and routes them mathematically through the execution pipeline at zero-overhead.
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

* **[Architecture & Extensibility](docs/architecture.md)**: Deep dive into the host-guest engine, builtins, custom editions, and appliances.
* **[Conway's Game of Life: Architectural Tiers & Trade-Offs](docs/GAME_OF_LIFE.md)**: Flagship case study demonstrating the 3 tiers of implementation (from Category Theory Comonads to 100% Pure DSL), detailing effort, intimidation, performance, and the Clean Core Prevention Rule.
* **[DSL Guide](docs/dsl_guide.md)**: A complete handbook on how to write `gatwd` scripts.
* **[Formal Grammar](docs/grammar.md)**: The EBNF specification for the DSL.
* **[Native Adapters, Editions & Appliances](docs/native_adapters.md)**: How to write pure Go adapters, register third-party types/steps, and package standalone appliances.
* **[Agentic Workflows & AI](docs/agentic_workflows.md)**: Harnessing the DSL for LLMs, autonomous agents, and K8s Operators.
* **[How-To: Kubernetes Edition & Operator](docs/k8s_howto.md)**: Define tasks with scripts, extend the prelude and graph model (e.g. Secret → database → DBMS), and build, package and deploy the binaries.
* **[Kubernetes Graph: Design & Implementation Record](docs/k8s_graph_design.md)**: The generic graph core and the Kubernetes edition built on it, phase by phase.
* **[Language Roadmap](docs/language_roadmap.md)**: First-class functions, gradual typing and the expression-level function matrix.
* **[Development Playbook](docs/development_playbook.md)**: How work is done here — delivery protocol, gates, traps, performance and handoff protocols.

## Contributing & AI Agents

Start with [`CONTRIBUTING.md`](CONTRIBUTING.md). Rules for humans and AI assistants alike live in [`AGENTS.md`](AGENTS.md) (eight mandates, layers, prevention rules), [`STYLEGUIDE.md`](STYLEGUIDE.md) and [`PROMPTS.md`](PROMPTS.md); reusable skills and task prompts are in [`.agents/`](.agents/README.md). Architecture rules are enforced by tests (`internal/archtest`) and by the DSL gate (`./run_dsl_tests.sh`), not just by prose.


## Testing & Verification

gatwd maintains a 100% self-asserting native DSL test suite alongside standard Go tests:

```bash
# Run all Go package tests, example tests and the architecture guards (layers + Clean Core):
go test ./...

# Run the exhaustive self-asserting DSL suite; it fails on [ERROR], on statements that
# produce no output (a vacuous assert), and on `gatwd --check` structural errors:
./run_dsl_tests.sh

# Benchmark the Kubernetes graph pipeline on synthetic clusters:
benchmarks/run_k8s_graph.sh
```

Every language construct, stream transducer, persistent collection (`HAMT`, `Set`, `Chunk`, `List`), comonad zipper, optics lens (`focus`, `update`), and visual card layout is verified end-to-end within [`tests/dsl_exhaustive_test.gatwd`](tests/dsl_exhaustive_test.gatwd).

## Interactive Tour

Run the interactive tutorials in `examples/tour/` to learn both the language and functional programming (`examples/new_features/showcase.gatwd` is a self-asserting tour of the newest features).

```bash
go build -o gatwd ./cmd/gatwd/...

./gatwd examples/tour/01_basics.gatwd
./gatwd examples/tour/12_turing_complete.gatwd
./gatwd examples/tour/17_rbtree_tomfoolery.gatwd
./gatwd examples/tour/20_visual_tomfoolery.gatwd
```
