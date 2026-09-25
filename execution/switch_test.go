package execution

import (
	"testing"

	"github.com/cainlara/gogit-branch/core"
	"github.com/cainlara/gogit-branch/model"
)

const testHash = "abcdef1234567890abcdef1234567890abcdef12"

func mustRemoteRef(t *testing.T, refName string) core.RemoteRef {
	t.Helper()

	ref := core.NewRemoteRef(refName, testHash)
	if ref == nil {
		t.Fatalf("NewRemoteRef(%q) = nil", refName)
	}

	return *ref
}

func displayNames(branches []model.Branch) []string {
	names := make([]string, len(branches))
	for i, b := range branches {
		names[i] = b.GetDisplayName()
	}

	return names
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func TestBuildSwitchBranchList(t *testing.T) {
	tests := []struct {
		name          string
		localBranches []model.Branch
		remoteRefs    []core.RemoteRef
		wantDisplay   []string
		wantNames     []string
		wantRemote    []bool
	}{
		{
			name:          "pulled branch appears once as local (rule A)",
			localBranches: []model.Branch{*model.NewBranch("feature-pulled", "", "", false)},
			remoteRefs:    []core.RemoteRef{mustRemoteRef(t, "origin/feature-pulled"), mustRemoteRef(t, "origin/feature-b")},
			wantDisplay:   []string{"feature-pulled", "feature-b"},
			wantNames:     []string{"feature-pulled", "feature-b"},
			wantRemote:    []bool{false, true},
		},
		{
			name:          "unique remote kept with bare name",
			localBranches: []model.Branch{*model.NewBranch("local-only", "", "", false)},
			remoteRefs:    []core.RemoteRef{mustRemoteRef(t, "origin/feature-b")},
			wantDisplay:   []string{"local-only", "feature-b"},
			wantNames:     []string{"local-only", "feature-b"},
			wantRemote:    []bool{false, true},
		},
		{
			name:          "dual-remote collision qualified on both rows (rule C)",
			localBranches: []model.Branch{},
			remoteRefs:    []core.RemoteRef{mustRemoteRef(t, "origin/feature-x"), mustRemoteRef(t, "upstream/feature-x")},
			wantDisplay:   []string{"origin/feature-x", "upstream/feature-x"},
			wantNames:     []string{"feature-x", "feature-x"},
			wantRemote:    []bool{true, true},
		},
		{
			name:          "locals ordered strictly before remotes (rule D)",
			localBranches: []model.Branch{*model.NewBranch("b-local", "", "", false), *model.NewBranch("a-local", "", "", false)},
			remoteRefs:    []core.RemoteRef{mustRemoteRef(t, "origin/zeta"), mustRemoteRef(t, "origin/alpha")},
			wantDisplay:   []string{"b-local", "a-local", "zeta", "alpha"},
			wantNames:     []string{"b-local", "a-local", "zeta", "alpha"},
			wantRemote:    []bool{false, false, true, true},
		},
		{
			name:          "collision only among remote survivors — local name never qualified",
			localBranches: []model.Branch{*model.NewBranch("shared", "", "", false)},
			remoteRefs:    []core.RemoteRef{mustRemoteRef(t, "origin/shared"), mustRemoteRef(t, "upstream/feature-y")},
			wantDisplay:   []string{"shared", "feature-y"},
			wantNames:     []string{"shared", "feature-y"},
			wantRemote:    []bool{false, true},
		},
		{
			name:          "empty inputs yield empty list",
			localBranches: []model.Branch{},
			remoteRefs:    []core.RemoteRef{},
			wantDisplay:   []string{},
			wantNames:     []string{},
			wantRemote:    []bool{},
		},
		{
			name:          "all-local input unchanged",
			localBranches: []model.Branch{*model.NewBranch("b-main", "", "", false), *model.NewBranch("other", "", "", false)},
			remoteRefs:    []core.RemoteRef{},
			wantDisplay:   []string{"b-main", "other"},
			wantNames:     []string{"b-main", "other"},
			wantRemote:    []bool{false, false},
		},
		{
			name:          "current branch excluded from list but still shadows its remote ref",
			localBranches: []model.Branch{*model.NewBranch("main", "", "", true), *model.NewBranch("other", "", "", false)},
			remoteRefs:    []core.RemoteRef{mustRemoteRef(t, "origin/main")},
			wantDisplay:   []string{"other"},
			wantNames:     []string{"other"},
			wantRemote:    []bool{false},
		},
		{
			name:          "current branch alone yields no selectable rows",
			localBranches: []model.Branch{*model.NewBranch("main", "", "", true)},
			remoteRefs:    []core.RemoteRef{},
			wantDisplay:   []string{},
			wantNames:     []string{},
			wantRemote:    []bool{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildSwitchBranchList(tt.localBranches, tt.remoteRefs)

			if !equalStrings(displayNames(got), tt.wantDisplay) {
				t.Errorf("displayNames = %v, want %v", displayNames(got), tt.wantDisplay)
			}

			if len(got) != len(tt.wantNames) {
				t.Fatalf("len(got) = %d, want %d", len(got), len(tt.wantNames))
			}

			for i, b := range got {
				if b.GetName() != tt.wantNames[i] {
					t.Errorf("got[%d].GetName() = %q, want %q", i, b.GetName(), tt.wantNames[i])
				}

				if b.IsRemoteBranch() != tt.wantRemote[i] {
					t.Errorf("got[%d].IsRemoteBranch() = %v, want %v", i, b.IsRemoteBranch(), tt.wantRemote[i])
				}

				if b.IsRemoteBranch() {
					if b.GetRemoteRef() == "" {
						t.Errorf("got[%d] remote entry with empty remoteRef", i)
					}
					if b.GetFullHash() != testHash || b.GetShortHash() != testHash[:7] {
						t.Errorf("got[%d] hashes = (%q, %q), want (%q, %q)", i, b.GetFullHash(), b.GetShortHash(), testHash, testHash[:7])
					}
				} else if b.GetRemoteRef() != "" {
					t.Errorf("got[%d] local entry with non-empty remoteRef %q", i, b.GetRemoteRef())
				}
			}
		})
	}
}

func TestBuildSwitchBranchListDisplayNameUniqueness(t *testing.T) {
	// data-model invariant 4: no two rendered rows share a displayName —
	// local-wins filtering plus collision qualification must guarantee it
	// even when the same remote name exists on two remotes AND a local row
	// shadows one of them.
	locals := []model.Branch{*model.NewBranch("shared", "", "", false)}
	remotes := []core.RemoteRef{
		mustRemoteRef(t, "origin/shared"),
		mustRemoteRef(t, "origin/dup"),
		mustRemoteRef(t, "upstream/dup"),
	}

	got := buildSwitchBranchList(locals, remotes)

	seen := make(map[string]bool, len(got))
	for _, b := range got {
		dn := b.GetDisplayName()
		if dn == "" {
			t.Fatalf("empty displayName in composed list")
		}
		if seen[dn] {
			t.Errorf("duplicate displayName %q", dn)
		}
		seen[dn] = true
	}

	want := []string{"shared", "origin/dup", "upstream/dup"}
	if !equalStrings(displayNames(got), want) {
		t.Errorf("displayNames = %v, want %v", displayNames(got), want)
	}
}
