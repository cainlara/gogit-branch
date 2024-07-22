package core

import (
	"os"

	"cainlara/gogit-branch/model"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func GetBranches() ([]model.Branch, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	return readBranches(currentDir)
}

func readBranches(path string) ([]model.Branch, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, err
	}

	iter, err := repo.Branches()
	if err != nil {
		return nil, err
	}

	var branches []model.Branch

	err = iter.ForEach(func(c *plumbing.Reference) error {
		b := model.NewBranch(string(c.Name()), c.Name().Short(), c.Hash().String()[:7], c.Hash().String())
		branches = append(branches, *b)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return branches, nil
}
