package execution

import (
	"cainlara/gogit-branch/core"
	"fmt"

	"github.com/fatih/color"
)

func ListCurrentBranches() {
	fmt.Println()
	color.Green("Listing branches")
	core.GetBranches()
}
