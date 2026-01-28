package version

import (
	"fmt"
	"runtime/debug"
	"strings"
)

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
	Dirty   = "false"
)

func Short() string {
	return fmt.Sprintf("%s %s", appName(), Version)
}

func Long() string {
	parts := []string{fmt.Sprintf("%s %s", appName(), Version)}
	if Commit != "none" {
		parts = append(parts, "commit "+Commit)
	}
	if Date != "unknown" {
		parts = append(parts, "built "+Date)
	}
	if Dirty == "true" {
		parts = append(parts, "dirty")
	}

	if bi, ok := debug.ReadBuildInfo(); ok && bi != nil {
		if vcs := vcsSummary(bi); vcs != "" {
			parts = append(parts, vcs)
		}
	}

	return strings.Join(parts, ", ")
}

func appName() string { return "gogit-branch" }

func vcsSummary(bi *debug.BuildInfo) string {
	var rev, t, mod string
	for _, s := range bi.Settings {
		switch s.Key {
		case "vcs.revision":
			rev = s.Value
		case "vcs.time":
			t = s.Value
		case "vcs.modified":
			mod = s.Value
		}
	}
	if rev == "" && t == "" && mod == "" {
		return ""
	}
	out := []string{}
	if rev != "" {
		if len(rev) > 7 {
			rev = rev[:7]
		}
		out = append(out, "vcs "+rev)
	}
	if t != "" {
		out = append(out, "vcs-time "+t)
	}
	if mod == "true" {
		out = append(out, "vcs-modified")
	}
	return strings.Join(out, " ")
}
