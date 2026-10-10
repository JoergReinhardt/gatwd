# Style Guide: Gophers All The Way Down

This project walks the line between pure functional programming theory (Category Theory) and pragmatic, modern Go (1.18+ Generics and 1.23+ Iterators). 

When contributing to this project, adhere to the following rules:

## 1. Go Idiomatic First
While the concepts are functional, the syntax must remain strictly idiomatic Go. 
- Use standard Go formatting (`gofmt`).
- Error handling should use multiple return values or `Result[T]` types, avoiding panics.
- Use native Go language features (`iter.Seq`, Generics `[T any]`) rather than deep interface boxing or manual closures whenever possible.
- If there is a conflict between "how Haskell does it" and "how Go does it" (e.g., naming conventions, package layouts), **Go idiomatic conventions win**.

## 2. Naming: The Category Theory Theme
We intentionally use the vocabulary of functional programming and Category Theory to name our abstractions. This is the domain language of the project.
- **Do use:** `Functor`, `Applicative`, `Monad`, `Bind`, `Map`, `Fold`, `FlatMap`, `Maybe`, `Either`, `Sum`, `Product`.
- **Do not** invent Go-specific names for mathematically established concepts (e.g., do not name `Bind` as `ChainExecutions`). 

## 3. Mitigating "Smarty-Pants" Syndrome (The Golden Rule of Docs)
Category Theory vocabulary can be intimidating. We mitigate this through **exceptional, educational documentation**. 

Our standard is the Go standard library's `regexp` package, which beautifully teaches the reader what a regular language actually is. **Do not take functional concepts for granted.** 

### Documentation Requirements:
1. **Explain the Concept:** If you write a type called `Monad`, the `godoc` must explain what a Monad is conceptually in simple terms (e.g., "A context that wraps a value and allows sequencing dependent operations safely").
2. **Explain the Mechanics:** Explain *how* the implementation works (e.g., "This uses an `iter.Seq` to lazily evaluate the operations without blowing up the stack").
3. **Provide Examples:** Every core interface and function must include a practical, readable Go example showing why it is useful.

### Example of Good Documentation:
```go
// Bind (often called FlatMap) chains two operations together. 
//
// In functional programming, this is the defining operation of a "Monad". 
// Imagine you have a lazy sequence of Hosts, and you want to look up the 
// Network Interfaces for each Host. Because the lookup function itself returns 
// a sequence of Interfaces, a simple Map would result in a sequence of sequences.
//
// Bind solves this by applying the function and immediately flattening the 
// resulting sequences into a single continuous stream. 
// 
// Example:
//    // hosts is an iter.Seq[Host]
//    // getInterfaces is func(Host) iter.Seq[Interface]
//    interfaces := Bind(hosts, getInterfaces)
func Bind[T, U any](seq iter.Seq[T], f func(T) iter.Seq[U]) iter.Seq[U] {
    // ...
}
```
