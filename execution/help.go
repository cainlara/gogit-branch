package execution

import (
	"fmt"

	"github.com/fatih/color"
)

const banner = `                             ███   █████   
                            ░░░   ░░███    
  ███████  ██████   ███████ ████  ███████  
 ███░░███ ███░░███ ███░░███░░███ ░░░███░   
░███ ░███░███ ░███░███ ░███ ░███   ░███    
░███ ░███░███ ░███░███ ░███ ░███   ░███ ███
░░███████░░██████ ░░███████ █████  ░░█████ 
 ░░░░░███ ░░░░░░   ░░░░░███░░░░░    ░░░░░  
 ███ ░███          ███ ░███                 
░░██████          ░░██████                 
 ░░░░░░            ░░░░░░                  
 v0.2.0`

func PrintHelp(withError bool) {
	fmt.Println()

	if withError {
		color.Red("This humble tool can't understand what you are trying to do.")
		fmt.Println("Use 'help' (or 'h') argument to see the available commands.")
	} else {
		fmt.Println(banner)
		fmt.Println()
		color.Green("Usage:")
		fmt.Println("list, ls:\t\tList all the branches in the current working directory.")
		fmt.Println("switch, sw:\t\tList all the branches available to switch.")
		fmt.Println("delete, del:\t\tList all the branches available to delete.")
		fmt.Println("batch-delete, bd:\tList all the available branches, allowing you to select multiple branches to delete.")
		fmt.Println("help, h:\t\tPrint this help.")
	}

	fmt.Println()
}
