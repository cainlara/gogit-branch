package execution

import (
	"cainlara/gogit-branch/core"
	"cainlara/gogit-branch/model"
	"fmt"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
)

func ListAndDeleteBranch() error {
	fmt.Println()
	color.Cyan("Deleting branch")

	branches, err := core.GetBranches()
	if err != nil {
		return err
	}

	selectedBranch, err := listBranchesAndSelectTarget(branches)
	if err != nil {
		return err
	}

	if confirmDeleteSelectedBranch(selectedBranch) {
		return core.PerformDeleteBranch(selectedBranch)
	}

	color.Blue("Deletion aborted")

	return nil
}

func confirmDeleteSelectedBranch(branch model.Branch) bool {
	prompt := promptui.Prompt{
		Label:     fmt.Sprintf("Are you sure you want to delete %s (%s) ?", branch.GetShortName(), branch.GetFullHash()),
		IsConfirm: true,
	}

	result, _ := prompt.Run()

	// if err != nil {
	// 	fmt.Printf("Prompt failed %v\n", err)
	// 	return
	// }

	fmt.Printf("You choose %q\n", result)

	return false
}
