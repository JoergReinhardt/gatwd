package cli

import "testing"

func TestGlobValues(t *testing.T) {
	in := NewInterpreter()
	for _, tc := range []struct{ src, want string }{
		{`glob_match "/repo" "/repo"`, "true"},
		{`glob_match "/repo" "/repo/a/b"`, "true"},
		{`glob_match "/repo" "/repox"`, "false"},
		{`glob_match "/repo/**/overlays/staging" "/repo/sip/overlays/staging"`, "true"},
		{`glob_match "/repo/**/overlays/staging" "/repo/a/b/overlays/staging"`, "true"},
		{`glob_match "/repo/**/overlays/staging" "/repo/overlays/staging"`, "true"},
		{`glob_match "/repo/**/overlays/staging" "/repo/sip/overlays/production"`, "false"},
		{`glob_match "/repo/sip*/overlays/staging" "/repo/sipwise-broker/overlays/staging"`, "true"},
		{`glob_match "/repo/sip*/overlays/staging" "/repo/other/overlays/staging"`, "false"},
		{`glob_match "**" "/anything/at/all"`, "true"},
		{`glob_root "/repo/**/overlays/staging"`, "/repo"},
		{`glob_root "/repo"`, "/repo"},
		{`glob_root "sip*"`, "."},
		{`"/repo/sipwise-broker/overlays/staging" |> glob_match "/repo/sip*/overlays/staging"`, "true"},
	} {
		if got := displayText(evalOne(t, in, tc.src)); got != tc.want {
			t.Errorf("%s = %q, want %q", tc.src, got, tc.want)
		}
	}
}
