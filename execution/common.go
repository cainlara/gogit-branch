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

// listBranchesAndSelectTarget renders the single-select prompt shared by the
// switch and delete flows. showOriginMarkers selects the template variant:
//
//   - false (delete): byte-for-byte the historical templates — names cyan,
//     hashes red, no origin markers — so delete's rendering stays untouched
//     (FR-001, research D6).
//   - true (switch): rows carry a plain-text "l "/"r " marker (local /
//     remote-only) that survives color-disabled environments, remote rows
//     render their GetDisplayName() (bare or qualified, FR-011) in magenta —
//     visibly distinct from the cyan locals (FR-002's both-cues rule), the
//     Cancel dummy carries no marker, and the Selected echo prints the same
//     marker + display name so the confirmation identifies what was picked
//     (FR-002, contract §2).
func listBranchesAndSelectTarget(options []model.Branch, icon string, showOriginMarkers bool) (model.Branch, error) {
	templates := targetSelectTemplates(showOriginMarkers, icon)

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

// targetSelectTemplates builds the promptui templates for the single-select
// prompt (extracted so the rendering contract is unit-testable without a TTY,
// research D10).
func targetSelectTemplates(showOriginMarkers bool, icon string) *promptui.SelectTemplates {
	var activeCopy, inactiveCopy, selectedCopy string

	if showOriginMarkers {
		activeCopy = fmt.Sprintf(`%s {{if .IsDummyBranch}}{{ .GetName | cyan }}  Pick To Abort {{else}}{{if .IsRemoteBranch}}r {{ .GetDisplayName | magenta }}{{else}}l {{ .GetDisplayName | cyan }}{{end}} ({{ .GetFullHash | red }}){{end}}`, icon)
		inactiveCopy = `  {{if .IsDummyBranch}}{{ .GetName | cyan }}{{else}}{{if .IsRemoteBranch}}r {{ .GetDisplayName | magenta }}{{else}}l {{ .GetDisplayName | cyan }}{{end}} ({{ .GetShortHash | red }}){{end}}`
		selectedCopy = fmt.Sprintf(`{{if .IsDummyBranch}}Operation Cancelled {{else}}%s {{if .IsRemoteBranch}}r{{else}}l{{end}} {{ .GetDisplayName | green}} Selected{{end}}`, icon)
	} else {
		activeCopy = fmt.Sprintf("%s {{ .GetName | cyan }} {{if .IsDummyBranch}} Pick To Abort {{else}}({{ .GetFullHash | red }}){{end}}", icon)
		inactiveCopy = "  {{ .GetName | cyan }} {{if .IsDummyBranch}}{{else}}({{ .GetShortHash | red }}){{end}}"
		selectedCopy = fmt.Sprintf("{{if .IsDummyBranch}}Operation Cancelled {{else}}%s {{ .GetName | green}} Selected{{end}}", icon)
	}

	return &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   activeCopy,
		Inactive: inactiveCopy,
		Selected: selectedCopy,
	}
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
