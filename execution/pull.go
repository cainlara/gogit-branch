package execution

import (
	"fmt"

	"github.com/cainlara/gogit-branch/core"

	"github.com/fatih/color"
)

// PullCurrentBranch updates the current branch from its upstream in a single
// invocation: it announces and runs the fetch step FIRST, and only proceeds to
// the pull step when that fetch succeeded (FR-002, FR-003 — a failed fetch
// returns here, so GitClient.Pull is unreachable). On success it reports
// whether remote commits were integrated or the branch was already up to date
// (FR-004). Failures are returned rather than printed here (Constitution IV —
// main.go owns all error output).
func PullCurrentBranch(gitClient *core.GitClient) error {
	fmt.Println()
	color.Cyan("Pulling branch")
	color.Cyan("Fetching remote updates...")

	if err := gitClient.Fetch(); err != nil {
		return err
	}

	updated, err := gitClient.Pull()
	if err != nil {
		return err
	}

	if updated {
		color.Green(fmt.Sprintf("%s Pulled current branch from the remote", EMOJI_ROCKET))
	} else {
		color.Green(fmt.Sprintf("%s Branch already up to date", EMOJI_HERB))
	}

	return nil
}
