package main

import (
	"fmt"
	"os"

	"cainlara/gogit-branch/execution"

	"github.com/fatih/color"
)

const (
	MODE_LIST_LONG  = "list"
	MODE_LIST_SHORT = "ls"
	MODE_HELP_LONG  = "help"
	MODE_HELP_SHORT = "h"
)

func main() {
	if len(os.Args) < 2 {
		execution.ListCurrentBranches()

		return
	}

	argsWithoutProg := os.Args[1:]

	triggerExecution(argsWithoutProg)
}

func triggerExecution(args []string) {
	if len(args) > 1 {
		execution.PrintHelp(true)
	}

	arg := args[0]

	var err error

	switch arg {
	case MODE_LIST_LONG, MODE_LIST_SHORT:
		err = execution.ListCurrentBranches()
	case MODE_HELP_LONG, MODE_HELP_SHORT:
		execution.PrintHelp(false)
	default:
		execution.PrintHelp(true)
	}

	if err != nil {
		color.Red(fmt.Sprintf("Error Retrieving Path: %v\n", err))
	}
}
