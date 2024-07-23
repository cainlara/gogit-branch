package execution

import (
	"cainlara/gogit-branch/model"
	"fmt"

	"github.com/manifoldco/promptui"
)

const (
	EMOJI_HERB  = "\U0001F33F"
	EMOJI_SKULL = "\U0001F571"
)

func listBranchesAndSelectTarget(options []model.Branch, icon string) (model.Branch, error) {
	activeCopy := fmt.Sprintf("%s {{ .GetShortName | cyan }} ({{ .GetFullHash | red }})", icon)
	selectedCopy := fmt.Sprintf("%s {{ .GetRefName | green}} Selected", icon)

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   activeCopy,
		Inactive: "  {{ .GetShortName | cyan }} ({{ .GetShortHash | red }})",
		Selected: selectedCopy,
	}

	prompt := promptui.Select{
		Label:     "Select Target Branch",
		Items:     options,
		Templates: templates,
	}

	i, _, err := prompt.Run()
	if err != nil {
		return model.Branch{}, err
	}

	selectedBranch := options[i]

	return selectedBranch, nil
}
