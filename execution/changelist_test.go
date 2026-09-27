package execution

import (
	"testing"

	"github.com/cainlara/gogit-branch/model"
	"github.com/fatih/color"
)

// Pin NoColor so marker/note assertions stay byte-exact even when tests run
// inside a real terminal (fatih/color otherwise injects escape sequences).
func init() {
	color.NoColor = true
}

func changeListFixture() []model.FileChange {
	modified := model.NewFileChange("mod.txt", model.CHANGE_KIND_MODIFIED)
	modified.AddLine(model.NewLineChange(model.LINE_KIND_REMOVED, "beta"))
	modified.AddLine(model.NewLineChange(model.LINE_KIND_ADDED, "BETA"))
	modified.AddLine(model.NewLineChange(model.LINE_KIND_MODIFIED, "gamma"))

	deleted := model.NewFileChange("del.txt", model.CHANGE_KIND_DELETED)
	deleted.AddLine(model.NewLineChange(model.LINE_KIND_REMOVED, "keep"))
	deleted.AddLine(model.NewLineChange(model.LINE_KIND_REMOVED, "me"))

	binary := model.NewFileChange("real.bin", model.CHANGE_KIND_MODIFIED)
	binary.SetBinary(true)

	renamed := model.NewFileChange("moved.txt", model.CHANGE_KIND_RENAMED)
	renamed.SetRenamedFrom("move.txt")

	added := model.NewFileChange("added.txt", model.CHANGE_KIND_ADDED)
	added.AddLine(model.NewLineChange(model.LINE_KIND_ADDED, "fresh"))

	return []model.FileChange{*modified, *deleted, *binary, *renamed, *added}
}

func TestComposeChangeListEmptyInputIsZeroBytes(t *testing.T) {
	if got := composeChangeList(nil); got != "" {
		t.Errorf("composeChangeList(nil) = %q, want empty string", got)
	}

	if got := composeChangeList([]model.FileChange{}); got != "" {
		t.Errorf("composeChangeList(empty) = %q, want empty string", got)
	}
}

func TestComposeChangeListSectionFormat(t *testing.T) {
	modified := model.NewFileChange("mod.txt", model.CHANGE_KIND_MODIFIED)
	modified.AddLine(model.NewLineChange(model.LINE_KIND_ADDED, "delta"))

	want := "mod.txt\n+ delta\n"

	if got := composeChangeList([]model.FileChange{*modified}); got != want {
		t.Errorf("composeChangeList() = %q, want %q", got, want)
	}
}

func TestComposeChangeListBlankLineBetweenSections(t *testing.T) {
	got := composeChangeList(changeListFixture())

	want := "mod.txt\n" +
		"- beta\n" +
		"+ BETA\n" +
		"m gamma\n" +
		"\n" +
		"del.txt\n" +
		"(deleted)\n" +
		"\n" +
		"real.bin\n" +
		"(binary file — line-level detail unavailable)\n" +
		"\n" +
		"moved.txt\n" +
		"(renamed from move.txt)\n" +
		"\n" +
		"added.txt\n" +
		"+ fresh\n"

	if got != want {
		t.Errorf("composeChangeList() =\n%q\nwant\n%q", got, want)
	}
}

func TestComposeFileSectionNotes(t *testing.T) {
	deleted := model.NewFileChange("del.txt", model.CHANGE_KIND_DELETED)
	deleted.AddLine(model.NewLineChange(model.LINE_KIND_REMOVED, "suppressed"))

	if got, want := composeFileSection(*deleted), "del.txt\n(deleted)"; got != want {
		t.Errorf("deleted section = %q, want %q", got, want)
	}

	binary := model.NewFileChange("real.bin", model.CHANGE_KIND_MODIFIED)
	binary.SetBinary(true)

	if got, want := composeFileSection(*binary), "real.bin\n(binary file — line-level detail unavailable)"; got != want {
		t.Errorf("binary section = %q, want %q", got, want)
	}

	renamed := model.NewFileChange("new/name.txt", model.CHANGE_KIND_RENAMED)
	renamed.SetRenamedFrom("old/name.txt")

	if got, want := composeFileSection(*renamed), "new/name.txt\n(renamed from old/name.txt)"; got != want {
		t.Errorf("pure rename section = %q, want %q", got, want)
	}
}

func TestComposeFileSectionRenameWithEditsShowsEntriesOnly(t *testing.T) {
	renamed := model.NewFileChange("new.txt", model.CHANGE_KIND_RENAMED)
	renamed.SetRenamedFrom("old.txt")
	renamed.AddLine(model.NewLineChange(model.LINE_KIND_MODIFIED, "KEEP"))

	want := "new.txt\nm KEEP"

	if got := composeFileSection(*renamed); got != want {
		t.Errorf("rename+edit section = %q, want %q", got, want)
	}
}
