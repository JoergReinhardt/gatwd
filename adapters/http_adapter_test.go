package adapters_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/joergreinhardt/gatwd/adapters"
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
)

func TestNativeHTTPAdapter(t *testing.T) {
	// 1. Create a mock HTTP server simulating a microservice registry
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "service-a.local")
		fmt.Fprintln(w, "service-b.local")
		fmt.Fprintln(w, "service-c.local")
	}))
	defer mockServer.Close()

	const IsService = dsltypes.BitFlag(1 << 7)

	// 2. Initialize the Native Adapter with a persistent client
	adapter := adapters.NativeHTTPAdapter(mockServer.Client(), IsService)

	// 3. Register it with the Engine
	engine := evaluator.NewEngine[string]()
	const IsURL = dsltypes.BitFlag(1 << 8)
	engine.Register(IsURL, adapter)

	// 4. Create a stream with the root URL
	initialStream := func(yield func(dsltypes.ASTNode[string]) bool) {
		yield(dsltypes.NewASTNode(mockServer.URL, IsURL))
	}

	// 5. Evaluate!
	resultStream := engine.Evaluate(initialStream)

	var results []string
	for node := range resultStream {
		if !node.Match(IsService) {
			t.Fatalf("Expected output to have IsService flag")
		}
		results = append(results, node.Value)
	}

	if len(results) != 3 {
		t.Fatalf("Expected 3 services discovered, got %d", len(results))
	}
	if results[0] != "service-a.local" {
		t.Fatalf("Unexpected execution output: %s", results[0])
	}
}
