package core

import (
	"errors"
	"os/exec"
	"strconv"
	"strings"

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
		return pushError(output, err)
	}

	branchOut, err := g.runGitCommand("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return err
	}

	branchName := strings.TrimSpace(string(branchOut))

	out, err = g.runGitCommandCombinedOutput("push", "--set-upstream", "origin", branchName)
	if err != nil {
		return pushError(string(out), err)
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

// pushError turns a failed push's combined output into an error carrying
// git's own message. Unlike the other GitClient commands, a rejected
// `git push`'s output can start with a "To <remote>" progress line before
// the error:/fatal: line, so this checks for the marker anywhere in the
// output rather than only as a prefix.
func pushError(output string, err error) error {
	if strings.Contains(output, OUTPUT_ERROR_PREFIX) || strings.Contains(output, OUTPUT_FATAL_PREFIX) {
		return errors.New(output)
	}

	return err
}
