// Package git provides utilities for git operations used by mortal-prompter.
package git

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// Error types for git operations
var (
	// ErrGitNotInstalled is returned when git is not installed or not found in PATH.
	ErrGitNotInstalled = errors.New("git is not installed or not found in PATH")

	// ErrNotGitRepo is returned when the working directory is not inside a git repository.
	ErrNotGitRepo = errors.New("not a git repository (or any of the parent directories)")

	// ErrNoChanges is returned when there are no changes to commit.
	ErrNoChanges = errors.New("no changes to commit")
)

// Git provides git operations for a specific working directory.
type Git struct {
	workDir string
}

// New creates a new Git instance for the specified working directory.
func New(workDir string) *Git {
	return &Git{
		workDir: workDir,
	}
}

// GetUnstagedDiff returns the diff of unstaged changes (git diff).
func (g *Git) GetUnstagedDiff() (string, error) {
	if !g.IsGitRepo() {
		return "", ErrNotGitRepo
	}
	return g.runGitCommand("diff")
}

// GetStagedDiff returns the diff of staged changes (git diff --staged).
func (g *Git) GetStagedDiff() (string, error) {
	if !g.IsGitRepo() {
		return "", ErrNotGitRepo
	}
	return g.runGitCommand("diff", "--staged")
}

// GetAllDiff returns the diff of all uncommitted changes (git diff HEAD).
func (g *Git) GetAllDiff() (string, error) {
	if !g.IsGitRepo() {
		return "", ErrNotGitRepo
	}
	return g.runGitCommand("diff", "HEAD")
}

// StageAll stages all changes including untracked files (git add -A).
func (g *Git) StageAll() error {
	_, err := g.runGitCommand("add", "-A")
	return err
}

// Commit creates a commit with the specified message.
func (g *Git) Commit(message string) error {
	if message == "" {
		return errors.New("commit message cannot be empty")
	}

	// Check if there are changes to commit
	hasChanges, err := g.HasUncommittedChanges()
	if err != nil {
		return fmt.Errorf("failed to check for uncommitted changes: %w", err)
	}
	if !hasChanges {
		return ErrNoChanges
	}

	_, err = g.runGitCommand("commit", "-m", message)
	return err
}

// GetCurrentBranch returns the name of the current git branch.
func (g *Git) GetCurrentBranch() (string, error) {
	output, err := g.runGitCommand("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// HasUncommittedChanges returns true if there are uncommitted changes in the working directory.
// This includes both staged and unstaged changes, as well as untracked files.
func (g *Git) HasUncommittedChanges() (bool, error) {
	// git status --porcelain returns empty output if there are no changes
	output, err := g.runGitCommand("status", "--porcelain")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(output) != "", nil
}

// IsGitRepo returns true if the working directory is inside a git repository.
func (g *Git) IsGitRepo() bool {
	_, err := g.runGitCommand("rev-parse", "--git-dir")
	return err == nil
}

// runGitCommand executes a git command with the provided arguments.
// It sets the working directory and captures both stdout and stderr.
func (g *Git) runGitCommand(args ...string) (string, error) {
	// Check if git is installed
	if _, err := exec.LookPath("git"); err != nil {
		return "", ErrGitNotInstalled
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = g.workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		// Check if this is a "not a git repository" error
		stderrStr := stderr.String()
		stderrLower := strings.ToLower(stderrStr)
		if strings.Contains(stderrLower, "not a git repository") {
			return "", ErrNotGitRepo
		}

		// For other errors, wrap with context from stderr
		if stderrStr != "" {
			return "", fmt.Errorf("git %s failed: %s", args[0], strings.TrimSpace(stderrStr))
		}
		return "", fmt.Errorf("git %s failed: %w", args[0], err)
	}

	return stdout.String(), nil
}

// WorkDir returns the working directory configured for this Git instance.
func (g *Git) WorkDir() string {
	return g.workDir
}
