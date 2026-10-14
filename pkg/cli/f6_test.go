package cli

import (
	"fmt"
	"strings"
	"testing"
)

// F6: env_get is the env capability's first consumer; denial and absence are visible values.
func TestEnvGetPolicy(t *testing.T) {
	t.Setenv("GATWD_F6_PROBE", "hello")
	for _, tc := range []struct {
		policy Policy
		step   string
		want   string
	}{
		{DenyAll(), "env_get GATWD_F6_PROBE", "[ERROR] step 'env_get' requires the env capability (policy denies it)"},
		{Policy{AllowFS: true, AllowNet: true, AllowExec: true}, "env_get GATWD_F6_PROBE", "[ERROR] step 'env_get' requires the env capability (policy denies it)"},
		{Policy{AllowEnv: true}, "env_get GATWD_F6_PROBE", "Just(hello)"},
		{Policy{AllowEnv: true}, "env_get GATWD_F6_UNSET", "Nothing"},
		{Policy{AllowEnv: true}, "env_get", "[ERROR] env_get needs a variable name (env_get HOME)"},
	} {
		in, _ := quietInterpreter()
		in.SetPolicy(tc.policy)
		got := evalStrings(in, spawnPfx+tc.step)
		if len(got) != 1 || got[0] != tc.want {
			t.Errorf("%v %q: got %q, want [%q]", tc.policy, tc.step, got, tc.want)
		}
	}
}

func TestRatioEncodesExactlyAndLayoutParseIsLoud(t *testing.T) {
	in, _ := quietInterpreter()
	for script, want := range map[string]string{
		`json_encode 3r4`:                                    `"3r4"`,
		`json_encode [1r3, 1r2]`:                             `["1r3","1r2"]`,
		`time_parse_with "%Y-%m-%d %z" "2026-02-30 Z"`:       "Nothing",
		`time_parse_with "%Y-%m-%d" "2026-02-03"`:            "Nothing", // no zone
		`time_parse_with "%Y-%m-%d %H %z" "2026-02-03 25 Z"`: "Nothing",
	} {
		got := in.EvaluateToValues(script)
		if len(got) != 1 || !strings.Contains(strings.ReplaceAll(toS(got[0]), " ", ""), strings.ReplaceAll(want, " ", "")) {
			t.Errorf("%s = %v, want %s", script, got, want)
		}
	}
	for _, s := range []string{`time_parse_with "%Q" "x"`} {
		got := in.EvaluateToValues(s)
		if len(got) == 1 && !strings.Contains(toS(got[0]), "rror") && !strings.Contains(toS(got[0]), "unknown") && !strings.Contains(toS(got[0]), "Nothing") {
			t.Errorf("%s must be loud or Nothing, got %v", s, got)
		}
	}
}

func toS(v any) string { return fmt.Sprint(v) }
