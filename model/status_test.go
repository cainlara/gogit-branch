package model

import "testing"

func TestFileStatusCategory(t *testing.T) {
	tests := []struct {
		name         string
		stagedCode   byte
		unstagedCode byte
		want         string
	}{
		{"unstaged modified", ' ', 'M', CATEGORY_MODIFIED},
		{"staged modified", 'M', ' ', CATEGORY_MODIFIED},
		{"staged deleted", 'D', ' ', CATEGORY_DELETED},
		{"unstaged deleted", ' ', 'D', CATEGORY_DELETED},
		{"staged new (added)", 'A', ' ', CATEGORY_NEW},
		{"renamed", 'R', ' ', CATEGORY_OTHER},
		{"copied", ' ', 'C', CATEGORY_OTHER},
		{"deleted takes priority over modified", 'D', 'M', CATEGORY_DELETED},
		{"new takes priority over modified", 'A', 'M', CATEGORY_NEW},
		{"unmerged both added (conflict)", 'A', 'A', CATEGORY_OTHER},
		{"unmerged both deleted (conflict)", 'D', 'D', CATEGORY_OTHER},
		{"unmerged marker U", 'U', 'U', CATEGORY_OTHER},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFileStatus("path.txt", tt.stagedCode, tt.unstagedCode)

			if got := f.Category(); got != tt.want {
				t.Fatalf("Category() with codes (%q, %q) = %q, want %q", tt.stagedCode, tt.unstagedCode, got, tt.want)
			}
		})
	}
}

func TestFileStatusStateAnnotation(t *testing.T) {
	tests := []struct {
		name         string
		stagedCode   byte
		unstagedCode byte
		want         string
	}{
		{"staged only", 'M', ' ', "staged"},
		{"unstaged only", ' ', 'M', "unstaged"},
		{"staged and unstaged", 'M', 'M', "staged + unstaged"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFileStatus("path.txt", tt.stagedCode, tt.unstagedCode)

			if got := f.StateAnnotation(); got != tt.want {
				t.Fatalf("StateAnnotation() with codes (%q, %q) = %q, want %q", tt.stagedCode, tt.unstagedCode, got, tt.want)
			}
		})
	}
}

func TestRepositoryStatusHelpers(t *testing.T) {
	status := NewRepositoryStatus()

	if status.HasTrackedChanges() {
		t.Fatal("expected no tracked changes on a fresh RepositoryStatus")
	}

	if status.HasUntrackedFiles() {
		t.Fatal("expected no untracked files on a fresh RepositoryStatus")
	}

	status.AddTracked(*NewFileStatus("a.txt", 'M', ' '))

	if !status.HasTrackedChanges() {
		t.Fatal("expected HasTrackedChanges() to be true after AddTracked")
	}

	status.AddUntracked("b.txt")

	if !status.HasUntrackedFiles() {
		t.Fatal("expected HasUntrackedFiles() to be true after AddUntracked")
	}
}
