<!--
Sync Impact Report
- Version change: 1.0.0 → 1.1.0 (MINOR: materially expanded network-safety guidance and
  broader destructive-action coverage; no numbered principle removed or redefined)
- Modified principles:
  - III. Confirm Before Destructive Actions (NON-NEGOTIABLE) → title unchanged; now also
    covers `reset`'s confirm flow, and the confirmation mechanism is corrected to the shared
    `confirmYesNo` helper (promptui `IsConfirm: true` is intentionally avoided)
  - V. Minimal Dependencies & Idiomatic Go Simplicity → dependency list completed with
    `golang.org/x/term` (raw key-press input for the `status` view)
  - Additional Constraints (network policy) → "never makes network calls" replaced by a
    scoped policy: network only via `git` itself where a command requires it
    (`push`/`pull`/`fetch`, plus `switch`'s MUST-run refresh on remote-only selection with
    abort-on-failure); list rendering and local-branch selections MUST stay offline
- Added sections: none
- Removed sections: none
- Modified sections (accuracy fixes):
  - Development Workflow → command dispatch table completed (added `create`/`c`,
    `status`/`st`, `push`/`p`, `pull`/`pl`, `log`/`l`, `reset`/`r`); test-suite line updated
    to reflect the existing unit tests for pure, non-git-invoking logic
- Templates requiring updates: none — plan/spec/tasks templates and command files derive
  from this file at runtime (no hardcoded principle text to sync)
- Follow-up TODOs: none (no placeholders deferred; RATIFICATION_DATE unchanged: 2024-07-22)
-->

# GoGit Branch Manager Constitution

## Core Principles

### I. Shell Out, Don't Reimplement Git
All Git interaction MUST happen by invoking the local `git` binary via `os/exec`; the tool
MUST NOT embed a Git implementation or Git library. `core.GitClient` (`core/git_client.go`)
is the single, exclusive place that shells out to `git` and parses its output (e.g. stripping
the `* ` current-branch prefix from `git branch` output, detecting `error:`-prefixed stderr).
No other package may invoke `os/exec` for Git commands.
Rationale: keeps the tool lightweight, transparent, and trivially compatible with whatever
`git` the user already has installed, instead of tracking a second implementation of Git
semantics.

### II. Strict One-Way Layered Architecture
The codebase is organized into four packages with a strict dependency direction:
`main` → `execution` → `core`/`model`. `core` MUST NOT import `execution`, and no lower
layer may depend on a higher one. `main.go` parses `os.Args`, maps the subcommand to an
`execution` function, and constructs one `core.GitClient`; `execution/*.go` holds one file
per subcommand, each exposing a single entry point; `model/branch.go` holds plain data with
no exported fields and no dependency on `core` or `execution`.
Rationale: keeps Git plumbing, business logic, and CLI wiring independently reasoned about,
testable, and safe to change without hidden coupling.

### III. Confirm Before Destructive Actions (NON-NEGOTIABLE)
`delete` and `batch-delete` use `git branch -D` (force delete, ignoring unmerged-changes
safety), and `reset` reverts tracked changes with `git reset --hard` (plus `git clean -fd`
when invoked with `--hard`). Every flow that has destructive work to perform MUST prompt for
an explicit `yes`/`y` confirmation before invoking `core.GitClient` — via the shared
`confirmYesNo` helper in `execution/common.go`, which accepts `y`/`yes` case-insensitively
and intentionally avoids promptui's `IsConfirm: true` mode (its built-in accept check rejects
`yes`, which would let the rendered prompt and the actual outcome disagree). `reset` may skip
the prompt only when there is nothing to revert. This confirm-before-destructive pattern MUST
be preserved in any change that touches these flows.
Rationale: force-delete and hard-reset are irreversible and can silently discard unmerged or
uncommitted work if left unguarded; an explicit confirmation is the only safety net.

### IV. Centralized, Predictable Error Handling
`execution` functions MUST return `error` rather than printing failures themselves;
`main.go` is the single place that prints errors returned from execution functions. The
only sanctioned exception is the two delete flows, which print an abort message directly
on user cancellation. New subcommands MUST follow this same convention.
Rationale: one predictable place to reason about user-facing failure output avoids scattered,
inconsistent error presentation.

### V. Minimal Dependencies & Idiomatic Go Simplicity
The dependency footprint MUST stay small and deliberate — currently `fatih/color`,
`jedib0t/go-pretty/v6`, `manifoldco/promptui`, and `golang.org/x/term` (raw key-press input
for the `status` view). New dependencies (in particular a Git library or a heavyweight
CLI/UI framework) require explicit justification against this principle. All code MUST pass
`go fmt ./...` and `go vet ./...`.
Rationale: a small, fast, easily-auditable CLI is the project's core value proposition;
every added dependency erodes that.

## Additional Constraints

- Version metadata (`version.Version`, `version.Commit`, `version.Date`, `version.Dirty`) is
  injected at build time via `-ldflags -X` in `scripts/build.sh`, with `"dev"`/`"none"`/
  `"unknown"` defaults for a plain `go build`/`go run`, and a fallback to Go's own build-info
  VCS stamp (`debug.ReadBuildInfo`) when linker-injected values aren't set.
  `scripts/release_check.sh` MUST verify the current commit is on an exact tag before any
  release build.
- Interactive selection flows inject a sentinel `Branch` into the options list rather than
  special-casing an "exit" path: single-select flows (`switch`, `delete`) append a
  `NewDummyBranch("Cancel ...")` entry; the multi-select flow (`batch-delete`) prepends a
  `NewDoneBranch("Done")` entry and recurses until "Done" is chosen. New interactive flows
  MUST follow this same sentinel pattern instead of introducing bespoke exit handling.
- Licensed under MIT. No data generated by the tool itself leaves the local machine: all Git
  interaction happens by shelling out to the local `git` binary (Principle I).
- All network activity is performed by `git` itself and only where a command requires it:
  `push`, `pull`, `fetch`, and — in `switch` — the refresh that MUST run when a remote-only
  branch is selected (before the checkout, so the user lands on the branch's current remote
  tip; if that refresh fails, the switch MUST abort with an error and no branch change).
  Rendering any branch list MUST NOT perform network operations, and selecting a local
  branch in `switch` MUST NOT perform any sync/fetch — the tool only switches to it and
  restores that branch's local state.

## Development Workflow

- Build via `go build -o gogit-branch .`, or `./scripts/build.sh [output-name]` for
  version-stamped builds mirroring CI/release builds.
- `go fmt ./...` and `go vet ./...` MUST pass before work is considered complete.
- Unit tests exist for pure, non-git-invoking logic only (branch-name validation,
  status-line categorization, branch-line parsing); there is no end-to-end harness, so
  git-invoking flows are validated manually per each feature's quickstart. New user-facing
  functionality SHOULD add tests per the README's contribution guidelines, and any
  user-facing change MUST update the relevant documentation (README.md and/or CLAUDE.md).
- One subcommand = one file under `execution/`, with a single exported entry point wired
  into `main.go`'s command dispatch table (`list`/`ls`, `switch`/`sw`, `delete`/`del`,
  `batch-delete`/`bd`, `create`/`c`, `status`/`st`, `push`/`p`, `pull`/`pl`, `log`/`l`,
  `reset`/`r`, `version`/`v`, `help`/`h`).

## Governance

This constitution supersedes ad hoc practice for this repository. Amendments require
updating this file plus propagating consequential changes to any dependent Spec-Kit
templates (plan/spec/tasks templates, command files) in the same change. Versioning follows
semantic versioning: MAJOR for backward-incompatible principle removals or redefinitions,
MINOR for new or materially expanded principles, PATCH for clarifications and wording fixes.
All PRs and reviews MUST be checked against these principles, with particular attention to
the confirm-before-destructive-delete rule (Principle III) and the one-way layered
architecture (Principle II). Complexity or new dependencies that conflict with Principle V
must be justified in the PR description before merging. Use CLAUDE.md for day-to-day runtime
development guidance; this constitution governs when the two conflict.

**Version**: 1.1.0 | **Ratified**: 2024-07-22 | **Last Amended**: 2026-09-25
