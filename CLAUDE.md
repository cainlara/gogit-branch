# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project overview

GoGit Branch Manager (`gogit-branch`) is a small Go CLI that wraps common Git branch operations (list, switch, delete, batch-delete) behind an interactive terminal UI. It shells out to the local `git` binary rather than using a Git library.

## Commands

```bash
# Build a plain binary
go build -o gogit-branch .

# Build with version metadata baked in (mirrors CI/release builds)
./scripts/build.sh [output-name]      # defaults to output name "gogit-branch"

# Verify the current commit is on an exact tag before a release build
./scripts/release_check.sh

# Run the tool locally without building
go run . <command>

# Format / vet
go fmt ./...
go vet ./...
```

There is no test suite in this repository (`go test ./...` finds no tests).

Version info (`version.Version`, `version.Commit`, `version.Date`, `version.Dirty`) is injected at build time via `-ldflags -X` in `scripts/build.sh`; running via `go build`/`go run` directly leaves these at their `"dev"`/`"none"`/`"unknown"` defaults.

## Architecture

The code is organized into four packages with a strict dependency direction: `main` → `execution` → `core`/`model`. There is intentionally no dependency the other way (e.g. `core` never imports `execution`).

- **`main.go`** — parses `os.Args`, maps the single subcommand (`list`/`ls`, `switch`/`sw`, `delete`/`del`, `batch-delete`/`bd`, `create`/`c`, `version`/`v`, `help`/`h`) to an `execution` function, and constructs one `core.GitClient`. Errors returned from `execution` functions are printed here (centralized error handling — `execution` functions return `error` rather than printing failures themselves, except for the two "delete" flows which print an abort message directly on user cancellation).
- **`core/git_client.go`** — `GitClient` is the only place that shells out to `git` (via `os/exec`). It auto-detects the repo root with `git rev-parse --show-toplevel` when constructed with an empty path (`core.NewGitClient("")`), and exposes `Branches`, `Checkout`, `DeleteBranch`, `DeleteBranches`. All git-command output parsing (e.g. stripping the `* ` current-branch prefix from `git branch` output, detecting `error:`-prefixed stderr) lives here.
- **`model/branch.go`** — `Branch` is a plain data holder with getters/setters (no exported fields) representing a branch's name/hashes plus UI-only state (`isSelected`, `isDone`, `isDummy` for the "Cancel" sentinel entries used in select prompts). `NewDummyBranch`/`NewDoneBranch` construct sentinel entries injected into the list passed to `promptui`, not real branches from git.
- **`execution/*.go`** — one file per subcommand, each exposing a single entry point called from `main.go` (`ListCurrentBranches`, `BrowseAndSwitchBranches`, `ListAndDeleteBranch`, `ListAndDeleteBranches`, `CreateAndSwitchBranch`, `PrintHelp`, `ShowVersion`). `common.go` holds the shared `promptui.Select` wiring used by both single-select (switch/delete) and multi-select (batch-delete) flows, plus the emoji/banner constants. `create.go`'s `CreateAndSwitchBranch` is the exception to the `Select`-based UX: it takes a new branch name via a freeform `promptui.Prompt` instead, since there is nothing existing to select from — unless the caller already supplied the name as a CLI argument (`create <name>` / `c <name>`), in which case `main.go` passes it through as a non-nil `*string` and the prompt is skipped entirely. A supplied name is trimmed and validated (rejecting empty or dash-led values) in `create.go`'s `validateBranchName`; the interactive-prompt path is intentionally left unvalidated/unchanged.
- **`version/version.go`** — version string formatting; also falls back to Go's own build-info VCS stamp (`debug.ReadBuildInfo`) when linker-injected values aren't set.

### Interactive selection flow

Both single-select and multi-select prompts inject a sentinel `Branch` into the options list rather than special-casing an "exit" path:
- Single-select (`switch`, `delete`) appends a `NewDummyBranch("Cancel ...")` entry; selecting it short-circuits the caller via `IsDummyBranch()`.
- Multi-select (`batch-delete`) prepends a `NewDoneBranch("Done")` entry and recurses in `listBranchesAndSelectMultipleTargets` (toggling `isSelected` on each non-Done pick) until "Done" is chosen, then returns everything with `IsSelected() == true`.

### Destructive operations

`delete` and `batch-delete` both call `git branch -D` (force delete, ignores unmerged-changes safety) and always prompt for an explicit `yes`/`y` confirmation (`promptui.Prompt{IsConfirm: true}`) before invoking `core.GitClient`. Preserve this confirm-before-force-delete pattern when touching these flows.
