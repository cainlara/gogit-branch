package execution

import (
	"cainlara/gogit-branch/core"
	"cainlara/gogit-branch/model"
	"errors"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
)

func ListAndDeleteBranches(gitClient *core.GitClient) error {
	fmt.Println()
	color.Cyan("Deleting branches")

	branches, err := gitClient.Branches(false)
	if err != nil {
		return err
	}

	if len(branches) <= 0 {
		return errors.New("no branches to select from")
	}

	doneBranch := model.NewDoneBranch("Done")
	branches = append([]model.Branch{*doneBranch}, branches...)

	selectedBranches, err := listBranchesAndSelectMultipleTargets(0, branches, EMOJI_SKULL)
	if err != nil {
		return err
	}

	if len(selectedBranches) > 0 {
		if confirmDeleteSelectedBranches(selectedBranches) {
			return gitClient.DeleteBranches(selectedBranches)
		} else {
			color.Blue("\nDeletion aborted")
		}
	}

	return nil
}

func confirmDeleteSelectedBranches(selectedBranches []model.Branch) bool {
	branchNames := make([]string, 0, len(selectedBranches))
	for _, branch := range selectedBranches {
		branchNames = append(branchNames, branch.String())
	}

	confirmMessage := fmt.Sprintf("Confirm deletion of selected branches: %s", strings.Join(branchNames, ", "))

	prompt := promptui.Prompt{
		Label:     confirmMessage,
		IsConfirm: true,
	}

	result, _ := prompt.Run()

	return result == "yes" || result == "y"
}
