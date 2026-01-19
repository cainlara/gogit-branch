package core

import (
	"cainlara/gogit-branch/model"
	"os/exec"
	"strings"
)

const CURRENT_BRANCH_PREFIX = "* "

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

func (g *GitClient) Branches() ([]model.Branch, error) {
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

		branch := model.NewBranch(branchName, branchName, "", "", isCurrent)

		branches = append(branches, *branch)
	}

	return branches, nil
}
