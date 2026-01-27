package model

type Branch struct {
	name          string
	shortHash     string
	fullHash      string
	currentBranch bool
	isDummy       bool
	isSelected    bool
	isDone        bool
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

func (b Branch) String() string {
	return b.name + " (" + b.shortHash + ")"
}
