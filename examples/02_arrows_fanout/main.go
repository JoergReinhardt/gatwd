package main

import (
	"fmt"
	"iter"

	"github.com/joergreinhardt/gatwd/arrows"
	"github.com/joergreinhardt/gatwd/dsltypes"
)

func main() {
	fmt.Println("=== Operator: Kleisli Arrow Fan-Out (&&&) ===")

	// Kleisli Arrows represent A -> Seq[B]. This perfectly matches our Adapters!
	
	// Adapter A: Fetches user profile
	getProfile := func(node dsltypes.ASTNode[string]) iter.Seq[string] {
		return func(yield func(string) bool) {
			yield(node.Value + " -> [Profile: Premium User]")
		}
	}

	// Adapter B: Fetches user billing history
	getBilling := func(node dsltypes.ASTNode[string]) iter.Seq[string] {
		return func(yield func(string) bool) {
			yield(node.Value + " -> [Billing: $42.00]")
		}
	}

	// Combine them into a FanOut Arrow (&&&)
	// This routes the incoming user node to BOTH adapters, and zips the results together!
	fanOutArrow := arrows.FanOut(getProfile, getBilling)

	userNode := dsltypes.NewASTNode("User_Joerg", 1)
	
	fmt.Println("Routing UserNode through FanOut Arrow (getProfile &&& getBilling)...")
	for pair := range fanOutArrow(userNode) {
		fmt.Printf("Result Pair: \n Left: %s\n Right: %s\n", pair.Fst, pair.Snd)
	}
}
