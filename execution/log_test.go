package execution

import (
	"strings"
	"testing"
)

func TestValidateLimit(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    int
		wantErr bool
	}{
		{name: "empty string", raw: "", wantErr: true},
		{name: "whitespace only", raw: "   ", wantErr: true},
		{name: "non-numeric text", raw: "abc", wantErr: true},
		{name: "fractional value", raw: "1.5", wantErr: true},
		{name: "zero", raw: "0", wantErr: true},
		{name: "negative number", raw: "-3", wantErr: true},
		{name: "valid number", raw: "5", want: 5},
		{name: "leading plus accepted", raw: "+5", want: 5},
		{name: "surrounding whitespace accepted", raw: " 5 ", want: 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateLimit(tt.raw)

			if tt.wantErr {
				if err == nil {
					t.Errorf("validateLimit(%q) = %d, nil; want an error", tt.raw, got)
				}

				return
			}

			if err != nil {
				t.Errorf("validateLimit(%q) returned unexpected error: %v", tt.raw, err)
			}

			if got != tt.want {
				t.Errorf("validateLimit(%q) = %d, want %d", tt.raw, got, tt.want)
			}
		})
	}
}

func TestSelectTrailingNote(t *testing.T) {
	tests := []struct {
		name           string
		actualCount    int
		effectiveLimit int
		explicit       bool
		wantEmpty      bool
		wantContains   string
	}{
		{name: "fewer available than default limit", actualCount: 5, effectiveLimit: 20, explicit: false, wantContains: "Showing all logs"},
		{name: "fewer available than explicit limit", actualCount: 5, effectiveLimit: 1000, explicit: true, wantContains: "Showing all logs"},
		{name: "default limit satisfied", actualCount: 20, effectiveLimit: 20, explicit: false, wantContains: "Defaulted to 20"},
		{name: "explicit limit satisfied", actualCount: 5, effectiveLimit: 5, explicit: true, wantEmpty: true},
		{name: "boundary: actualCount equals effectiveLimit, explicit", actualCount: 5, effectiveLimit: 5, explicit: true, wantEmpty: true},
		{name: "boundary: actualCount equals effectiveLimit, default", actualCount: 20, effectiveLimit: 20, explicit: false, wantContains: "Defaulted to 20"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := selectTrailingNote(tt.actualCount, tt.effectiveLimit, tt.explicit)

			if tt.wantEmpty {
				if got != "" {
					t.Errorf("selectTrailingNote(%d, %d, %v) = %q, want empty", tt.actualCount, tt.effectiveLimit, tt.explicit, got)
				}

				return
			}

			if got == "" || !strings.Contains(got, tt.wantContains) {
				t.Errorf("selectTrailingNote(%d, %d, %v) = %q, want to contain %q", tt.actualCount, tt.effectiveLimit, tt.explicit, got, tt.wantContains)
			}
		})
	}
}
