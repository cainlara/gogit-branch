package execution

import (
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/term"
)

const (
	// progressTickInterval is the repaint cadence while a stage runs. Every
	// tick advances the spinner, so the screen visibly changes far more often
	// than SC-005's 0.5s bound — even when git reports no new percent (the
	// "dormant network stretch" case, research D7). Percent updates arriving
	// between ticks are coalesced by the mutex-protected state (last value
	// wins at the next repaint) to avoid flicker.
	progressTickInterval = 100 * time.Millisecond

	// progressStageFetching/progressStagePulling are the two stage labels the
	// bar displays (spec FR-003, contract B3). They ride on the bar's own line
	// and are independent of the static stdout lines in execution/pull.go
	// (contract S4 — the bar never replaces them).
	progressStageFetching = "Fetching remote updates"
	progressStagePulling  = "Pulling remote updates"

	// progressBarWidth is the fill width of the bracketed bar in percent mode.
	progressBarWidth = 10

	// progressSpinnerFrames are ASCII spinner frames used in both forms so the
	// screen visibly changes even while a reported percent is stagnant
	// (SC-005 — a change at least every 0.5s; ASCII avoids font/emoji
	// portability issues across terminals).
	progressSpinnerFrames = "|/-\\"
)

// composeProgressFrame builds one bar line. With hasPercent it renders the
// determinate form `<spinner> <stage> [====>] NN%` where the displayed percent
// is capped at 99 while the stage is active (contract B1/B7 — a completed
// state must never be shown before completion, FR-008); otherwise it renders
// the animated form `<spinner> <stage>` with NO numeric percent at all
// (contract B2 — the renderer never fabricates numbers, SC-008). Pure function
// so it is directly unit-testable (research D2).
func composeProgressFrame(stage string, percent int, hasPercent bool, frame int) string {
	spin := string(progressSpinnerFrames[frame%len(progressSpinnerFrames)])

	if !hasPercent {
		return spin + " " + stage
	}

	if percent > 99 {
		percent = 99
	}

	filled := percent * progressBarWidth / 100
	bar := strings.Repeat("=", filled) + ">" + strings.Repeat(" ", progressBarWidth-filled-1)

	return spin + " " + stage + " [" + bar + "] " + strconv.Itoa(percent) + "%"
}

// clearSequence builds the bytes that erase a previously rendered frame of
// the given length from the status line: carriage return, space pad, carriage
// return — leaving the cursor at column 0 of a blank line, ready for the
// final message (contract L1). Pure function; contains only `\r` and spaces,
// never bar characters.
func clearSequence(renderedLen int) string {
	if renderedLen <= 0 {
		return ""
	}

	return "\r" + strings.Repeat(" ", renderedLen) + "\r"
}

// progressRenderer draws a single self-overwriting progress line to its
// writer (os.Stderr by default — the "status stream" from clarification
// Q2 → A, contract §1). It is purely visual: it never prompts, reads input,
// delays completion, or writes to any stream other than its own (FR-011,
// contract L5). All state is mutex-guarded because core invokes the progress
// callback from the stderr reader goroutine (research D3).
type progressRenderer struct {
	mu sync.Mutex

	// enabled is the interactive-status-stream gate (FR-009, contract S1/S3,
	// research D4): when the status stream (stderr) is not an interactive
	// terminal, every operation is a strict no-op producing ZERO bytes on any
	// stream (SC-004). Interactivity is checked exactly once at construction.
	enabled bool

	out        io.Writer
	stage      string
	percent    int
	hasPercent bool
	frame      int
	active     bool
	rendered   int // length of the frame currently occupying the line

	stop    chan struct{}
	stopped bool
}

// newProgressRenderer creates an ENABLED renderer writing to out (nil →
// os.Stderr). The writer is injectable so tests can assert exactly which
// bytes are produced (research D2); the interactive gate lives in
// newStatusStreamRenderer.
func newProgressRenderer(out io.Writer) *progressRenderer {
	if out == nil {
		out = os.Stderr
	}

	return &progressRenderer{out: out, enabled: true}
}

// newStatusStreamRenderer is the production constructor: the bar renders to
// the status stream (stderr, clarification Q2 → A) and ONLY when that stream
// is an interactive terminal (FR-009, contract S1, research D4). When stderr
// is redirected/piped the returned renderer is disabled and never writes a
// byte — even if stdout is still a terminal (contract S3), and vice versa the
// bar still shows when stdout is redirected but stderr is a terminal
// (contract S2). Uses the already-required golang.org/x/term (Constitution V,
// precedent: execution/status.go).
func newStatusStreamRenderer() *progressRenderer {
	r := newProgressRenderer(os.Stderr)
	r.enabled = term.IsTerminal(int(os.Stderr.Fd()))

	return r
}

// onProgress matches core.ProgressFunc: it latches the first percent seen for
// the current stage (hasPercent stays true for the rest of the stage so the
// bar never flickers back to the animated form mid-transfer) and ignores
// percent-less fragments (the ticker, not git, drives visible liveness).
func (r *progressRenderer) onProgress(percent int, hasPercent bool) {
	if !hasPercent {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.active {
		return
	}

	r.percent = percent
	r.hasPercent = true
}

// start opens the stage: it resets per-stage percent state and renders the
// FIRST frame synchronously (contract B4/FR-002 — visible within 1s; in
// practice within microseconds of the git command starting).
func (r *progressRenderer) start(stage string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.enabled {
		return
	}

	r.stage = stage
	r.percent = 0
	r.hasPercent = false
	r.active = true

	if r.stop == nil {
		r.stop = make(chan struct{})
		go r.tickLoop(r.stop)
	}

	r.renderLocked()
}

// tickLoop repaints every progressTickInterval until finish closes stop.
// It owns no state of its own — each iteration takes the renderer lock,
// advances the spinner, and repaints (research D7).
func (r *progressRenderer) tickLoop(stop <-chan struct{}) {
	ticker := time.NewTicker(progressTickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			r.mu.Lock()
			r.tickLocked()
			r.mu.Unlock()
		}
	}
}

// setStage switches the displayed stage label in place (FR-003/contract B3),
// resetting the percent latch because git reports percentages per stage, and
// re-renders immediately so the change does not wait for the next tick.
func (r *progressRenderer) setStage(stage string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.active {
		return
	}

	r.stage = stage
	r.percent = 0
	r.hasPercent = false

	r.renderLocked()
}

// finish clears the line and deactivates the renderer. It is idempotent
// (safe to call from the deferred panic guard and the explicit pre-result
// call), never waits for a tick (contract L2 — no delay added to
// completion), and after it returns the status line carries zero bar bytes
// (contract L1, FR-004).
func (r *progressRenderer) finish() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.active {
		return
	}

	r.active = false

	if r.stop != nil && !r.stopped {
		close(r.stop)
		r.stopped = true
	}

	if r.rendered > 0 {
		_, _ = io.WriteString(r.out, clearSequence(r.rendered))
		r.rendered = 0
	}
}

// tickLocked advances the animation frame and repaints. Called by the ticker
// (T008) and directly after state changes; the caller must hold r.mu.
func (r *progressRenderer) tickLocked() {
	if !r.active {
		return
	}

	r.frame++
	r.renderLocked()
}

// renderLocked paints the current frame in place: clear the previous frame
// (pad to its length), then write the new one. Caller must hold r.mu.
func (r *progressRenderer) renderLocked() {
	frame := composeProgressFrame(r.stage, r.percent, r.hasPercent, r.frame)

	_, _ = io.WriteString(r.out, clearSequence(r.rendered)+frame)
	r.rendered = len(frame)
}
