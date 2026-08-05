package model

const (
	CATEGORY_MODIFIED = "modified"
	CATEGORY_DELETED  = "deleted"
	CATEGORY_NEW      = "new"
	CATEGORY_OTHER    = "other"

	NO_CODE = ' '
)

type FileStatus struct {
	path         string
	stagedCode   byte
	unstagedCode byte
}

func NewFileStatus(path string, stagedCode, unstagedCode byte) *FileStatus {
	f := new(FileStatus)

	f.path = path
	f.stagedCode = stagedCode
	f.unstagedCode = unstagedCode

	return f
}

func (f FileStatus) GetPath() string {
	return f.path
}

func (f FileStatus) GetStagedCode() byte {
	return f.stagedCode
}

func (f FileStatus) GetUnstagedCode() byte {
	return f.unstagedCode
}

// conflictCombos lists staged/unstaged code pairs that git itself reports for
// unmerged (conflicted) paths; these are bucketed as "other" rather than
// letting either half's letter (e.g. 'A' or 'D') drive a misleading category.
var conflictCombos = map[[2]byte]bool{
	{'D', 'D'}: true,
	{'A', 'A'}: true,
	{'U', 'U'}: true,
	{'A', 'U'}: true,
	{'U', 'A'}: true,
	{'D', 'U'}: true,
	{'U', 'D'}: true,
}

var categoryPriority = map[string]int{
	CATEGORY_DELETED:  4,
	CATEGORY_NEW:      3,
	CATEGORY_OTHER:    2,
	CATEGORY_MODIFIED: 1,
}

func codeCategory(code byte) (string, bool) {
	switch code {
	case 'D':
		return CATEGORY_DELETED, true
	case 'A':
		return CATEGORY_NEW, true
	case 'M':
		return CATEGORY_MODIFIED, true
	case 'R', 'C', 'U':
		return CATEGORY_OTHER, true
	default:
		return "", false
	}
}

// Category returns this file's primary status color category, per the
// deleted > new > other > modified priority rule (see data-model.md /
// research.md) used when the staged and unstaged codes differ.
func (f FileStatus) Category() string {
	if conflictCombos[[2]byte{f.stagedCode, f.unstagedCode}] {
		return CATEGORY_OTHER
	}

	best := ""

	for _, code := range []byte{f.stagedCode, f.unstagedCode} {
		category, ok := codeCategory(code)
		if !ok {
			continue
		}

		if best == "" || categoryPriority[category] > categoryPriority[best] {
			best = category
		}
	}

	if best == "" {
		return CATEGORY_OTHER
	}

	return best
}

// StateAnnotation reports whether this file has a staged change, an
// unstaged change, or both — preserving the distinction FR-004 requires,
// without splitting a dual-state file into two separate entries.
func (f FileStatus) StateAnnotation() string {
	staged := f.stagedCode != NO_CODE
	unstaged := f.unstagedCode != NO_CODE

	switch {
	case staged && unstaged:
		return "staged + unstaged"
	case staged:
		return "staged"
	case unstaged:
		return "unstaged"
	default:
		return ""
	}
}

type RepositoryStatus struct {
	branch    string
	detached  bool
	upstream  string
	ahead     int
	behind    int
	tracked   []FileStatus
	untracked []string
}

func NewRepositoryStatus() *RepositoryStatus {
	r := new(RepositoryStatus)

	r.tracked = make([]FileStatus, 0)
	r.untracked = make([]string, 0)

	return r
}

func (r RepositoryStatus) GetBranch() string {
	return r.branch
}

func (r *RepositoryStatus) SetBranch(branch string) {
	r.branch = branch
}

func (r RepositoryStatus) IsDetached() bool {
	return r.detached
}

func (r *RepositoryStatus) SetDetached(detached bool) {
	r.detached = detached
}

func (r RepositoryStatus) GetUpstream() string {
	return r.upstream
}

func (r *RepositoryStatus) SetUpstream(upstream string) {
	r.upstream = upstream
}

func (r RepositoryStatus) GetAhead() int {
	return r.ahead
}

func (r *RepositoryStatus) SetAhead(ahead int) {
	r.ahead = ahead
}

func (r RepositoryStatus) GetBehind() int {
	return r.behind
}

func (r *RepositoryStatus) SetBehind(behind int) {
	r.behind = behind
}

func (r RepositoryStatus) GetTracked() []FileStatus {
	return r.tracked
}

func (r *RepositoryStatus) AddTracked(file FileStatus) {
	r.tracked = append(r.tracked, file)
}

func (r RepositoryStatus) GetUntracked() []string {
	return r.untracked
}

func (r *RepositoryStatus) AddUntracked(path string) {
	r.untracked = append(r.untracked, path)
}

func (r RepositoryStatus) HasTrackedChanges() bool {
	return len(r.tracked) > 0
}

func (r RepositoryStatus) HasUntrackedFiles() bool {
	return len(r.untracked) > 0
}
