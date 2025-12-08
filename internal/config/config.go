// Package config handles CLI configuration and flag parsing for mortal-prompter.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/diegoram/mortal-prompter/internal/fighters"
	"github.com/spf13/cobra"
)

// Default values for configuration options
const (
	DefaultMaxIterations = 10
	DefaultOutputDir     = ".mortal-prompter"
	DefaultCommitMessage = "feat: implemented via mortal-prompter"
)

// Config holds all configuration options for mortal-prompter.
type Config struct {
	// Prompt is the initial prompt to send to the implementer (required in CLI mode)
	Prompt string

	// WorkDir is the working directory for git operations and CLI execution
	WorkDir string

	// MaxIterations is the maximum number of rounds before requiring confirmation
	MaxIterations int

	// Interactive enables interactive mode, prompting for confirmation each round
	Interactive bool

	// Verbose enables detailed output logging
	Verbose bool

	// OutputDir is the directory for logs and reports
	OutputDir string

	// AutoCommit enables automatic git commit on successful completion
	AutoCommit bool

	// CommitMessage is the base message for auto-commits
	CommitMessage string

	// NoTUI disables the TUI and uses CLI mode instead
	NoTUI bool

	// Implementer is the fighter type used as implementer (claude, codex, gemini)
	Implementer fighters.FighterType

	// Reviewer is the fighter type used as reviewer (claude, codex, gemini)
	Reviewer fighters.FighterType
}

// New creates a new Config with default values.
func New() *Config {
	return &Config{
		WorkDir:       ".",
		MaxIterations: DefaultMaxIterations,
		OutputDir:     DefaultOutputDir,
		CommitMessage: DefaultCommitMessage,
		Implementer:   fighters.FighterTypeClaude,
		Reviewer:      fighters.FighterTypeCodex,
	}
}

// BindFlags binds the configuration flags to a Cobra command.
// This sets up all CLI flags and their descriptions.
func (c *Config) BindFlags(cmd *cobra.Command) {
	flags := cmd.Flags()

	flags.StringVarP(&c.Prompt, "prompt", "p", "",
		"Initial prompt for the implementer (required in CLI mode)")

	flags.StringVarP(&c.WorkDir, "dir", "d", ".",
		"Working directory for git operations")

	flags.IntVarP(&c.MaxIterations, "max-iterations", "m", DefaultMaxIterations,
		"Maximum number of iterations before requiring confirmation")

	flags.BoolVarP(&c.Interactive, "interactive", "i", false,
		"Interactive mode - prompt for confirmation each round")

	flags.BoolVarP(&c.Verbose, "verbose", "v", false,
		"Enable verbose/detailed output")

	flags.StringVarP(&c.OutputDir, "output", "o", DefaultOutputDir,
		"Directory for logs and reports")

	flags.BoolVar(&c.AutoCommit, "auto-commit", false,
		"Automatically commit changes on successful completion")

	flags.StringVar(&c.CommitMessage, "commit-message", DefaultCommitMessage,
		"Base message for auto-commits")

	flags.BoolVar(&c.NoTUI, "no-tui", false,
		"Disable TUI and use CLI mode (requires -p/--prompt)")

	// Fighter selection flags
	var implementer, reviewer string
	flags.StringVar(&implementer, "implementer", "claude",
		"Fighter to use as implementer (claude, codex, gemini)")
	flags.StringVar(&reviewer, "reviewer", "codex",
		"Fighter to use as reviewer (claude, codex, gemini)")

	// Store the string values to be parsed in a PreRun hook
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		var err error
		c.Implementer, err = parseFighterType(implementer)
		if err != nil {
			return fmt.Errorf("invalid implementer: %w", err)
		}
		c.Reviewer, err = parseFighterType(reviewer)
		if err != nil {
			return fmt.Errorf("invalid reviewer: %w", err)
		}
		return nil
	}
}

// parseFighterType converts a string to a FighterType
func parseFighterType(s string) (fighters.FighterType, error) {
	switch strings.ToLower(s) {
	case "claude":
		return fighters.FighterTypeClaude, nil
	case "codex":
		return fighters.FighterTypeCodex, nil
	case "gemini":
		return fighters.FighterTypeGemini, nil
	default:
		return "", fmt.Errorf("unknown fighter type: %s (valid: claude, codex, gemini)", s)
	}
}

// Validate checks that the configuration is valid and returns an error if not.
func (c *Config) Validate() error {
	// Prompt is only required in CLI mode (--no-tui or when -p is provided)
	if c.NoTUI && c.Prompt == "" {
		return errors.New("prompt is required in CLI mode: use -p or --prompt to specify")
	}

	if c.MaxIterations < 1 {
		return errors.New("max-iterations must be at least 1")
	}

	// Resolve and validate working directory
	absWorkDir, err := filepath.Abs(c.WorkDir)
	if err != nil {
		return errors.New("invalid working directory: " + err.Error())
	}

	info, err := os.Stat(absWorkDir)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.New("working directory does not exist: " + absWorkDir)
		}
		return errors.New("cannot access working directory: " + err.Error())
	}

	if !info.IsDir() {
		return errors.New("working directory path is not a directory: " + absWorkDir)
	}

	// Update to absolute path
	c.WorkDir = absWorkDir

	// Resolve output directory relative to working directory
	if !filepath.IsAbs(c.OutputDir) {
		c.OutputDir = filepath.Join(c.WorkDir, c.OutputDir)
	}

	return nil
}

// EnsureOutputDir creates the output directory if it doesn't exist.
func (c *Config) EnsureOutputDir() error {
	return os.MkdirAll(c.OutputDir, 0755)
}
