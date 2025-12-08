package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestNew(t *testing.T) {
	cfg := New()

	if cfg.WorkDir != "." {
		t.Errorf("expected WorkDir to be '.', got %q", cfg.WorkDir)
	}
	if cfg.MaxIterations != DefaultMaxIterations {
		t.Errorf("expected MaxIterations to be %d, got %d", DefaultMaxIterations, cfg.MaxIterations)
	}
	if cfg.OutputDir != DefaultOutputDir {
		t.Errorf("expected OutputDir to be %q, got %q", DefaultOutputDir, cfg.OutputDir)
	}
	if cfg.CommitMessage != DefaultCommitMessage {
		t.Errorf("expected CommitMessage to be %q, got %q", DefaultCommitMessage, cfg.CommitMessage)
	}
	if cfg.Interactive {
		t.Error("expected Interactive to be false")
	}
	if cfg.Verbose {
		t.Error("expected Verbose to be false")
	}
	if cfg.AutoCommit {
		t.Error("expected AutoCommit to be false")
	}
}

func TestValidate_MissingPrompt(t *testing.T) {
	cfg := New()
	// In TUI mode (default), prompt is not required at validation time
	// Only required in CLI mode (NoTUI = true)
	cfg.NoTUI = true
	err := cfg.Validate()
	if err == nil {
		t.Error("expected error for missing prompt in CLI mode")
	}
	if err.Error() != "prompt is required in CLI mode: use -p or --prompt to specify" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestValidate_PromptNotRequiredInTUIMode(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "mortal-prompter-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := New()
	cfg.WorkDir = tmpDir
	// NoTUI defaults to false (TUI mode)
	// Prompt should not be required in TUI mode
	err = cfg.Validate()
	if err != nil {
		t.Errorf("unexpected error in TUI mode without prompt: %v", err)
	}
}

func TestValidate_InvalidMaxIterations(t *testing.T) {
	cfg := New()
	cfg.Prompt = "test prompt"
	cfg.MaxIterations = 0

	err := cfg.Validate()
	if err == nil {
		t.Error("expected error for invalid max-iterations")
	}
	if err.Error() != "max-iterations must be at least 1" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestValidate_InvalidWorkDir(t *testing.T) {
	cfg := New()
	cfg.Prompt = "test prompt"
	cfg.WorkDir = "/nonexistent/path/that/does/not/exist"

	err := cfg.Validate()
	if err == nil {
		t.Error("expected error for nonexistent work directory")
	}
}

func TestValidate_Success(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "mortal-prompter-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := New()
	cfg.Prompt = "implement something"
	cfg.WorkDir = tmpDir

	err = cfg.Validate()
	if err != nil {
		t.Errorf("unexpected validation error: %v", err)
	}

	// Check that WorkDir is now absolute
	if !filepath.IsAbs(cfg.WorkDir) {
		t.Error("expected WorkDir to be absolute after validation")
	}

	// Check that OutputDir is resolved relative to WorkDir
	expectedOutputDir := filepath.Join(tmpDir, DefaultOutputDir)
	if cfg.OutputDir != expectedOutputDir {
		t.Errorf("expected OutputDir to be %q, got %q", expectedOutputDir, cfg.OutputDir)
	}
}

func TestBindFlags(t *testing.T) {
	cfg := New()
	cmd := &cobra.Command{Use: "test"}
	cfg.BindFlags(cmd)

	// Test that all flags are registered
	flags := []string{"prompt", "dir", "max-iterations", "interactive", "verbose", "output", "auto-commit", "commit-message", "no-tui"}
	for _, flag := range flags {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("expected flag %q to be registered", flag)
		}
	}

	// Test short flags
	shortFlags := map[string]string{
		"p": "prompt",
		"d": "dir",
		"m": "max-iterations",
		"i": "interactive",
		"v": "verbose",
		"o": "output",
	}
	for short, long := range shortFlags {
		f := cmd.Flags().ShorthandLookup(short)
		if f == nil {
			t.Errorf("expected short flag -%s to be registered", short)
		} else if f.Name != long {
			t.Errorf("expected short flag -%s to map to --%s, got --%s", short, long, f.Name)
		}
	}
}

func TestEnsureOutputDir(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "mortal-prompter-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := New()
	cfg.OutputDir = filepath.Join(tmpDir, "test-output", "nested")

	err = cfg.EnsureOutputDir()
	if err != nil {
		t.Errorf("unexpected error ensuring output dir: %v", err)
	}

	// Verify directory was created
	info, err := os.Stat(cfg.OutputDir)
	if err != nil {
		t.Errorf("output directory was not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected output path to be a directory")
	}
}

func TestValidate_WorkDirIsFile(t *testing.T) {
	// Create a temporary file (not directory)
	tmpFile, err := os.CreateTemp("", "mortal-prompter-test")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	cfg := New()
	cfg.Prompt = "test prompt"
	cfg.WorkDir = tmpFile.Name()

	err = cfg.Validate()
	if err == nil {
		t.Error("expected error when work directory is a file")
	}
}
