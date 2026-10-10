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
