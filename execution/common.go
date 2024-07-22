package execution

import (
	"cainlara/gogit-branch/model"

	"github.com/manifoldco/promptui"
)

func listBranchesAndSelectTarget(options []model.Branch) (model.Branch, error) {
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U0001F33F {{ .GetShortName | cyan }} ({{ .GetFullHash | red }})",
		Inactive: "  {{ .GetShortName | cyan }} ({{ .GetShortHash | red }})",
		// Selected: "Switching to \U0001F33F {{ .GetRefName | green}}",
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
