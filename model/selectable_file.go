package model

// SelectableFile is one untracked file row in the status command's
// select-untracked-files dialog: the path plus its toggle state (data-model.md).
// Plain data, same convention as Branch — unexported fields, accessors only.
type SelectableFile struct {
	path       string
	isSelected bool
}

// NewSelectableFile constructs a deselected row — the user opts files in
// one by one (or via Toggle all), so nothing is pre-selected.
func NewSelectableFile(path string) *SelectableFile {
	f := new(SelectableFile)

	f.path = path
	f.isSelected = false

	return f
}

func (f SelectableFile) GetPath() string {
	return f.path
}

func (f SelectableFile) IsSelected() bool {
	return f.isSelected
}

func (f *SelectableFile) SetSelected(selected bool) {
	f.isSelected = selected
}
