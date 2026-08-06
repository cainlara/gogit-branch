package execution

import "testing"

func TestNextTabStop(t *testing.T) {
	tests := []struct {
		name   string
		column int
		want   int
	}{
		{name: "zero", column: 0, want: 8},
		{name: "mid stop", column: 5, want: 8},
		{name: "exactly on a stop", column: 8, want: 16},
		{name: "past a stop", column: 11, want: 16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := nextTabStop(tt.column, STATUS_TAB_WIDTH); got != tt.want {
				t.Errorf("nextTabStop(%d, %d) = %d, want %d", tt.column, STATUS_TAB_WIDTH, got, tt.want)
			}
		})
	}
}

func TestTargetColumn(t *testing.T) {
	tests := []struct {
		name string
		rows []renderRow
		want int
	}{
		{name: "empty", rows: []renderRow{}, want: 0},
		{
			name: "mixed label lengths, longest is untracked",
			rows: []renderRow{
				{label: "(modified)"},
				{label: "(new)"},
				{label: "(untracked)"},
			},
			want: 16,
		},
		{
			name: "single short label",
			rows: []renderRow{
				{label: "(new)"},
			},
			want: 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := targetColumn(tt.rows, 2, STATUS_TAB_WIDTH); got != tt.want {
				t.Errorf("targetColumn(%v, 2, %d) = %d, want %d", tt.rows, STATUS_TAB_WIDTH, got, tt.want)
			}
		})
	}
}

func TestPadding(t *testing.T) {
	tests := []struct {
		name      string
		prefixLen int
		target    int
		want      string
	}{
		{name: "one tab reaches target", prefixLen: 12, target: 16, want: "\t"},
		{name: "short label needs two tabs to reach a farther target", prefixLen: 7, target: 16, want: "\t\t"},
		{name: "prefix already at target", prefixLen: 16, target: 16, want: ""},
		{name: "prefix lands exactly on a stop", prefixLen: 9, target: 16, want: "\t"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := padding(tt.prefixLen, tt.target, STATUS_TAB_WIDTH); got != tt.want {
				t.Errorf("padding(%d, %d, %d) = %q, want %q", tt.prefixLen, tt.target, STATUS_TAB_WIDTH, got, tt.want)
			}
		})
	}
}
