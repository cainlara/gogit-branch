package execution

import (
	"strings"
	"testing"
)

func TestParseCloneArgs(t *testing.T) {
	cases := []struct {
		name      string
		args      []string
		wantErr   bool
		errSubstr string
		wantURL   string
		wantName  string
		wantMail  string
		wantAnon  bool
	}{
		{name: "no args", args: nil, wantErr: true, errSubstr: "repository URL required"},
		{name: "flag first", args: []string{"-anon"}, wantErr: true, errSubstr: "repository URL required"},
		{name: "empty url", args: []string{""}, wantErr: true, errSubstr: "repository URL required"},
		{name: "url only", args: []string{"https://h/r.git"}, wantURL: "https://h/r.git"},
		{name: "url plus anon", args: []string{"https://h/r.git", "-anon"}, wantURL: "https://h/r.git", wantAnon: true},
		{name: "url plus pair", args: []string{"https://h/r.git", "Jane Doe", "j@e.x"},
			wantURL: "https://h/r.git", wantName: "Jane Doe", wantMail: "j@e.x"},
		{name: "url pair plus anon", args: []string{"u", "n", "e", "-anon"}, wantErr: true,
			errSubstr: "conflicting arguments"},
		{name: "lone name", args: []string{"u", "OnlyName"}, wantErr: true, errSubstr: "provided together"},
		{name: "anon plus lone name", args: []string{"u", "-anon", "n"}, wantErr: true,
			errSubstr: "conflicting arguments"},
		{name: "surplus token", args: []string{"u", "n", "e", "extra"}, wantErr: true,
			errSubstr: "unexpected arguments"},
		{name: "empty name value", args: []string{"u", "", "e"}, wantErr: true, errSubstr: "non-empty"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := parseCloneArgs(c.args)

			if c.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (parsed=%+v)", got)
				}

				if c.errSubstr != "" && !strings.Contains(err.Error(), c.errSubstr) {
					t.Errorf("error %q does not contain %q", err.Error(), c.errSubstr)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.URL != c.wantURL || got.Name != c.wantName || got.Email != c.wantMail || got.Anon != c.wantAnon {
				t.Errorf("parseCloneArgs(%q) = %+v, want URL=%q Name=%q Email=%q Anon=%v",
					c.args, got, c.wantURL, c.wantName, c.wantMail, c.wantAnon)
			}
		})
	}
}
