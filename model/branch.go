package model

type Branch struct {
	name          string
	shortHash     string
	fullHash      string
	currentBranch bool
	isDummy       bool
	isSelected    bool
	isDone        bool
	isRemote      bool
	remoteRef     string
	displayName   string
}

func NewBranch(name, shortHash, fullHash string, currentBranch bool) *Branch {
	b := new(Branch)

	b.name = name
	b.shortHash = shortHash
	b.fullHash = fullHash
	b.currentBranch = currentBranch
	b.isDummy = false
	b.isSelected = false
	b.isDone = false
	b.isRemote = false
	b.remoteRef = ""
	b.displayName = name

	return b
}

// NewRemoteBranch constructs a remote-only entry: name is the LOCAL branch
// name (the checkout target), remoteRef the full remote-tracking ref it
// starts from (e.g. "origin/feature-x"), and displayName the text the list
// renders (bare local name, or remoteRef when the same local name is offered
// by two or more remotes — research D2/D3, data-model.md).
func NewRemoteBranch(localName, remoteRef, displayName, shortHash, fullHash string) *Branch {
	b := new(Branch)

	b.name = localName
	b.shortHash = shortHash
	b.fullHash = fullHash
	b.currentBranch = false
	b.isDummy = false
	b.isSelected = false
	b.isDone = false
	b.isRemote = true
	b.remoteRef = remoteRef
	b.displayName = displayName

	return b
}

func NewDummyBranch(name string) *Branch {
	b := new(Branch)

	b.name = name
	b.shortHash = ""
	b.fullHash = ""
	b.currentBranch = false
	b.isDummy = true
	b.isSelected = false
	b.isDone = false
	b.isRemote = false
	b.remoteRef = ""
	b.displayName = name

	return b
}

func NewDoneBranch(name string) *Branch {
	b := new(Branch)

	b.name = name
	b.shortHash = ""
	b.fullHash = ""
	b.currentBranch = false
	b.isDummy = false
	b.isSelected = false
	b.isDone = true
	b.isRemote = false
	b.remoteRef = ""
	b.displayName = name

	return b
}

func (b Branch) GetName() string {
	return b.name
}

func (b *Branch) SetName(name string) {
	b.name = name
}

func (b Branch) GetShortHash() string {
	return b.shortHash
}

func (b *Branch) SetShortHash(shortHash string) {
	b.shortHash = shortHash
}

func (b Branch) GetFullHash() string {
	return b.fullHash
}

func (b *Branch) SetFullHash(fullHash string) {
	b.fullHash = fullHash
}

func (b Branch) IsCurrentBranch() bool {
	return b.currentBranch
}

func (b *Branch) SetCurrentBranch(currentBranch bool) {
	b.currentBranch = currentBranch
}

func (b Branch) IsDummyBranch() bool {
	return b.isDummy
}

func (b Branch) IsSelected() bool {
	return b.isSelected
}

func (b *Branch) SetSelected(selected bool) {
	b.isSelected = selected
}

func (b Branch) IsDone() bool {
	return b.isDone
}

// IsRemoteBranch reports whether this entry represents a remote-only branch
// (FR-001). False for local rows and for the Cancel/done sentinels, so the
// marker logic can exclude them cleanly (research D6, data-model invariant 3).
func (b Branch) IsRemoteBranch() bool {
	return b.isRemote
}

// GetRemoteRef returns the full remote-tracking ref (e.g. "origin/feature-x")
// used as the checkout startpoint for remote entries; empty for local/dummy.
func (b Branch) GetRemoteRef() string {
	return b.remoteRef
}

// GetDisplayName returns the text the selection list renders: the local name
// for local/dummy rows, possibly a qualified "<remote>/<name>" for remote
// entries (FR-011). Never pass this to git — GetName is the checkout target.
func (b Branch) GetDisplayName() string {
	return b.displayName
}

func (b Branch) String() string {
	return b.name + " (" + b.shortHash + ")"
}
