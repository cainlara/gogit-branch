package execution

import (
	"cainlara/gogit-branch/core"
	"errors"
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

	if len(branches) <= 0 {
		return errors.New("no branches to select from")
	}

	selectedBranch, err := listBranchesAndSelectTarget(branches, EMOJI_HERB)
	if err != nil {
		return err
	}

	return core.PerformSwitch(selectedBranch)
}
