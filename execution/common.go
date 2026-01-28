package execution

import (
	"fmt"

	"github.com/cainlara/gogit-branch/model"

	"github.com/manifoldco/promptui"
)

const (
	EMOJI_HERB  = "\U0001F33F"
	EMOJI_SKULL = "\U0001F480"
)

func listBranchesAndSelectTarget(options []model.Branch, icon string) (model.Branch, error) {
	activeCopy := fmt.Sprintf("%s {{ .GetName | cyan }} {{if .IsDummyBranch}} Pick To Abort {{else}}({{ .GetFullHash | red }}){{end}}", icon)
	inactiveCopy := "  {{ .GetName | cyan }} {{if .IsDummyBranch}}{{else}}({{ .GetShortHash | red }}){{end}}"
	selectedCopy := fmt.Sprintf("{{if .IsDummyBranch}}Operation Cancelled {{else}}%s {{ .GetName | green}} Selected{{end}}", icon)

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   activeCopy,
		Inactive: inactiveCopy,
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

func listBranchesAndSelectMultipleTargets(selectedPos int, options []model.Branch, icon string) ([]model.Branch, error) {
	activeCopy := fmt.Sprintf("%s [{{if .IsSelected}}x{{else}} {{end}}] {{ .GetName | cyan }} ({{if .IsDone}}Pick to finish selection{{else}}{{ .GetFullHash | red }}{{end}})", icon)
	inactiveCopy := " [{{if .IsSelected}}x{{else}} {{end}}] {{ .GetName | cyan }} {{if .IsDummyBranch}}{{else}}({{ .GetShortHash | red }}){{end}}"

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   activeCopy,
		Inactive: inactiveCopy,
	}

	prompt := promptui.Select{
		Label:        "Select Target Branches",
		Items:        options,
		Templates:    templates,
		Size:         5,
		CursorPos:    selectedPos,
		HideSelected: true,
	}

	selectionIdx, _, err := prompt.Run()
	if err != nil {
		return nil, err
	}

	if !options[selectionIdx].IsDone() {
		options[selectionIdx].SetSelected(!options[selectionIdx].IsSelected())

		return listBranchesAndSelectMultipleTargets(selectionIdx, options, icon)
	}

	var selectedBranches []model.Branch
	for _, branch := range options {
		if branch.IsSelected() {
			selectedBranches = append(selectedBranches, branch)
		}
	}

	return selectedBranches, nil
}
