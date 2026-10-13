package cli

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// Tests for the Appliance side of the execution policy (W7): an override is a script that did
// not come from the command line, so it starts with no capability and only the process's own
// flags can grant one.

func TestApplianceDefaultsToDenyAll(t *testing.T) {
	app, err := NewAppliance(WithEmbeddedScript(toyRules), WithRequiredDefs("entry"))
	if err != nil {
		t.Fatal(err)
	}
	if app.Policy() != DenyAll() {
		t.Fatalf("Appliance policy = %v, want none", app.Policy())
	}
	if app.Interpreter().Policy() != DenyAll() {
		t.Fatalf("the rules' interpreter runs under %v, want none", app.Interpreter().Policy())
	}
	app2, _ := NewAppliance(WithEmbeddedScript(toyRules), WithPolicy(Policy{AllowExec: true}))
	if app2.Interpreter().Policy() != (Policy{AllowExec: true}) {
		t.Fatalf("WithPolicy did not reach the interpreter: %v", app2.Interpreter().Policy())
	}
}

func TestApplianceFlagsMapToThePolicy(t *testing.T) {
	var got Policy
	tree := NewApplianceTree(ApplianceSpec{
		Name: "toy",
		Bind: func(cmd *cobra.Command) ApplianceSetup {
			return func(s ApplianceSettings) (*Appliance, string, error) {
				got = s.Policy
				return nil, "", UsageError(errors.New("stop after reading the settings"))
			}
		},
	})
	rows := []struct {
		argv []string
		want Policy
	}{
		{[]string{"check"}, DenyAll()},
		{[]string{"check", "--allow-exec"}, Policy{AllowExec: true}},
		{[]string{"check", "--allow-fs"}, Policy{AllowFS: true}},
		{[]string{"check", "--allow-net"}, Policy{AllowNet: true}},
		{[]string{"check", "--allow-env"}, Policy{AllowEnv: true}},
		{[]string{"--allow-exec", "--allow-net"}, Policy{AllowExec: true, AllowNet: true}},
		{[]string{"check", "--allow-exec=false"}, DenyAll()},
	}
	for _, r := range rows {
		got = Policy{AllowExec: true, AllowFS: true, AllowNet: true, AllowEnv: true} // a stale value must not survive
		toyRun(t, tree, r.argv...)
		if got != r.want {
			t.Errorf("%v: policy = %v, want %v", r.argv, got, r.want)
		}
	}
}

func TestApplianceTreeOverrideExecIsDeniedUnlessAllowed(t *testing.T) {
	dir := t.TempDir()
	// a def that needs the exec capability when it runs ...
	deferred := filepath.Join(dir, "deferred.gatwd")
	os.WriteFile(deferred, []byte("def shout = echo hello-from-override\n"), 0o644)
	// ... and a top-level statement that needs it while the override loads
	atLoad := filepath.Join(dir, "atload.gatwd")
	os.WriteFile(atLoad, []byte("spawn StrVal: x |> echo ran-at-load\n"), 0o644)

	tree := toyTree(toyRules)
	src := "spawn StrVal: s |> shout"

	code, out, errb := toyRun(t, tree, "check", "--source", src, "--rules", deferred)
	if code != 1 || !strings.Contains(out, "[ERROR] step 'echo' requires the exec capability (policy denies it)") || strings.Contains(out, "hello-from-override") {
		t.Fatalf("denied run: code %d, stdout %q, stderr %q", code, out, errb)
	}
	code, out, errb = toyRun(t, tree, "check", "--source", src, "--rules", deferred, "--allow-exec")
	if code != 0 || !strings.Contains(out, "hello-from-override") || strings.Contains(out, "[ERROR]") {
		t.Fatalf("with --allow-exec: code %d, stdout %q, stderr %q", code, out, errb)
	}
	// another capability does not help
	code, out, _ = toyRun(t, tree, "check", "--source", src, "--rules", deferred, "--allow-fs", "--allow-net", "--allow-env")
	if code != 1 || !strings.Contains(out, "requires the exec capability") {
		t.Fatalf("--allow-fs/net/env must not grant exec: code %d, stdout %q", code, out)
	}

	// a top-level statement is caught at load time: the Appliance refuses to start
	code, out, errb = toyRun(t, tree, "check", "--source", "entry 1", "--rules", atLoad)
	if code != 1 || !strings.Contains(errb, "step 'echo' requires the exec capability (policy denies it)") || out != "" {
		t.Fatalf("top-level exec in an override: code %d, stdout %q, stderr %q", code, out, errb)
	}
	code, out, errb = toyRun(t, tree, "check", "--source", "entry 1", "--rules", atLoad, "--allow-exec")
	if code != 0 {
		t.Fatalf("top-level exec with --allow-exec: code %d, stdout %q, stderr %q", code, out, errb)
	}
}

func TestApplianceOverrideCannotWidenThePolicy(t *testing.T) {
	// Everything a rule file can say that might look like granting itself a capability.
	override := writeFile(t, "widen.gatwd", strings.Join([]string{
		"const AllowExec = true",
		"const Policy = TrustedPolicy",
		"def AllowExec = true",
		"def allow_exec = true",
		"def shout = echo widened",
		"",
	}, "\n"))
	app, err := NewAppliance(WithEmbeddedScript(toyRules), WithConfigMapOverride(override), WithRequiredDefs("entry"))
	if err != nil {
		t.Fatal(err)
	}
	if app.Interpreter().Policy() != DenyAll() {
		t.Fatalf("the override changed the policy to %v", app.Interpreter().Policy())
	}
	vals := app.Execute("spawn StrVal: s |> shout")
	if len(vals) != 1 {
		t.Fatalf("values = %v", vals)
	}
	var ce *CapabilityError
	if err, ok := vals[0].(error); !ok || !errors.As(err, &ce) || ce.Capability != CapExec {
		t.Fatalf("shout = %v (%T), want a capability error", vals[0], vals[0])
	}
	// a hot reload keeps the policy of the process
	if err := app.Reload(); err != nil {
		t.Fatal(err)
	}
	if app.Interpreter().Policy() != DenyAll() {
		t.Fatal("Reload widened the policy")
	}
}

func TestApplianceEditionStepsAreNotGatedByThePolicy(t *testing.T) {
	// the operator's crawl and `manifests` are Edition steps: DenyAll must not break them
	app, err := NewAppliance(
		WithEmbeddedScript(toyRules),
		WithStep(MapStep("wired", func(v any) any { return "wired:" + v.(string) })),
	)
	if err != nil {
		t.Fatal(err)
	}
	if got := app.Execute("spawn StrVal: a |> wired"); len(got) != 1 || got[0] != "wired:a" {
		t.Fatalf("an Edition step under the default policy: %v", got)
	}
}
