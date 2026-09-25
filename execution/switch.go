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

// refreshRemoteStateBeforeListing runs feature 013's pre-list refresh
// (contract §1 steps 2–5, research D3/D5) BEFORE any branch data is read:
//
//  1. probe `git remote` — a local read. A probe error RETURNS (contract G5:
//     unexpected local failure, not an FR-006 refresh failure); zero remotes
//     returns nil immediately with no banner and no network call, so
//     no-remote repositories stay byte-identical to a pre-network run
//     (FR-005/G4, research D2);
//  2. announce the refresh (FR-003/SC-002: feedback within one second);
//  3. run `git fetch --all --prune --progress` through the existing progress
//     renderer (reused from the pull flow, research D5), finishing the bar
//     explicitly so no warning or list byte can ever interleave with a
//     half-drawn bar (contract W2);
//  4. on fetch failure, emit the one-time warning event (FR-006/Q1→A,
//     contract §2 W1–W5: yellow tool line + git's outputError text exactly
//     once, bar already finished) and return nil — the caller continues with
//     the last-known ref list. Only a `Remotes()` probe failure still
//     returns an error (contract G5).
func refreshRemoteStateBeforeListing(gitClient *core.GitClient) error {
	remotes, err := gitClient.Remotes()
	if err != nil {
		return err
	}

	if len(remotes) == 0 {
		return nil
	}

	color.Cyan("Fetching remote updates...")

	bar := newStatusStreamRenderer()
	defer bar.finish()

	stopInterruptWatch := clearBarOnInterrupt(bar)
	defer stopInterruptWatch()

	bar.start(progressStageFetching)

	fetchErr := gitClient.FetchAllWithProgress(bar.onProgress)
	bar.finish() // W2: bar cleared before any warning byte can print

	if fetchErr != nil {
		// FR-006/Q1→A warn-and-continue: one yellow tool line (W3 — visibly
		// distinct from main.go's red Operation Failed:) followed by git's
		// own outputError-formatted text exactly once (W1, data-model §4),
		// then nil so the caller keeps going (W4/W5). Several failed remotes
		// still yield ONE event: the single `git fetch --all` failure already
		// aggregates their attribution (research D1/N2).
		color.Yellow(refreshWarningLine())
		fmt.Println(fetchErr.Error())
		return nil
	}

	return nil
}

// refreshWarningLine is the fixed yellow tool line for a failed pre-list
// refresh (contract §2, data-model §4). Pure — no git detail, no printing —
// so the caller appends outputError's multi-line text separately
// (execution/switch_test.go).
func refreshWarningLine() string {
	return "⚠ Refresh failed, showing last known branches"
}

func BrowseAndSwitchBranches(gitClient *core.GitClient) error {
	fmt.Println()
	color.Cyan("Switching branches")

	// Refresh every configured remote FIRST, so the list below observes the
	// post-refresh ref state (FR-001/FR-002, contract G1/G2 — research D3).
	if err := refreshRemoteStateBeforeListing(gitClient); err != nil {
		return err
	}

	// Branches(true): the current branch's name must participate in the
	// local-wins filter even though the builder excludes it from the list.
	branches, err := gitClient.Branches(true)
	if err != nil {
		return err
	}

	// Local read of the remote-tracking refs refreshed above — deleted
	// upstream branches are already pruned out of this snapshot (FR-002,
	// contract G1; feature 013 supersedes 012's offline-only FR-007).
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

	// Selection-time refresh for remote picks, kept unchanged alongside the
	// pre-list refresh (feature 013 clarification Q2→A, FR-009): abort before
	// checkout on any fetch failure (012 FR-013 — current branch and working
	// tree untouched). Local picks never fetch at selection time, so they work
	// even when every remote is unreachable (012 scenario 11).
	if selectedBranch.IsRemoteBranch() {
		if err := gitClient.Fetch(); err != nil {
			return err
		}
	}

	return gitClient.Checkout(selectedBranch)
}
