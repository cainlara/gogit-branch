package execution

import (
	"strings"
	"testing"
)

func TestValidateBranchName(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantName      string
		wantErr       bool
		wantErrSubstr string
	}{
		{
			name:          "empty string",
			input:         "",
			wantErr:       true,
			wantErrSubstr: "empty",
		},
		{
			name:          "whitespace only",
			input:         "   ",
			wantErr:       true,
			wantErrSubstr: "empty",
		},
		{
			name:          "single dash led",
			input:         "-x",
			wantErr:       true,
			wantErrSubstr: "-x",
		},
		{
			name:          "double dash led",
			input:         "--foo",
			wantErr:       true,
			wantErrSubstr: "--foo",
		},
		{
			name:     "valid name",
			input:    "feature/my-branch",
			wantName: "feature/my-branch",
		},
		{
			name:     "valid name with surrounding whitespace",
			input:    "  feature/my-branch  ",
			wantName: "feature/my-branch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateBranchName(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("validateBranchName(%q) = nil error, want error", tt.input)
				}

				if !strings.Contains(err.Error(), tt.wantErrSubstr) {
					t.Fatalf("validateBranchName(%q) error = %q, want it to contain %q", tt.input, err.Error(), tt.wantErrSubstr)
				}

				return
			}

			if err != nil {
				t.Fatalf("validateBranchName(%q) unexpected error: %v", tt.input, err)
			}

			if got != tt.wantName {
				t.Fatalf("validateBranchName(%q) = %q, want %q", tt.input, got, tt.wantName)
			}
		})
	}
}
