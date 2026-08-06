package execution

import (
	"fmt"

	"github.com/cainlara/gogit-branch/core"

	"github.com/fatih/color"
)

// PushCurrentBranch pushes the current branch's commits to the remote,
// establishing tracking first if none exists yet (see core.GitClient.Push).
func PushCurrentBranch(gitClient *core.GitClient) error {
	fmt.Println()
	color.Cyan("Pushing branch")

	if err := gitClient.Push(); err != nil {
		return err
	}

	color.Green(fmt.Sprintf("%s Pushed current branch to the remote", EMOJI_ROCKET))

	return nil
}
