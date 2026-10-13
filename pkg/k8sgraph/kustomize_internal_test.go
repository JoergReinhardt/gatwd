package k8sgraph

import "testing"

// TestIsRemoteRef pins the remote-reference test against kustomize's own rules (its
// internal NewRepoSpecFromURL / IsRemoteFile): everything kustomize would clone or
// download is remote, every plain path is local.
func TestIsRemoteRef(t *testing.T) {
	remote := []string{
		"https://github.com/org/repo//base?ref=v1",
		"HTTP://example.com/patch.yaml",
		"ssh://git@github.com/org/repo.git",
		"file:///srv/git/repo.git",
		"git::https://example.com/org/repo.git",
		"git@github.com:org/repo.git//base",
		"user@host.example:org/repo.git",
		"github.com/org/repo//base",
		"GitHub.com:org/repo",
	}
	local := []string{
		"", "base", "./base", "../base", "../../base", "overlays/prod", "patch.yaml", "dir/sub/file.yaml",
		"/abs/path/base", "my.repo.d/base", "github.company.local/x", "key=values.env",
	}
	for _, s := range remote {
		if !isRemoteRef(s) {
			t.Errorf("isRemoteRef(%q) = false, want true", s)
		}
	}
	for _, s := range local {
		if isRemoteRef(s) {
			t.Errorf("isRemoteRef(%q) = true, want false", s)
		}
	}
}
