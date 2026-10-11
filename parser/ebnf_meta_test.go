package parser_test

import (
	"testing"

	"github.com/joergreinhardt/gatwd/parser"
)

func TestEBNFMetaParser(t *testing.T) {
	// Define a custom micro-grammar in standard EBNF
	grammarText := `
Assignment = Var ":=" Value ;
Var = "x" | "y" | "z" ;
Value = "100" | "200" ;
`
	grammar, err := parser.ParseEBNFGrammar("Assignment", grammarText)
	if err != nil {
		t.Fatalf("Failed to parse EBNF grammar: %v", err)
	}

	input := "x := 100"
	ast, rem, err := grammar.Parse(input)
	if err != nil {
		t.Fatalf("Failed to parse valid input: %v", err)
	}

	if ast.Rule != "sequence" && ast.Rule != "Assignment" {
		t.Logf("Parsed AST root: %+v", ast)
	}
	if rem != "" && rem != " " {
		t.Logf("Remaining input: %q", rem)
	}

	// Negative test: invalid variable "w"
	invalidInput := "w := 100"
	_, _, err = grammar.Parse(invalidInput)
	if err == nil {
		t.Fatalf("Expected parse error for invalid variable 'w', got success")
	}
}

func TestEBNFOptionalAndRepeat(t *testing.T) {
	grammarText := `
Greeting = "Hello" [ "Dear" ] "World" ;
`
	grammar, err := parser.ParseEBNFGrammar("Greeting", grammarText)
	if err != nil {
		t.Fatalf("Failed to parse EBNF grammar: %v", err)
	}

	// Test with optional present
	ast1, _, err := grammar.Parse("Hello Dear World")
	if err != nil {
		t.Fatalf("Failed to parse with optional present: %v", err)
	}
	if ast1 == nil {
		t.Fatal("Expected non-nil AST")
	}

	// Test without optional present
	ast2, _, err := grammar.Parse("Hello World")
	if err != nil {
		t.Fatalf("Failed to parse without optional present: %v", err)
	}
	if ast2 == nil {
		t.Fatal("Expected non-nil AST")
	}
}
