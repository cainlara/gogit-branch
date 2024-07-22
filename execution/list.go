package execution

import (
	"cainlara/gogit-branch/core"
	"cainlara/gogit-branch/model"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/table"
)

func ListCurrentBranches() error {
	fmt.Println()
	color.Green("Listing branches")

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
		t.AppendHeader(table.Row{"#", "Branch Name", "Current Hash"})

		for index, branch := range branches {
			t.AppendRows([]table.Row{
				{index + 1, branch.GetShortName(), branch.GetShortHash()},
			})
		}

		t.Render()
	}
}
