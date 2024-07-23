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

	selectedBranch, err := listBranchesAndSelectTarget(branches, EMOJI_SKULL)
	if err != nil {
		return err
	}

	if confirmDeleteSelectedBranch(selectedBranch) {
		return core.PerformDeleteBranch(selectedBranch)
	} else {
		color.Blue("Deletion aborted")
	}

	return nil
}

func confirmDeleteSelectedBranch(branch model.Branch) bool {
	prompt := promptui.Prompt{
		Label:     fmt.Sprintf("Are you sure you want to delete %s (%s) [Type yes or y to continue or anything else to cancel]?", branch.GetShortName(), branch.GetShortHash()),
		IsConfirm: true,
	}

	result, _ := prompt.Run()

	return result == "yes" || result == "y"
}
