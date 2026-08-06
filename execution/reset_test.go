package execution

import "testing"

func TestValidateResetFlag(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    bool
		wantErr bool
	}{
		{name: "exactly --hard", raw: "--hard", want: true},
		{name: "different flag", raw: "--force", wantErr: true},
		{name: "single dash variant", raw: "-hard", wantErr: true},
		{name: "wrong casing", raw: "--Hard", wantErr: true},
		{name: "arbitrary text", raw: "banana", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateResetFlag(tt.raw)

			if tt.wantErr {
				if err == nil {
					t.Errorf("validateResetFlag(%q) = %v, nil; want an error", tt.raw, got)
				}

				return
			}

			if err != nil {
				t.Errorf("validateResetFlag(%q) returned unexpected error: %v", tt.raw, err)
			}

			if got != tt.want {
				t.Errorf("validateResetFlag(%q) = %v, want %v", tt.raw, got, tt.want)
			}
		})
	}
}
