package execution

import "testing"

func TestSummarizeBranchNamesForConfirm(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  string
	}{
		{name: "zero names", input: []string{}, want: ""},
		{name: "one name", input: []string{"a"}, want: "a"},
		{name: "exactly five names", input: []string{"a", "b", "c", "d", "e"}, want: "a, b, c, d, e"},
		{
			name:  "six names",
			input: []string{"a", "b", "c", "d", "e", "f"},
			want:  "a, b, c, d, e, and 1 more",
		},
		{
			name:  "ten names",
			input: []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"},
			want:  "a, b, c, d, e, and 5 more",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := summarizeBranchNamesForConfirm(tt.input)

			if got != tt.want {
				t.Fatalf("summarizeBranchNamesForConfirm(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
