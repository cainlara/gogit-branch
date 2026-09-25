package execution

import (
	"bytes"
	"strings"
	"testing"
)

func TestDisabledRendererWritesZeroBytes(t *testing.T) {
	// A disabled renderer is what the production gate (FR-009, contract S1/S3)
	// yields when the status stream is not an interactive terminal: every
	// operation must be a strict no-op on every stream (SC-004).
	var buf bytes.Buffer

	r := &progressRenderer{out: &buf, enabled: false}

	r.start("Fetching remote updates")
	r.setStage("Pulling remote updates")
	r.onProgress(42, true)
	r.finish()

	if buf.Len() != 0 {
		t.Errorf("disabled renderer wrote %d bytes (%q), want 0 (SC-004)", buf.Len(), buf.String())
	}
}

func TestEnabledRendererWritesOnlyToItsOwnWriter(t *testing.T) {
	var buf bytes.Buffer

	r := newProgressRenderer(&buf)

	r.start("Fetching remote updates")
	r.onProgress(42, true)

	// Percent updates coalesce to the next tick by design (research D7) —
	// advance one frame deterministically instead of sleeping.
	r.mu.Lock()
	r.tickLocked()
	r.mu.Unlock()

	r.finish()

	out := buf.String()

	if !strings.Contains(out, "Fetching remote updates") {
		t.Errorf("enabled renderer output %q missing stage label", out)
	}

	if !strings.Contains(out, "42%") {
		t.Errorf("enabled renderer output %q missing reported percent", out)
	}

	// finish must leave only CR/spaces after the final frame — no bar text.
	lastCR := strings.LastIndex(out, "\r")
	if lastCR >= 0 && strings.ContainsAny(out[lastCR+1:], "[]%") {
		t.Errorf("trailing segment %q contains bar characters after final clear", out[lastCR+1:])
	}
}

func TestComposeProgressFramePercentForm(t *testing.T) {
	tests := []struct {
		name    string
		percent int
		want    string
	}{
		{
			name:    "mid-transfer percent renders bar and number",
			percent: 42,
			want:    "42%",
		},
		{
			name:    "zero percent renders empty bar with cursor",
			percent: 0,
			want:    "0%",
		},
		{
			name:    "reported 100 is capped at 99 while stage active",
			percent: 100,
			want:    "99%",
		},
		{
			name:    "oversized value capped at 99",
			percent: 1000,
			want:    "99%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frame := composeProgressFrame("Fetching remote updates", tt.percent, true, 0)

			if !strings.Contains(frame, tt.want) {
				t.Errorf("frame %q does not contain %q", frame, tt.want)
			}

			if !strings.Contains(frame, "Fetching remote updates [") {
				t.Errorf("frame %q missing stage label and bar", frame)
			}

			if strings.Contains(frame, "100%") {
				t.Errorf("frame %q shows a completed state while stage is active (B7/FR-008)", frame)
			}
		})
	}
}

func TestComposeProgressFrameAnimatedForm(t *testing.T) {
	frame := composeProgressFrame("Pulling remote updates", 0, false, 0)

	if strings.Contains(frame, "%") {
		t.Errorf("animated frame %q must not contain any numeric percent (SC-008/B2)", frame)
	}

	if strings.Contains(frame, "[") {
		t.Errorf("animated frame %q must not contain a bar", frame)
	}

	if !strings.Contains(frame, "Pulling remote updates") {
		t.Errorf("animated frame %q missing stage label", frame)
	}
}

func TestComposeProgressFrameSpinnerAdvances(t *testing.T) {
	first := composeProgressFrame("Fetching remote updates", 0, false, 0)
	second := composeProgressFrame("Fetching remote updates", 0, false, 1)

	if first == second {
		t.Errorf("consecutive frames identical (%q) — no visible change for SC-005", first)
	}
}

func TestClearSequence(t *testing.T) {
	if got := clearSequence(0); got != "" {
		t.Errorf("clearSequence(0) = %q, want empty (nothing rendered)", got)
	}

	got := clearSequence(25)

	if got != "\r"+strings.Repeat(" ", 25)+"\r" {
		t.Errorf("clearSequence(25) = %q, want CR + 25 spaces + CR", got)
	}

	for _, r := range got {
		if r != '\r' && r != ' ' {
			t.Errorf("clearSequence contains non-clear rune %q", r)
		}
	}
}
