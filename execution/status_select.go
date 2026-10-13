package execution

import (
	"github.com/cainlara/gogit-branch/model"

	"github.com/manifoldco/promptui"
)

// Row kinds for the select-untracked-files dialog: the two sentinels
// (constitution: sentinel-based interactive flows) plus the real file rows.
const (
	selectRowKindFile = iota
	selectRowKindToggleAll
	selectRowKindDone
)

// selectRow is one dialog row rendered by promptui: a real untracked file
// or one of the two sentinels. Exported methods (not fields) feed the
// templates, per promptui's reflection rules.
type selectRow struct {
	kind int
	file model.SelectableFile
}

func (r selectRow) IsFile() bool {
	return r.kind == selectRowKindFile
}

func (r selectRow) IsSelected() bool {
	return r.kind == selectRowKindFile && r.file.IsSelected()
}

// Label is the rendered row text: the file path, or the sentinel name.
func (r selectRow) Label() string {
	switch r.kind {
	case selectRowKindToggleAll:
		return "Toggle all"
	case selectRowKindDone:
		return "Done"
	default:
		return r.file.GetPath()
	}
}

// buildSelectRows builds the dialog list: Toggle all, Done, then one file
// row per untracked path in status order, all starting deselected
// (data-model.md).
func buildSelectRows(paths []string) []selectRow {
	rows := make([]selectRow, 0, len(paths)+2)
	rows = append(rows,
		selectRow{kind: selectRowKindToggleAll},
		selectRow{kind: selectRowKindDone},
	)

	for _, path := range paths {
		rows = append(rows, selectRow{kind: selectRowKindFile, file: *model.NewSelectableFile(path)})
	}

	return rows
}

// toggleAllRows flips every file row's selection state in one action
// (FR-004); sentinels are untouched.
func toggleAllRows(rows []selectRow) {
	for i := range rows {
		if rows[i].kind == selectRowKindFile {
			rows[i].file.SetSelected(!rows[i].file.IsSelected())
		}
	}
}

// selectedPaths returns the selected file paths in list order. Empty result
// ≡ cancel (FR-007).
func selectedPaths(rows []selectRow) []string {
	var paths []string

	for _, row := range rows {
		if row.kind == selectRowKindFile && row.file.IsSelected() {
			paths = append(paths, row.file.GetPath())
		}
	}

	return paths
}

// selectUntrackedFiles runs the recursive multi-select dialog (research D1,
// contract §3): picking a file row toggles it, Toggle all flips every file
// row, Done returns the selected paths in list order. A promptui error
// (Ctrl+C / interrupt) propagates so the caller can map it to the cancel
// path (FR-007).
func selectUntrackedFiles(cursorPos int, rows []selectRow) ([]string, error) {
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "[{{ if .IsSelected }}x{{ else }} {{ end }}] {{ .Label | cyan }}",
		Inactive: "[{{ if .IsSelected }}x{{ else }} {{ end }}] {{ .Label }}",
	}

	prompt := promptui.Select{
		Label:        "Select untracked files to commit",
		Items:        rows,
		Templates:    templates,
		Size:         5,
		CursorPos:    cursorPos,
		HideSelected: true,
	}

	selectionIdx, _, err := prompt.Run()
	if err != nil {
		return nil, err
	}

	switch rows[selectionIdx].kind {
	case selectRowKindToggleAll:
		toggleAllRows(rows)

		return selectUntrackedFiles(selectionIdx, rows)
	case selectRowKindDone:
		return selectedPaths(rows), nil
	default:
		rows[selectionIdx].file.SetSelected(!rows[selectionIdx].file.IsSelected())

		return selectUntrackedFiles(selectionIdx, rows)
	}
}
