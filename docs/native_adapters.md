# Native Go Adapters & The Closure Pattern

While `gatwd`'s DSL allows users to spawn generic CLI execution wrappers (`interactive.GenericExecAdapter`) on the fly, **Native Go Adapters** are the gold standard for extending the Host Engine.

Native adapters allow you to interact directly with Go's standard library (or heavy third-party SDKs like `k8s.io/client-go`) without paying the performance cost of shelling out to external binaries.

## The Problem: State vs. Purity

The `gatwd` evaluator loop requires all adapters to conform to a strict, pure mathematical function signature (the IO Monad):
```go
type Adapter[T any] func(ASTNode[T]) iter.Seq[ASTNode[T]]
```

If the signature only accepts an `ASTNode`, how do you pass in a database connection pool (`*sql.DB`), a customized HTTP client (`*http.Client`), or authentication tokens?

## The Solution: The Closure Pattern

We solve this using **Higher-Order Functions** and **Closures**. 

You write a function that accepts your heavy, stateful Go dependencies and returns the pure `Adapter` function back to the engine.

### Reference Implementation: Native HTTP Crawler
*(You can view the full code in `adapters/http_adapter.go`)*

```go
func NativeHTTPAdapter(client *http.Client, targetFlag dsltypes.BitFlag) evaluator.Adapter[string] {
    
    // 1. The outer function holds the state.
    // The `client` (with its connection pool, TLS certs, and timeouts) is 
    // captured in this closure.

    // 2. We return the pure monadic function expected by the Evaluator.
    return func(node dsltypes.ASTNode[string]) iter.Seq[dsltypes.ASTNode[string]] {
        return func(yield func(dsltypes.ASTNode[string]) bool) {
            
            // 3. Perform the Native Go side-effect
            resp, err := client.Get(node.Value)
            if err != nil { return }
            defer resp.Body.Close()

            // 4. Lazily yield results directly into the DSL Trampoline
            scanner := bufio.NewScanner(resp.Body)
            for scanner.Scan() {
                childNode := dsltypes.NewASTNode(scanner.Text(), targetFlag)
                
                if !yield(childNode) {
                    // Safe short-circuit! If the evaluator stops reading 
                    // (e.g. ParMap cancels context), we immediately close 
                    // the response body via defer and exit.
                    return 
                }
            }
        }
    }
}
```

## Why this is Architecturally Flawless

1. **Massive Performance:** The `*http.Client` (and its underlying TCP connection pool) is initialized exactly once. When the graph crawler evaluates 1,000,000 URLs, it reuses the exact same optimized Go client instead of spawning 1,000,000 `curl` processes.
2. **Absolute Purity:** The heavy state management, error handling, and resource teardown (`defer resp.Body.Close()`) is completely isolated inside the closure. The Evaluator Engine remains mathematically pristine.
3. **Short-Circuit Safety:** Because it uses Go 1.23 `iter.Seq`, if the engine cancels the graph traversal early (e.g., stopping after finding 10 servers), the `!yield()` check fails, immediately triggering the `defer` block to close the network connection. **No goroutine leaks!**

## Writing Your Own

To write a Native Adapter for AWS, Kubernetes, or Postgres, simply follow the closure pattern:
1. Pass your `Session` or `Client` into an outer wrapper function.
2. Return `evaluator.Adapter[string]`.
3. Use `yield(childNode)` to bridge the native Go structs back into the DSL's `ASTNode` graph!
