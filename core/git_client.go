package core

import (
	"errors"
	"os/exec"
	"strings"

	"github.com/cainlara/gogit-branch/model"
)

const (
	CURRENT_BRANCH_PREFIX = "* "
	OUTPUT_ERROR_PREFIX   = "error:"
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
