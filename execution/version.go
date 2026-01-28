package execution

import (
	"fmt"

	"github.com/cainlara/gogit-branch/version"
)

func ShowVersion() {
	fmt.Printf("Version: %s\n", version.Long())
	fmt.Printf("Built: %s\n", version.Date)
}
