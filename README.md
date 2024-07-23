# GoGit Utility Belt
A _git_ utility wrapping the git commands I use the most, written purely in __Go__.

# Available commands
- list, ls: List all the branches in the current working directory.
- switch, sw: List all the available branches and allows you to pick one to switch to.
- delete, del: List all the available branches and allows you to pick one to delete it.

# Disclaimer
When using the `switch` (or `sw`) mode, the switch operation is __forced__, meaning all unstaged/uncommitted changes will be lost. This is a _non-reversible_ operation. Be warned !
