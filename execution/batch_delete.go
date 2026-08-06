package execution

import (
	"errors"
	"fmt"
	"strings"

	"github.com/cainlara/gogit-branch/core"
	"github.com/cainlara/gogit-branch/model"

	"github.com/fatih/color"
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

// maxConfirmBranchNames bounds how many selected branch names are shown in
// the batch-delete confirmation label, regardless of how many were actually
// selected — an unbounded, data-driven label can grow arbitrarily long and
// wrap unpredictably across the terminal's width (FR-007).
const maxConfirmBranchNames = 5

// summarizeBranchNamesForConfirm joins names for display in the
// batch-delete confirmation label, capping the shown names at
// maxConfirmBranchNames and appending an "and N more" suffix when more were
// selected. The full selection is still passed to gitClient.DeleteBranches
// regardless of how many names appear here.
func summarizeBranchNamesForConfirm(names []string) string {
	if len(names) <= maxConfirmBranchNames {
		return strings.Join(names, ", ")
	}

	shown := strings.Join(names[:maxConfirmBranchNames], ", ")

	return fmt.Sprintf("%s, and %d more", shown, len(names)-maxConfirmBranchNames)
}

func confirmDeleteSelectedBranches(selectedBranches []model.Branch) bool {
	branchNames := make([]string, 0, len(selectedBranches))
	for _, branch := range selectedBranches {
		branchNames = append(branchNames, branch.String())
	}

	label := fmt.Sprintf("Confirm deletion of selected branches: %s", summarizeBranchNamesForConfirm(branchNames))

	return confirmYesNo(label)
}
