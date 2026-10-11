package adapters_test

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/joergreinhardt/gatwd/adapters"
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/functions"
)

func TestListenerAndDial_StreamingIntegration(t *testing.T) {
	// 1. Start a local TCP listener
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen on tcp: %v", err)
	}
	defer ln.Close()

	addr := ln.Addr().String()
	flag := dsltypes.BitFlag(1 << 3)

	// 2. Start an echo server goroutine using ListenerToSeq and ConnToLines
	serverDone := make(chan struct{})
	go func() {
		defer close(serverDone)
		connStream := adapters.ListenerToSeq(ln, flag)
		for node := range connStream {
			if node.Value.IsRight() {
				val, _ := node.Value.UnwrapRight()
				conn := val.(net.Conn)
				// Read line from client and echo it back with uppercase prefix
				lineStream := adapters.ConnToLines(conn, flag)
				for lineNode := range lineStream {
					if lineNode.Value.IsRight() {
						lineVal, _ := lineNode.Value.UnwrapRight()
						text := lineVal.(string)
						_, _ = fmt.Fprintf(conn, "ECHO:%s\n", text)
					}
				}
			}
		}
	}()

	// 3. Use StepDial to send requests over the network
	dialStep := adapters.StepDial("tcp", addr, 2*time.Second)
	engine := evaluator.NewEngine[functions.Either[error, any]]()

	trigger := func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
		reqNode := dsltypes.NewASTNode(
			functions.Right[error, any](any("hello_gatwd")),
			flag,
		)
		yield(reqNode)
	}

	resultStream := engine.EvaluatePipeline(trigger, []evaluator.PipelineStep[functions.Either[error, any]]{dialStep})

	var gotResponses []string
	for node := range resultStream {
		if node.Value.IsLeft() {
			err, _ := node.Value.UnwrapLeft()
			t.Fatalf("unexpected dial error: %v", err)
		}
		val, _ := node.Value.UnwrapRight()
		gotResponses = append(gotResponses, val.(string))
	}

	if len(gotResponses) != 1 {
		t.Fatalf("expected 1 response, got %d", len(gotResponses))
	}
	if gotResponses[0] != "ECHO:hello_gatwd" {
		t.Fatalf("expected 'ECHO:hello_gatwd', got %q", gotResponses[0])
	}

	// 4. Close listener to ensure ListenerToSeq terminates cleanly
	_ = ln.Close()
	select {
	case <-serverDone:
		// Clean server shutdown
	case <-time.After(1 * time.Second):
		t.Fatalf("timed out waiting for server to shut down cleanly")
	}
}
