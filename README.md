# GoGit Branch Manager

<div align="center">

**A powerful, interactive Git branch management CLI tool written in Go**

[![Go Version](https://img.shields.io/badge/Go-1.22.2+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

</div>

---

## 📖 Overview

**GoGit Branch Manager** is a streamlined command-line utility that simplifies Git branch operations through an intuitive, interactive interface. Built with Go, it wraps commonly-used Git commands into a cohesive tool that enhances developer productivity by providing visual feedback, interactive selection menus, and safety confirmations.

This tool is inspired by the amazing package [froggit](https://github.com/thewizardshell/froggit). Please go ahead and check it out!

### Key Features

- 🌿 **Interactive Branch Switching** - Browse and switch between branches with an elegant selection interface
- 🗑️ **Safe Branch Deletion** - Delete branches with confirmation prompts to prevent accidental data loss
- 📊 **Visual Branch Listing** - Display all branches in a formatted table with commit hashes
- 🎨 **Colorized Output** - Enhanced readability with color-coded terminal output
- ⚡ **Fast & Lightweight** - Minimal dependencies, compiled binary for instant execution
- 🔍 **Auto Git Root Detection** - Automatically detects and operates from the Git repository root

---

## 🚀 Installation

### Prerequisites

- Go 1.22.2 or higher
- Git installed and accessible in your PATH

### Build from Source

```bash
# Clone the repository
git clone https://github.com/cainlara/gogit-branch.git
cd gogit-branch

# Build the binary using the provided build script (recommended)
./scripts/build.sh
```

> [!TIP]
> Use `./scripts/build.sh [output-name]` rather than a plain `go build` — it stamps the
> binary with version metadata (git tag/commit, build date, and a dirty-tree flag) via
> `-ldflags`, so the `version`/`v` command reports something meaningful instead of the
> `dev`/`none`/`unknown` defaults. It defaults to producing a binary named `gogit`; pass a
> different name as the first argument if you want another one (e.g.
> `./scripts/build.sh gogit`).

```bash
# (Optional) Move to a directory in your PATH
mv gogit /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/cainlara/gogit-branch@latest
```

### Check installation
```bash
gogit
```

---

## 📚 Usage

### Command Syntax

```bash
gogit [command]
```

If no command is provided, the tool defaults to show current version.

### Available Commands

| Command | Alias | Description |
|---------|-------|-------------|
| [`list`](#list-all-branches) | `ls` | Display all branches in a formatted table with current branch indicator and commit hashes |
| [`switch`](#switch-branches) | `sw` | Interactively browse and switch to a different branch |
| [`delete`](#delete-a-branch) | `del` | Interactively select and delete a branch with confirmation |
| [`batch-delete`](#batch-delete-branches) | `bd` | Interactively select and delete multiple branches with confirmation |
| [`create`](#create-a-new-branch) | `c` | Create a new branch and switch to it (optionally pass the branch name directly) |
| [`status`](#view-and-act-on-repository-status) | `st` | Show a colorized, key-driven view of tracked/untracked changes, with quick commit actions |
| [`push`](#push-the-current-branch) | `p` | Push the current branch, automatically setting the upstream first if it isn't tracked yet |
| [`log`](#view-recent-commit-history) | `l` | Show recent commit history in a compact, colorized one-line-per-commit format (optionally pass a limit) |
| [`version`](#show-current-version) | `v` | Show the version of this humble tool. |
| `help` | `h` | Display usage information and available commands |

### Examples

#### Show Current Version
```bash
gogit version
```

**Output:**
```
Version: v1.0.0
Built: 2026-01-28T19:54:08Z
```

#### List All Branches
```bash
# Using full command
gogit list

# Using alias
gogit ls

# Default behavior (no arguments)
gogit
```

**Output:**
```
Listing branches
┌─────────┬─────────────┬──────────────┐
│ CURRENT │ BRANCH NAME │ CURRENT HASH │
├─────────┼─────────────┼──────────────┤
│ *       │ main        │ a1b2c3d      │
│         │ feature-x   │ e4f5g6h      │
│         │ bugfix-y    │ i7j8k9l      │
└─────────┴─────────────┴──────────────┘
```

#### Switch Branches
```bash
gogit switch
# or
gogit sw
```

**Interactive prompt:**
```
Switching branches
Select Target Branch
🌿 feature-x (e4f5g6h1234567890abcdef1234567890abcdef)
  bugfix-y (i7j8k9l)
  Cancel Switch
```

#### Delete a Branch
```bash
gogit delete
# or
gogit del
```

**Interactive prompt with confirmation:**
```
Deleting branch
Select Target Branch
💀 feature-x (e4f5g6h1234567890abcdef1234567890abcdef)
  bugfix-y (i7j8k9l)
  Cancel Delete
Are you sure you want to delete feature-x (e4f5g6h)? 
[Type yes or y to continue or n to cancel]
```

#### Batch Delete Branches
```bash
gogit batch-delete
# or
gogit bd
```

**Interactive multi-selection prompt:**
```
Deleting branches
Select Target Branches (Press Space to select, Enter when done)
  Done
💀 [x] feature-x (e4f5g6h1234567890abcdef1234567890abcdef)
[x] bugfix-y (i7j8k9l)
[ ] old-feature (m1n2o3p)
```

**Confirmation prompt after selection:**
```
Confirm deletion of selected branches: feature-x (e4f5g6h), bugfix-y (i7j8k9l)
[Type yes or y to continue or n to cancel]
```

> [!TIP]
> Use the batch-delete command to efficiently clean up multiple branches at once. Press **Enter** to toggle selection. Select "Done" when you've finished choosing branches.

#### Create a New Branch
```bash
gogit create
# or
gogit c
```

**Interactive prompt:**
```
Creating branch
✔ Branch name: feature-x
🌿 Created and switched to feature-x
```

You can also pass the branch name directly to skip the prompt entirely:
```bash
gogit create feature-x
# or
gogit c feature-x
```

**Output:**
```
Creating branch
🌿 Created and switched to feature-x
```

When passed as an argument, the name is trimmed of surrounding whitespace and rejected (with
a clear error) if it is empty or starts with a dash; the interactive prompt is unaffected by
this validation. Either way, if a branch with that name already exists, `create` reports the
conflict and leaves your repository untouched — use `switch` to move to an existing branch
instead.

#### View and Act on Repository Status
```bash
gogit status
# or
gogit st
```

**Output:**
```
On branch main (tracking origin/main, ahead 1)

Tracked files
  tracked1.txt (modified) [staged + unstaged]
  tracked2.txt (deleted) [unstaged]

Untracked files
  untracked.txt

Options
    (c)ommit all tracked files  |  (a)dd untracked files and then commit all  |  (e)xit
```

This mirrors everything a plain `git status` would tell you — current branch (or detached-HEAD
state), upstream tracking/ahead-behind counts, and every tracked file colored by its status
(modified, deleted, new, or other), with an annotation when a file has both a staged and an
unstaged change — plus untracked files in their own distinct color. Nothing is left out.

Press a single key to act, no Enter required:
- **`c`** — prompts for a commit message and commits all tracked changes
- **`a`** — prompts for a commit message, stages every tracked and untracked change, and
  commits everything together
- **`e`** — exits immediately with no changes
- Any other key is ignored

A commit message is required for both `c` and `a`; an empty (or whitespace-only) message is
rejected with a clear error and nothing is committed — for `a`, nothing is staged either.

#### Push the Current Branch
```bash
gogit push
# or
gogit p
```

**Output (branch already tracked):**
```
Pushing branch
🚀 Pushed current branch to the remote
```

If the current branch has no upstream configured yet (e.g. it was just created), `push`
automatically sets it — equivalent to `git push --set-upstream origin <branch>` — before
publishing, so you never see git's "no upstream branch" error. Any other failure (no remote
configured, a rejected/diverged push, detached HEAD, etc.) is surfaced exactly as git reports
it; `push` never force-pushes or otherwise bypasses git's own safety checks.

#### View Recent Commit History
```bash
gogit log
# or
gogit l
```

**Output:**
```
a1b2c3d - Improve error messages (2 hours ago) <Jane Doe>
e4f5g6h - Fix a typo (1 day ago) <John Smith>
...
Defaulted to 20 entries because no limit was provided
```

Each line shows the commit's short hash, subject, relative date, and author, colorized
exactly like `git log --pretty=format:"%Cred%h%Creset - %s %Cgreen(%cr)
%C(bold blue)<%an>%Creset"`. With no argument, up to the 20 most recent commits are shown; if
the repository has fewer than 20 commits, every commit is shown and the trailing note reads
"Showing all logs" instead.

You can pass a limit to see a different number of commits:
```bash
gogit log 5
# or
gogit l 5
```

When the supplied limit is fully satisfied (i.e. at least that many commits exist), no
trailing note is shown. Only a positive whole number is accepted — anything else (text, a
decimal, zero, or a negative number) is rejected with a clear error before any git command
runs.

---

## ⚠️ Important Warnings

> [!CAUTION]
> **Branch Deletion**
> 
> The `delete` command uses `git branch -D` (force delete), which will delete branches even if they contain unmerged changes. Always verify you're deleting the correct branch.

---

## 🔧 Dependencies

| Package | Purpose |
|---------|---------|
| [fatih/color](https://github.com/fatih/color) | Terminal color output |
| [jedib0t/go-pretty/v6](https://github.com/jedib0t/go-pretty) | Table rendering |
| [manifoldco/promptui](https://github.com/manifoldco/promptui) | Interactive prompts and selection menus |

---

## 🛠️ Development

### Code Organization

The codebase follows clean architecture principles:

1. **Separation of Concerns** - Git operations, UI logic, and data models are isolated
2. **Dependency Injection** - `GitClient` is passed to execution functions
3. **Error Propagation** - Errors bubble up to the main function for centralized handling
4. **Immutable Models** - Branch struct uses getters/setters for controlled access

---

## 🤝 Contributing

Contributions are welcome! Here's how you can help:

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Commit** your changes (`git commit -m 'Add amazing feature'`)
4. **Push** to the branch (`git push origin feature/amazing-feature`)
5. **Open** a Pull Request

### Development Guidelines

- Follow Go best practices and idioms
- Add tests for new functionality
- Update documentation for user-facing changes
- Ensure all tests pass before submitting PR

---

## 📄 License

This project is licensed under the **MIT License** - see the [LICENSE](LICENSE) file for details.

Copyright (c) 2024 Jose Lara

---

## 🙏 Acknowledgments

- Built with [Go](https://go.dev/)
- Interactive UI powered by [promptui](https://github.com/manifoldco/promptui)
- Table rendering by [go-pretty](https://github.com/jedib0t/go-pretty)
- Colorful output via [color](https://github.com/fatih/color)

---

## 📞 Support

If you encounter any issues or have questions:

- **Issues**: [GitHub Issues](https://github.com/cainlara/gogit-branch/issues)
- **Discussions**: [GitHub Discussions](https://github.com/cainlara/gogit-branch/discussions)

---

<div align="center">

**Made with ❤️ by [Jose Lara](https://github.com/cainlara)**

⭐ Star this repository if you find it helpful!

</div>
