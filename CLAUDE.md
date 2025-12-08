# CLAUDE.md

This file provides guidance to Claude Code when working with code in this repository.

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
cmd/mortal-prompter/main.go    # Entry point, CLI/TUI mode selection
internal/
├── orchestrator/              # Main "combat" loop between LLMs
│   └── orchestrator.go        # Battle logic, round management
├── fighters/
│   ├── fighter.go             # Fighter interface definition
│   ├── claude.go              # Claude Code CLI wrapper
│   └── codex.go               # OpenAI Codex CLI wrapper
├── tui/                       # Terminal UI (Bubble Tea)
│   ├── model.go               # TUI state model
│   ├── view.go                # Rendering logic
│   ├── update.go              # Event handling
│   ├── events.go              # Custom event types
│   ├── observer.go            # Observer for orchestrator events
│   ├── styles.go              # Lip Gloss styling
│   ├── keys.go                # Key bindings
│   └── components/
│       └── healthbar.go       # Health bar component
├── git/                       # Git operations (diff, commit)
├── logger/                    # Terminal + file logging
├── reporter/                  # Markdown battle report generator
└── config/                    # Configuration and flag parsing
pkg/types/                     # Shared types (Round, SessionResult, etc.)
```

## Core Flow

1. User provides initial prompt (via TUI or CLI flag)
2. Claude Code executes task (via `claude -p "<prompt>" --dangerously-skip-permissions`)
3. Capture `git diff` of changes
4. Codex reviews diff and identifies issues (via `codex review`)
5. If issues found: construct new prompt for Claude with issues list, repeat
6. If no issues (LGTM): finish successfully
7. After max iterations: prompt for manual confirmation

## Key Types

- `Orchestrator`: Manages the battle loop, holds references to fighters and tracks rounds
- `Round`: Records each iteration's prompt, output, diff, review, and issues
- `ReviewResult`: Codex's parsed review with `HasIssues` flag and issue list
- `SessionResult`: Final outcome with all rounds, success status, and timing
- `Observer`: Interface for TUI to receive real-time updates from orchestrator

## TUI Architecture

The TUI uses the Elm architecture via Bubble Tea:

- **Model**: Holds all UI state (current phase, prompt, battle progress, logs)
- **Update**: Handles messages and user input, returns new model
- **View**: Renders the current model to the terminal
- **Observer**: Channel-based pattern to receive orchestrator events asynchronously

Phases: `PhasePrompt` → `PhaseBattle` → `PhaseComplete`

## Issue Detection

- Codex outputs "LGTM: No issues found" when code passes review
- Issues are parsed from lines starting with "ISSUE:"

## Dependencies

- `github.com/spf13/cobra` - CLI framework
- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/lipgloss` - TUI styling
- `github.com/charmbracelet/bubbles` - TUI components
- `github.com/fatih/color` - Terminal colors (CLI mode)
- `github.com/briandowns/spinner` - Loading spinners (CLI mode)

## Output

Session logs and reports are written to `.mortal-prompter/`:
- `session-{timestamp}.log` - Detailed log file
- `report-{timestamp}.md` - Markdown battle report

## Modes

- **TUI Mode** (default): Interactive terminal UI with prompt input and battle visualization
- **CLI Mode**: Non-interactive, requires `-p` flag or `--no-tui`
