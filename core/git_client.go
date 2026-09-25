package core

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/cainlara/gogit-branch/model"
)

const (
	CURRENT_BRANCH_PREFIX = "* "
	OUTPUT_ERROR_PREFIX   = "error:"
	OUTPUT_FATAL_PREFIX   = "fatal:"

	STATUS_BRANCH_LINE_PREFIX = "## "
	STATUS_DETACHED_HEAD      = "HEAD (no branch)"
	STATUS_UNTRACKED_PREFIX   = "??"

	NO_UPSTREAM_BRANCH_MARKER = "has no upstream branch"

	// ALREADY_UP_TO_DATE_MARKER is git's stable message for a `git pull` that
	// had nothing to integrate. Substring-matched (not exact) so the legacy
	// hyphenated spelling "Already up-to-date." also matches — same
	// English-message-matching approach as NO_UPSTREAM_BRANCH_MARKER
	// (research.md D2).
	ALREADY_UP_TO_DATE_MARKER = "Already up to date"
)

type GitClient struct {
	Path string
}

func NewGitClient(path string) *GitClient {
	if path == "" {
		if root, err := getGitRoot(); err == nil {
			path = root
		}
	}

	return &GitClient{Path: path}
}

func getGitRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()

	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

func (g *GitClient) runGitCommand(args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)

	if g.Path != "" {
		cmd.Dir = g.Path
	}

	return cmd.Output()
}

func (g *GitClient) runGitCommandCombinedOutput(args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)

	if g.Path != "" {
		cmd.Dir = g.Path
	}

	return cmd.CombinedOutput()
}

func (g *GitClient) Branches(includeCurrent bool) ([]model.Branch, error) {
	out, err := g.runGitCommand("branch")

	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")

	branches := make([]model.Branch, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		var branchName string
		var isCurrent bool

		if strings.HasPrefix(line, CURRENT_BRANCH_PREFIX) {
			branchName = strings.TrimPrefix(line, CURRENT_BRANCH_PREFIX)
			isCurrent = true
		} else {
			branchName = strings.TrimSpace(line)
		}

		if isCurrent && !includeCurrent {
			continue
		}

		branch := model.NewBranch(branchName, "", "", isCurrent)

		hashOutput, err := g.runGitCommand("rev-parse", branch.GetName())
		if err != nil {
			return nil, err
		}

		fullHash := strings.TrimSpace(string(hashOutput))
		shortHash := fullHash[:7]

		branch.SetFullHash(fullHash)
		branch.SetShortHash(shortHash)

		branches = append(branches, *branch)
	}

	return branches, nil
}

func (g *GitClient) Checkout(branch model.Branch) error {
	out, err := g.runGitCommandCombinedOutput("checkout", branch.GetName())
	if err != nil {
		output := string(out)

		if strings.HasPrefix(output, OUTPUT_ERROR_PREFIX) {
			return errors.New(output)
		}

		return err
	}

	return nil
}

func (g *GitClient) CreateAndSwitchBranch(branchName string) error {
	out, err := g.runGitCommandCombinedOutput("checkout", "-b", branchName)
	if err != nil {
		output := string(out)

		if strings.HasPrefix(output, OUTPUT_ERROR_PREFIX) || strings.HasPrefix(output, OUTPUT_FATAL_PREFIX) {
			return errors.New(output)
		}

		return err
	}

	return nil
}

func (g *GitClient) DeleteBranch(branch model.Branch) error {
	out, err := g.runGitCommandCombinedOutput("branch", "-D", branch.GetName())
	if err != nil {
		output := string(out)

		if strings.HasPrefix(output, OUTPUT_ERROR_PREFIX) {
			return errors.New(output)
		}

		return err
	}

	return nil
}

func (g *GitClient) DeleteBranches(branches []model.Branch) error {
	args := make([]string, 0, len(branches)+2)
	args = append(args, "branch")
	args = append(args, "-D")

	for _, branch := range branches {
		args = append(args, branch.GetName())
	}

	out, err := g.runGitCommandCombinedOutput(args...)
	if err != nil {
		output := string(out)

		if strings.HasPrefix(output, OUTPUT_ERROR_PREFIX) {
			return errors.New(output)
		}

		return err
	}

	return nil
}

// Status runs `git status --porcelain --branch` and parses its output into a
// model.RepositoryStatus. All porcelain parsing lives here, per this
// package's role as the sole place that shells out to and interprets git.
func (g *GitClient) Status() (*model.RepositoryStatus, error) {
	out, err := g.runGitCommand("status", "--porcelain", "--branch")
	if err != nil {
		return nil, err
	}

	status := model.NewRepositoryStatus()

	lines := strings.Split(strings.TrimRight(string(out), "\n"), "\n")

	for i, line := range lines {
		if i == 0 && strings.HasPrefix(line, STATUS_BRANCH_LINE_PREFIX) {
			parseStatusBranchLine(strings.TrimPrefix(line, STATUS_BRANCH_LINE_PREFIX), status)

			continue
		}

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, STATUS_UNTRACKED_PREFIX) {
			status.AddUntracked(strings.TrimSpace(line[len(STATUS_UNTRACKED_PREFIX):]))

			continue
		}

		if len(line) < 3 {
			continue
		}

		stagedCode := line[0]
		unstagedCode := line[1]
		path := line[3:]

		status.AddTracked(*model.NewFileStatus(path, stagedCode, unstagedCode))
	}

	return status, nil
}

// parseStatusBranchLine parses the content after "## " from
// `git status --porcelain --branch`, e.g. "main", "HEAD (no branch)", or
// "main...origin/main [ahead 1, behind 2]".
func parseStatusBranchLine(line string, status *model.RepositoryStatus) {
	if line == STATUS_DETACHED_HEAD {
		status.SetDetached(true)
		status.SetBranch(STATUS_DETACHED_HEAD)

		return
	}

	parts := strings.SplitN(line, "...", 2)
	status.SetBranch(parts[0])

	if len(parts) < 2 {
		return
	}

	rest := parts[1]

	bracketStart := strings.Index(rest, " [")
	if bracketStart == -1 {
		status.SetUpstream(rest)

		return
	}

	status.SetUpstream(rest[:bracketStart])

	tracking := strings.TrimSuffix(rest[bracketStart+2:], "]")
	if tracking == "gone" {
		return
	}

	for _, part := range strings.Split(tracking, ", ") {
		if n, ok := strings.CutPrefix(part, "ahead "); ok {
			if value, err := strconv.Atoi(n); err == nil {
				status.SetAhead(value)
			}
		} else if n, ok := strings.CutPrefix(part, "behind "); ok {
			if value, err := strconv.Atoi(n); err == nil {
				status.SetBehind(value)
			}
		}
	}
}

// CommitTracked commits all tracked changes (staged and unstaged) using the
// given message, mirroring `git commit -a -m <message>`.
func (g *GitClient) CommitTracked(message string) error {
	out, err := g.runGitCommandCombinedOutput("commit", "-a", "-m", message)
	if err != nil {
		output := string(out)

		if strings.HasPrefix(output, OUTPUT_ERROR_PREFIX) || strings.HasPrefix(output, OUTPUT_FATAL_PREFIX) {
			return errors.New(output)
		}

		return err
	}

	return nil
}

// AddAllAndCommit explicitly stages every tracked and untracked change, then
// commits everything staged using the given message. If staging succeeds but
// the commit fails, the staged files are left staged — no rollback is
// attempted (FR-013).
func (g *GitClient) AddAllAndCommit(message string) error {
	if out, err := g.runGitCommandCombinedOutput("add", "-A"); err != nil {
		output := string(out)

		if strings.HasPrefix(output, OUTPUT_ERROR_PREFIX) || strings.HasPrefix(output, OUTPUT_FATAL_PREFIX) {
			return errors.New(output)
		}

		return err
	}

	out, err := g.runGitCommandCombinedOutput("commit", "-m", message)
	if err != nil {
		output := string(out)

		if strings.HasPrefix(output, OUTPUT_ERROR_PREFIX) || strings.HasPrefix(output, OUTPUT_FATAL_PREFIX) {
			return errors.New(output)
		}

		return err
	}

	return nil
}

// Push runs `git push` for the current branch. If that fails specifically
// because the current branch has no upstream configured yet, it resolves
// the current branch name and retries once as
// `git push --set-upstream origin <branch>`, so a brand-new local branch
// publishes in a single call instead of surfacing the "no upstream branch"
// error. Any other failure (no remote, non-fast-forward rejection,
// detached HEAD, etc.) is returned as-is, with no retry and no force-push.
func (g *GitClient) Push() error {
	out, err := g.runGitCommandCombinedOutput("push")
	if err == nil {
		return nil
	}

	output := string(out)

	if !hasNoUpstreamBranchError(output) {
		return outputError(output, err)
	}

	branchOut, err := g.runGitCommand("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return err
	}

	branchName := strings.TrimSpace(string(branchOut))

	out, err = g.runGitCommandCombinedOutput("push", "--set-upstream", "origin", branchName)
	if err != nil {
		return outputError(string(out), err)
	}

	return nil
}

// hasNoUpstreamBranchError reports whether a failed `git push`'s combined
// output matches git's stable "has no upstream branch" message — the one
// specific failure Push retries with --set-upstream. Every other failure
// is left untouched.
func hasNoUpstreamBranchError(output string) bool {
	return strings.Contains(output, NO_UPSTREAM_BRANCH_MARKER)
}

// outputError turns a failed git command's combined output into an error
// carrying git's own message. Network-facing commands (push, fetch, pull) can
// emit progress lines ("To <remote>…", "From <remote>…") BEFORE the
// error:/fatal: line — unlike the checkout/log/reset family, whose failures
// emit that prefix as their very first line — so this checks for the marker
// anywhere in the output rather than only as a prefix (research.md D3).
//
// Marker-free output is still carried when present: `git pull`'s
// no-upstream and detached-HEAD failures print actionable English with NO
// error:/fatal: prefix at all (implementation-time discovery, research.md D3),
// and surfacing that beats exec's bare "exit status 128" (SC-005). Only an
// empty output falls back to the raw command error.
func outputError(output string, err error) error {
	if strings.Contains(output, OUTPUT_ERROR_PREFIX) || strings.Contains(output, OUTPUT_FATAL_PREFIX) {
		return errors.New(output)
	}

	if trimmed := strings.TrimSpace(output); trimmed != "" {
		return errors.New(trimmed)
	}

	return err
}

// ProgressFunc receives one progress update parsed from a git command's stderr
// stream. hasPercent reports whether this fragment carried a numeric completion
// percentage (percent is then a value in 0..100, clamped); hasPercent=false
// means the fragment contained no percentage and exists only as an activity
// signal — the renderer must never treat it as a number (spec FR-008/SC-008).
// Defined here, in core, because percent extraction is git-output parsing
// (research.md D3).
type ProgressFunc func(percent int, hasPercent bool)

var (
	// progressPercentRe matches numeric percent tokens git emits in progress
	// updates ("Receiving objects:  42% (35/83)").
	progressPercentRe = regexp.MustCompile(`\d+%`)

	// sidebandProgressRe matches the non-percent sideband progress lines
	// (`--progress`-only output such as "remote: Enumerating objects: 153,
	// done." / "remote: Total 153 (delta 0)…"). Verified experimentally
	// (research.md D8): plain (no --progress) non-TTY runs never emit these,
	// so they must not leak into error text either.
	sidebandProgressRe = regexp.MustCompile(`^(?:remote: )?(?:Enumerating|Counting|Compressing|Receiving|Resolving|Updating|Total)\b`)
)

// parseProgressPercent extracts the LAST percent token from a stderr fragment
// (last-wins across interleaved sub-phases) and clamps values above 100 to
// 100. Returns hasPercent=false when the fragment carries no token at all
// (research.md D6, data-model.md "Percent Token").
func parseProgressPercent(fragment string) (int, bool) {
	matches := progressPercentRe.FindAllString(fragment, -1)
	if len(matches) == 0 {
		return 0, false
	}

	value, err := strconv.Atoi(strings.TrimSuffix(matches[len(matches)-1], "%"))
	if err != nil {
		return 0, false
	}

	if value > 100 {
		value = 100
	}

	return value, true
}

// isProgressFragment reports whether a stderr fragment is git progress noise
// rather than actionable output. Progress-shaped fragments are excluded from
// the accumulated output that outputError wraps, so failure messages stay
// byte-identical to the pre-feature (no --progress) behavior; fragments
// carrying error:/fatal: markers are NEVER dropped, so real git failures —
// including `remote: error:` lines from pre-receive hooks — survive
// (research.md D8 amendment, spec FR-005/SC-003).
func isProgressFragment(fragment string) bool {
	if strings.Contains(fragment, OUTPUT_ERROR_PREFIX) || strings.Contains(fragment, OUTPUT_FATAL_PREFIX) {
		return false
	}

	if progressPercentRe.MatchString(fragment) {
		return true
	}

	return sidebandProgressRe.MatchString(strings.TrimSpace(fragment))
}

// runGitCommandWithProgress runs `git <args>` with stdout and stderr streamed
// instead of buffered. Every stderr fragment (split on both \r and \n — git
// progress uses carriage returns as separators) is reported to onProgress
// (may be nil), and the command's full text output is accumulated so callers
// can apply the exact same outputError/ALREADY_UP_TO_DATE_MARKER processing
// they do on CombinedOutput results (research.md D8). Progress-shaped stderr
// fragments are excluded from that accumulation (isProgressFragment). Output
// ordering is deterministic: processed stderr lines are appended as they
// arrive, raw stdout is queued and appended only after the stderr reader
// finishes. That reproduces CombinedOutput's single-pipe time order without a
// two-reader race, because git block-buffers its stdout when piped and flushes
// it at process exit while stderr is written unbuffered mid-run — pull's
// stderr error block therefore always precedes the stdout "Updating a..b"
// (research.md D8 amendment 2, quickstart P13/010 S9).
func (g *GitClient) runGitCommandWithProgress(onProgress ProgressFunc, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)

	if g.Path != "" {
		cmd.Dir = g.Path
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	var mu sync.Mutex
	var outBuf strings.Builder

	// appendOut appends processed stderr text immediately (stderr streams
	// unbuffered mid-run, so it leads CombinedOutput's time order too).
	appendOut := func(s string) {
		mu.Lock()
		outBuf.WriteString(s)
		mu.Unlock()
	}

	// stdoutChunks holds raw stdout in arrival order; it is appended AFTER the
	// stderr reader finishes (see runGitCommandWithProgress's doc), because
	// git block-buffers stdout and flushes it at exit — late in time order,
	// and a live two-reader append would race on near-simultaneous writes
	// (research.md D8 amendment 2). Only this goroutine writes the slice; it
	// is read after wg.Wait, which synchronizes the happens-before edge.
	var stdoutChunks []string

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		buf := make([]byte, 4096)

		for {
			n, readErr := stdoutPipe.Read(buf)

			if n > 0 {
				stdoutChunks = append(stdoutChunks, string(buf[:n]))
			}

			if readErr != nil {
				break
			}
		}
	}()

	go func() {
		defer wg.Done()

		var partial []byte

		// pendingCR tracks a \r delimiter so that the \n of a \r\n pair is
		// swallowed (one line, not two) while a standalone \n delimiting an
		// EMPTY fragment is a genuine blank line in git's message and must
		// survive — git's no-tracking/detached-HEAD texts are paragraph-
		// formatted and baseline CombinedOutput preserves those blank lines
		// byte-for-byte (FR-005, SC-003).
		pendingCR := false

		emit := func(text string) {
			if text == "" {
				return
			}

			if onProgress != nil {
				percent, hasPercent := parseProgressPercent(text)
				onProgress(percent, hasPercent)
			}

			if !isProgressFragment(text) {
				appendOut(text + "\n")
			}
		}

		buf := make([]byte, 4096)

		for {
			n, readErr := stderrPipe.Read(buf)

			if n > 0 {
				partial = append(partial, buf[:n]...)

				for {
					idx := bytes.IndexAny(partial, "\r\n")
					if idx < 0 {
						break
					}

					delim := partial[idx]
					frag := string(partial[:idx])
					partial = partial[idx+1:]

					if delim == '\r' {
						pendingCR = true
						emit(frag)
						continue
					}

					// delim == '\n'
					if pendingCR {
						pendingCR = false
					} else if frag == "" {
						appendOut("\n")
						continue
					}

					emit(frag)
				}
			}

			if readErr != nil {
				break
			}
		}

		// Final delimiter-less remainder: append raw (git chose not to end
		// with a newline — don't add one), still filtering progress and
		// still feeding percent updates to the callback.
		if len(partial) > 0 {
			text := string(partial)

			if onProgress != nil {
				percent, hasPercent := parseProgressPercent(text)
				onProgress(percent, hasPercent)
			}

			if !isProgressFragment(text) {
				appendOut(text)
			}
		}
	}()

	wg.Wait()
	waitErr := cmd.Wait()

	// Deterministic tail: raw stdout (git's exit-time flush) after all
	// processed stderr — see the function doc for why this beats a live
	// interleave.
	for _, chunk := range stdoutChunks {
		outBuf.WriteString(chunk)
	}

	out := outBuf.String()

	return []byte(out), waitErr
}

// Fetch runs `git fetch`, updating the remote-tracking refs for the configured
// remote without touching the working tree or the current branch. It is the
// first step of the pull command: execution/pull.go only proceeds to Pull()
// when this returns nil (FR-002, FR-003). Failures (no remote, unreachable
// remote, auth errors) are wrapped via outputError so git's actionable message
// survives even when preceded by progress output (research.md D3).
func (g *GitClient) Fetch() error {
	out, err := g.runGitCommandCombinedOutput("fetch")
	if err != nil {
		return outputError(string(out), err)
	}

	return nil
}

// FetchWithProgress is Fetch with git's `--progress` forced on and stderr
// streamed live to onProgress (may be nil). `--progress` is required because
// GitClient never attaches a TTY, so git would otherwise silence its own
// progress output entirely (research.md D1). Error wrapping is identical to
// Fetch — progress-shaped fragments are stripped before outputError sees the
// text, keeping failure messages byte-identical to the non-progress path
// (research.md D8 amendment, FR-005/SC-003).
func (g *GitClient) FetchWithProgress(onProgress ProgressFunc) error {
	out, err := g.runGitCommandWithProgress(onProgress, "fetch", "--progress")
	if err != nil {
		return outputError(string(out), err)
	}

	return nil
}

// Pull runs a plain `git pull` for the current branch — deliberately WITHOUT
// --force or --rebase, so git's own safety checks (non-fast-forward refusal,
// uncommitted-change overwrite refusal) stay intact (contract §2 guarantee G2).
// On success it reports whether remote commits were actually integrated:
// updated=false exactly when git prints ALREADY_UP_TO_DATE_MARKER, updated=true
// otherwise (FR-004, research.md D2, data-model.md "Pull Outcome"). On failure
// the branch is left as git left it — Pull never rolls back or retries.
func (g *GitClient) Pull() (bool, error) {
	out, err := g.runGitCommandCombinedOutput("pull")
	if err != nil {
		return false, outputError(string(out), err)
	}

	return !strings.Contains(string(out), ALREADY_UP_TO_DATE_MARKER), nil
}

// PullWithProgress is Pull with git's `--progress` forced on and stderr
// streamed live to onProgress (may be nil). Same guarantees as Pull: no
// --force/--rebase, identical outcome classification via
// ALREADY_UP_TO_DATE_MARKER (always on stdout, never stripped), and identical
// error wrapping (research.md D8 amendment).
func (g *GitClient) PullWithProgress(onProgress ProgressFunc) (bool, error) {
	out, err := g.runGitCommandWithProgress(onProgress, "pull", "--progress")
	if err != nil {
		return false, outputError(string(out), err)
	}

	return !strings.Contains(string(out), ALREADY_UP_TO_DATE_MARKER), nil
}

// Log runs `git log`, capped at limit entries, using the fixed one-line
// colorized format this tool always shows. `--color=always` is required,
// not cosmetic: GitClient never attaches a real TTY to git (it always
// shells out via os/exec), so git's own color auto-detection would
// otherwise silently strip every %C(...) placeholder in the format string
// (research.md). On success, the trimmed output is split into one string
// per commit, since the requested format produces exactly one line per
// commit with no header/footer noise.
func (g *GitClient) Log(limit int) ([]string, error) {
	out, err := g.runGitCommandCombinedOutput(
		"log",
		"--color=always",
		"--pretty=format:%Cred%h%Creset - %s %Cgreen(%cr) %C(bold blue)<%an>%Creset",
		fmt.Sprintf("-%d", limit),
	)
	if err != nil {
		output := string(out)

		if strings.HasPrefix(output, OUTPUT_ERROR_PREFIX) || strings.HasPrefix(output, OUTPUT_FATAL_PREFIX) {
			return nil, errors.New(output)
		}

		return nil, err
	}

	trimmed := strings.TrimRight(string(out), "\n")
	if trimmed == "" {
		return []string{}, nil
	}

	return strings.Split(trimmed, "\n"), nil
}

// Reset always runs `git reset --hard`, reverting all tracked changes to the
// current branch's latest commit. This happens regardless of removeUntracked
// (a coincidental naming overlap with git's own --hard reset mode, not a
// semantic one — see research.md): FR-004 requires even the no-flag mode to
// fully revert the working tree, which plain `git reset` (index-only) would
// not do. When removeUntracked is true, an additional `git clean -fd` runs to
// remove untracked files too. Mirrors AddAllAndCommit's two-step, no-rollback
// pattern: if the first command succeeds but the second fails, the first
// step's effect is not undone.
func (g *GitClient) Reset(removeUntracked bool) error {
	if out, err := g.runGitCommandCombinedOutput("reset", "--hard"); err != nil {
		output := string(out)

		if strings.HasPrefix(output, OUTPUT_ERROR_PREFIX) || strings.HasPrefix(output, OUTPUT_FATAL_PREFIX) {
			return errors.New(output)
		}

		return err
	}

	if !removeUntracked {
		return nil
	}

	if out, err := g.runGitCommandCombinedOutput("clean", "-fd"); err != nil {
		output := string(out)

		if strings.HasPrefix(output, OUTPUT_ERROR_PREFIX) || strings.HasPrefix(output, OUTPUT_FATAL_PREFIX) {
			return errors.New(output)
		}

		return err
	}

	return nil
}
