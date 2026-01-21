package execution

import (
	"cainlara/gogit-branch/core"
	"cainlara/gogit-branch/model"
	"errors"
	"fmt"

	"github.com/fatih/color"
)

func BrowseAndSwitchBranches(gitClient *core.GitClient) error {
	fmt.Println()
	color.Cyan("Switching branches")

	branches, err := gitClient.Branches(false)
	if err != nil {
		return err
	}

	if len(branches) <= 0 {
		return errors.New("no branches to select from")
	}

	branches = append(branches, *model.NewBranch("Cancel Switch", "", "", false))

	selectedBranch, err := listBranchesAndSelectTarget(branches, EMOJI_HERB)
	if err != nil {
		return err
	}

	if selectedBranch.GetName() == "Cancel Switch" {
		return nil
	}

	return gitClient.Checkout(selectedBranch)
}
