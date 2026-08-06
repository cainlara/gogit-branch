package execution

import "testing"

func TestIsAffirmativeAnswer(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "lowercase y", input: "y", want: true},
		{name: "lowercase yes", input: "yes", want: true},
		{name: "uppercase Y", input: "Y", want: true},
		{name: "mixed case Yes", input: "Yes", want: true},
		{name: "uppercase YES", input: "YES", want: true},
		{name: "surrounding whitespace", input: " yes ", want: true},
		{name: "decline n", input: "n", want: false},
		{name: "decline no", input: "no", want: false},
		{name: "empty input", input: "", want: false},
		{name: "near-miss yy", input: "yy", want: false},
		{name: "near-miss ye", input: "ye", want: false},
		{name: "near-miss yess", input: "yess", want: false},
		{name: "arbitrary text", input: "whatever", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isAffirmativeAnswer(tt.input)

			if got != tt.want {
				t.Fatalf("isAffirmativeAnswer(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
