# Native Go Adapters, Gatwd Editions & Appliances

The `gatwd` DSL engine acts as an algebraic routing pipeline. It uses Go's `iter.Seq` trampolines to stream generic `ASTNode[any]` pointers. 

Because the stream relies on `any`, **you can yield and process actual, raw Go memory pointers (structs) directly through the functional graph at bare-metal speed.**

---

## 1. Registering Native Types (`Interpreter.RegisterType`)

When building a custom CLI, server, or operator, you can seamlessly register third-party Go constructors into the DSL Parser. 

```go
package main

import (
	"math/big"
	"github.com/joergreinhardt/gatwd/pkg/cli"
)

func main() {
	interpreter := cli.NewInterpreter()
	
	// Register math/big.Int!
	interpreter.RegisterType("GoBigInt", func(s string) any {
		n := new(big.Int)
		n.SetString(s, 10)
		return n // Yields *big.Int directly to the DSL stream!
	})
	
	interpreter.LoadScript("script.gatwd")
}
```

Inside your `script.gatwd`:
```haskell
-- Define a logical Alias targeting the Go Type Factory
type Balance = GoBigInt

-- The DSL dynamically invokes big.Int.SetString("999999")!
spawn Balance: 999999 |> math_add 1
```

---

## 2. Ergonomic Step Constructors (`MapStep`, `FilterStep`, `ArgStep`)

Contributors and authors do **not** need to write complex `iter.Seq` closures, unpack `functions.Either[error, any]`, or manage AST nodes manually. Gatwd provides high-level functional constructors:

* **`cli.MapStep(name, fn)`**: Transforms values directly:
  ```go
  interp.RegisterStep(cli.MapStep("mask_secret", func(val any) any {
      return fmt.Sprintf("***[%v]***", val)
  }))
  ```
* **`cli.FilterStep(name, predicate)`**: Keeps items matching a predicate:
  ```go
  interp.RegisterStep(cli.FilterStep("only_numbers", func(val any) bool {
      _, ok := val.(int)
      return ok
  }))
  ```
* **`cli.ArgStep(prefix, fn)`**: Parses step parameters (e.g. `bigint_mult 10`):
  ```go
  interp.RegisterStep(cli.ArgStep("bigint_mult", func(arg string, val any) (any, error) {
      factor, _ := new(big.Int).SetString(arg, 10)
      bi := val.(*big.Int)
      return new(big.Int).Mul(bi, factor), nil
  }))
  ```
* **`cli.SinkStep(name, fn)`**: Performs side effects (logging, metrics) without mutating items.

*(For advanced use cases requiring full control over the `iter.Seq` stream trampoline or error channels, raw `StepHandler` remains fully accessible).*

---

## 3. Gatwd Editions vs. Gatwd Appliances

When extending Gatwd with domain-specific logic, implementations fall into two packaging patterns:

### Gatwd Editions (or Distributions)
A custom Go binary (e.g., `cmd/gatwd-cloud` or `cmd/gatwd-db`) that imports `github.com/joergreinhardt/gatwd/pkg/cli`, bundles third-party Go libraries (AWS SDK, database drivers, Kubernetes `client-go`), and exposes custom types and transducers via `RegisterType` and `RegisterStep`. Users run arbitrary external `.gatwd` scripts against this enriched runtime.

### Gatwd Appliances
A self-contained, single-purpose release (such as a Kubernetes Operator, event processing microservice, or compliance daemon) that compiles pre-written `.gatwd` DSL rules directly into the static binary using `//go:embed` and `interp.LoadScriptFS(fsys, path)`:
- Ships with embedded default rules and zero file dependencies.
- **Mutable Overrides & Hot-Reloading**: Allows live cluster administrators to override rules via mounted ConfigMaps (`/etc/gatwd/rules.gatwd`). When the ConfigMap updates, the appliance re-reads the script and atomically swaps the active interpreter (`atomic.Pointer[cli.Interpreter]`) with zero downtime.
- See [`examples/hot_reload_appliance/main.go`](../examples/hot_reload_appliance/main.go) for an executable reference implementation of hot-reloading, and [`examples/kubernetes_repl/main.go`](../examples/kubernetes_repl/main.go) for Kubernetes dynamic crawling.
- See [docs/architecture.md Section 5.5](architecture.md#55-zero-downtime-hot-reloading--configmap-mutation-architecture) for full architectural specifications.



---

## 4. Domain Editions: Adapters, Shims and Embedded Preludes

A *domain* Edition (Kubernetes is the reference) is assembled from four generic pieces; none of them puts domain logic in Go or in core. The full recipe is the `edition-appliance` skill (`.agents/skills/edition-appliance/SKILL.md`).

**1. An adapter does the real I/O** and yields *structured* values, not strings (`evaluator.Adapter[any]`: `func(node) iter.Seq[node]`). Errors become `Left` nodes. Keep third-party SDK imports confined to the adapter file, and give the adapter an **offline twin** that yields the same payload shape, so downstream DSL cannot tell live from offline.

**2. `cli.AdapterStep(keyword, adapter)` lifts any adapter to a pipeline step** (it applies the `EitherT` lift and `StepAdapter` for you):

```go
e := cli.NewEdition("gatwd-k8s").
    WithStep(cli.AdapterStep("k8s_pods", adapters.NativeKubernetesResourceAdapter(client, podsGVR)))
```

**3. A thin shim package** (`pkg/<domain>`) constructs the adapters and exposes one assembly function. It contains no algorithm:

```go
func ApplyEdition(e *cli.Edition, client dynamic.Interface) *cli.Edition {
    e = e.WithStep(ManifestStep("manifests")).WithStep(KustomizeStep("kustomize"))
    if client != nil { // nil client ⇒ offline-only edition (CI, tests)
        for _, c := range CoreCrawls { e = e.WithStep(CrawlStep(c.Keyword, client, c.GVR)) }
    }
    return e.WithPreludeFS(preludeFS, PreludePath)
}
```

**4. An embedded DSL prelude defines the domain.** `Edition.WithPrelude(script)` / `WithPreludeFS(fsys, path)` evaluate a `.gatwd` library after the steps are registered; the prelude may `import` its sibling files from the same `embed.FS`. Vocabulary (`type NodeColor = Pod | Service | …`), ingest, relations and rules all live there, written with closures, folds and comprehensions (see [`dsl_guide.md`](dsl_guide.md) §§19–22).

**Merging producers.** `(k8s_pods <+> k8s_services <+> k8s_deployments)` feeds one downstream pipeline from many adapters; the prelude names the list once (`def k8s_resources = (…)`).

**Testing contract.** Use fixtures and a fake client built from them, a parity test that runs the same prelude over live and offline sources, a negative fixture that violates every rule, and skip (not fail) when an optional external binary is absent. See `pkg/k8sgraph/k8sgraphtest` and `pkg/k8sgraph/live_parity_test.go`.
