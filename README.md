# Mortal Prompter

<img width="591" height="291" alt="image" src="https://github.com/user-attachments/assets/d7269d78-eef3-45c0-94c1-83c4903cfd89" />

<img width="410" height="689" alt="image" src="https://github.com/user-attachments/assets/fb99a937-980d-458e-98f0-d660d290f4ae" />


> *"FINISH HIM!"* - When your code finally passes code review

A CLI tool that orchestrates an iterative development and code review loop between AI coding assistants. Choose your fighters from **Claude Code**, **OpenAI Codex**, or **Google Gemini** and watch them battle it out to produce flawless code.

## Overview

Mortal Prompter acts as a referee in a code review battle arena where you choose your fighters:
- **Implementer**: Executes development tasks and fixes issues
- **Reviewer**: Reviews code changes and identifies problems

### Available Fighters

| Fighter | Description | Can Implement | Can Review |
|---------|-------------|:-------------:|:----------:|
| **Claude** | Anthropic's Claude Code CLI | ✅ | ✅ |
| **Codex** | OpenAI's Codex CLI | ✅ | ✅ |
| **Gemini** | Google's Gemini CLI | ✅ | ✅ |

By default, **Claude** is the implementer and **Codex** is the reviewer, but you can mix and match any combination!

The battle continues until the reviewer gives a "LGTM" (Looks Good To Me) or the iteration limit is reached.

## Features

- Interactive TUI with arcade-style visuals
- CLI mode for scripting and automation
- Real-time battle progress with health bars
- Automatic git diff capture between rounds
- Detailed session logs and markdown battle reports
- Auto-commit option for successful sessions
- Configurable iteration limits

## How It Works

```
User provides prompt
        │
┌───────▼───────────────┐
│      ROUND N          │
├───────────────────────┤
│ 1. Claude Code builds │
│ 2. Capture git diff   │
│ 3. Codex reviews diff │
│ 4. Issues found?      │
│    - Yes → new round  │
│    - No → FINISH HIM! │
└───────────────────────┘
        │
  FLAWLESS VICTORY
```

## Installation

### Prerequisites

- **Go 1.21+** (for building from source)
- **Git** - [Install Git](https://git-scm.com/downloads)
- **Claude Code CLI** - [Install Claude Code](https://docs.anthropic.com/en/docs/claude-code)
- **OpenAI Codex CLI** - [Install Codex](https://github.com/openai/codex)

### With Go

```bash
go install github.com/diegoram/mortal-prompter/cmd/mortal-prompter@latest
```

### With Homebrew (macOS/Linux)

```bash
brew tap diegoram/tap
brew install mortal-prompter
```

### Direct Script (macOS/Linux)

```bash
curl -sSL https://raw.githubusercontent.com/diegoram/mortal-prompter/main/scripts/install.sh | bash
```

### Manual Download

Download the latest release from [GitHub Releases](https://github.com/diegoram/mortal-prompter/releases).

## Usage

### TUI Mode (Default)

Simply run without arguments to launch the interactive TUI:

```bash
mortal-prompter
```

The TUI provides:
- Text input for your development prompt
- Real-time battle visualization
- Health bars showing iteration progress
- Live output from both fighters

### CLI Mode

Use `-p` flag or `--no-tui` for non-interactive mode:

```bash
# Basic usage (Claude implements, Codex reviews)
mortal-prompter -p "implement user authentication"

# Choose your fighters
mortal-prompter -p "add feature" --implementer gemini --reviewer claude

# Gemini vs Gemini battle
mortal-prompter -p "refactor module" --implementer gemini --reviewer gemini

# Codex implements, Claude reviews
mortal-prompter -p "fix bug" --implementer codex --reviewer claude

# With auto-commit on success
mortal-prompter -p "add input validation" --auto-commit

# Verbose and interactive mode
mortal-prompter -p "refactor auth module" -v -i

# In specific directory with custom iteration limit
mortal-prompter -p "add unit tests" -d ./backend -m 5

# Show version
mortal-prompter --version
```

### Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--prompt` | `-p` | Initial prompt for the implementer | - |
| `--implementer` | - | Fighter for implementation (claude, codex, gemini) | `claude` |
| `--reviewer` | - | Fighter for code review (claude, codex, gemini) | `codex` |
| `--dir` | `-d` | Working directory | `.` |
| `--max-iterations` | `-m` | Max iterations before confirmation | `10` |
| `--interactive` | `-i` | Prompt for confirmation each round | `false` |
| `--verbose` | `-v` | Enable detailed output | `false` |
| `--output` | `-o` | Directory for logs and reports | `.mortal-prompter` |
| `--auto-commit` | - | Auto-commit on success | `false` |
| `--commit-message` | - | Base commit message | `feat: implemented via mortal-prompter` |
| `--no-tui` | - | Disable TUI, use CLI mode | `false` |
| `--version` | - | Show version info | - |

## Output

Session artifacts are saved to `.mortal-prompter/`:

- `session-{timestamp}.log` - Detailed session log with the original prompt and all battle activity
- `report-{timestamp}.md` - Markdown battle report

### Monitoring a Live Session

You can monitor an active battle in real-time from another terminal using `tail -f`:

```bash
# Watch the latest log file
tail -f .mortal-prompter/session-*.log

# Or find the most recent one
tail -f $(ls -t .mortal-prompter/session-*.log | head -1)
```

This is useful for:
- **Postmortem analysis**: Review what happened after a session
- **Live debugging**: Watch the battle unfold in a separate terminal
- **CI/CD pipelines**: Capture detailed logs for automated runs

### Log Contents

The session log includes:
- **Original prompt**: The exact prompt provided by the user
- **Round activity**: Each round's implementation and review details
- **Git diffs**: Changes made in each iteration
- **Issues found**: All issues identified by the reviewer
- **Timing**: Duration of each phase

### Example CLI Output

```
═══════════════════════════════════════════════════════════
  MORTAL PROMPTER - ROUND 1
═══════════════════════════════════════════════════════════

  CLAUDE CODE enters the arena...
  Executing task...
  CLAUDE CODE finishes! (took 45s)

  Changes detected: 5 files modified

  CODEX enters the arena...
  Reviewing changes...
  CODEX found 3 issues!

   ISSUE 1: Missing error handling in auth.go:45
   ISSUE 2: SQL injection vulnerability in users.go:23
   ISSUE 3: Unused variable in main.go:12

  Preparing next round...
═══════════════════════════════════════════════════════════
```

## Development

```bash
# Clone the repo
git clone https://github.com/diegoram/mortal-prompter.git
cd mortal-prompter

# Build
make build

# Run tests
make test

# Install locally
make install

# Build for all platforms
make build-all

# Dry run release
make release-dry
```

### Project Structure

```
cmd/mortal-prompter/       # CLI entry point
internal/
├── orchestrator/          # Main battle loop between LLMs
├── fighters/              # Fighter implementations (Claude, Codex, Gemini)
├── tui/                   # Terminal UI with Bubble Tea
├── git/                   # Git operations (diff, commit)
├── logger/                # Logging with arcade-style output
├── reporter/              # Markdown battle report generator
└── config/                # Configuration and flag parsing
pkg/types/                 # Shared types
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) for CLI
- TUI powered by [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- Inspired by Mortal Kombat's iconic battle format
