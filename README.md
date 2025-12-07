# Mortal Prompter

> *"FINISH HIM!"* - When your code finally passes code review

CLI tool that orchestrates an iterative development and code review loop between **Claude Code** and **OpenAI Codex**.

## How It Works

```
User sends initial prompt
        │
┌───────▼───────────────┐
│   ROUND N             │
├───────────────────────┤
│ 1. Claude Code builds │
│ 2. Capture git diff   │
│ 3. Codex reviews diff │
│ 4. Issues found?      │
│    - Yes → new round  │
│    - No → FINISH HIM! │
└───────────────────────┘
        │
Final commit + report
```

1. You send a development prompt
2. **Claude Code** implements the changes
3. **Codex** reviews the code and finds issues
4. **Claude Code** fixes the issues
5. Repeat until **FLAWLESS VICTORY**

## Installation

<!-- TODO: Add installation verification -->

### With Go

```bash
go install github.com/minimalart/mortal-prompter/cmd/mortal-prompter@latest
```

### With Homebrew (macOS/Linux)

```bash
brew tap minimalart/tap
brew install mortal-prompter
```

### Direct Script (macOS/Linux)

```bash
curl -sSL https://raw.githubusercontent.com/minimalart/mortal-prompter/main/scripts/install.sh | bash
```

### Manual Download

Download the latest release from [GitHub Releases](https://github.com/minimalart/mortal-prompter/releases).

## Usage

```bash
# Basic usage
mortal-prompter -p "implement user authentication"

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
| `--prompt` | `-p` | Initial prompt for Claude Code (required) | - |
| `--dir` | `-d` | Working directory | `.` |
| `--max-iterations` | `-m` | Max iterations before confirmation | `10` |
| `--interactive` | `-i` | Prompt for confirmation each round | `false` |
| `--verbose` | `-v` | Enable detailed output | `false` |
| `--output` | `-o` | Directory for logs and reports | `.mortal-prompter` |
| `--auto-commit` | - | Auto-commit on success | `false` |
| `--commit-message` | - | Base commit message | `feat: implemented via mortal-prompter` |
| `--version` | - | Show version info | - |

## Output

Session artifacts are saved to `.mortal-prompter/`:

- `session-{timestamp}.log` - Detailed session log
- `report-{timestamp}.md` - Markdown battle report

### Example Output

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

## Requirements

- Git installed and configured
- [Claude Code CLI](https://docs.anthropic.com/claude-code) installed (`claude`)
- [OpenAI Codex CLI](https://openai.com/codex) installed (`codex`)

## Development

```bash
# Clone the repo
git clone https://github.com/minimalart/mortal-prompter.git
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

## License

MIT
