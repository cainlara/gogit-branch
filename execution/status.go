package execution

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/cainlara/gogit-branch/core"
	"github.com/cainlara/gogit-branch/model"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"golang.org/x/term"
)

const (
	KEY_COMMIT = 'c'
	KEY_ADD    = 'a'
	KEY_EXIT   = 'e'
)

// readKeyPress puts stdin into raw mode, reads exactly one byte with no
// Enter required, restores the terminal, and returns the byte read.
func readKeyPress() (byte, error) {
	fd := int(os.Stdin.Fd())

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return 0, err
	}
	defer term.Restore(fd, oldState)

	buf := make([]byte, 1)

	if _, err := os.Stdin.Read(buf); err != nil {
		return 0, err
	}

	return buf[0], nil
}

func categoryColor(category string) *color.Color {
	switch category {
	case model.CATEGORY_MODIFIED:
		return color.New(color.FgYellow)
	case model.CATEGORY_DELETED:
		return color.New(color.FgRed)
	case model.CATEGORY_NEW:
		return color.New(color.FgGreen)
	default:
		return color.New(color.FgMagenta)
	}
}

func untrackedColor() *color.Color {
	return color.New(color.FgCyan)
}

// printBranchInfo renders the current branch (or detached-HEAD indicator)
// plus any upstream tracking/ahead-behind counts, exactly mirroring what
// `git status` itself would report (FR-004).
func printBranchInfo(status *model.RepositoryStatus) {
	if status.IsDetached() {
		fmt.Printf("On %s\n", status.GetBranch())

		return
	}

	line := fmt.Sprintf("On branch %s", status.GetBranch())

	if status.GetUpstream() != "" {
		line += fmt.Sprintf(" (tracking %s", status.GetUpstream())

		if status.GetAhead() > 0 || status.GetBehind() > 0 {
			line += ", "

			if status.GetAhead() > 0 {
				line += "ahead " + strconv.Itoa(status.GetAhead())
			}

			if status.GetAhead() > 0 && status.GetBehind() > 0 {
				line += ", "
			}

			if status.GetBehind() > 0 {
				line += "behind " + strconv.Itoa(status.GetBehind())
			}
		}

		line += ")"
	}

	fmt.Println(line)
}

// printTrackedFiles renders the tracked-files section: one line per file,
// colored by its primary status category, annotated when it has both a
// staged and an unstaged change (FR-002, FR-004).
func printTrackedFiles(status *model.RepositoryStatus) {
	color.Cyan("Tracked files")

	if !status.HasTrackedChanges() {
		fmt.Println("  (nothing to show)")

		return
	}

	for _, file := range status.GetTracked() {
		line := categoryColor(file.Category()).Sprintf("  %s (%s)", file.GetPath(), file.Category())

		if annotation := file.StateAnnotation(); annotation != "" {
			line += fmt.Sprintf(" [%s]", annotation)
		}

		fmt.Println(line)
	}
}

// printUntrackedFiles renders the untracked-files section, all in one color
// distinct from every tracked-file category color (FR-003).
func printUntrackedFiles(status *model.RepositoryStatus) {
	color.Cyan("Untracked files")

	if !status.HasUntrackedFiles() {
		fmt.Println("  (nothing to show)")

		return
	}

	for _, path := range status.GetUntracked() {
		fmt.Println(untrackedColor().Sprintf("  %s", path))
	}
}

// printOptions renders the third section: the three fixed actions (FR-005).
func printOptions() {
	color.Cyan("Options")
	fmt.Println("  1 (c)ommit all tracked files")
	fmt.Println("  2 (a)dd untracked files and the commit all")
	fmt.Println("  3 (e)xit")
}

// promptForCommitMessage asks the user for a commit message and trims
// leading/trailing whitespace before returning it (FR-010).
func promptForCommitMessage() (string, error) {
	prompt := promptui.Prompt{
		Label: "Commit message",
	}

	message, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return message, nil
}

func requireNonEmptyMessage(message string) error {
	if strings.TrimSpace(message) == "" {
		return fmt.Errorf("a commit message is required")
	}

	return nil
}

// handleCommitTracked implements the 'c' action: commit all tracked changes.
func handleCommitTracked(gitClient *core.GitClient, status *model.RepositoryStatus) error {
	if !status.HasTrackedChanges() {
		color.Yellow("Nothing to commit — there are no tracked changes.")

		return nil
	}

	message, err := promptForCommitMessage()
	if err != nil {
		return err
	}

	if err := requireNonEmptyMessage(message); err != nil {
		return err
	}

	if err := gitClient.CommitTracked(strings.TrimSpace(message)); err != nil {
		return err
	}

	color.Green("%s Committed all tracked changes", EMOJI_HERB)

	return nil
}

// handleAddAndCommit implements the 'a' action: stage everything (tracked +
// untracked) and commit it together. Falls back to handleCommitTracked's
// behavior when there are no untracked files to add.
func handleAddAndCommit(gitClient *core.GitClient, status *model.RepositoryStatus) error {
	if !status.HasUntrackedFiles() {
		return handleCommitTracked(gitClient, status)
	}

	message, err := promptForCommitMessage()
	if err != nil {
		return err
	}

	if err := requireNonEmptyMessage(message); err != nil {
		return err
	}

	if err := gitClient.AddAllAndCommit(strings.TrimSpace(message)); err != nil {
		return err
	}

	color.Green("%s Staged untracked files and committed everything", EMOJI_HERB)

	return nil
}

// ShowStatusAndOperate renders the repository's working tree state as three
// sections and lets the user act on it via a single key press.
func ShowStatusAndOperate(gitClient *core.GitClient) error {
	fmt.Println()

	status, err := gitClient.Status()
	if err != nil {
		return err
	}

	printBranchInfo(status)
	fmt.Println()
	printTrackedFiles(status)
	fmt.Println()
	printUntrackedFiles(status)
	fmt.Println()
	printOptions()

	for {
		key, err := readKeyPress()
		if err != nil {
			return err
		}

		switch key {
		case KEY_COMMIT:
			return handleCommitTracked(gitClient, status)
		case KEY_ADD:
			return handleAddAndCommit(gitClient, status)
		case KEY_EXIT:
			return nil
		default:
			continue
		}
	}
}
