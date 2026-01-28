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
	if len(args) > 1 {
		execution.PrintHelp(true, true)
		return
	}

	arg := args[0]

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
	case MODE_VERSION_LONG, MODE_VERSION_SHORT:
		execution.ShowVersion()
	default:
		execution.PrintHelp(true, false)
	}

	if err != nil {
		color.Red(fmt.Sprintf("Operation Failed: %v\n", err))
	}
}
