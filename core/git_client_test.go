package core

import (
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
