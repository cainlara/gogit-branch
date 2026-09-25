package execution

import (
	"fmt"
	"os"
	"os/signal"

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
func PullCurrentBranch(gitClient *core.GitClient) error {
	fmt.Println()
	color.Cyan("Pulling branch")
	color.Cyan("Fetching remote updates...")

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
		color.Green(fmt.Sprintf("%s Pulled current branch from the remote", EMOJI_ROCKET))
	} else {
		color.Green(fmt.Sprintf("%s Branch already up to date", EMOJI_HERB))
	}

	return nil
}

// clearBarOnInterrupt ensures Ctrl+C clears the progress line before the
// process dies (FR-010, US3, contract L3, research D5): while registered, a
// SIGINT first runs the renderer's finish(), then the default disposition is
// restored (signal.Reset) and SIGINT is re-raised to this process — so
// termination behaves exactly as it would have without the handler (same
// signal, same shell-visible exit semantics, no swallowed interrupt), with no
// half-drawn bar remnant left for the next prompt. The returned cleanup
// function unregisters the watch when the command completes normally; a
// SIGINT arriving after that runs with default handling, unchanged.
func clearBarOnInterrupt(bar *progressRenderer) func() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	done := make(chan struct{})

	go func() {
		select {
		case <-sigCh:
			bar.finish()
			signal.Reset(os.Interrupt)

			if proc, err := os.FindProcess(os.Getpid()); err == nil {
				_ = proc.Signal(os.Interrupt)
			}
		case <-done:
		}
	}()

	return func() {
		signal.Stop(sigCh)
		close(done)
	}
}
