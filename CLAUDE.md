# CLAUDE.md

This file provides guidance to Claude Code when working with code in this repository.

## Project Overview

Mortal Prompter is a Go CLI tool that orchestrates an iterative development and code review loop between AI coding assistants. It acts as an "arbiter" in a Mortal Kombat-style battle where a configurable **Implementer** builds code and a **Reviewer** reviews it, continuing until no issues are found or the iteration limit is reached.

### Available Fighters

Users can choose any combination of fighters for implementation and review:

| Fighter | Type | Description |
|---------|------|-------------|
| `claude` | Claude Code CLI | Anthropic's Claude Code |
| `codex` | OpenAI Codex CLI | OpenAI's Codex |
| `gemini` | Google Gemini CLI | Google's Gemini |

**Default configuration:** Claude (implementer) vs Codex (reviewer)

**CLI flags:**
- `--implementer <fighter>` - Choose the implementing fighter (default: claude)
- `--reviewer <fighter>` - Choose the reviewing fighter (default: codex)

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
│   ├── codex.go               # OpenAI Codex CLI wrapper
│   └── gemini.go              # Google Gemini CLI wrapper
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
2. User selects implementer and reviewer fighters (via TUI or CLI flags `--implementer`, `--reviewer`)
3. Implementer executes task (e.g., `claude -p "<prompt>" --dangerously-skip-permissions`)
4. Capture `git diff` of changes
5. Reviewer reviews diff and identifies issues (e.g., `codex review`)
6. If issues found: construct new prompt for implementer with issues list, repeat
7. If no issues (LGTM): finish successfully
8. After max iterations: prompt for manual confirmation

## Key Types

- `Orchestrator`: Manages the battle loop, holds references to fighters and tracks rounds
- `Round`: Records each iteration's prompt, output, diff, review, and issues
- `ReviewResult`: Codex's parsed review with `HasIssues` flag and issue list
- `SessionResult`: Final outcome with all rounds, success status, and timing
- `Observer`: Interface for TUI to receive real-time updates from orchestrator

## TUI Architecture

The TUI uses the Elm architecture via Bubble Tea:

- **Model** (`model.go`): Holds all UI state (current view, fighter states, battle progress)
- **Update** (`update.go`): Handles messages and user input, returns new model
- **View** (`view.go`): Renders the current model to the terminal
- **Observer** (`observer.go`): Channel-based pattern to receive orchestrator events asynchronously

### View States Flow

```
ViewFighterSelect → ViewPrompt → ViewBattle → ViewResults
                                      ↓
                              ViewConfirmation
```

### CRITICAL: Adding New Views

When adding a new `ViewState`, you MUST update ALL of these:

1. **model.go**: Add the new `ViewState` constant
2. **view.go**: Add `case ViewNewState:` in `View()` and implement `viewNewState()`
3. **update.go**: Add `case ViewNewState:` in `handleKeyMsg()` and implement `handleNewStateKeys()`

**Checklist for new views:**
- [ ] Added ViewState constant in model.go
- [ ] Added case in View() switch
- [ ] Added viewXxx() render function
- [ ] Added case in handleKeyMsg() switch
- [ ] Added handleXxxKeys() function for keyboard input
- [ ] Initialized any new model fields in NewModel()

### Event Handling Pattern

When the orchestrator sends events via Observer:
1. Events arrive through `eventChan`
2. `waitForEvent()` wraps them in `eventMsg`
3. `handleEvent()` processes them and updates model state
4. Must return `tea.Batch(m.spinner.Tick, tick(), waitForEvent(...))` to keep listening

### Fighter State Updates

Fighter states (`FighterIdle`, `FighterActive`, `FighterFinished`) are updated by:
- `EventFighterEnter`: Sets fighter to Active
- `EventFighterAction`: Updates currentAction and fighter states based on `payload.Fighter`
- `EventFighterFinish`: Sets fighter to Finished

The `FighterActionPayload` includes both `Fighter` (who) and `Action` (what) fields.

## Common Pitfalls to Avoid

1. **Missing keyboard handler**: Every ViewState needs a case in `handleKeyMsg()` or keyboard won't work
2. **Stale field names**: Model fields were refactored from `claudeState/codexState` to `implementerState/reviewerState`. Always check current field names in model.go
3. **Event payload missing data**: If UI shows wrong state, check that event payloads include all needed fields (e.g., `FighterActionPayload.Fighter`)
4. **UI alignment with ANSI**: Use `len(plainText)` for padding calculations, not styled text (ANSI codes break width)
5. **Two-phase TUI**: The app runs TWO TUIs - first for prompt input (quits on submit), second for battle display. Don't wait for events in the first TUI.

## Issue Detection

- Reviewer outputs "LGTM: No issues found" when code passes review
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
- `session-{timestamp}.log` - Detailed log file with original prompt and all activity
- `report-{timestamp}.md` - Markdown battle report

### Log File Usage

The log file is useful for:
- **Postmortem analysis**: Review complete session after it ends
- **Live monitoring**: Use `tail -f .mortal-prompter/session-*.log` from another terminal
- **Debugging**: Contains the original prompt, all rounds, diffs, and issues

The log captures:
1. Original user prompt (stored at session start)
2. Each round's implementation output
3. Git diffs between rounds
4. Reviewer feedback and issues
5. Timing information

## Modes

- **TUI Mode** (default): Interactive terminal UI with prompt input and battle visualization
- **CLI Mode**: Non-interactive, requires `-p` flag or `--no-tui`
