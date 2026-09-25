package execution

import (
	"strings"
	"testing"
	"text/template"

	"github.com/cainlara/gogit-branch/model"

	"github.com/manifoldco/promptui"
)

const (
	ansiCyan    = "\x1b[36m"
	ansiMagenta = "\x1b[35m"
	ansiGreen   = "\x1b[32m"
	ansiReset   = "\x1b[0m"
)

func renderTemplateString(t *testing.T, tplText string, data model.Branch) string {
	t.Helper()

	tpl, err := template.New("").Funcs(promptui.FuncMap).Parse(tplText)
	if err != nil {
		t.Fatalf("parse template %q: %v", tplText, err)
	}

	var buf strings.Builder
	if err := tpl.Execute(&buf, data); err != nil {
		t.Fatalf("execute template for %q: %v", data.GetName(), err)
	}

	return buf.String()
}

func stripANSI(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			j := i + 2
			for j < len(s) && s[j] != 'm' {
				j++
			}
			if j < len(s) {
				i = j + 1
				continue
			}
		}
		b.WriteByte(s[i])
		i++
	}

	return b.String()
}

// The delete (and every other non-switch) flow must keep rendering exactly the
// historical templates — FR-001 scope isolation, research D6, contract §2 R3.
func TestTargetSelectTemplatesFalseVariantByteIdentical(t *testing.T) {
	templates := targetSelectTemplates(false, EMOJI_SKULL)

	wantActive := EMOJI_SKULL + " {{ .GetName | cyan }} {{if .IsDummyBranch}} Pick To Abort {{else}}({{ .GetFullHash | red }}){{end}}"
	wantInactive := "  {{ .GetName | cyan }} {{if .IsDummyBranch}}{{else}}({{ .GetShortHash | red }}){{end}}"
	wantSelected := "{{if .IsDummyBranch}}Operation Cancelled {{else}}" + EMOJI_SKULL + " {{ .GetName | green}} Selected{{end}}"

	if templates.Active != wantActive {
		t.Errorf("Active = %q, want %q", templates.Active, wantActive)
	}
	if templates.Inactive != wantInactive {
		t.Errorf("Inactive = %q, want %q", templates.Inactive, wantInactive)
	}
	if templates.Selected != wantSelected {
		t.Errorf("Selected = %q, want %q", templates.Selected, wantSelected)
	}
	if strings.Contains(templates.Active, "magenta") || strings.Contains(templates.Active, "l {{") {
		t.Errorf("false variant must not carry marker/color markup: %q", templates.Active)
	}
}

func TestTargetSelectTemplatesMarkerVariant(t *testing.T) {
	templates := targetSelectTemplates(true, EMOJI_HERB)

	hash40 := "1111111111111111111111111111111111111111"

	local := *model.NewBranch("local-only", "7dc4bec", hash40, false)
	remoteUnique := *model.NewRemoteBranch("feature-b", "origin/feature-b", "feature-b", "f86d618", hash40)
	remoteQualified := *model.NewRemoteBranch("feature-x", "origin/feature-x", "origin/feature-x", "1b3a089", hash40)
	dummy := *model.NewDummyBranch("Cancel Switch")

	t.Run("active local row carries l marker and cyan name", func(t *testing.T) {
		got := renderTemplateString(t, templates.Active, local)

		if !strings.Contains(got, "🌿 l ") || !strings.Contains(got, ansiCyan+"local-only"+ansiReset) {
			t.Errorf("active local = %q, want herb + 'l ' + cyan local-only", got)
		}
		if strings.Contains(got, "r local-only") {
			t.Errorf("local row must not carry r marker: %q", got)
		}
	})

	t.Run("active remote rows carry r marker and magenta name", func(t *testing.T) {
		gotUnique := renderTemplateString(t, templates.Active, remoteUnique)
		if !strings.Contains(gotUnique, "🌿 r ") || !strings.Contains(gotUnique, ansiMagenta+"feature-b"+ansiReset) {
			t.Errorf("active remote unique = %q, want herb + 'r ' + magenta feature-b", gotUnique)
		}

		gotQualified := renderTemplateString(t, templates.Active, remoteQualified)
		if !strings.Contains(gotQualified, ansiMagenta+"origin/feature-x"+ansiReset) {
			t.Errorf("active remote qualified = %q, want magenta origin/feature-x", gotQualified)
		}
	})

	t.Run("inactive rows keep marker outside the color span (text fallback)", func(t *testing.T) {
		gotRemote := renderTemplateString(t, templates.Inactive, remoteQualified)
		if !strings.Contains(gotRemote, "  r ") {
			t.Errorf("inactive remote = %q, want two-space + 'r ' prefix", gotRemote)
		}
		if !strings.Contains(gotRemote, ansiMagenta+"origin/feature-x"+ansiReset) {
			t.Errorf("inactive remote name not magenta: %q", gotRemote)
		}

		stripped := stripANSI(gotRemote)
		if !strings.Contains(stripped, "r origin/feature-x") {
			t.Errorf("color-stripped inactive remote lost its marker/name: %q", stripped)
		}

		gotLocal := renderTemplateString(t, templates.Inactive, local)
		if !strings.Contains(gotLocal, "  l "+ansiCyan+"local-only"+ansiReset) {
			t.Errorf("inactive local = %q, want 'l ' + cyan local-only", gotLocal)
		}
	})

	t.Run("dummy carries no marker in any row", func(t *testing.T) {
		active := renderTemplateString(t, templates.Active, dummy)
		inactive := renderTemplateString(t, templates.Inactive, dummy)

		if !strings.Contains(active, EMOJI_HERB+" "+ansiCyan+"Cancel Switch"+ansiReset+"  Pick To Abort") {
			t.Errorf("active dummy = %q, want herb + cyan Cancel Switch + Pick To Abort", active)
		}
		if !strings.Contains(inactive, "  "+ansiCyan+"Cancel Switch"+ansiReset) {
			t.Errorf("inactive dummy = %q, want two-space + cyan name", inactive)
		}

		for name, got := range map[string]string{"active": active, "inactive": inactive} {
			if strings.Contains(got, " l ") || strings.Contains(got, " r ") {
				t.Errorf("dummy %s row must carry no origin marker: %q", name, got)
			}
		}
	})

	t.Run("selected echo shows marker + display name for every pick", func(t *testing.T) {
		gotRemote := renderTemplateString(t, templates.Selected, remoteQualified)
		wantRemote := EMOJI_HERB + " r " + ansiGreen + "origin/feature-x" + ansiReset + " Selected"
		if gotRemote != wantRemote {
			t.Errorf("selected remote = %q, want %q", gotRemote, wantRemote)
		}

		gotLocal := renderTemplateString(t, templates.Selected, local)
		wantLocal := EMOJI_HERB + " l " + ansiGreen + "local-only" + ansiReset + " Selected"
		if gotLocal != wantLocal {
			t.Errorf("selected local = %q, want %q", gotLocal, wantLocal)
		}

		gotDummy := renderTemplateString(t, templates.Selected, dummy)
		if gotDummy != "Operation Cancelled " {
			t.Errorf("selected dummy = %q, want %q", gotDummy, "Operation Cancelled ")
		}
	})

	t.Run("colors are distinct between local and remote rows", func(t *testing.T) {
		gotLocal := renderTemplateString(t, templates.Active, local)
		gotRemote := renderTemplateString(t, templates.Active, remoteUnique)

		if !strings.Contains(gotLocal, ansiCyan) || strings.Contains(gotLocal, ansiMagenta) {
			t.Errorf("local row must be cyan, never magenta: %q", gotLocal)
		}
		if !strings.Contains(gotRemote, ansiMagenta) || strings.Contains(gotRemote, ansiCyan) {
			t.Errorf("remote row must be magenta, never cyan: %q", gotRemote)
		}
	})
}
