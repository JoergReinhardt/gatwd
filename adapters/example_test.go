package adapters_test

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/joergreinhardt/gatwd/adapters"
	"github.com/joergreinhardt/gatwd/dsltypes"
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
