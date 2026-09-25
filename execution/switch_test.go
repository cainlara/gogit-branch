package execution

import (
	"strings"
	"testing"
)

// TestRefreshWarningLine locks the FR-006 tool line (contract §2 W1/W3,
// data-model §4): a single fixed sentence, verbatim and stable across calls,
// free of git's multi-line detail (the caller appends outputError's text
// itself) and free of main.go's fatal "Operation Failed:" phrasing, so a
// continued run can never look like an aborted one.
func TestRefreshWarningLine(t *testing.T) {
	t.Run("fixed single-line message returned verbatim", func(t *testing.T) {
		want := "⚠ Refresh failed, showing last known branches"
		if got := refreshWarningLine(); got != want {
			t.Fatalf("refreshWarningLine() = %q, want %q", got, want)
		}
	})

	t.Run("stable across calls and never multi-line", func(t *testing.T) {
		first := refreshWarningLine()
		for i := 0; i < 3; i++ {
			if got := refreshWarningLine(); got != first {
				t.Fatalf("call %d returned %q, want stable %q", i, got, first)
			}
		}
		if strings.ContainsAny(first, "\r\n") {
			t.Fatalf("warning line must be single-line, got %q", first)
		}
	})

	t.Run("no Operation Failed fatal phrasing", func(t *testing.T) {
		if strings.Contains(refreshWarningLine(), "Operation Failed") {
			t.Fatalf("warning line must not use the fatal style, got %q", refreshWarningLine())
		}
	})

	t.Run("independent of git detail markers", func(t *testing.T) {
		// The helper must not embed git's output — the caller prints
		// outputError's text separately, exactly once.
		for _, marker := range []string{"fatal:", "error:", "exit status", "git "} {
			if strings.Contains(strings.ToLower(refreshWarningLine()), marker) {
				t.Fatalf("warning line must not contain git detail marker %q, got %q", marker, refreshWarningLine())
			}
		}
	})
}
