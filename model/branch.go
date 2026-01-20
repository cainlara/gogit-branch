package model

type Branch struct {
	name          string
	shortHash     string
	fullHash      string
	currentBranch bool
}

func NewBranch(name, shortHash, fullHash string, currentBranch bool) *Branch {
	b := new(Branch)

	b.name = name
	b.shortHash = shortHash
	b.fullHash = fullHash
	b.currentBranch = currentBranch

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
