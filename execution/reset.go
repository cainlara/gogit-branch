package execution

import (
	"fmt"

	"github.com/cainlara/gogit-branch/core"

	"github.com/fatih/color"
)

// validateResetFlag reports whether raw is exactly "--hard", rejecting any
// other value with an error naming the offending input.
func validateResetFlag(raw string) (bool, error) {
	if raw == "--hard" {
		return true, nil
	}

	return false, fmt.Errorf("flag %q is not accepted; --hard is the only accepted flag", raw)
}

// confirmReset warns the user that the action cannot be undone and that
// uncommitted changes (and, if removeUntracked, untracked files too) will
// be lost, then requires an explicit yes/y confirmation — mirroring
// execution/delete.go's confirmDeleteSelectedBranch pattern (FR-002, FR-003).
func confirmReset(removeUntracked bool) bool {
	message := "This cannot be undone: uncommitted changes will be lost"
	if removeUntracked {
		message = "This cannot be undone: uncommitted changes AND untracked files will be lost"
	}

	label := fmt.Sprintf("%s. Are you sure you want to continue [Type yes or y to continue or n to cancel]?", message)

	return confirmYesNo(label)
}

// ResetToLatestCommit reverts the working tree to the current branch's
// latest commit after an explicit confirmation. When flag is nil, only
// tracked changes are reverted; when flag is the literal string "--hard",
// untracked files are removed too. Any other flag value is rejected.
func ResetToLatestCommit(gitClient *core.GitClient, flag *string) error {
	removeUntracked := false

	if flag != nil {
		validated, err := validateResetFlag(*flag)
		if err != nil {
			return err
		}

		removeUntracked = validated
	}

	status, err := gitClient.Status()
	if err != nil {
		return err
	}

	nothingToReset := !status.HasTrackedChanges()
	if removeUntracked {
		nothingToReset = nothingToReset && !status.HasUntrackedFiles()
	}

	if nothingToReset {
		color.Blue("Nothing to reset")

		return nil
	}

	if !confirmReset(removeUntracked) {
		color.Blue("\nReset aborted")

		return nil
	}

	if err := gitClient.Reset(removeUntracked); err != nil {
		return err
	}

	if removeUntracked {
		color.Green("%s Reverted tracked changes and removed untracked files", EMOJI_HERB)
	} else {
		color.Green("%s Reverted tracked changes to the latest commit", EMOJI_HERB)
	}

	return nil
}
