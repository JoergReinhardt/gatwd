package functions_test

import (
	"context"
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/functions"
)

func TestStateMonad(t *testing.T) {
	// Standard counter state: State[int, string]
	increment := func(label string) functions.State[int, string] {
		return func(s int) (string, int) {
			return label, s + 1
		}
	}

	pipeline := functions.BindState(increment("step1"), func(lbl1 string) functions.State[int, string] {
		return functions.BindState(increment("step2"), func(lbl2 string) functions.State[int, string] {
			return functions.ReturnState[int, string](lbl1 + " -> " + lbl2)
		})
	})

	val, finalState := pipeline.Run(10)
	if finalState != 12 {
		t.Errorf("expected final state 12, got %d", finalState)
	}
	if val != "step1 -> step2" {
		t.Errorf("expected 'step1 -> step2', got %q", val)
	}

	// Test Get and Put
	modifyPipeline := functions.BindState(functions.Get[int](), func(curr int) functions.State[int, int] {
		return functions.BindState(functions.Put(curr*5), func(_ struct{}) functions.State[int, int] {
			return functions.ReturnState[int, int](curr + 1)
		})
	})
	resVal, resState := modifyPipeline.Run(4)
	if resVal != 5 || resState != 20 {
		t.Errorf("expected val 5 and state 20, got val %d, state %d", resVal, resState)
	}
}

func TestStateContext(t *testing.T) {
	step := func(amount int) functions.StateContext[int, int] {
		return func(ctx context.Context, s int) (int, int, error) {
			select {
			case <-ctx.Done():
				return 0, s, ctx.Err()
			default:
				return s + amount, s + amount, nil
			}
		}
	}

	ctx := context.Background()
	chained := functions.BindStateContext(step(10), func(val int) functions.StateContext[int, int] {
		return step(val * 2)
	})

	v, s, err := chained.Run(ctx, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s != 45 || v != 45 {
		t.Errorf("expected 45, got %d (val %d)", s, v)
	}

	// Test cancellation
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel() // canceled immediately
	_, sCan, errCan := chained.Run(canceledCtx, 5)
	if errCan == nil {
		t.Errorf("expected context cancellation error")
	}
	if sCan != 5 {
		t.Errorf("expected state preserved at 5, got %d", sCan)
	}
}

func TestReaderMonad(t *testing.T) {
	type Config struct {
		Host string
		Port int
	}

	getHost := functions.Asks(func(c Config) string { return c.Host })
	getPort := functions.Asks(func(c Config) int { return c.Port })

	renderURL := functions.BindReader(getHost, func(h string) functions.Reader[Config, string] {
		return functions.BindReader(getPort, func(p int) functions.Reader[Config, string] {
			return functions.ReturnReader[Config, string](h + ":" + strings.TrimSpace(string(rune('0'+p))))
		})
	})

	url := renderURL.Run(Config{Host: "localhost", Port: 8})
	if url != "localhost:8" {
		t.Errorf("expected 'localhost:8', got %q", url)
	}
}

func TestWriterMonad(t *testing.T) {
	concatLog := func(a, b []string) []string {
		return append(append([]string{}, a...), b...)
	}

	w1 := functions.Writer[[]string, int]{Value: 10, Log: []string{"computed 10"}}
	w2 := functions.BindWriter(w1, concatLog, func(v int) functions.Writer[[]string, int] {
		return functions.Writer[[]string, int]{Value: v * 2, Log: []string{"doubled to 20"}}
	})

	val, log := w2.Run()
	if val != 20 {
		t.Errorf("expected val 20, got %d", val)
	}
	if len(log) != 2 || log[0] != "computed 10" || log[1] != "doubled to 20" {
		t.Errorf("expected 2 log entries, got %v", log)
	}
}

func TestContinuationMonad_CallCC(t *testing.T) {
	// Function that divides 100 by divisor, using CallCC to escape early on zero
	safeDiv := func(n, d int) functions.Cont[string, int] {
		return functions.CallCC(func(exit func(int) functions.Cont[string, int]) functions.Cont[string, int] {
			if d == 0 {
				return exit(-1) // early escape!
			}
			return functions.ReturnCont[string, int](n / d)
		})
	}

	// Case 1: normal division
	res1 := safeDiv(100, 4).Run(func(val int) string {
		if val == -1 {
			return "div by zero error"
		}
		return "result ok"
	})
	if res1 != "result ok" {
		t.Errorf("expected 'result ok', got %q", res1)
	}

	// Case 2: zero divisor with early exit short-circuit
	res2 := safeDiv(100, 0).Run(func(val int) string {
		if val == -1 {
			return "div by zero error"
		}
		return "result ok"
	})
	if res2 != "div by zero error" {
		t.Errorf("expected 'div by zero error', got %q", res2)
	}
}
