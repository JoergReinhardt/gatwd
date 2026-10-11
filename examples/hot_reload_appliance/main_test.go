package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAppliance_HotReloadingAndFallback(t *testing.T) {
	tmpDir := t.TempDir()
	overrideFile := filepath.Join(tmpDir, "rules.gatwd")

	// 1. Initialize appliance with non-existent override file -> must load embedded fallback
	app, err := NewAppliance(overrideFile)
	if err != nil {
		t.Fatalf("failed to initialize appliance: %v", err)
	}

	thresh, ok := app.GetConstant("THRESHOLD")
	if !ok || thresh != "80" {
		t.Fatalf("expected embedded default THRESHOLD '80', got %v", thresh)
	}

	res := app.Execute("spawn HealthStatus: OK |> eval_status")
	if len(res) == 0 || res[0] != "OK" {
		t.Fatalf("expected execution to yield 'OK', got %v", res)
	}

	// 2. Simulate Kubernetes ConfigMap update by writing to overrideFile
	updatedDSL := `
type HealthStatus = OK | CRITICAL
const THRESHOLD = 42
def eval_status = assert CRITICAL
`
	if err := os.WriteFile(overrideFile, []byte(updatedDSL), 0644); err != nil {
		t.Fatalf("failed to write simulated ConfigMap file: %v", err)
	}

	// Trigger hot-reload
	if err := app.Reload(); err != nil {
		t.Fatalf("failed to hot-reload: %v", err)
	}

	threshUpdated, ok := app.GetConstant("THRESHOLD")
	if !ok || threshUpdated != "42" {
		t.Fatalf("expected updated THRESHOLD '42', got %v", threshUpdated)
	}

	resUpdated := app.Execute("spawn HealthStatus: CRITICAL |> eval_status")
	if len(resUpdated) == 0 || resUpdated[0] != "CRITICAL" {
		t.Fatalf("expected execution after reload to yield 'CRITICAL', got %v", resUpdated)
	}

	// 3. Simulate ConfigMap deletion -> must fall back cleanly to embedded rules
	if err := os.Remove(overrideFile); err != nil {
		t.Fatalf("failed to remove override file: %v", err)
	}

	if err := app.Reload(); err != nil {
		t.Fatalf("failed to reload after deleting override: %v", err)
	}

	threshRevert, ok := app.GetConstant("THRESHOLD")
	if !ok || threshRevert != "80" {
		t.Fatalf("expected reverted THRESHOLD '80', got %v", threshRevert)
	}
}
