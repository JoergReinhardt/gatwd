package cli

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// Tests for the two write capabilities (CapClusterApply, CapClusterRollout): denied by
// default everywhere, including for the trusted command-line script, and enforced at every
// step an Edition wraps in GatedStep.

var writeCapabilities = []Capability{CapClusterApply, CapClusterRollout}

func TestWriteCapabilitiesAreDeniedByDefault(t *testing.T) {
	for _, c := range writeCapabilities {
		if DenyAll().Allows(c) {
			t.Errorf("DenyAll allows %s", c)
		}
		if TrustedPolicy().Allows(c) {
			t.Errorf("TrustedPolicy allows %s: a write to a remote system needs an explicit opt-in even for the user's own script", c)
		}
		in := NewInterpreter()
		if in.Policy().Allows(c) {
			t.Errorf("NewInterpreter (the trusted command-line engine) allows %s", c)
		}
		if got := TrustedPolicy().String(); strings.Contains(got, string(c)) {
			t.Errorf("TrustedPolicy lists %s: %q", c, got)
		}
	}
	// ... and the machine-local capabilities stay granted to the trusted script
	for _, c := range []Capability{CapExec, CapFS, CapNet, CapEnv} {
		if !TrustedPolicy().Allows(c) {
			t.Errorf("TrustedPolicy lost %s", c)
		}
	}
}

func TestWriteCapabilitiesAreSeparateSwitches(t *testing.T) {
	apply := Policy{AllowClusterApply: true}
	rollout := Policy{AllowClusterRollout: true}
	if !apply.Allows(CapClusterApply) || apply.Allows(CapClusterRollout) || apply.Allows(CapExec) {
		t.Errorf("%v grants the wrong set", apply)
	}
	if !rollout.Allows(CapClusterRollout) || rollout.Allows(CapClusterApply) || rollout.Allows(CapFS) {
		t.Errorf("%v grants the wrong set", rollout)
	}
	if got := apply.String(); got != "cluster-apply" {
		t.Errorf("String = %q", got)
	}
	both := TrustedPolicy().Grant(CapClusterApply).Grant(CapClusterRollout)
	if got := both.String(); got != "exec,fs,net,env,cluster-apply,cluster-rollout" {
		t.Errorf("String = %q", got)
	}
	if DenyAll().Grant(Capability("bogus")) != DenyAll() {
		t.Error("Grant of an unknown capability must change nothing")
	}
	for _, c := range allCapabilities {
		if !DenyAll().Grant(c).Allows(c) {
			t.Errorf("Grant(%s) does not allow it", c)
		}
	}
}

func TestGatedStepDeniesAWriteStepAndNeverStartsIt(t *testing.T) {
	for _, c := range writeCapabilities {
		t.Run(string(c), func(t *testing.T) {
			var ran int
			write := ArgStep("write_probe", func(arg string, v any) (any, error) {
				ran++
				return fmt.Sprintf("wrote:%s:%v", arg, v), nil
			})
			in, _ := quietInterpreter()
			in.RegisterStep(GatedStep(c, write))

			// the trusted script (NewInterpreter) is denied: this is the whole point
			got := evalStrings(in, spawnPfx+`write_probe 3`)
			want := fmt.Sprintf("[ERROR] step 'write_probe' requires the %s capability (policy denies it)", c)
			if len(got) != 1 || got[0] != want || ran != 0 {
				t.Fatalf("trusted policy: got %q (ran %d), want exactly [%q] and no run", got, ran, want)
			}
			in.SetPolicy(DenyAll())
			if got := evalStrings(in, spawnPfx+`write_probe 3`); len(got) != 1 || got[0] != want || ran != 0 {
				t.Fatalf("DenyAll: got %q (ran %d)", got, ran)
			}
			// another write capability does not help
			other := CapClusterApply
			if c == CapClusterApply {
				other = CapClusterRollout
			}
			in.SetPolicy(DenyAll().Grant(other).Grant(CapExec).Grant(CapFS))
			if got := evalStrings(in, spawnPfx+`write_probe 3`); len(got) != 1 || got[0] != want || ran != 0 {
				t.Fatalf("wrong capability granted: got %q (ran %d)", got, ran)
			}
			// the granted capability lets it through, with its argument
			in.SetPolicy(DenyAll().Grant(c))
			got = evalStrings(in, spawnPfx+`write_probe 3`)
			if len(got) != 1 || got[0] != "wrote:3:seed" || ran != 1 {
				t.Fatalf("granted: got %q (ran %d)", got, ran)
			}
			// a handler that does not match the step is not claimed by the wrapper
			if _, ok := GatedStep(c, write)("some_other_step", in); ok {
				t.Fatal("GatedStep claimed a step its handler does not match")
			}
		})
	}
}

func TestGatedStepChecksThePolicyWhenTheStreamRuns(t *testing.T) {
	var ran bool
	in, _ := quietInterpreter()
	in.RegisterStep(GatedStep(CapClusterApply, MapStep("write_probe", func(v any) any { ran = true; return v })))
	in.SetPolicy(DenyAll().Grant(CapClusterApply))
	in.EvaluateLine(`def push = write_probe`)
	if got := evalStrings(in, spawnPfx+`push`); len(got) != 1 || got[0] != "seed" || !ran {
		t.Fatalf("granted: %q ran=%v", got, ran) // also warms the compiled cache
	}
	ran = false
	in.SetPolicy(TrustedPolicy())
	got := evalStrings(in, spawnPfx+`push`)
	if ran || len(got) != 1 || !strings.Contains(got[0], "requires the cluster-apply capability") {
		t.Fatalf("a cached def ran after the capability was withdrawn: %q ran=%v", got, ran)
	}
}

func TestDeniedWriteCapabilityIsAStructuredDiagnostic(t *testing.T) {
	in, _ := quietInterpreter()
	in.RegisterStep(GatedStep(CapClusterRollout, MapStep("write_probe", func(v any) any { return v })))
	vals := in.EvaluateToValues(spawnPfx + `write_probe`)
	if len(vals) != 1 {
		t.Fatalf("values %v", vals)
	}
	err, ok := vals[0].(error)
	if !ok {
		t.Fatalf("want an error value, got %T %v", vals[0], vals[0])
	}
	f := FailureOf(err)
	if f.Kind != FailCapabilityDenied || f.Step != "write_probe" || f.Capability != "cluster-rollout" {
		t.Fatalf("diagnostic = %+v", f)
	}
	var sb strings.Builder
	classify(vals).writeJSON(&sb)
	want := `"diagnostics":[{"kind":"capability_denied","message":"step 'write_probe' requires the cluster-rollout capability (policy denies it)","step":"write_probe","capability":"cluster-rollout"}]`
	if !strings.Contains(sb.String(), want) {
		t.Fatalf("report %q lacks %q", sb.String(), want)
	}
}

// ---- the opt-in flag of an Edition ----

func TestEditionCapabilityFlagIsTheOnlyWayToGrantAWriteCapability(t *testing.T) {
	ed := NewEdition("gatwd-captest").
		WithPrelude("type StrVal = String").
		WithStep(GatedStep(CapClusterApply, MapStep("apply_probe", func(any) any { return "applied" }))).
		WithStep(GatedStep(CapClusterRollout, MapStep("rollout_probe", func(any) any { return "rolled" }))).
		WithCapabilityFlag("allow-cluster-apply", CapClusterApply, "let scripts apply").
		WithCapabilityFlag("allow-cluster-rollout", CapClusterRollout, "let scripts roll out")
	run := func(argv ...string) (int, string) {
		var out, errb strings.Builder
		code := ed.Command().Execute(argv, strings.NewReader(""), &out, &errb)
		return code, out.String() + errb.String()
	}
	const apply = "spawn StrVal: x |> apply_probe"
	const rollout = "spawn StrVal: x |> rollout_probe"
	denyA := "step 'apply_probe' requires the cluster-apply capability (policy denies it)"
	denyR := "step 'rollout_probe' requires the cluster-rollout capability (policy denies it)"

	// no flag: the user's own -e script is denied
	if _, out := run("eval", "-e", apply, "-e", rollout); !strings.Contains(out, denyA) || !strings.Contains(out, denyR) || strings.Contains(out, "applied") || strings.Contains(out, "rolled") {
		t.Fatalf("without a flag both writes must be denied: %q", out)
	}
	// each flag grants exactly its own capability
	_, out := run("--allow-cluster-apply", "eval", "-e", apply, "-e", rollout)
	if !strings.Contains(out, "applied") || strings.Contains(out, denyA) || !strings.Contains(out, denyR) {
		t.Fatalf("--allow-cluster-apply: %q", out)
	}
	_, out = run("eval", "--allow-cluster-rollout", "-e", apply, "-e", rollout)
	if !strings.Contains(out, "rolled") || strings.Contains(out, denyR) || !strings.Contains(out, denyA) {
		t.Fatalf("--allow-cluster-rollout: %q", out)
	}
	// state does not leak into the next tree
	if _, out := run("eval", "-e", apply); !strings.Contains(out, denyA) {
		t.Fatalf("a flag leaked into the next command line: %q", out)
	}
	if code, out := run("--help"); code != 0 || !strings.Contains(out, "--allow-cluster-apply") || !strings.Contains(out, "denied by default") {
		t.Fatalf("help must list the flag: %q", out)
	}
}

// ---- the Appliance flags ----

func TestApplianceFlagsMapToTheWriteCapabilities(t *testing.T) {
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
		{[]string{"check", "--allow-cluster-apply"}, Policy{AllowClusterApply: true}},
		{[]string{"check", "--allow-cluster-rollout"}, Policy{AllowClusterRollout: true}},
		{[]string{"--allow-cluster-apply", "--allow-cluster-rollout", "--allow-exec"}, Policy{AllowClusterApply: true, AllowClusterRollout: true, AllowExec: true}},
		{[]string{"check", "--allow-fs", "--allow-net", "--allow-env", "--allow-exec"}, Policy{AllowFS: true, AllowNet: true, AllowEnv: true, AllowExec: true}},
	}
	for _, r := range rows {
		got = TrustedPolicy().Grant(CapClusterApply).Grant(CapClusterRollout) // a stale value must not survive
		toyRun(t, tree, r.argv...)
		if got != r.want {
			t.Errorf("%v: policy = %v, want %v", r.argv, got, r.want)
		}
	}
}
