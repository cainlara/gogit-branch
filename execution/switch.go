package execution

import (
	"errors"
	"fmt"

	"github.com/cainlara/gogit-branch/core"
	"github.com/cainlara/gogit-branch/model"

	"github.com/fatih/color"
)

// buildSwitchBranchList composes the switch selection list from ALL local
// branches (the current branch included — its name must still shadow remote
// refs, otherwise a repo on `main` would offer a bogus `r main` row backed by
// origin/main; implementation-time discovery against quickstart §1) and the
// parsed remote-tracking refs. Pure presentation logic — no GitClient, no
// I/O — so every rule is unit-testable (research D5/D10):
//
//	A. local-wins filter (FR-004): a remote ref whose local name already
//	   exists locally is dropped, so a pulled branch appears once, as local;
//	C. collision qualification (FR-011): a local name surviving from 2+
//	   remotes renders qualified as "<remote>/<name>" for every occurrence,
//	   all others keep the bare local name;
//	D. ordering (FR-012): all non-current locals in their existing order
//	   first (the current branch is excluded from the list, as today), then
//	   the surviving remote-only entries in input order. The Cancel sentinel
//	   is appended by the caller afterwards, last (L5).
//
// The emptiness rule (E / FR-005) is the caller's: it checks the returned
// slice before appending the sentinel.
func buildSwitchBranchList(localBranches []model.Branch, remoteRefs []core.RemoteRef) []model.Branch {
	localNames := make(map[string]bool, len(localBranches))
	for _, branch := range localBranches {
		localNames[branch.GetName()] = true
	}

	survivors := make([]core.RemoteRef, 0, len(remoteRefs))
	for _, ref := range remoteRefs {
		if !localNames[ref.GetLocalName()] {
			survivors = append(survivors, ref)
		}
	}

	nameCounts := make(map[string]int, len(survivors))
	for _, ref := range survivors {
		nameCounts[ref.GetLocalName()]++
	}

	list := make([]model.Branch, 0, len(localBranches)+len(survivors))
	for _, branch := range localBranches {
		if !branch.IsCurrentBranch() {
			list = append(list, branch)
		}
	}

	for _, ref := range survivors {
		localName := ref.GetLocalName()
		displayName := localName
		if nameCounts[localName] > 1 {
			displayName = ref.GetRefName()
		}

		list = append(list, *model.NewRemoteBranch(localName, ref.GetRefName(), displayName, ref.GetShortHash(), ref.GetFullHash()))
	}

	return list
}

func BrowseAndSwitchBranches(gitClient *core.GitClient) error {
	fmt.Println()
	color.Cyan("Switching branches")

	// Branches(true): the current branch's name must participate in the
	// local-wins filter even though the builder excludes it from the list.
	branches, err := gitClient.Branches(true)
	if err != nil {
		return err
	}

	// Offline local read of the cached remote-tracking refs — the list itself
	// never touches the network (FR-007; research D1).
	remoteRefs, err := gitClient.RemoteRefs()
	if err != nil {
		return err
	}

	options := buildSwitchBranchList(branches, remoteRefs)

	if len(options) <= 0 {
		return errors.New("no branches to select from")
	}

	cancelOption := model.NewDummyBranch("Cancel Switch")
	options = append(options, *cancelOption)

	selectedBranch, err := listBranchesAndSelectTarget(options, EMOJI_HERB, true)
	if err != nil {
		return err
	}

	if selectedBranch.IsDummyBranch() {
		return nil
	}

	// Refresh only for remote picks, and abort before checkout on any fetch
	// failure (FR-013 — current branch and working tree untouched). Local
	// picks never fetch, so they work even when every remote is unreachable
	// (scenario 11, research D4).
	if selectedBranch.IsRemoteBranch() {
		if err := gitClient.Fetch(); err != nil {
			return err
		}
	}

	return gitClient.Checkout(selectedBranch)
}
