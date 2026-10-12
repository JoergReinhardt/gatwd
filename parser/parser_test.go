package parser_test

import (
	"testing"

	"github.com/joergreinhardt/gatwd/parser"
)

func TestParserAndEvaluator(t *testing.T) {
	script := `
type Server = LinuxHost | WindowsHost
type Client = Android | iOS
type Endpoint = Server | Client

spawn LinuxHost: 192.168.0.1
spawn Android: phone-1
`
	p := parser.NewParser()
	astStream := p.Parse(script)

	// We expect 2 nodes spawned. Let's run them through an Engine to see if the dynamic BitFlag matched.
	
	count := 0
	for node := range astStream {
		count++
		if node.Value == "192.168.0.1" {
			// It should have the flag for LinuxHost
			// (we can't easily assert exactly which bit it is, but it exists!)
			if node.DSLType == 0 {
				t.Fatalf("Bitflag not allocated")
			}
		}
	}

	if count != 2 {
		t.Fatalf("Expected 2 nodes spawned, got %d", count)
	}
}

func TestParser_SumTypeVariants(t *testing.T) {
	script := `
type Status = Active | Paused | Stopped
spawn Status: Active
spawn Status: Stopped(404)
`
	p := parser.NewParser()
	var nodes []any
	for node := range p.Parse(script) {
		nodes = append(nodes, node.Value)
	}

	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}

	activeStr := nodes[0].(interface{ String() string }).String()
	if activeStr != "Active" {
		t.Errorf("expected 'Active', got %q", activeStr)
	}

	stoppedStr := nodes[1].(interface{ String() string }).String()
	if stoppedStr != "Stopped(404)" {
		t.Errorf("expected 'Stopped(404)', got %q", stoppedStr)
	}
}

func TestParser_IsConstructorAndMakeVariant(t *testing.T) {
	p := parser.NewParser()
	// Register the sum type so its branches become known constructors.
	for range p.Parse("type Color = Red | Green | Blue\n") {
	}

	// IsConstructor: true only for declared branches, not the type itself or
	// arbitrary identifiers (which the match evaluator treats as fresh variables).
	for _, ctor := range []string{"Red", "Green", "Blue"} {
		if !p.IsConstructor(ctor) {
			t.Errorf("expected %q to be a constructor", ctor)
		}
	}
	for _, notCtor := range []string{"Color", "live", "Purple", "x"} {
		if p.IsConstructor(notCtor) {
			t.Errorf("expected %q NOT to be a constructor", notCtor)
		}
	}

	// MakeVariant: boxes a declared branch as a Variant carrying the branch tag.
	v, ok := p.MakeVariant("Color", "Green")
	if !ok || !v.Is("Green") {
		t.Fatalf("MakeVariant(Color, Green) = %+v, %v; want Variant{Tag:Green}, true", v, ok)
	}
	// A value that is not a branch of the type yields no variant.
	if _, ok := p.MakeVariant("Color", "Purple"); ok {
		t.Errorf("MakeVariant(Color, Purple) should fail: Purple is not a branch")
	}
	// A non-sum type name yields no variant.
	if _, ok := p.MakeVariant("NotAType", "Green"); ok {
		t.Errorf("MakeVariant(NotAType, ...) should fail")
	}
}
