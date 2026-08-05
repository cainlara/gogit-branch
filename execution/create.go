package execution

import (
	"fmt"
	"strings"

	"github.com/cainlara/gogit-branch/core"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
)

const ALREADY_EXISTS_MARKER = "already exists"

// validateBranchName trims leading/trailing whitespace from name and rejects
// a resulting empty string or a name starting with a dash (which git would
// otherwise interpret as a flag). Only the command-line-argument path calls
// this; the interactive prompt below is intentionally left untouched.
func validateBranchName(name string) (string, error) {
	trimmed := strings.TrimSpace(name)

	if trimmed == "" {
		return "", fmt.Errorf("branch name cannot be empty")
	}

	if strings.HasPrefix(trimmed, "-") {
		return "", fmt.Errorf("branch name %q cannot start with a dash", trimmed)
	}

	return trimmed, nil
}

// CreateAndSwitchBranch creates and switches to a new branch. When name is
// non-nil (a branch name argument was supplied on the command line), it is
// validated and used directly with no interactive prompt shown. When name is
// nil (no argument was supplied), the existing interactive prompt runs
// exactly as before.
func CreateAndSwitchBranch(gitClient *core.GitClient, name *string) error {
	fmt.Println()
	color.Cyan("Creating branch")

	var branchName string

	if name != nil {
		validated, err := validateBranchName(*name)
		if err != nil {
			return err
		}

		branchName = validated
	} else {
		prompt := promptui.Prompt{
			Label: "Branch name",
		}

		promptedName, err := prompt.Run()
		if err != nil {
			return err
		}

		branchName = promptedName
	}

	if err := gitClient.CreateAndSwitchBranch(branchName); err != nil {
		if strings.Contains(err.Error(), ALREADY_EXISTS_MARKER) {
			return fmt.Errorf("branch %s already exists; use 'switch' (or 'sw') to move to it instead", branchName)
		}

		return err
	}

	color.Green(fmt.Sprintf("%s Created and switched to %s", EMOJI_HERB, branchName))

	return nil
}
