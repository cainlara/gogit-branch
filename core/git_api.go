package core

import (
	"fmt"
	"os"

	"cainlara/gogit-branch/model"

	"github.com/fatih/color"
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

func PerformSwitch(selectedBranch model.Branch) error {
	path, err := os.Getwd()
	if err != nil {
		return err
	}

	repo, err := git.PlainOpen(path)
	if err != nil {
		return err
	}

	workTree, err := repo.Worktree()
	if err != nil {
		return err
	}

	err = workTree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName(selectedBranch.GetRefName()),
		Force:  true,
	})
	if err != nil {
		return err
	}

	color.Green(fmt.Sprintf("Switched to Branch %s\n", selectedBranch.GetShortName()))

	return nil
}

func PerformDeleteBranch(selectedBranch model.Branch) error {

	return nil
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
