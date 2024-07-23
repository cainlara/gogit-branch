package model

type Branch struct {
	refName       string
	shortName     string
	shortHash     string
	fullHash      string
	currentBranch bool
}

func NewBranch(refName, shortName, shortHash, fullHash string, currentBranch bool) *Branch {
	b := new(Branch)

	b.refName = refName
	b.shortName = shortName
	b.shortHash = shortHash
	b.fullHash = fullHash
	b.currentBranch = currentBranch

	return b
}

func (b Branch) GetRefName() string {
	return b.refName
}

func (b *Branch) SetRefName(refName string) {
	b.refName = refName
}

func (b Branch) GetShortName() string {
	return b.shortName
}

func (b *Branch) SetShortName(shortName string) {
	b.shortName = shortName
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
