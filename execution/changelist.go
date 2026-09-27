package execution

import (
	"strings"

	"github.com/cainlara/gogit-branch/model"
	"github.com/fatih/color"
)

const (
	// Note texts are byte-exact per contract N2 (research D6): notes fill the
	// empty-section slot for files that carry no line entries.
	deletedFileNote = "(deleted)"
	binaryFileNote  = "(binary file — line-level detail unavailable)"
)

func renamedFileNote(oldPath string) string {
	return "(renamed from " + oldPath + ")"
}

// markerFor maps a line change kind to its single display marker
// (spec FR-004/FR-005: one marker per entry, `m` only for modified).
func markerFor(kind string) string {
	switch kind {
	case model.LINE_KIND_ADDED:
		return "+"
	case model.LINE_KIND_REMOVED:
		return "-"
	case model.LINE_KIND_MODIFIED:
		return "m"
	default:
		return "?"
	}
}

// coloredMarker wraps markerFor in the spec's marker colors (FR-005: added
// green, removed red, modified yellow). fatih/color suppresses its escape
// sequences automatically when stdout is not a terminal, so redirected output
// stays byte-comparable; unit tests pin color.NoColor explicitly.
func coloredMarker(kind string) string {
	switch kind {
	case model.LINE_KIND_ADDED:
		return color.GreenString("%s", markerFor(kind))
	case model.LINE_KIND_REMOVED:
		return color.RedString("%s", markerFor(kind))
	case model.LINE_KIND_MODIFIED:
		return color.YellowString("%s", markerFor(kind))
	default:
		return markerFor(kind)
	}
}

// composeFileSection renders one file's section: path heading, then either the
// status note or the marked line entries — never both (contract N1). A deleted
// file always shows its note (FR-013, suppressing the parser's removal lines),
// a binary change always shows the no-detail note, and a pure rename shows the
// renamed-from note while a rename with edits shows only its entries
// (research D6).
func composeFileSection(f model.FileChange) string {
	lines := make([]string, 0, len(f.GetLines())+2)
	lines = append(lines, f.GetPath())

	switch {
	case f.GetKind() == model.CHANGE_KIND_DELETED:
		lines = append(lines, color.RedString("%s", deletedFileNote))
	case f.IsBinary():
		lines = append(lines, color.YellowString("%s", binaryFileNote))
	case f.GetKind() == model.CHANGE_KIND_RENAMED && len(f.GetLines()) == 0:
		lines = append(lines, color.YellowString("%s", renamedFileNote(f.GetRenamedFrom())))
	default:
		for _, line := range f.GetLines() {
			lines = append(lines, coloredMarker(line.GetKind())+" "+line.GetContent())
		}
	}

	return strings.Join(lines, "\n")
}

// composeChangeList renders the whole list: one section per file, exactly one
// blank line between sections, and ZERO bytes for an empty input (SC-004,
// contract P4/F3). Pure — unit-tested in changelist_test.go.
func composeChangeList(files []model.FileChange) string {
	if len(files) == 0 {
		return ""
	}

	sections := make([]string, 0, len(files))

	for _, f := range files {
		sections = append(sections, composeFileSection(f))
	}

	return strings.Join(sections, "\n\n") + "\n"
}
