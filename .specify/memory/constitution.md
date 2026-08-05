<!--
Sync Impact Report
- Version change: [TEMPLATE UNFILLED] → 1.0.0 (initial ratification)
- Modified principles: n/a (first fill of the template)
- Added sections:
  - Core Principles: Shell Out, Don't Reimplement Git; Strict One-Way Layered Architecture;
    Confirm Before Destructive Actions (NON-NEGOTIABLE); Centralized, Predictable Error Handling;
    Minimal Dependencies & Idiomatic Go Simplicity
  - Additional Constraints (build/version/UX/license constraints)
  - Development Workflow
  - Governance (amendment procedure, versioning policy, compliance review)
- Removed sections: none
- Templates requiring updates:
  - .specify/templates/plan-template.md ✅ (generic "Constitution Check" gate derives from this file at plan time; no hardcoded principle text to update)
  - .specify/templates/spec-template.md ✅ (language-agnostic; no changes required)
  - .specify/templates/tasks-template.md ✅ (language-agnostic; no changes required)
  - .specify/templates/commands/*.md ⚠ pending (directory does not exist in this project checkout; nothing to sync)
  - README.md / CLAUDE.md ✅ (principles below were derived from, and remain consistent with, both files)
- Follow-up TODOs:
  - None. RATIFICATION_DATE was derived from the repository's initial commit (2024-07-22).
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
safety). Both flows MUST always prompt for an explicit `yes`/`y` confirmation
(`promptui.Prompt{IsConfirm: true}`) before invoking `core.GitClient`, and this
confirm-before-force-delete pattern MUST be preserved in any change that touches these flows.
Rationale: force-delete is irreversible and can silently discard unmerged work if left
unguarded; an explicit confirmation is the only safety net.

### IV. Centralized, Predictable Error Handling
`execution` functions MUST return `error` rather than printing failures themselves;
`main.go` is the single place that prints errors returned from execution functions. The
only sanctioned exception is the two delete flows, which print an abort message directly
on user cancellation. New subcommands MUST follow this same convention.
Rationale: one predictable place to reason about user-facing failure output avoids scattered,
inconsistent error presentation.

### V. Minimal Dependencies & Idiomatic Go Simplicity
The dependency footprint MUST stay small and deliberate — currently `fatih/color`,
`jedib0t/go-pretty/v6`, and `manifoldco/promptui`. New dependencies (in particular a Git
library or a heavyweight CLI/UI framework) require explicit justification against this
principle. All code MUST pass `go fmt ./...` and `go vet ./...`.
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
- Licensed under MIT. No user data leaves the local machine — the tool only shells out to
  the local `git` binary and never makes network calls.

## Development Workflow

- Build via `go build -o gogit-branch .`, or `./scripts/build.sh [output-name]` for
  version-stamped builds mirroring CI/release builds.
- `go fmt ./...` and `go vet ./...` MUST pass before work is considered complete.
- There is currently no test suite (`go test ./...` finds no tests). New user-facing
  functionality SHOULD add tests per the README's contribution guidelines, and any
  user-facing change MUST update the relevant documentation (README.md and/or CLAUDE.md).
- One subcommand = one file under `execution/`, with a single exported entry point wired
  into `main.go`'s command dispatch table (`list`/`ls`, `switch`/`sw`, `delete`/`del`,
  `batch-delete`/`bd`, `version`/`v`, `help`/`h`).

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

**Version**: 1.0.0 | **Ratified**: 2024-07-22 | **Last Amended**: 2026-08-05
