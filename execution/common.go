package execution

import (
	"fmt"
	"strings"

	"github.com/cainlara/gogit-branch/model"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
)

const (
	EMOJI_HERB   = "\U0001F33F"
	EMOJI_SKULL  = "\U0001F480"
	EMOJI_ROCKET = "\U0001F680"
	BANNER       = `                             ███   █████   
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
`
)

// isAffirmativeAnswer reports whether raw, normalized case-insensitively and
// trimmed of incidental surrounding whitespace, is exactly "y" or "yes"
// (FR-002, FR-003). Anything else is a decline (FR-004) — both outcomes are
// two branches of this same single comparison, not independent checks.
func isAffirmativeAnswer(raw string) bool {
	normalized := strings.ToLower(strings.TrimSpace(raw))

	return normalized == "y" || normalized == "yes"
}

// confirmYesNo runs a freeform (non-IsConfirm) prompt and decides
// accept/decline itself via isAffirmativeAnswer, so the rendered outcome and
// the actual decision can never disagree the way promptui's own IsConfirm
// mode does (its built-in check only ever treats an exact "y" as accepted,
// silently rendering "yes" as rejected even though this tool treats it as
// accepted too). On acceptance, an explicit confirmation line is printed;
// on decline, nothing is printed here — each caller keeps its own distinct
// abort message (FR-006).
func confirmYesNo(label string) bool {
	prompt := promptui.Prompt{
		Label: label,
	}

	result, _ := prompt.Run()

	accepted := isAffirmativeAnswer(result)
	if accepted {
		color.Green("\nConfirmed")
	}

	return accepted
}

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
