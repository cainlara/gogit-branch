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
//
// The progress bar (feature 011) is purely visual and additive: the static
// lines above are printed exactly as before (contract S4), the bar renders on
// the status stream alongside them, and it is explicitly finished BEFORE any
// result or error message can appear (FR-004/SC-002, contract L1) — the
// deferred finish() is an idempotent panic guard behind that explicit call.
// The git steps themselves are unchanged: no --force/--rebase, identical
// outcome classification and error wrapping (FR-005/FR-006).
//
// Feature 016 adds the change list: the HEAD baseline is captured before the
// fetch, and after the bar finishes a successful "updated" pull prints the
// change list BEFORE the success message — frozen order progress bar → change
// list → success message (FR-001a, clarification 2026-09-27 → B). Every list
// step degrades to silence (no baseline, up-to-date, diff failure, zero
// sections) so the descriptive output can never alter the pull's outcome
// (FR-009, research D5); no prompt is involved (FR-011).
func PullCurrentBranch(gitClient *core.GitClient) error {
	fmt.Println()
	color.Cyan("Pulling branch")
	color.Cyan("Fetching remote updates...")

	baseline, _ := gitClient.HeadCommit()

	bar := newStatusStreamRenderer()
	defer bar.finish()

	stopInterruptWatch := clearBarOnInterrupt(bar)
	defer stopInterruptWatch()

	bar.start(progressStageFetching)

	if err := gitClient.FetchWithProgress(bar.onProgress); err != nil {
		bar.finish()
		return err
	}

	bar.setStage(progressStagePulling)

	updated, err := gitClient.PullWithProgress(bar.onProgress)
	bar.finish()

	if err != nil {
		return err
	}

	if updated {
		printChangeList(gitClient, baseline)
		color.Green(fmt.Sprintf("%s Pulled current branch from the remote", EMOJI_ROCKET))
	} else {
		color.Green(fmt.Sprintf("%s Branch already up to date", EMOJI_HERB))
	}

	return nil
}

// printChangeList renders the change list between the cleared progress bar and
// the success message (FR-001a). It is descriptive-only by contract: a missing
// baseline (zero-commit repository), an unchanged HEAD, a failed diff, or a
// diff with zero sections each print nothing and never surface an error
// (FR-007, FR-008, FR-009 — research D5). Pre-existing uncommitted work cannot
// appear because only the two commit endpoints are diffed (FR-006).
func printChangeList(gitClient *core.GitClient, baseline string) {
	if baseline == "" {
		return
	}

	head, err := gitClient.HeadCommit()
	if err != nil || head == baseline {
		return
	}

	changes, err := gitClient.ChangeList(baseline, head)
	if err != nil {
		return
	}

	if text := composeChangeList(changes); text != "" {
		fmt.Print(text)
	}
}
