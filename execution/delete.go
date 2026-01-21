package execution

import (
	"cainlara/gogit-branch/core"
	"cainlara/gogit-branch/model"
	"errors"
	"fmt"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
)

func ListAndDeleteBranch(gitClient *core.GitClient) error {
	fmt.Println()
	color.Cyan("Deleting branch")

	branches, err := gitClient.Branches(false)
	if err != nil {
		return err
	}

	if len(branches) <= 0 {
		return errors.New("no branches to select from")
	}

	cancelOption := model.NewDummyBranch("Cancel Delete")
	branches = append(branches, *cancelOption)

	selectedBranch, err := listBranchesAndSelectTarget(branches, EMOJI_SKULL)
	if err != nil {
		return err
	}

	if selectedBranch.IsDummyBranch() {
		return nil
	}

	if confirmDeleteSelectedBranch(selectedBranch) {
		return gitClient.DeleteBranch(selectedBranch)
	} else {
		color.Blue("\nDeletion aborted")
	}

	return nil
}

func confirmDeleteSelectedBranch(branch model.Branch) bool {
	prompt := promptui.Prompt{
		Label:     fmt.Sprintf("Are you sure you want to delete %s (%s) [Type yes or y to continue or n to cancel]?", branch.GetName(), branch.GetShortHash()),
		IsConfirm: true,
	}

	result, _ := prompt.Run()

	return result == "yes" || result == "y"
}
