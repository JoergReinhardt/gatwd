package adapters_test

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/joergreinhardt/gatwd/adapters"
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/visualizer"
)

func ExampleScannerToSeq() {
	input := "token1\ntoken2\ntoken3"
	scanner := bufio.NewScanner(strings.NewReader(input))
	flag := dsltypes.BitFlag(1)

	seq := adapters.ScannerToSeq(scanner, flag)

	for node := range seq {
		val, _ := node.Value.UnwrapRight()
		fmt.Printf("%s\n", string(val.([]byte)))
	}

	// Output:
	// token1
	// token2
	// token3
}

func ExampleScannerToSeq_visualized() {
	input := "node-01:ready"
	scanner := bufio.NewScanner(strings.NewReader(input))
	flag := dsltypes.BitFlag(1)

	type StatusRecord struct {
		Node  string
		State string
	}

	seq := adapters.ScannerToSeq(scanner, flag)
	opts := visualizer.ParseOptions("--curved --border")

	for node := range seq {
		val, _ := node.Value.UnwrapRight()
		parts := strings.Split(string(val.([]byte)), ":")
		rec := StatusRecord{Node: parts[0], State: parts[1]}
		fmt.Print(visualizer.Format(rec, opts))
	}

	// Output:
	// ╭──────────────────────╮
	// │ ╰─ StatusRecord      │
	// │     ├─ Node: node-01 │
	// │     ╰─ State: ready  │
	// ╰──────────────────────╯
}
