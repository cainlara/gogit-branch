package core

import (
	"os"

	"cainlara/gogit-branch/model"

	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jedib0t/go-pretty/v6/table"
)

func GetBranches() error {
	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}

	branches, err := readBranches(currentDir)
	if err != nil {
		return err
	}

	printAsTable(branches)

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

func printAsTable(branches []model.Branch) {
	if len(branches) == 0 {
		color.Red("No branches to list")
	} else {
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"#", "Branch Name", "Current Hash"})

		for index, branch := range branches {
			t.AppendRows([]table.Row{
				{index + 1, branch.GetRefName(), branch.GetShortHash()},
			})
		}

		t.Render()
	}
}
