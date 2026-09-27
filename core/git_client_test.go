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

func TestParseRemoteRefs(t *testing.T) {
	hashA := "1111111111111111111111111111111111111111"
	hashB := "2222222222222222222222222222222222222222"
	hashC := "3333333333333333333333333333333333333333"

	t.Run("symref row dropped, ordinary rows kept with names split", func(t *testing.T) {
		output := hashA + " origin/HEAD origin/main\n" +
			hashB + " origin/feature-x \n"

		refs := parseRemoteRefs(output)

		if len(refs) != 1 {
			t.Fatalf("len(refs) = %d, want 1 (origin/HEAD dropped)", len(refs))
		}

		if refs[0].GetRemoteName() != "origin" || refs[0].GetLocalName() != "feature-x" {
			t.Errorf("split = (%q, %q), want (origin, feature-x)", refs[0].GetRemoteName(), refs[0].GetLocalName())
		}

		if refs[0].GetRefName() != "origin/feature-x" {
			t.Errorf("GetRefName() = %q, want %q", refs[0].GetRefName(), "origin/feature-x")
		}

		if refs[0].GetFullHash() != hashB || refs[0].GetShortHash() != hashB[:7] {
			t.Errorf("hashes = (%q, %q), want (%q, %q)", refs[0].GetFullHash(), refs[0].GetShortHash(), hashB, hashB[:7])
		}
	})

	t.Run("multiple remotes kept in input order", func(t *testing.T) {
		output := hashA + " origin/feature-x \n" +
			hashB + " upstream/feature-x \n" +
			hashC + " origin/feature-b \n"

		refs := parseRemoteRefs(output)

		want := []string{"origin/feature-x", "upstream/feature-x", "origin/feature-b"}
		if len(refs) != len(want) {
			t.Fatalf("len(refs) = %d, want %d", len(refs), len(want))
		}

		for i, w := range want {
			if refs[i].GetRefName() != w {
				t.Errorf("refs[%d] = %q, want %q", i, refs[i].GetRefName(), w)
			}
		}
	})

	t.Run("malformed lines are skipped", func(t *testing.T) {
		output := "not-a-ref\n" +
			hashA + " no-slash-at-all \n" +
			hashB[:3] + " origin/short \n" +
			hashC + " origin/good \n"

		refs := parseRemoteRefs(output)

		if len(refs) != 1 {
			t.Fatalf("len(refs) = %d, want 1 (only origin/good survives)", len(refs))
		}

		if refs[0].GetRefName() != "origin/good" {
			t.Errorf("refs[0] = %q, want %q", refs[0].GetRefName(), "origin/good")
		}
	})

	t.Run("empty output yields empty list", func(t *testing.T) {
		refs := parseRemoteRefs("")

		if len(refs) != 0 {
			t.Errorf("len(refs) = %d, want 0", len(refs))
		}
	})
}

func TestParseRemoteNames(t *testing.T) {
	t.Run("multiple names preserved in order", func(t *testing.T) {
		names := parseRemoteNames("broken\norigin\nother\n")

		want := []string{"broken", "origin", "other"}
		if len(names) != len(want) {
			t.Fatalf("len(names) = %d, want %d", len(names), len(want))
		}

		for i, w := range want {
			if names[i] != w {
				t.Errorf("names[%d] = %q, want %q", i, names[i], w)
			}
		}
	})

	t.Run("empty output yields empty slice", func(t *testing.T) {
		names := parseRemoteNames("")

		if len(names) != 0 {
			t.Errorf("len(names) = %d, want 0", len(names))
		}
	})

	t.Run("blank lines skipped, surrounding whitespace trimmed", func(t *testing.T) {
		names := parseRemoteNames("\n  origin  \n\n\tupstream\t\n")

		want := []string{"origin", "upstream"}
		if len(names) != len(want) {
			t.Fatalf("len(names) = %d, want %d", len(names), len(want))
		}

		for i, w := range want {
			if names[i] != w {
				t.Errorf("names[%d] = %q, want %q", i, names[i], w)
			}
		}
	})

	t.Run("CRLF line endings tolerated", func(t *testing.T) {
		names := parseRemoteNames("origin\r\nupstream\r\n")

		want := []string{"origin", "upstream"}
		if len(names) != len(want) {
			t.Fatalf("len(names) = %d, want %d", len(names), len(want))
		}

		for i, w := range want {
			if names[i] != w {
				t.Errorf("names[%d] = %q, want %q", i, names[i], w)
			}
		}
	})
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

func TestTargetDirFromURL(t *testing.T) {
	cases := []struct {
		name string
		url  string
		want string
	}{
		{"https with .git", "https://github.com/cainlara/gogit-branch.git", "gogit-branch"},
		{"https without .git", "https://github.com/cainlara/gogit-branch", "gogit-branch"},
		{"trailing slash", "https://github.com/cainlara/gogit-branch/", "gogit-branch"},
		{"trailing slash with .git", "https://github.com/cainlara/gogit-branch.git/", "gogit-branch"},
		{"ssh url", "ssh://git@github.com/cainlara/gogit-branch.git", "gogit-branch"},
		{"scp style with path", "git@github.com:cainlara/gogit-branch.git", "gogit-branch"},
		{"scp style no path", "github.com:gogit-branch.git", "gogit-branch"},
		{"local path", "/tmp/clone-fixtures/origin", "origin"},
		{"local path trailing slash", "/tmp/clone-fixtures/origin/", "origin"},
		{"whitespace padded", "  https://github.com/x/y.git  ", "y"},
		{"port in url", "https://host:8080/repos/y.git", "y"},
		{"empty", "", ""},
		{"only slashes", "///", ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := TargetDirFromURL(c.url); got != c.want {
				t.Errorf("TargetDirFromURL(%q) = %q, want %q", c.url, got, c.want)
			}
		})
	}
}

func TestParseUnifiedDiff(t *testing.T) {
	tests := []struct {
		name string
		text string
		want []model.FileChange
	}{
		{
			name: "modified file with replaced and added lines",
			text: "diff --git a/f1.txt b/f1.txt\n" +
				"index de98044..bb79aec 100644\n" +
				"--- a/f1.txt\n" +
				"+++ b/f1.txt\n" +
				"@@ -2 +2 @@ a\n" +
				"-b\n" +
				"+B\n" +
				"@@ -3,0 +4 @@ c\n" +
				"+D\n",
			want: []model.FileChange{
				fileWithLines("f1.txt", model.CHANGE_KIND_MODIFIED, "",
					model.NewLineChange(model.LINE_KIND_MODIFIED, "B"),
					model.NewLineChange(model.LINE_KIND_ADDED, "D"),
				),
			},
		},
		{
			name: "deleted file keeps its lines for the parser",
			text: "diff --git a/f2.txt b/f2.txt\n" +
				"deleted file mode 100644\n" +
				"index b77b4eb..0000000\n" +
				"--- a/f2.txt\n" +
				"+++ /dev/null\n" +
				"@@ -1,2 +0,0 @@\n" +
				"-x\n" +
				"-y\n",
			want: []model.FileChange{
				fileWithLines("f2.txt", model.CHANGE_KIND_DELETED, "",
					model.NewLineChange(model.LINE_KIND_REMOVED, "x"),
					model.NewLineChange(model.LINE_KIND_REMOVED, "y"),
				),
			},
		},
		{
			name: "added file",
			text: "diff --git a/f3.txt b/f3.txt\n" +
				"new file mode 100644\n" +
				"index 0000000..3e75765\n" +
				"--- /dev/null\n" +
				"+++ b/f3.txt\n" +
				"@@ -0,0 +1 @@\n" +
				"+new\n",
			want: []model.FileChange{
				fileWithLines("f3.txt", model.CHANGE_KIND_ADDED, "",
					model.NewLineChange(model.LINE_KIND_ADDED, "new"),
				),
			},
		},
		{
			name: "pure rename has no lines and records the old path",
			text: "diff --git a/pure.txt b/renamed.txt\n" +
				"similarity index 100%\n" +
				"rename from pure.txt\n" +
				"rename to renamed.txt\n",
			want: []model.FileChange{
				fileWithLines("renamed.txt", model.CHANGE_KIND_RENAMED, "pure.txt"),
			},
		},
		{
			name: "rename with edits keeps rename kind plus line entries",
			text: "diff --git a/old.txt b/new.txt\n" +
				"similarity index 50%\n" +
				"rename from old.txt\n" +
				"rename to new.txt\n" +
				"index 1111111..2222222 100644\n" +
				"--- a/old.txt\n" +
				"+++ b/new.txt\n" +
				"@@ -1 +1 @@\n" +
				"-keep\n" +
				"+KEEP\n",
			want: []model.FileChange{
				fileWithLines("new.txt", model.CHANGE_KIND_RENAMED, "old.txt",
					model.NewLineChange(model.LINE_KIND_MODIFIED, "KEEP"),
				),
			},
		},
		{
			name: "binary change marks isBinary with no line entries",
			text: "diff --git a/real.bin b/real.bin\n" +
				"index c52e049..4891f69 100644\n" +
				"Binary files a/real.bin and b/real.bin differ\n",
			want: []model.FileChange{
				fileWithLines("real.bin", model.CHANGE_KIND_MODIFIED, ""),
			},
		},
		{
			name: "no-newline marker never becomes an entry",
			text: "diff --git a/nl.txt b/nl.txt\n" +
				"index 1111111..2222222 100644\n" +
				"--- a/nl.txt\n" +
				"+++ b/nl.txt\n" +
				"@@ -1 +1 @@\n" +
				"-BIN123\n" +
				"\\ No newline at end of file\n" +
				"+BIN456\n" +
				"\\ No newline at end of file\n",
			want: []model.FileChange{
				fileWithLines("nl.txt", model.CHANGE_KIND_MODIFIED, "",
					model.NewLineChange(model.LINE_KIND_MODIFIED, "BIN456"),
				),
			},
		},
		{
			name: "multiple sections keep diff emission order",
			text: "diff --git a/a.txt b/a.txt\n" +
				"index 1111111..2222222 100644\n" +
				"--- a/a.txt\n" +
				"+++ b/a.txt\n" +
				"@@ -1 +1 @@\n" +
				"-one\n" +
				"+two\n" +
				"diff --git b/b.txt b/b.txt\n" +
				"new file mode 100644\n" +
				"index 0000000..3333333\n" +
				"--- /dev/null\n" +
				"+++ b/b.txt\n" +
				"@@ -0,0 +1 @@\n" +
				"+fresh\n",
			want: []model.FileChange{
				fileWithLines("a.txt", model.CHANGE_KIND_MODIFIED, "",
					model.NewLineChange(model.LINE_KIND_MODIFIED, "two"),
				),
				fileWithLines("b.txt", model.CHANGE_KIND_ADDED, "",
					model.NewLineChange(model.LINE_KIND_ADDED, "fresh"),
				),
			},
		},
		{
			name: "path containing spaces splits at the last b/",
			text: "diff --git a/my file.txt b/my file.txt\n" +
				"index 1111111..2222222 100644\n" +
				"--- a/my file.txt\n" +
				"+++ b/my file.txt\n" +
				"@@ -1 +1 @@\n" +
				"-old\n" +
				"+new\n",
			want: []model.FileChange{
				fileWithLines("my file.txt", model.CHANGE_KIND_MODIFIED, "",
					model.NewLineChange(model.LINE_KIND_MODIFIED, "new"),
				),
			},
		},
		{
			name: "quoted header paths are unquoted",
			text: "diff --git \"a/sp\\tace.txt\" \"b/sp\\tace.txt\"\n" +
				"index 1111111..2222222 100644\n" +
				"--- \"a/sp\\tace.txt\"\n" +
				"+++ \"b/sp\\tace.txt\"\n" +
				"@@ -1 +1 @@\n" +
				"-old\n" +
				"+new\n",
			want: []model.FileChange{
				fileWithLines("sp\tace.txt", model.CHANGE_KIND_MODIFIED, "",
					model.NewLineChange(model.LINE_KIND_MODIFIED, "new"),
				),
			},
		},
		{
			name: "empty text yields no sections",
			text: "",
			want: []model.FileChange{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseUnifiedDiff(tt.text)

			if len(got) != len(tt.want) {
				t.Fatalf("parseUnifiedDiff() returned %d sections, want %d", len(got), len(tt.want))
			}

			for i := range got {
				if got[i].GetPath() != tt.want[i].GetPath() {
					t.Errorf("section %d path = %q, want %q", i, got[i].GetPath(), tt.want[i].GetPath())
				}

				if got[i].GetKind() != tt.want[i].GetKind() {
					t.Errorf("section %d kind = %q, want %q", i, got[i].GetKind(), tt.want[i].GetKind())
				}

				if got[i].GetRenamedFrom() != tt.want[i].GetRenamedFrom() {
					t.Errorf("section %d renamedFrom = %q, want %q", i, got[i].GetRenamedFrom(), tt.want[i].GetRenamedFrom())
				}

				binaryWant := tt.want[i].IsBinary()
				if tt.name == "binary change marks isBinary with no line entries" {
					binaryWant = true
				}

				if got[i].IsBinary() != binaryWant {
					t.Errorf("section %d isBinary = %v, want %v", i, got[i].IsBinary(), binaryWant)
				}

				lines := got[i].GetLines()
				wantLines := tt.want[i].GetLines()

				if len(lines) != len(wantLines) {
					t.Fatalf("section %d has %d lines, want %d", i, len(lines), len(wantLines))
				}

				for j := range lines {
					if lines[j].GetKind() != wantLines[j].GetKind() || lines[j].GetContent() != wantLines[j].GetContent() {
						t.Errorf("section %d line %d = (%q, %q), want (%q, %q)",
							i, j, lines[j].GetKind(), lines[j].GetContent(),
							wantLines[j].GetKind(), wantLines[j].GetContent())
					}
				}
			}
		})
	}
}

func fileWithLines(path, kind, renamedFrom string, lines ...model.LineChange) model.FileChange {
	f := model.NewFileChange(path, kind)

	if renamedFrom != "" {
		f.SetRenamedFrom(renamedFrom)
	}

	for _, line := range lines {
		f.AddLine(line)
	}

	return *f
}

func TestPairChangeBlock(t *testing.T) {
	tests := []struct {
		name    string
		removed []string
		added   []string
		want    []model.LineChange
	}{
		{
			name:    "equal counts pair index-wise into modified entries",
			removed: []string{"a", "b"},
			added:   []string{"A", "B"},
			want: []model.LineChange{
				model.NewLineChange(model.LINE_KIND_MODIFIED, "A"),
				model.NewLineChange(model.LINE_KIND_MODIFIED, "B"),
			},
		},
		{
			name:    "more removals than additions keeps leftover removals after pairs",
			removed: []string{"x", "y", "z"},
			added:   []string{"X"},
			want: []model.LineChange{
				model.NewLineChange(model.LINE_KIND_MODIFIED, "X"),
				model.NewLineChange(model.LINE_KIND_REMOVED, "y"),
				model.NewLineChange(model.LINE_KIND_REMOVED, "z"),
			},
		},
		{
			name:    "more additions than removals keeps leftover additions after pairs",
			removed: []string{"x"},
			added:   []string{"X", "Y", "Z"},
			want: []model.LineChange{
				model.NewLineChange(model.LINE_KIND_MODIFIED, "X"),
				model.NewLineChange(model.LINE_KIND_ADDED, "Y"),
				model.NewLineChange(model.LINE_KIND_ADDED, "Z"),
			},
		},
		{
			name:    "removal-only block stays pure removals",
			removed: []string{"gone"},
			added:   nil,
			want: []model.LineChange{
				model.NewLineChange(model.LINE_KIND_REMOVED, "gone"),
			},
		},
		{
			name:    "addition-only block stays pure additions",
			removed: nil,
			added:   []string{"fresh"},
			want: []model.LineChange{
				model.NewLineChange(model.LINE_KIND_ADDED, "fresh"),
			},
		},
		{
			name:    "both sides empty yields nothing",
			removed: nil,
			added:   nil,
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pairChangeBlock(tt.removed, tt.added)

			if len(got) != len(tt.want) {
				t.Fatalf("pairChangeBlock() returned %d lines, want %d", len(got), len(tt.want))
			}

			for i := range got {
				if got[i].GetKind() != tt.want[i].GetKind() || got[i].GetContent() != tt.want[i].GetContent() {
					t.Errorf("line %d = (%q, %q), want (%q, %q)",
						i, got[i].GetKind(), got[i].GetContent(),
						tt.want[i].GetKind(), tt.want[i].GetContent())
				}
			}
		})
	}
}
