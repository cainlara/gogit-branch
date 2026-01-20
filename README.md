# GoGit Branch Manager

<div align="center">

**A powerful, interactive Git branch management CLI tool written in Go**

[![Go Version](https://img.shields.io/badge/Go-1.22.2+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

</div>

---

## 📖 Overview

**GoGit Branch Manager** is a streamlined command-line utility that simplifies Git branch operations through an intuitive, interactive interface. Built with Go, it wraps commonly-used Git commands into a cohesive tool that enhances developer productivity by providing visual feedback, interactive selection menus, and safety confirmations.

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

# Build the binary
go build -o gogit-branch

# (Optional) Move to a directory in your PATH
mv gogit-branch /usr/local/bin/
```

### Using Go Install

```bash
go install cainlara/gogit-branch@latest
```

---

## 📚 Usage

### Command Syntax

```bash
gogit-branch [command]
```

If no command is provided, the tool defaults to listing all branches.

### Available Commands

| Command | Alias | Description |
|---------|-------|-------------|
| `list` | `ls` | Display all branches in a formatted table with current branch indicator and commit hashes |
| `switch` | `sw` | Interactively browse and switch to a different branch |
| `delete` | `del` | Interactively select and delete a branch with confirmation |
| `help` | `h` | Display usage information and available commands |

### Examples

#### List All Branches
```bash
# Using full command
gogit-branch list

# Using alias
gogit-branch ls

# Default behavior (no arguments)
gogit-branch
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
gogit-branch switch
# or
gogit-branch sw
```

**Interactive prompt:**
```
Switching branches
Select Target Branch
🌿 feature-x (e4f5g6h1234567890abcdef1234567890abcdef)
  bugfix-y (i7j8k9l)
```

#### Delete a Branch
```bash
gogit-branch delete
# or
gogit-branch del
```

**Interactive prompt with confirmation:**
```
Deleting branch
Select Target Branch
💀 feature-x (e4f5g6h1234567890abcdef1234567890abcdef)
  bugfix-y (i7j8k9l)

Are you sure you want to delete feature-x (e4f5g6h1234567890abcdef1234567890abcdef)? 
[Type yes or y to continue or anything else to cancel]
```

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

### Running Tests

```bash
go test ./...
```

### Building for Multiple Platforms

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o gogit-branch-linux

# macOS
GOOS=darwin GOARCH=amd64 go build -o gogit-branch-macos

# Windows
GOOS=windows GOARCH=amd64 go build -o gogit-branch.exe
```

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
