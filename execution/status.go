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

	STATUS_TAB_WIDTH = 8
)

// renderRow is one tracked or untracked file's plain-text rendering
// ingredients, built before any tab-padding or coloring is applied (see
// data-model.md's renderRow entity).
type renderRow struct {
	label      string
	filename   string
	annotation string
	colorer    *color.Color
}

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

// upstreamColor renders the "has upstream" indicator, distinct from every
// other color used on this screen (research.md).
func upstreamColor() *color.Color {
	return color.New(color.FgBlue)
}

// noUpstreamColor renders the "no upstream" indicator, distinct from
// upstreamColor and from every other color used on this screen (research.md).
func noUpstreamColor() *color.Color {
	return color.New(color.FgHiBlack)
}

// nextTabStop returns the smallest multiple of tabWidth strictly greater
// than column — i.e. where a single tab character would land the cursor
// if it currently sits at column.
func nextTabStop(column, tabWidth int) int {
	return (column/tabWidth + 1) * tabWidth
}

// targetColumn returns the shared column every row's file name should
// start at: the smallest multiple of tabWidth strictly greater than the
// longest indent+label prefix across every row. Returns 0 for an empty
// slice (nothing to align).
func targetColumn(rows []renderRow, indent, tabWidth int) int {
	maxPrefix := 0

	for _, row := range rows {
		prefix := indent + len(row.label)

		if prefix > maxPrefix {
			maxPrefix = prefix
		}
	}

	if maxPrefix == 0 {
		return 0
	}

	return nextTabStop(maxPrefix, tabWidth)
}

// padding returns exactly as many tab characters as are needed to advance
// from prefixLen to target, per research.md's two-pass alignment decision.
func padding(prefixLen, target, tabWidth int) string {
	tabs := ""

	for pos := prefixLen; pos < target; pos = nextTabStop(pos, tabWidth) {
		tabs += "\t"
	}

	return tabs
}

// printBranchInfo renders the current branch (or detached-HEAD indicator)
// plus an explicit, colored upstream-tracking indicator: blue with the
// tracked remote (and any ahead/behind counts) when one is configured, or
// gray stating that none is set when it isn't (FR-006 through FR-010).
// Detached HEAD is unaffected — there is no branch for either state to
// describe.
func printBranchInfo(status *model.RepositoryStatus) {
	if status.IsDetached() {
		fmt.Printf("On %s\n", status.GetBranch())

		return
	}

	fmt.Printf("On branch %s ", status.GetBranch())

	if status.GetUpstream() == "" {
		fmt.Println(noUpstreamColor().Sprint("(no remote upstream set)"))

		return
	}

	segment := fmt.Sprintf("(tracking %s", status.GetUpstream())

	if status.GetAhead() > 0 || status.GetBehind() > 0 {
		segment += ", "

		if status.GetAhead() > 0 {
			segment += "ahead " + strconv.Itoa(status.GetAhead())
		}

		if status.GetAhead() > 0 && status.GetBehind() > 0 {
			segment += ", "
		}

		if status.GetBehind() > 0 {
			segment += "behind " + strconv.Itoa(status.GetBehind())
		}
	}

	segment += ")"

	fmt.Println(upstreamColor().Sprint(segment))
}

// buildTrackedRows converts each tracked file into a renderRow, labeling it
// by its primary status category and preserving its staged/unstaged
// annotation (FR-004 mirror target for buildUntrackedRows).
func buildTrackedRows(status *model.RepositoryStatus) []renderRow {
	rows := make([]renderRow, 0, len(status.GetTracked()))

	for _, file := range status.GetTracked() {
		rows = append(rows, renderRow{
			label:      fmt.Sprintf("(%s)", file.Category()),
			filename:   file.GetPath(),
			annotation: file.StateAnnotation(),
			colorer:    categoryColor(file.Category()),
		})
	}

	return rows
}

// buildUntrackedRows converts each untracked path into a renderRow with an
// explicit "(untracked)" label, mirroring the tracked-file labels (FR-004).
func buildUntrackedRows(status *model.RepositoryStatus) []renderRow {
	rows := make([]renderRow, 0, len(status.GetUntracked()))

	for _, path := range status.GetUntracked() {
		rows = append(rows, renderRow{
			label:    "(untracked)",
			filename: path,
			colorer:  untrackedColor(),
		})
	}

	return rows
}

// printFileSection renders one section's header followed by its rows, each
// tab-aligned to the shared target column computed across both sections
// (FR-001, FR-002, FR-003), or the existing "(nothing to show)" placeholder
// when there are no rows. The plain-text line is fully assembled before a
// single color wrap is applied, so ANSI codes never affect the tab-count
// computation (research.md).
func printFileSection(header string, rows []renderRow, indent, target int) {
	color.Cyan(header)

	if len(rows) == 0 {
		fmt.Println("  (nothing to show)")

		return
	}

	prefix := strings.Repeat(" ", indent)

	for _, row := range rows {
		line := prefix + row.label + padding(indent+len(row.label), target, STATUS_TAB_WIDTH) + row.filename

		if row.annotation != "" {
			line += fmt.Sprintf(" [%s]", row.annotation)
		}

		fmt.Println(row.colorer.Sprint(line))
	}
}

// printOptions renders the third section: the three fixed actions (FR-005),
// on a single line for at-a-glance scanning.
func printOptions() {
	color.Cyan("Options")
	fmt.Println("  (c)ommit all tracked files  |  (a)dd untracked files and then commit all  |  (e)xit")
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

	trackedRows := buildTrackedRows(status)
	untrackedRows := buildUntrackedRows(status)
	target := targetColumn(append(append([]renderRow{}, trackedRows...), untrackedRows...), 2, STATUS_TAB_WIDTH)

	printBranchInfo(status)
	fmt.Println()
	printFileSection("Tracked files", trackedRows, 2, target)
	fmt.Println()
	printFileSection("Untracked files", untrackedRows, 2, target)
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
