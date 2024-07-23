package execution

import (
	"cainlara/gogit-branch/core"
	"cainlara/gogit-branch/model"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
)

func ListCurrentBranches() error {
	fmt.Println()
	color.Cyan("Listing branches")

	branches, err := core.GetBranches()
	if err != nil {
		return err
	}

	printAsTable(branches)
	return nil
}

func printAsTable(branches []model.Branch) {
	if len(branches) == 0 {
		color.Red("No branches to list")
	} else {
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"", "Branch Name", "Current Hash"})

		for _, branch := range branches {
			isCurrent := ""

			if branch.IsCurrentBranch() {
				isCurrent = "*"
			}

			t.AppendRow(table.Row{isCurrent, branch.GetShortName(), branch.GetShortHash()})
		}

		t.Render()
	}
}
