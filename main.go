package main

import (
	"fmt"
	"os"

	"github.com/cainlara/gogit-branch/core"
	"github.com/cainlara/gogit-branch/execution"

	"github.com/fatih/color"
)

const (
	MODE_LIST_LONG          = "list"
	MODE_LIST_SHORT         = "ls"
	MODE_HELP_LONG          = "help"
	MODE_HELP_SHORT         = "h"
	MODE_SWITCH_LONG        = "switch"
	MODE_SWITCH_SHORT       = "sw"
	MODE_DELETE_LONG        = "delete"
	MODE_DELETE_SHORT       = "del"
	MODE_BATCH_DELETE_LONG  = "batch-delete"
	MODE_BATCH_DELETE_SHORT = "bd"
	MODE_CREATE_LONG        = "create"
	MODE_CREATE_SHORT       = "c"
	MODE_STATUS_LONG        = "status"
	MODE_STATUS_SHORT       = "st"
	MODE_PUSH_LONG          = "push"
	MODE_PUSH_SHORT         = "p"
	MODE_LOG_LONG           = "log"
	MODE_LOG_SHORT          = "l"
	MODE_PULL_LONG          = "pull"
	MODE_PULL_SHORT         = "pl"
	MODE_RESET_LONG         = "reset"
	MODE_RESET_SHORT        = "r"
	MODE_CLONE_LONG         = "clone"
	MODE_CLONE_SHORT        = "cl"
	MODE_VERSION_LONG       = "version"
	MODE_VERSION_SHORT      = "v"
)

func main() {
	gitClient := core.NewGitClient("")

	if len(os.Args) < 2 {
		execution.PrintHelp(false, false)

		return
	}

	argsWithoutProg := os.Args[1:]

	triggerExecution(argsWithoutProg, gitClient)
}

func triggerExecution(args []string, gitClient *core.GitClient) {
	arg := args[0]
	isCreate := arg == MODE_CREATE_LONG || arg == MODE_CREATE_SHORT
	isLog := arg == MODE_LOG_LONG || arg == MODE_LOG_SHORT
	isReset := arg == MODE_RESET_LONG || arg == MODE_RESET_SHORT
	acceptsOptionalArg := isCreate || isLog || isReset
	isClone := arg == MODE_CLONE_LONG || arg == MODE_CLONE_SHORT

	// The create/c, log/l, and reset/r subcommands each accept one optional
	// extra argument (a branch name, a commit-count limit, or the --hard
	// flag); every other subcommand keeps rejecting any extra argument.
	// clone/cl is the exception with a RANGE: 0..4 extras (URL, optional
	// name+email pair, optional -anon) — deeper combination validation lives
	// in execution/clone.go so bad shapes fail before any git runs
	// (spec FR-007, research D1/D6). A bare `clone` (no URL) also passes
	// through here so parseCloneArgs can report the missing URL itself.
	if isClone {
		if len(args) > 5 {
			execution.PrintHelp(true, true)

			return
		}
	} else if len(args) > 2 || (len(args) == 2 && !acceptsOptionalArg) {
		execution.PrintHelp(true, true)

		return
	}

	var err error

	switch arg {
	case MODE_LIST_LONG, MODE_LIST_SHORT:
		err = execution.ListCurrentBranches(gitClient)
	case MODE_HELP_LONG, MODE_HELP_SHORT:
		execution.PrintHelp(false, true)
	case MODE_SWITCH_LONG, MODE_SWITCH_SHORT:
		err = execution.BrowseAndSwitchBranches(gitClient)
	case MODE_DELETE_LONG, MODE_DELETE_SHORT:
		err = execution.ListAndDeleteBranch(gitClient)
	case MODE_BATCH_DELETE_LONG, MODE_BATCH_DELETE_SHORT:
		err = execution.ListAndDeleteBranches(gitClient)
	case MODE_CREATE_LONG, MODE_CREATE_SHORT:
		var name *string
		if len(args) == 2 {
			name = &args[1]
		}
		err = execution.CreateAndSwitchBranch(gitClient, name)
	case MODE_STATUS_LONG, MODE_STATUS_SHORT:
		err = execution.ShowStatusAndOperate(gitClient)
	case MODE_PUSH_LONG, MODE_PUSH_SHORT:
		err = execution.PushCurrentBranch(gitClient)
	case MODE_LOG_LONG, MODE_LOG_SHORT:
		var limit *string
		if len(args) == 2 {
			limit = &args[1]
		}
		err = execution.ShowLog(gitClient, limit)
	case MODE_PULL_LONG, MODE_PULL_SHORT:
		// Intentionally NOT added to acceptsOptionalArg: pull takes zero
		// arguments (FR-007), so the existing guard rejects any extra arg
		// before this branch is reached (research.md D5).
		err = execution.PullCurrentBranch(gitClient)
	case MODE_RESET_LONG, MODE_RESET_SHORT:
		var flag *string
		if len(args) == 2 {
			flag = &args[1]
		}
		err = execution.ResetToLatestCommit(gitClient, flag)
	case MODE_CLONE_LONG, MODE_CLONE_SHORT:
		err = execution.CloneRepository(gitClient, args[1:])
	case MODE_VERSION_LONG, MODE_VERSION_SHORT:
		execution.ShowVersion()
	default:
		// Unknown command: show the error AND the usage list (spec 014 US4
		// scenario 2 — the usage list, including clone, must be reachable
		// from the unrecognized-command path too).
		execution.PrintHelp(true, true)
	}

	if err != nil {
		color.Red(fmt.Sprintf("Operation Failed: %v\n", err))
	}
}
