# Native Go Adapters

The `gatwd` DSL engine acts purely as an algebraic routing pipeline. It uses Go's `iter.Seq` trampolines to stream generic `ASTNode[any]` pointers. 

Because the stream relies on `any`, **you can yield and process actual, raw Go memory pointers (structs) directly through the functional graph.**

## 1. Registering Native Types (`Interpreter.RegisterTypeFactory`)

When you invoke the CLI or build a `main.go`, you can seamlessly register third-party Go constructors into the DSL Parser. 

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

Now, inside your `script.gatwd`:
```haskell
-- Define a logical Alias that targets the Go Type Factory
type Balance = GoBigInt

-- The DSL will dynamically invoke big.Int.SetString("999999")!
spawn Balance: 999999 |> math_add 1
```

## 2. Registering Pipeline Transducers

You can register custom Go logic that processes these native structs by utilizing the `plugins` registry.

```go
package plugins

import (
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/functions"
)

func StepBigIntAdd(amountStr string) evaluator.PipelineStep[functions.Either[error, any]] {
	return func(stream iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]) iter.Seq[...] {
		return func(yield func(...) bool) {
			for node := range stream {
				if node.Value.IsLeft() { continue }
				
				// 1. Unwrap the dynamic Any pointer
				val, _ := node.Value.UnwrapRight()
				
				// 2. Safely type assert to your native Go Primitive!
				bigIntVal := val.(*big.Int)
				
				// 3. Do blazing fast, bare-metal Go computation
				// ...
				
				yield(newNode)
			}
		}
	}
}
```
