package main

import (
	"fmt"
	"iter"

	"github.com/joergreinhardt/gatwd/arrows"
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/visualizer"
)

func main() {
	header := visualizer.Format("Operator: Kleisli Arrow Fan-Out (&&&)", visualizer.ParseOptions("--heavy --border --color=cyan"))
	fmt.Print(header)

	// Adapter A: Fetches user profile
	getProfile := func(node dsltypes.ASTNode[string]) iter.Seq[string] {
		return func(yield func(string) bool) {
			yield(node.Value + " -> Premium User")
		}
	}

	// Adapter B: Fetches user billing history
	getBilling := func(node dsltypes.ASTNode[string]) iter.Seq[string] {
		return func(yield func(string) bool) {
			yield("$42.00 / month")
		}
	}

	// Combine them into a FanOut Arrow (&&&)
	fanOutArrow := arrows.FanOut(getProfile, getBilling)

	userFlag := dsltypes.BitFlag(1 << 3)
	dsltypes.RegisterTypeFlag(userFlag, "AccountSummary")
	userNode := dsltypes.NewASTNode("User_Joerg", userFlag)

	type FanOutResult struct {
		Subject string
		Profile string
		Billing string
	}

	for pair := range fanOutArrow(userNode) {
		rec := FanOutResult{
			Subject: userNode.Value,
			Profile: pair.Fst,
			Billing: pair.Snd,
		}
		fmt.Print(visualizer.FormatAST(userNode.DSLType, rec, visualizer.ParseOptions("--curved --border --color=green")))
	}
}
