package execution

import (
	"cainlara/gogit-branch/core"
	"cainlara/gogit-branch/model"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
)

func ListCurrentBranches(gitClient *core.GitClient) error {
	fmt.Println()
	color.Cyan("Listing branches")

	branches, err := gitClient.Branches(true)

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

		green := color.New(color.FgGreen).SprintFunc()
		yellow := color.New(color.FgYellow, color.Bold).SprintFunc()
		magenta := color.New(color.FgMagenta, color.Bold).SprintFunc()

		t.AppendHeader(table.Row{yellow("Current"), yellow("Branch Name"), yellow("Current Hash")})

		for _, branch := range branches {
			isCurrent := ""

			if branch.IsCurrentBranch() {
				isCurrent = "*"
			}

			t.AppendRow(table.Row{magenta(isCurrent), green(branch.GetName()), green(branch.GetShortHash())})
		}

		t.Render()
	}
}
