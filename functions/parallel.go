package functions

import (
	"context"
	"iter"
	"sync"
)

// ParMap executes a pure function concurrently over a lazy sequence.
//
// In functional programming, if a function is pure (has no side effects),
// evaluating it over a list sequentially is a waste of multi-core CPU time.
// ParMap evaluates `fn` across a worker pool concurrently.
//
// Note: This implementation does not guarantee the order of the output
// sequence. If order is required, a sequential Map must be used.
//
// The function elegantly handles Go 1.23 iterator semantics: if the consumer
// stops reading (yield returns false), a context cancellation propagates
// to all worker goroutines, preventing memory leaks.
//
// Example:
//
//	hosts := iter.Seq[Host]{...}
//	results := ParMap(hosts, 10, func(h Host) Result { return h.HeavyComputation() })
func ParMap[T, U any](seq iter.Seq[T], workers int, fn func(T) U) iter.Seq[U] {
	return func(yield func(U) bool) {
		if workers < 1 {
			workers = 1
		}

		in := make(chan T, workers)
		out := make(chan U, workers)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var wg sync.WaitGroup

		// Start Worker Pool
		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					select {
					case <-ctx.Done():
						return
					case val, ok := <-in:
						if !ok {
							return
						}
						res := fn(val)
						select {
						case <-ctx.Done():
							return
						case out <- res:
						}
					}
				}
			}()
		}

		// Start Producer
		go func() {
			defer close(in)
			for val := range seq {
				select {
				case <-ctx.Done():
					return
				case in <- val:
				}
			}
		}()

		// Start Closer
		go func() {
			wg.Wait()
			close(out)
		}()

		// Consume Output and Yield
		for res := range out {
			if !yield(res) {
				// Consumer stopped early; cancel context to tear down goroutines
				cancel()
				return
			}
		}
	}
}

// ParBind (Parallel FlatMap) executes a monadic bind concurrently.
//
// Like Bind, this function takes elements from a sequence, applies a function
// that returns a new sequence (a sub-sequence), and flattens all resulting
// sub-sequences into a single continuous stream.
//
// Because the work is distributed across a worker pool, the sub-sequences
// are evaluated concurrently. The output order is non-deterministic.
//
// Example:
//
//	// Fetching all interfaces from a list of hosts concurrently
//	allInterfaces := ParBind(hosts, 10, func(h Host) iter.Seq[Interface] {
//		return h.FetchInterfacesViaSSH()
//	})
func ParBind[T, U any](seq iter.Seq[T], workers int, fn func(T) iter.Seq[U]) iter.Seq[U] {
	return func(yield func(U) bool) {
		if workers < 1 {
			workers = 1
		}

		in := make(chan T, workers)
		out := make(chan U, workers)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var wg sync.WaitGroup

		// Start Worker Pool
		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					select {
					case <-ctx.Done():
						return
					case val, ok := <-in:
						if !ok {
							return
						}
						// fn(val) returns a lazy iter.Seq[U]
						subSeq := fn(val)
						for subVal := range subSeq {
							select {
							case <-ctx.Done():
								return
							case out <- subVal:
							}
						}
					}
				}
			}()
		}

		// Start Producer
		go func() {
			defer close(in)
			for val := range seq {
				select {
				case <-ctx.Done():
					return
				case in <- val:
				}
			}
		}()

		// Start Closer
		go func() {
			wg.Wait()
			close(out)
		}()

		// Consume Output and Yield
		for res := range out {
			if !yield(res) {
				cancel()
				return
			}
		}
	}
}
