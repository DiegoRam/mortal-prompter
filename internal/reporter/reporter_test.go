package reporter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/minimalart/mortal-prompter/pkg/types"
)

func TestNew(t *testing.T) {
	r := New("/tmp/test-output")
	if r == nil {
		t.Fatal("New() returned nil")
	}
	if r.outputDir != "/tmp/test-output" {
		t.Errorf("outputDir = %s, want /tmp/test-output", r.outputDir)
	}
}

func TestGenerateReport(t *testing.T) {
	// Create temp directory
	tempDir := t.TempDir()

	r := New(tempDir)

	result := &types.SessionResult{
		Success:       true,
		TotalRounds:   2,
		TotalDuration: 5 * time.Minute,
		Rounds: []types.Round{
			{
				Number:       1,
				ClaudePrompt: "implement user authentication",
				GitDiff:      "diff --git a/auth.go b/auth.go\n+++ b/auth.go\n+package auth",
				HasIssues:    true,
				Issues:       []string{"Missing error handling", "No input validation"},
				Duration:     2 * time.Minute,
			},
			{
				Number:       2,
				ClaudePrompt: "fix issues",
				GitDiff:      "diff --git a/auth.go b/auth.go\n+++ b/auth.go\n+// fixed",
				HasIssues:    false,
				Issues:       []string{},
				Duration:     1 * time.Minute,
			},
		},
		FinalDiff:     "diff --git a/auth.go b/auth.go\n+++ b/auth.go\n+complete implementation",
		FilesModified: []string{"auth.go", "main.go"},
	}

	reportPath, err := r.GenerateReport(result, "implement user authentication")
	if err != nil {
		t.Fatalf("GenerateReport() error = %v", err)
	}

	// Check file was created
	if _, err := os.Stat(reportPath); os.IsNotExist(err) {
		t.Fatalf("Report file was not created at %s", reportPath)
	}

	// Check filename format
	if !strings.HasPrefix(filepath.Base(reportPath), "report-") {
		t.Errorf("Report filename should start with 'report-', got %s", filepath.Base(reportPath))
	}
	if !strings.HasSuffix(reportPath, ".md") {
		t.Errorf("Report filename should end with '.md', got %s", reportPath)
	}

	// Read and verify content
	content, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("Failed to read report file: %v", err)
	}

	contentStr := string(content)

	// Check required sections
	requiredSections := []string{
		"# Mortal Prompter - Battle Report",
		"## Summary",
		"## Round History",
		"## Final Changes",
		"## Files Modified",
	}

	for _, section := range requiredSections {
		if !strings.Contains(contentStr, section) {
			t.Errorf("Report missing section: %s", section)
		}
	}

	// Check content details
	if !strings.Contains(contentStr, "implement user authentication") {
		t.Error("Report should contain initial prompt")
	}
	if !strings.Contains(contentStr, "Total Rounds:** 2") {
		t.Error("Report should contain total rounds")
	}
	if !strings.Contains(contentStr, "SUCCESS") {
		t.Error("Report should indicate success")
	}
	if !strings.Contains(contentStr, "Missing error handling") {
		t.Error("Report should contain issues from round 1")
	}
	if !strings.Contains(contentStr, "`auth.go`") {
		t.Error("Report should list modified files")
	}
}

func TestGenerateReportFailure(t *testing.T) {
	tempDir := t.TempDir()
	r := New(tempDir)

	result := &types.SessionResult{
		Success:       false,
		TotalRounds:   1,
		TotalDuration: 1 * time.Minute,
		Rounds: []types.Round{
			{
				Number:    1,
				HasIssues: true,
				Issues:    []string{"Unresolved issue"},
			},
		},
	}

	reportPath, err := r.GenerateReport(result, "some task")
	if err != nil {
		t.Fatalf("GenerateReport() error = %v", err)
	}

	content, _ := os.ReadFile(reportPath)
	if !strings.Contains(string(content), "ABORTED") {
		t.Error("Failed session should show ABORTED result")
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

func TestTruncatePrompt(t *testing.T) {
	tests := []struct {
		name     string
		prompt   string
		maxLen   int
		expected string
	}{
		{
			name:     "short prompt unchanged",
			prompt:   "short prompt",
			maxLen:   50,
			expected: "short prompt",
		},
		{
			name:     "long prompt truncated",
			prompt:   "this is a very long prompt that should be truncated",
			maxLen:   20,
			expected: "this is a very lo...",
		},
		{
			name:     "newlines removed",
			prompt:   "line1\nline2\nline3",
			maxLen:   50,
			expected: "line1 line2 line3",
		},
		{
			name:     "whitespace normalized",
			prompt:   "word1   word2    word3",
			maxLen:   50,
			expected: "word1 word2 word3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncatePrompt(tt.prompt, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncatePrompt() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestCountFilesInDiff(t *testing.T) {
	tests := []struct {
		name     string
		diff     string
		expected int
	}{
		{
			name:     "empty diff",
			diff:     "",
			expected: 0,
		},
		{
			name: "single file",
			diff: `diff --git a/file.go b/file.go
+++ b/file.go`,
			expected: 1,
		},
		{
			name: "multiple files",
			diff: `diff --git a/file1.go b/file1.go
+++ b/file1.go
diff --git a/file2.go b/file2.go
+++ b/file2.go
diff --git a/file3.go b/file3.go
+++ b/file3.go`,
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countFilesInDiff(tt.diff)
			if result != tt.expected {
				t.Errorf("countFilesInDiff() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestGenerateReportCreatesDirectory(t *testing.T) {
	tempDir := t.TempDir()
	nestedDir := filepath.Join(tempDir, "nested", "output")

	r := New(nestedDir)

	result := &types.SessionResult{
		Success:     true,
		TotalRounds: 1,
		Rounds:      []types.Round{{Number: 1}},
	}

	_, err := r.GenerateReport(result, "test")
	if err != nil {
		t.Fatalf("GenerateReport() should create nested directories, got error: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(nestedDir); os.IsNotExist(err) {
		t.Error("Output directory was not created")
	}
}

func TestGenerateReportEmptyRounds(t *testing.T) {
	tempDir := t.TempDir()
	r := New(tempDir)

	result := &types.SessionResult{
		Success:       false,
		TotalRounds:   0,
		TotalDuration: 0,
		Rounds:        []types.Round{},
	}

	reportPath, err := r.GenerateReport(result, "test prompt")
	if err != nil {
		t.Fatalf("GenerateReport() error = %v", err)
	}

	content, _ := os.ReadFile(reportPath)
	contentStr := string(content)

	// Should still have basic structure
	if !strings.Contains(contentStr, "## Summary") {
		t.Error("Report should have summary even with no rounds")
	}
	if !strings.Contains(contentStr, "*No files modified*") {
		t.Error("Report should indicate no files modified")
	}
}
