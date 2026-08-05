package execution

import (
	"fmt"
	"strings"

	"github.com/cainlara/gogit-branch/core"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
)

const ALREADY_EXISTS_MARKER = "already exists"

func CreateAndSwitchBranch(gitClient *core.GitClient) error {
	fmt.Println()
	color.Cyan("Creating branch")

	prompt := promptui.Prompt{
		Label: "Branch name",
	}

	name, err := prompt.Run()
	if err != nil {
		return err
	}

	if err := gitClient.CreateAndSwitchBranch(name); err != nil {
		if strings.Contains(err.Error(), ALREADY_EXISTS_MARKER) {
			return fmt.Errorf("branch %s already exists; use 'switch' (or 'sw') to move to it instead", name)
		}

		return err
	}

	color.Green(fmt.Sprintf("%s Created and switched to %s", EMOJI_HERB, name))

	return nil
}
