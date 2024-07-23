package execution

import (
	"cainlara/gogit-branch/core"
	"cainlara/gogit-branch/model"
	"errors"
	"fmt"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
)

func ListAndDeleteBranch() error {
	fmt.Println()
	color.Cyan("Deleting branch")

	branches, err := core.GetBranches(true)
	if err != nil {
		return err
	}

	if len(branches) <= 0 {
		return errors.New("no branches to select from")
	}

	selectedBranch, err := listBranchesAndSelectTarget(branches, EMOJI_SKULL)
	if err != nil {
		return err
	}

	if confirmDeleteSelectedBranch(selectedBranch) {
		return core.PerformDeleteBranch(selectedBranch)
	} else {
		color.Blue("\nDeletion aborted")
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
