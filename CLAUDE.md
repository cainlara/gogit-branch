# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project overview

GoGit Branch Manager (`gogit-branch`) is a small Go CLI that wraps common Git branch operations (list, switch, delete, batch-delete, create) plus a colorized, key-driven repository status view behind an interactive terminal UI. It shells out to the local `git` binary rather than using a Git library.

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

Unit tests exist for pure, non-git-invoking logic only (branch-name validation in `execution/create_test.go`, porcelain status-line categorization in `model/status_test.go`, and branch-line parsing in `core/git_client_test.go`); there is no end-to-end/integration test harness, so anything that shells out to `git` or reads the terminal is validated manually per each feature's `quickstart.md`.

Version info (`version.Version`, `version.Commit`, `version.Date`, `version.Dirty`) is injected at build time via `-ldflags -X` in `scripts/build.sh`; running via `go build`/`go run` directly leaves these at their `"dev"`/`"none"`/`"unknown"` defaults.

## Architecture

The code is organized into four packages with a strict dependency direction: `main` → `execution` → `core`/`model`. There is intentionally no dependency the other way (e.g. `core` never imports `execution`).

- **`main.go`** — parses `os.Args`, maps the single subcommand (`list`/`ls`, `switch`/`sw`, `delete`/`del`, `batch-delete`/`bd`, `create`/`c`, `status`/`st`, `push`/`p`, `log`/`l`, `reset`/`r`, `version`/`v`, `help`/`h`) to an `execution` function, and constructs one `core.GitClient`. Errors returned from `execution` functions are printed here (centralized error handling — `execution` functions return `error` rather than printing failures themselves, except for the two "delete" flows and `reset`'s cancellation message which print directly on user cancellation).
- **`core/git_client.go`** — `GitClient` is the only place that shells out to `git` (via `os/exec`). It auto-detects the repo root with `git rev-parse --show-toplevel` when constructed with an empty path (`core.NewGitClient("")`), and exposes `Branches`, `Checkout`, `DeleteBranch`, `DeleteBranches`, `CreateAndSwitchBranch`, `Status`, `CommitTracked`, `AddAllAndCommit`, `Push`, `Log`, `Reset`. All git-command output parsing (e.g. stripping the `* ` current-branch prefix from `git branch` output, detecting `error:`-prefixed stderr, and parsing `git status --porcelain --branch`'s branch/tracking line and per-file XY codes) lives here.
- **`model/branch.go`** — `Branch` is a plain data holder with getters/setters (no exported fields) representing a branch's name/hashes plus UI-only state (`isSelected`, `isDone`, `isDummy` for the "Cancel" sentinel entries used in select prompts). `NewDummyBranch`/`NewDoneBranch` construct sentinel entries injected into the list passed to `promptui`, not real branches from git.
- **`model/status.go`** — `FileStatus` (path plus raw staged/unstaged XY codes) and `RepositoryStatus` (branch/detached/upstream/ahead/behind plus tracked `FileStatus` entries and untracked paths) are plain data holders, same convention as `Branch`. `FileStatus.Category()`/`StateAnnotation()` are derived (not stored) — they interpret the already-parsed codes into a color category (modified/deleted/new/other, picked by a fixed priority when staged and unstaged codes differ) and a staged/unstaged/both annotation, so a file is never silently reduced to only one of its two states.
- **`execution/*.go`** — one file per subcommand, each exposing a single entry point called from `main.go` (`ListCurrentBranches`, `BrowseAndSwitchBranches`, `ListAndDeleteBranch`, `ListAndDeleteBranches`, `CreateAndSwitchBranch`, `ShowStatusAndOperate`, `PushCurrentBranch`, `ShowLog`, `ResetToLatestCommit`, `PrintHelp`, `ShowVersion`). `log`/`l` and `reset`/`r` join `create`/`c` as the three subcommands that accept an optional positional CLI argument (a limit, or the `--hard` flag, respectively); `main.go`'s argument-count guard explicitly names all three as the only exceptions. `common.go` holds the shared `promptui.Select` wiring used by both single-select (switch/delete) and multi-select (batch-delete) flows, plus the emoji/banner constants. `create.go`'s `CreateAndSwitchBranch` is the exception to the `Select`-based UX: it takes a new branch name via a freeform `promptui.Prompt` instead, since there is nothing existing to select from — unless the caller already supplied the name as a CLI argument (`create <name>` / `c <name>`), in which case `main.go` passes it through as a non-nil `*string` and the prompt is skipped entirely. A supplied name is trimmed and validated (rejecting empty or dash-led values) in `create.go`'s `validateBranchName`; the interactive-prompt path is intentionally left unvalidated/unchanged. `status.go`'s `ShowStatusAndOperate` renders three sections (tracked files, untracked files, options) from `core.GitClient.Status()` and reads a single raw key press via `golang.org/x/term` (no Enter required) rather than using `promptui.Select`.
- **`version/version.go`** — version string formatting; also falls back to Go's own build-info VCS stamp (`debug.ReadBuildInfo`) when linker-injected values aren't set.

### Interactive selection flow

Both single-select and multi-select prompts inject a sentinel `Branch` into the options list rather than special-casing an "exit" path:
- Single-select (`switch`, `delete`) appends a `NewDummyBranch("Cancel ...")` entry; selecting it short-circuits the caller via `IsDummyBranch()`.
- Multi-select (`batch-delete`) prepends a `NewDoneBranch("Done")` entry and recurses in `listBranchesAndSelectMultipleTargets` (toggling `isSelected` on each non-Done pick) until "Done" is chosen, then returns everything with `IsSelected() == true`.

### Destructive operations

`delete` and `batch-delete` both call `git branch -D` (force delete, ignores unmerged-changes safety) and always prompt for an explicit `yes`/`y` confirmation (`promptui.Prompt{IsConfirm: true}`) before invoking `core.GitClient`. Preserve this confirm-before-force-delete pattern when touching these flows.

### Status command key-press menu

`status`/`st` is the one place in this tool that reads a single raw key press instead of using `promptui.Select` — `execution/status.go`'s `readKeyPress` puts stdin into raw mode via `golang.org/x/term` (`MakeRaw`/`Restore`), reads exactly one byte, and returns it; any byte other than `c`/`a`/`e` is ignored and the loop keeps reading. Choosing `c` or `a` prompts for a commit message via the existing `promptui.Prompt` pattern; an empty (post-trim) message is rejected before any git command runs. `c` commits tracked changes only (`git commit -a -m`); `a` explicitly stages tracked + untracked changes first (`git add -A`) and then commits (`git commit -m`) — it intentionally does not rely on `-a`'s auto-staging, since `-a` never stages new untracked files. If a git-level operation fails for a reason other than an empty message (e.g. a hook rejects it), the error is propagated like any other command's failure and no rollback of an already-completed step (e.g. staging in `AddAllAndCommit`) is attempted.

### Push command upstream handling

`push`/`p` (`core.GitClient.Push`) always attempts a plain `git push` first. Only if that fails with git's own, stable "has no upstream branch" message does it resolve the current branch name (`git rev-parse --abbrev-ref HEAD`) and retry once as `git push --set-upstream origin <branch>`, so a brand-new local branch publishes in a single command instead of surfacing that error to the user. Every other push failure (no remote configured, a non-fast-forward rejection, detached HEAD, etc.) is returned untouched — `Push` never force-pushes or otherwise bypasses git's safety checks. Because a rejected `git push`'s combined output can start with a `To <remote>` progress line before the `error:`/`fatal:` line (unlike every other `GitClient` command, whose failures emit that prefix as their very first line), `Push`'s error wrapping checks for the prefix anywhere in the output (`strings.Contains`) rather than only at the start (`strings.HasPrefix`).

### Log command forced coloring

`log`/`l` (`core.GitClient.Log`) runs `git log` with a fixed one-line `--pretty=format` (short hash, subject, relative date, author) and, critically, `--color=always`. That flag is not cosmetic: `GitClient` always shells out via `os/exec`, which never attaches a real TTY to `git`, so git's own default color auto-detection would otherwise silently strip every `%C(...)` placeholder in the format string — the command would never render any color at all without forcing it. `execution/log.go`'s `selectTrailingNote` decides between three mutually exclusive trailing notes ("Showing all logs" when fewer commits exist than the effective limit, a "defaulted to 20" note when no limit was supplied and the default genuinely constrained the output, or no note when an explicit limit was supplied and satisfied) by comparing the number of lines `Log` returned against the effective limit — no separate commit-count query is made.

### Reset command destructive-action handling

`reset`/`r` (`execution.ResetToLatestCommit`) extends this tool's confirm-before-destructive pattern (Principle III) to a third command alongside `delete`/`batch-delete`, via its own `confirmReset` function mirroring `delete.go`'s `confirmDeleteSelectedBranch`. It checks `model.RepositoryStatus.HasTrackedChanges()`/`HasUntrackedFiles()` (from the existing `Status()` call) to decide whether there's anything to revert for the requested mode *before* showing any warning — skipping the prompt entirely when there isn't, per the same auto-exit precedent `status` established. `core.GitClient.Reset(removeUntracked bool)` always runs `git reset --hard` regardless of the CLI's own `--hard` flag (the two share a name coincidentally, not semantically — even the no-flag mode must fully revert the working tree, which requires git's `--hard` reset mode); `git clean -fd` runs additionally only when the CLI flag was supplied. Mirrors `AddAllAndCommit`'s two-step, no-rollback pattern: if `reset --hard` succeeds but `clean -fd` subsequently fails, the already-reverted tracked changes are not rolled back. Verified during implementation that `git reset --hard` does not error on a repository with zero commits (unlike `log`'s very different "no commits" failure) — it succeeds as a no-op, so the existing "nothing to revert" check already produces the correct behavior with no special-casing needed.
