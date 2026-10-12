// Command ackermann is the native-Go baseline for the gatwd Ackermann benchmark.
//
// It computes the same A(3, n) workload as benchmarks/ackermann.gatwd using plain
// int64 recursion, so the two runtimes can be timed head-to-head. The ratio of DSL
// wall-clock to native wall-clock is the interpreter's "interpretation tax".
//
// Usage: go run ./benchmarks/native 6   (default n = 6, matching the .gatwd script)
package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// ackermann is the textbook two-argument Ackermann function on int64.
func ackermann(m, n int64) int64 {
	if m == 0 {
		return n + 1
	}
	if n == 0 {
		return ackermann(m-1, 1)
	}
	return ackermann(m-1, ackermann(m, n-1))
}

func main() {
	n := int64(6)
	if len(os.Args) > 1 {
		parsed, err := strconv.ParseInt(os.Args[1], 10, 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid n %q: %v\n", os.Args[1], err)
			os.Exit(1)
		}
		n = parsed
	}

	start := time.Now()
	got := ackermann(3, n)
	elapsed := time.Since(start)

	// Closed form: A(3, n) = 2^(n+3) - 3.
	want := int64(1)<<(uint(n)+3) - 3
	if got != want {
		fmt.Fprintf(os.Stderr, "A(3,%d) = %d, want %d\n", n, got, want)
		os.Exit(1)
	}

	fmt.Printf("A(3,%d) = %d  (native Go, %s)\n", n, got, elapsed)
}
