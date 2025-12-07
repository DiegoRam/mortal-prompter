# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Mortal Prompter is a Go CLI tool that orchestrates an iterative development and code review loop between Claude Code and OpenAI Codex. It acts as an "arbiter" in a Mortal Kombat-style battle where Claude Code implements tasks and Codex reviews the code, continuing until no issues are found or the iteration limit is reached.

## Build Commands

```bash
# Build for current platform
make build

# Install to system
make install

# Build for all platforms (darwin/linux amd64/arm64, windows amd64)
make build-all

# Run tests
make test

# Clean build artifacts
make clean

# Dry run release (test without publishing)
make release-dry
```

## Architecture

```
cmd/mortal-prompter/main.go    # Entry point with CLI flags
internal/
├── orchestrator/              # Main "combat" loop between LLMs
├── fighters/
│   ├── claude.go              # Wrapper for claude CLI
│   └── codex.go               # Wrapper for codex CLI
├── git/                       # Git operations (diff, commit)
├── logger/                    # Terminal + file logging with arcade-style output
└── config/                    # Configuration and flag parsing
pkg/types/                     # Shared types
```

### Core Flow

1. User provides initial prompt
2. Claude Code executes task (via `claude -p "<prompt>" --dangerously-skip-permissions`)
3. Capture `git diff` of changes
4. Codex reviews diff and identifies issues
5. If issues found: construct new prompt for Claude with issues list, repeat
6. If no issues (LGTM): finish successfully
7. After 10 iterations: prompt for manual confirmation

### Key Types

- `Orchestrator`: Manages the battle loop, holds references to fighters and tracks rounds
- `Round`: Records each iteration's prompt, output, diff, review, and issues
- `ReviewResult`: Codex's parsed review with `HasIssues` flag and issue list

### Issue Detection

- Codex outputs "LGTM: No issues found" when code passes review
- Issues are parsed from lines starting with "ISSUE:"

## Dependencies

- `github.com/spf13/cobra` - CLI framework
- `github.com/fatih/color` - Terminal colors
- `github.com/briandowns/spinner` - Loading spinners

## Output

Session logs and reports are written to `.mortal-prompter/`:
- `session-{timestamp}.log` - Detailed log file
- `report-{timestamp}.md` - Markdown battle report
