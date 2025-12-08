// Package config handles CLI configuration and flag parsing for mortal-prompter.
package config

import (
	"errors"
	"os"
	"path/filepath"

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
	// Prompt is the initial prompt to send to Claude Code (required in CLI mode)
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
}

// New creates a new Config with default values.
func New() *Config {
	return &Config{
		WorkDir:       ".",
		MaxIterations: DefaultMaxIterations,
		OutputDir:     DefaultOutputDir,
		CommitMessage: DefaultCommitMessage,
	}
}

// BindFlags binds the configuration flags to a Cobra command.
// This sets up all CLI flags and their descriptions.
func (c *Config) BindFlags(cmd *cobra.Command) {
	flags := cmd.Flags()

	flags.StringVarP(&c.Prompt, "prompt", "p", "",
		"Initial prompt for Claude Code (required)")

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
