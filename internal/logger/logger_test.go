package logger

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	t.Run("creates logger with new directory", func(t *testing.T) {
		tempDir := t.TempDir()
		outputDir := filepath.Join(tempDir, "test-output")

		l, err := New(outputDir, false)
		if err != nil {
			t.Fatalf("New() returned error: %v", err)
		}
		defer l.Close()

		// Verify directory was created
		info, err := os.Stat(outputDir)
		if err != nil {
			t.Fatalf("Output directory was not created: %v", err)
		}
		if !info.IsDir() {
			t.Fatal("Output path is not a directory")
		}
	})

	t.Run("creates log file with timestamp", func(t *testing.T) {
		tempDir := t.TempDir()

		l, err := New(tempDir, false)
		if err != nil {
			t.Fatalf("New() returned error: %v", err)
		}
		defer l.Close()

		logPath := l.GetLogFilePath()
		if logPath == "" {
			t.Fatal("Log file path is empty")
		}

		if !strings.HasPrefix(filepath.Base(logPath), "session-") {
			t.Errorf("Log file name does not start with 'session-': %s", logPath)
		}

		if !strings.HasSuffix(logPath, ".log") {
			t.Errorf("Log file name does not end with '.log': %s", logPath)
		}

		// Verify file exists
		if _, err := os.Stat(logPath); err != nil {
			t.Errorf("Log file does not exist: %v", err)
		}
	})

	t.Run("uses default directory when empty", func(t *testing.T) {
		// Change to temp directory for this test
		origDir, _ := os.Getwd()
		tempDir := t.TempDir()
		os.Chdir(tempDir)
		defer os.Chdir(origDir)

		l, err := New("", false)
		if err != nil {
			t.Fatalf("New() returned error: %v", err)
		}
		defer l.Close()

		if l.GetOutputDir() != ".mortal-prompter" {
			t.Errorf("Expected default output dir '.mortal-prompter', got '%s'", l.GetOutputDir())
		}

		// Cleanup
		os.RemoveAll(".mortal-prompter")
	})

	t.Run("sets verbose mode correctly", func(t *testing.T) {
		tempDir := t.TempDir()

		l, err := New(tempDir, true)
		if err != nil {
			t.Fatalf("New() returned error: %v", err)
		}
		defer l.Close()

		if !l.IsVerbose() {
			t.Error("Expected verbose to be true")
		}

		l2, err := New(filepath.Join(tempDir, "non-verbose"), false)
		if err != nil {
			t.Fatalf("New() returned error: %v", err)
		}
		defer l2.Close()

		if l2.IsVerbose() {
			t.Error("Expected verbose to be false")
		}
	})
}

func TestRoundStart(t *testing.T) {
	tempDir := t.TempDir()
	l, err := New(tempDir, false)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer l.Close()

	var stdout bytes.Buffer
	l.SetOutputWriters(&stdout, &bytes.Buffer{})

	l.RoundStart(1)

	output := stdout.String()
	if !strings.Contains(output, "ROUND 1") {
		t.Errorf("RoundStart output does not contain 'ROUND 1': %s", output)
	}
	if !strings.Contains(output, "MORTAL PROMPTER") {
		t.Errorf("RoundStart output does not contain 'MORTAL PROMPTER': %s", output)
	}
}

func TestFighterEnter(t *testing.T) {
	tempDir := t.TempDir()
	l, err := New(tempDir, false)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer l.Close()

	var stdout bytes.Buffer
	l.SetOutputWriters(&stdout, &bytes.Buffer{})

	l.FighterEnter("CLAUDE CODE")

	output := stdout.String()
	if !strings.Contains(output, "CLAUDE CODE") {
		t.Errorf("FighterEnter output does not contain fighter name: %s", output)
	}
	if !strings.Contains(output, "enters the arena") {
		t.Errorf("FighterEnter output does not contain 'enters the arena': %s", output)
	}
}

func TestFighterFinish(t *testing.T) {
	tempDir := t.TempDir()
	l, err := New(tempDir, false)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer l.Close()

	var stdout bytes.Buffer
	l.SetOutputWriters(&stdout, &bytes.Buffer{})

	l.FighterFinish("CLAUDE CODE", 45*time.Second)

	output := stdout.String()
	if !strings.Contains(output, "CLAUDE CODE") {
		t.Errorf("FighterFinish output does not contain fighter name: %s", output)
	}
	if !strings.Contains(output, "finishes") {
		t.Errorf("FighterFinish output does not contain 'finishes': %s", output)
	}
	if !strings.Contains(output, "45s") {
		t.Errorf("FighterFinish output does not contain duration: %s", output)
	}
}

func TestIssuesFound(t *testing.T) {
	tempDir := t.TempDir()
	l, err := New(tempDir, false)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer l.Close()

	var stdout bytes.Buffer
	l.SetOutputWriters(&stdout, &bytes.Buffer{})

	issues := []string{
		"Missing error handling in auth.go:45",
		"SQL injection vulnerability in users.go:23",
		"Unused variable in main.go:12",
	}

	l.IssuesFound(issues)

	output := stdout.String()
	if !strings.Contains(output, "3 issue(s)") {
		t.Errorf("IssuesFound output does not contain issue count: %s", output)
	}
	if !strings.Contains(output, "ISSUE 1") {
		t.Errorf("IssuesFound output does not contain 'ISSUE 1': %s", output)
	}
	if !strings.Contains(output, "Missing error handling") {
		t.Errorf("IssuesFound output does not contain first issue: %s", output)
	}
}

func TestNoIssues(t *testing.T) {
	tempDir := t.TempDir()
	l, err := New(tempDir, false)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer l.Close()

	var stdout bytes.Buffer
	l.SetOutputWriters(&stdout, &bytes.Buffer{})

	l.NoIssues()

	output := stdout.String()
	if !strings.Contains(output, "LGTM") {
		t.Errorf("NoIssues output does not contain 'LGTM': %s", output)
	}
	if !strings.Contains(output, "No issues found") {
		t.Errorf("NoIssues output does not contain 'No issues found': %s", output)
	}
}

func TestFinalVictory(t *testing.T) {
	tempDir := t.TempDir()
	l, err := New(tempDir, false)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer l.Close()

	var stdout bytes.Buffer
	l.SetOutputWriters(&stdout, &bytes.Buffer{})

	l.FinalVictory(3, 5*time.Minute)

	output := stdout.String()
	if !strings.Contains(output, "FLAWLESS VICTORY") {
		t.Errorf("FinalVictory output does not contain 'FLAWLESS VICTORY': %s", output)
	}
	if !strings.Contains(output, "Total Rounds:   3") {
		t.Errorf("FinalVictory output does not contain total rounds: %s", output)
	}
	if !strings.Contains(output, "5m") {
		t.Errorf("FinalVictory output does not contain duration: %s", output)
	}
	if !strings.Contains(output, "FINISH HIM") {
		t.Errorf("FinalVictory output does not contain 'FINISH HIM': %s", output)
	}
}

func TestError(t *testing.T) {
	tempDir := t.TempDir()
	l, err := New(tempDir, false)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer l.Close()

	var stderr bytes.Buffer
	l.SetOutputWriters(&bytes.Buffer{}, &stderr)

	testErr := errors.New("test error message")
	l.Error(testErr)

	output := stderr.String()
	if !strings.Contains(output, "ERROR") {
		t.Errorf("Error output does not contain 'ERROR': %s", output)
	}
	if !strings.Contains(output, "test error message") {
		t.Errorf("Error output does not contain error message: %s", output)
	}
}

func TestInfo(t *testing.T) {
	tempDir := t.TempDir()
	l, err := New(tempDir, false)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer l.Close()

	var stdout bytes.Buffer
	l.SetOutputWriters(&stdout, &bytes.Buffer{})

	l.Info("This is an info message")

	output := stdout.String()
	if !strings.Contains(output, "This is an info message") {
		t.Errorf("Info output does not contain message: %s", output)
	}
}

func TestDebug(t *testing.T) {
	t.Run("verbose mode enabled", func(t *testing.T) {
		tempDir := t.TempDir()
		l, err := New(tempDir, true) // verbose = true
		if err != nil {
			t.Fatalf("New() returned error: %v", err)
		}
		defer l.Close()

		var stdout bytes.Buffer
		l.SetOutputWriters(&stdout, &bytes.Buffer{})

		l.Debug("Debug message")

		output := stdout.String()
		if !strings.Contains(output, "[DEBUG]") {
			t.Errorf("Debug output does not contain '[DEBUG]': %s", output)
		}
		if !strings.Contains(output, "Debug message") {
			t.Errorf("Debug output does not contain message: %s", output)
		}
	})

	t.Run("verbose mode disabled", func(t *testing.T) {
		tempDir := t.TempDir()
		l, err := New(tempDir, false) // verbose = false
		if err != nil {
			t.Fatalf("New() returned error: %v", err)
		}
		defer l.Close()

		var stdout bytes.Buffer
		l.SetOutputWriters(&stdout, &bytes.Buffer{})

		l.Debug("Debug message")

		output := stdout.String()
		if output != "" {
			t.Errorf("Debug should not output when verbose is false, got: %s", output)
		}
	})
}

func TestFileLogging(t *testing.T) {
	tempDir := t.TempDir()
	l, err := New(tempDir, false)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	// Suppress terminal output
	l.SetOutputWriters(&bytes.Buffer{}, &bytes.Buffer{})

	// Perform various log operations
	l.RoundStart(1)
	l.FighterEnter("CLAUDE CODE")
	l.FighterFinish("CLAUDE CODE", 30*time.Second)
	l.IssuesFound([]string{"Test issue"})
	l.NoIssues()
	l.Info("Test info")
	l.Error(errors.New("test error"))

	logPath := l.GetLogFilePath()
	l.Close()

	// Read log file
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)

	// Verify log entries exist (without emojis)
	expectedEntries := []string{
		"ROUND 1 START",
		"CLAUDE CODE enters the arena",
		"CLAUDE CODE finished",
		"Issues found: 1",
		"ISSUE 1: Test issue",
		"LGTM - No issues found",
		"INFO: Test info",
		"ERROR: test error",
	}

	for _, entry := range expectedEntries {
		if !strings.Contains(logContent, entry) {
			t.Errorf("Log file does not contain '%s'. Content:\n%s", entry, logContent)
		}
	}

	// Verify log entries have timestamps
	if !strings.Contains(logContent, "[") {
		t.Error("Log entries do not appear to have timestamps")
	}
}

func TestClose(t *testing.T) {
	tempDir := t.TempDir()
	l, err := New(tempDir, false)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	logPath := l.GetLogFilePath()

	err = l.Close()
	if err != nil {
		t.Errorf("Close() returned error: %v", err)
	}

	// Verify file exists after close
	if _, err := os.Stat(logPath); err != nil {
		t.Errorf("Log file should still exist after close: %v", err)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{500 * time.Millisecond, "500ms"},
		{1 * time.Second, "1s"},
		{45 * time.Second, "45s"},
		{1 * time.Minute, "1m"},
		{5 * time.Minute, "5m"},
		{90 * time.Second, "1m 30s"},
		{5*time.Minute + 30*time.Second, "5m 30s"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatDuration(tt.duration)
			if result != tt.expected {
				t.Errorf("formatDuration(%v) = %s, want %s", tt.duration, result, tt.expected)
			}
		})
	}
}

func TestStripEmojis(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"\U0001F3AE MORTAL PROMPTER", " MORTAL PROMPTER"},
		{"\U0001F94A CLAUDE CODE enters", " CLAUDE CODE enters"},
		{"\u2705 Success!", " Success!"},
		{"No emojis here", "No emojis here"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := stripEmojis(tt.input)
			if result != tt.expected {
				t.Errorf("stripEmojis(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestChangesDetected(t *testing.T) {
	tempDir := t.TempDir()
	l, err := New(tempDir, false)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer l.Close()

	var stdout bytes.Buffer
	l.SetOutputWriters(&stdout, &bytes.Buffer{})

	l.ChangesDetected(5)

	output := stdout.String()
	if !strings.Contains(output, "5 file(s) modified") {
		t.Errorf("ChangesDetected output does not contain file count: %s", output)
	}
}

func TestPreparingNextRound(t *testing.T) {
	tempDir := t.TempDir()
	l, err := New(tempDir, false)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer l.Close()

	var stdout bytes.Buffer
	l.SetOutputWriters(&stdout, &bytes.Buffer{})

	l.PreparingNextRound()

	output := stdout.String()
	if !strings.Contains(output, "Preparing next round") {
		t.Errorf("PreparingNextRound output does not contain message: %s", output)
	}
}

func TestSpinnerStartStop(t *testing.T) {
	tempDir := t.TempDir()
	l, err := New(tempDir, false)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer l.Close()

	// Test that spinner methods don't panic
	l.StartSpinner("Loading...")

	// Give spinner a moment to start
	time.Sleep(50 * time.Millisecond)

	l.StopSpinner()

	// Test multiple stop calls don't panic
	l.StopSpinner()
	l.StopSpinner()
}

func TestGetters(t *testing.T) {
	tempDir := t.TempDir()
	outputDir := filepath.Join(tempDir, "custom-output")

	l, err := New(outputDir, true)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer l.Close()

	if l.GetOutputDir() != outputDir {
		t.Errorf("GetOutputDir() = %s, want %s", l.GetOutputDir(), outputDir)
	}

	if !l.IsVerbose() {
		t.Error("IsVerbose() should return true")
	}

	logPath := l.GetLogFilePath()
	if logPath == "" {
		t.Error("GetLogFilePath() returned empty string")
	}
	if !strings.HasPrefix(logPath, outputDir) {
		t.Errorf("Log file path should be in output directory: %s", logPath)
	}
}
