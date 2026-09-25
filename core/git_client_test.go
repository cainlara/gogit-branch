package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cainlara/gogit-branch/model"
)

func TestParseStatusBranchLine(t *testing.T) {
	tests := []struct {
		name         string
		line         string
		wantBranch   string
		wantDetached bool
		wantUpstream string
		wantAhead    int
		wantBehind   int
	}{
		{
			name:       "no upstream",
			line:       "main",
			wantBranch: "main",
		},
		{
			name:         "upstream, no divergence",
			line:         "main...origin/main",
			wantBranch:   "main",
			wantUpstream: "origin/main",
		},
		{
			name:         "upstream, ahead only",
			line:         "main...origin/main [ahead 1]",
			wantBranch:   "main",
			wantUpstream: "origin/main",
			wantAhead:    1,
		},
		{
			name:         "upstream, behind only",
			line:         "main...origin/main [behind 2]",
			wantBranch:   "main",
			wantUpstream: "origin/main",
			wantBehind:   2,
		},
		{
			name:         "upstream, ahead and behind",
			line:         "main...origin/main [ahead 1, behind 2]",
			wantBranch:   "main",
			wantUpstream: "origin/main",
			wantAhead:    1,
			wantBehind:   2,
		},
		{
			name:         "upstream gone",
			line:         "main...origin/main [gone]",
			wantBranch:   "main",
			wantUpstream: "origin/main",
		},
		{
			name:         "detached HEAD",
			line:         STATUS_DETACHED_HEAD,
			wantBranch:   STATUS_DETACHED_HEAD,
			wantDetached: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := model.NewRepositoryStatus()

			parseStatusBranchLine(tt.line, status)

			if status.GetBranch() != tt.wantBranch {
				t.Errorf("GetBranch() = %q, want %q", status.GetBranch(), tt.wantBranch)
			}

			if status.IsDetached() != tt.wantDetached {
				t.Errorf("IsDetached() = %v, want %v", status.IsDetached(), tt.wantDetached)
			}

			if status.GetUpstream() != tt.wantUpstream {
				t.Errorf("GetUpstream() = %q, want %q", status.GetUpstream(), tt.wantUpstream)
			}

			if status.GetAhead() != tt.wantAhead {
				t.Errorf("GetAhead() = %d, want %d", status.GetAhead(), tt.wantAhead)
			}

			if status.GetBehind() != tt.wantBehind {
				t.Errorf("GetBehind() = %d, want %d", status.GetBehind(), tt.wantBehind)
			}
		})
	}
}

func TestParseProgressPercent(t *testing.T) {
	tests := []struct {
		name           string
		fragment       string
		wantPercent    int
		wantHasPercent bool
	}{
		{
			name:           "typical receiving progress fragment",
			fragment:       "remote: Counting objects:  42% (35/83)",
			wantPercent:    42,
			wantHasPercent: true,
		},
		{
			name:           "hundred percent with done suffix",
			fragment:       "remote: Compressing objects: 100% (82/82), done.",
			wantPercent:    100,
			wantHasPercent: true,
		},
		{
			name:           "clamp oversized token to 100",
			fragment:       "Resolving deltas: 1000% (999/1000)",
			wantPercent:    100,
			wantHasPercent: true,
		},
		{
			name:           "no token reports false",
			fragment:       "remote: Enumerating objects: 153, done.",
			wantPercent:    0,
			wantHasPercent: false,
		},
		{
			name:           "multiple tokens last wins",
			fragment:       "Counting: 7% then 50% (5/10)",
			wantPercent:    50,
			wantHasPercent: true,
		},
		{
			name:           "empty fragment",
			fragment:       "",
			wantPercent:    0,
			wantHasPercent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			percent, hasPercent := parseProgressPercent(tt.fragment)

			if percent != tt.wantPercent {
				t.Errorf("parseProgressPercent(%q) percent = %d, want %d", tt.fragment, percent, tt.wantPercent)
			}

			if hasPercent != tt.wantHasPercent {
				t.Errorf("parseProgressPercent(%q) hasPercent = %v, want %v", tt.fragment, hasPercent, tt.wantHasPercent)
			}
		})
	}
}

func TestIsProgressFragment(t *testing.T) {
	tests := []struct {
		name     string
		fragment string
		want     bool
	}{
		{
			name:     "percent-bearing progress line is dropped",
			fragment: "remote: Counting objects:  42% (35/83)",
			want:     true,
		},
		{
			name:     "sideband enumerating line without percent is dropped",
			fragment: "remote: Enumerating objects: 153, done.",
			want:     true,
		},
		{
			name:     "sideband total line without percent is dropped",
			fragment: "remote: Total 153 (delta 0), reused 0 (delta 0)",
			want:     true,
		},
		{
			name:     "ref advertisement line is kept",
			fragment: "From /tmp/pbtest/remote",
			want:     false,
		},
		{
			name:     "marker-bearing line is never dropped",
			fragment: "remote: error: declined 100%",
			want:     false,
		},
		{
			name:     "fatal line is never dropped",
			fragment: "fatal: could not read Username for 'https://example.com'",
			want:     false,
		},
		{
			name:     "plain informational line is kept",
			fragment: "There is no tracking information for the current branch.",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isProgressFragment(tt.fragment); got != tt.want {
				t.Errorf("isProgressFragment(%q) = %v, want %v", tt.fragment, got, tt.want)
			}
		})
	}
}

func TestHasNoUpstreamBranchError(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   bool
	}{
		{
			name:   "no upstream branch message",
			output: "fatal: The current branch quickstart-demo has no upstream branch.\nTo push the current branch and set the remote as upstream, use\n\n    git push --set-upstream origin quickstart-demo\n",
			want:   true,
		},
		{
			name:   "unrelated fatal message",
			output: "fatal: 'origin' does not appear to be a git repository\n",
			want:   false,
		},
		{
			name:   "rejected push message",
			output: "To /tmp/scratch-remote.git\n ! [rejected]        brand-new -> brand-new (fetch first)\nerror: failed to push some refs to '/tmp/scratch-remote.git'\n",
			want:   false,
		},
		{
			name:   "empty output",
			output: "",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasNoUpstreamBranchError(tt.output); got != tt.want {
				t.Errorf("hasNoUpstreamBranchError(%q) = %v, want %v", tt.output, got, tt.want)
			}
		})
	}
}

// writeFakeGit places a fake `git` executable at the FRONT of PATH for the
// duration of the test (prepended, so scripts may still use ordinary tools
// like sleep while `git` resolves to the fake).
func writeFakeGit(t *testing.T, script string) {
	t.Helper()

	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "git"), []byte(script), 0o755); err != nil {
		t.Fatalf("write fake git: %v", err)
	}

	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

// Regression for feature 011: git's paragraph-formatted error messages (no
// tracking information, detached HEAD) contain blank lines, and the streamed
// runner must preserve them byte-for-byte while still filtering progress
// fragments and forwarding percent updates (FR-005, SC-003, contract §4).
func TestRunGitCommandWithProgressPreservesMessageShape(t *testing.T) {
	writeFakeGit(t, `#!/bin/sh
printf 'remote: Counting objects: 50%% (1/2), done.\rThere is no tracking information.\n\nPlease specify which branch you want to merge with.\n' >&2
exit 1
`)

	g := &GitClient{}

	var percents []int
	var flags []bool

	out, err := g.runGitCommandWithProgress(func(p int, has bool) {
		percents = append(percents, p)
		flags = append(flags, has)
	}, "fetch", "--progress")

	if err == nil {
		t.Fatal("want error from exit-1 fake git")
	}

	want := "There is no tracking information.\n\nPlease specify which branch you want to merge with.\n"

	if string(out) != want {
		t.Errorf("accumulated output = %q, want %q", string(out), want)
	}

	if strings.Contains(string(out), "Counting") || strings.Contains(string(out), "50%") {
		t.Errorf("progress fragment leaked into message: %q", string(out))
	}

	found := false

	for i, p := range percents {
		if p == 50 && flags[i] {
			found = true
		}
	}

	if !found {
		t.Errorf("percent callback never reported 50 (got %v/%v)", percents, flags)
	}
}

func TestRunGitCommandWithProgressNoTrailingNewline(t *testing.T) {
	writeFakeGit(t, `#!/bin/sh
printf 'tail without newline' >&2
exit 1
`)

	g := &GitClient{}

	out, err := g.runGitCommandWithProgress(nil, "fetch", "--progress")

	if err == nil {
		t.Fatal("want error")
	}

	if string(out) != "tail without newline" {
		t.Errorf("output = %q, want exactly %q (no newline added)", string(out), "tail without newline")
	}
}

func TestRunGitCommandWithProgressSuccessPassthrough(t *testing.T) {
	writeFakeGit(t, `#!/bin/sh
printf 'Already up to date.\n'
exit 0
`)

	g := &GitClient{}

	out, err := g.runGitCommandWithProgress(nil, "pull", "--progress")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(out) != "Already up to date.\n" {
		t.Errorf("output = %q, want %q", string(out), "Already up to date.\n")
	}
}

// Arrival-order interleaving: git block-buffers stdout when piped (flushing it
// last), so a real pull failure writes its stderr error block BEFORE the
// stdout "Updating a..b" line. The streamed runner must preserve that order
// (research D8 amendment 2) — stdout-first concatenation broke byte-identity
// with baseline CombinedOutput in quickstart P13 / 010 S9.
func TestRunGitCommandWithProgressKeepsArrivalOrder(t *testing.T) {
	writeFakeGit(t, `#!/bin/sh
printf 'stderr block line\n' >&2
sleep 0.2
printf 'Updating a1b2c3d..e4f5a6b\n'
exit 1
`)

	g := &GitClient{}

	out, err := g.runGitCommandWithProgress(nil, "pull", "--progress")

	if err == nil {
		t.Fatal("want error")
	}

	want := "stderr block line\nUpdating a1b2c3d..e4f5a6b\n"

	if string(out) != want {
		t.Errorf("output = %q, want arrival-ordered %q", string(out), want)
	}
}
