package parser_test

import (
	"fmt"

	"github.com/joergreinhardt/gatwd/parser"
	"github.com/joergreinhardt/gatwd/visualizer"
)

func ExampleCombinator_parseFanOut() {
	opts := visualizer.ParseOptions("--curved --border")

	arrowParser := parser.ParseFanOut()
	var left, right string
	for res := range arrowParser("(getProfile &&& getBilling)") {
		left = res.Fst.Fst
		right = res.Fst.Snd
		break
	}

	fmt.Print(visualizer.Format(map[string]string{
		"LeftArrow":  left,
		"RightArrow": right,
	}, opts))

	// Output:
	// ╭───────────────────────────────╮
	// │ ╰─ Record                     │
	// │     ├─ LeftArrow: getProfile  │
	// │     ╰─ RightArrow: getBilling │
	// ╰───────────────────────────────╯
}
