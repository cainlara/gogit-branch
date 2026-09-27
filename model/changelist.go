package model

const (
	// File-level change kinds produced by the pull change list's diff parse
	// (data-model.md "FileChange" field `kind`).
	CHANGE_KIND_ADDED    = "added"
	CHANGE_KIND_MODIFIED = "modified"
	CHANGE_KIND_DELETED  = "deleted"
	CHANGE_KIND_RENAMED  = "renamed"

	// Line-level change kinds (spec FR-004: exactly one marker per entry).
	LINE_KIND_ADDED    = "added"
	LINE_KIND_REMOVED  = "removed"
	LINE_KIND_MODIFIED = "modified"
)

// FileChange is one section of the pull change list: a file's path plus either
// its changed lines or a status note (deleted/binary/pure-rename). Plain data
// with no exported fields and no dependency on core or execution, per the
// model package convention (precedent: FileStatus).
type FileChange struct {
	path        string
	kind        string
	renamedFrom string
	isBinary    bool
	lines       []LineChange
}

func NewFileChange(path string, kind string) *FileChange {
	f := new(FileChange)

	f.path = path
	f.kind = kind
	f.lines = make([]LineChange, 0)

	return f
}

func (f FileChange) GetPath() string {
	return f.path
}

func (f *FileChange) SetPath(path string) {
	f.path = path
}

func (f FileChange) GetKind() string {
	return f.kind
}

func (f *FileChange) SetKind(kind string) {
	f.kind = kind
}

func (f FileChange) GetRenamedFrom() string {
	return f.renamedFrom
}

func (f FileChange) IsBinary() bool {
	return f.isBinary
}

func (f *FileChange) SetRenamedFrom(path string) {
	f.renamedFrom = path
}

func (f *FileChange) SetBinary(binary bool) {
	f.isBinary = binary
}

func (f FileChange) GetLines() []LineChange {
	return f.lines
}

func (f *FileChange) AddLine(line LineChange) {
	f.lines = append(f.lines, line)
}

// LineChange is one changed line of the change list: exactly one kind
// (added/removed/modified) plus the raw content without any marker.
type LineChange struct {
	kind    string
	content string
}

func NewLineChange(kind string, content string) LineChange {
	return LineChange{kind: kind, content: content}
}

func (l LineChange) GetKind() string {
	return l.kind
}

func (l LineChange) GetContent() string {
	return l.content
}
