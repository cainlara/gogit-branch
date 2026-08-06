package execution

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cainlara/gogit-branch/core"

	"github.com/fatih/color"
)

const DEFAULT_LOG_LIMIT = 20

// validateLimit trims leading/trailing whitespace from raw, strips an
// optional leading '+', and rejects an empty/non-digit remainder or a
// non-positive parsed value — mirroring execution/create.go's
// validateBranchName pattern.
func validateLimit(raw string) (int, error) {
	trimmed := strings.TrimSpace(raw)
	trimmed = strings.TrimPrefix(trimmed, "+")

	if trimmed == "" {
		return 0, fmt.Errorf("limit %q is not a positive whole number", raw)
	}

	for _, r := range trimmed {
		if r < '0' || r > '9' {
			return 0, fmt.Errorf("limit %q is not a positive whole number", raw)
		}
	}

	value, err := strconv.Atoi(trimmed)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("limit %q is not a positive whole number", raw)
	}

	return value, nil
}

// isDefaultNoteApplicable reports whether the "defaulted to N entries" note
// applies: no explicit limit was supplied, and the default genuinely
// constrained the output (at least as many commits exist as the default
// limit). Shared between selectTrailingNote and ShowLog so the condition
// that decides the note's text and the condition that decides its color
// never drift apart.
func isDefaultNoteApplicable(actualCount, effectiveLimit int, explicit bool) bool {
	return !explicit && actualCount >= effectiveLimit
}

// selectTrailingNote decides which of three mutually-exclusive notes (if
// any) should follow the printed commit list, per data-model.md's rule:
//   - fewer commits exist than the effective limit → "Showing all logs",
//     regardless of whether that limit was explicit or defaulted
//   - the default was used and genuinely constrained the output → the
//     "defaulted to 20" note
//   - an explicit limit was used and fully satisfied → no note
func selectTrailingNote(actualCount, effectiveLimit int, explicit bool) string {
	if actualCount < effectiveLimit {
		return "Showing all logs"
	}

	if isDefaultNoteApplicable(actualCount, effectiveLimit, explicit) {
		return fmt.Sprintf("\nDefaulted to %d entries because no limit was provided", effectiveLimit)
	}

	return ""
}

// ShowLog prints recent commit history using the tool's fixed colorized
// format. When limit is nil, it defaults to DEFAULT_LOG_LIMIT and notes so;
// when non-nil, it must be a positive whole number.
func ShowLog(gitClient *core.GitClient, limit *string) error {
	effectiveLimit := DEFAULT_LOG_LIMIT
	explicit := false

	if limit != nil {
		validated, err := validateLimit(*limit)
		if err != nil {
			return err
		}

		effectiveLimit = validated
		explicit = true
	}

	entries, err := gitClient.Log(effectiveLimit)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		fmt.Println(entry)
	}

	if note := selectTrailingNote(len(entries), effectiveLimit, explicit); note != "" {
		if isDefaultNoteApplicable(len(entries), effectiveLimit, explicit) {
			color.Cyan(note)
		} else {
			fmt.Println(note)
		}
	}

	return nil
}
