package execution

import (
	"fmt"

	"github.com/cainlara/gogit-branch/version"
	"github.com/fatih/color"
)

func PrintHelp(withError, withUsage bool) {
	fmt.Println()

	if withError {
		color.Red("This humble tool can't understand what you are trying to do.")
		fmt.Println("Use 'help' (or 'h') argument to see the available commands.")
	} else {
		fmt.Printf("%s%s\n", BANNER, version.Short())
		fmt.Println("Built: ", version.Date)

		if withUsage {
			fmt.Println()
			color.Green("Usage:")
			fmt.Println("batch-delete, bd:\tList all the available branches, allowing you to select multiple branches to delete.")
			fmt.Println("create, c:\t\tCreate a new branch and switch to it.")
			fmt.Println("delete, del:\t\tList all the branches available to delete.")
			fmt.Println("help, h:\t\tShow this help.")
			fmt.Println("list, ls:\t\tList all the branches in the current working directory.")
			fmt.Println("log, l:\t\tShow recent commit history, defaulting to the last 20 entries.")
			fmt.Println("push, p:\t\tPush the current branch, setting the upstream first if needed.")
			fmt.Println("reset, r:\t\tDANGER: irreversibly discard all uncommitted changes (add --hard to also remove untracked files).")
			fmt.Println("status, st:\t\tShow the repository status and commit tracked/untracked changes.")
			fmt.Println("switch, sw:\t\tList all the branches available to switch.")
			fmt.Println("version, v:\t\tShow the version of this humble tool.")
		}
	}

	fmt.Println()
}
