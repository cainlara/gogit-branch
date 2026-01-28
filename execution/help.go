package execution

import (
	"fmt"

	"github.com/fatih/color"
)

func PrintHelp(withError bool) {
	fmt.Println()

	if withError {
		color.Red("This tool can't understand what you are trying to do.")
		fmt.Println()
	}

	color.Green("Usage:")
	fmt.Println("list, ls:\t\tList all the branches in the current working directory.")
	fmt.Println("switch, sw:\t\tList all the branches available to switch.")
	fmt.Println("delete, del:\t\tList all the branches available to delete.")
	fmt.Println("batch-delete, bd:\tList all the branches available to delete.")
	fmt.Println("help, h:\t\tPrint this help.")
	fmt.Println()
}
