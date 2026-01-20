package execution

import (
	"cainlara/gogit-branch/core"
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

	selectedBranch, err := listBranchesAndSelectTarget(branches, EMOJI_HERB)
	if err != nil {
		return err
	}

	return gitClient.Checkout(selectedBranch)
}
