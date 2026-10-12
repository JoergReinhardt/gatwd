// Command crashprobe finds the recursion depth at which the native-Go Ackermann
// overflows the goroutine stack. Because Ackermann's compute cost (~4^n) outruns
// its stack depth (~2^n), the *default* 1 GB stack is computationally unreachable
// — so we cap the stack with debug.SetMaxStack to make depth the binding
// constraint, then report the max depth reached before the fatal overflow.
//
// Usage: go run ./benchmarks/native/crashprobe <n> <maxStackBytes>
//   n              A(3, n) input
//   maxStackBytes  goroutine stack cap (e.g. 1048576 for 1 MiB); 0 = Go default
//
// A completed run prints the peak depth; an overflow prints Go's fatal
// "goroutine stack exceeds" and exits non-zero (stack overflow is NOT
// recoverable in Go). Peak depth is reported via a deferred print that still
// runs for a clean return but is lost on overflow — so we also stream the
// running max to stderr periodically.
package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"strconv"
)

var (
	depth    int64
	maxDepth int64
)

func ackermann(m, n int64) int64 {
	depth++
	if depth > maxDepth {
		maxDepth = depth
	}
	defer func() { depth-- }()

	if m == 0 {
		return n + 1
	}
	if n == 0 {
		return ackermann(m-1, 1)
	}
	return ackermann(m-1, ackermann(m, n-1))
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: crashprobe <n> <maxStackBytes>")
		os.Exit(2)
	}
	n, _ := strconv.ParseInt(os.Args[1], 10, 64)
	maxStack, _ := strconv.ParseInt(os.Args[2], 10, 64)
	if maxStack > 0 {
		debug.SetMaxStack(int(maxStack))
	}

	fmt.Fprintf(os.Stderr, "probe: A(3,%d) maxStack=%d bytes\n", n, maxStack)
	got := ackermann(3, n)
	// Only reached if NO overflow occurred.
	fmt.Printf("A(3,%d) = %d  SURVIVED  peakDepth=%d\n", n, got, maxDepth)
}
