package execution

import (
	"cainlara/gogit-branch/core"
	"fmt"

	"github.com/fatih/color"
)

func BrowseAndSwitchBranches() error {
	fmt.Println()
	color.Cyan("Listing branches")

	branches, err := core.GetBranches(true)
	if err != nil {
		return err
	}

	selectedBranch, err := listBranchesAndSelectTarget(branches, EMOJI_HERB)
	if err != nil {
		return err
	}

	return core.PerformSwitch(selectedBranch)
}
