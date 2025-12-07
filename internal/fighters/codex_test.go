package fighters

import (
	"strings"
	"testing"
	"time"
)

func TestNewCodex(t *testing.T) {
	tests := []struct {
		name            string
		workDir         string
		timeout         time.Duration
		expectedTimeout time.Duration
	}{
		{
			name:            "with valid timeout",
			workDir:         "/tmp/test",
			timeout:         10 * time.Minute,
			expectedTimeout: 10 * time.Minute,
		},
		{
			name:            "with zero timeout uses default",
			workDir:         "/tmp/test",
			timeout:         0,
			expectedTimeout: DefaultTimeout,
		},
		{
			name:            "with negative timeout uses default",
			workDir:         "/tmp/test",
			timeout:         -1 * time.Minute,
			expectedTimeout: DefaultTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			codex := NewCodex(tt.workDir, tt.timeout)

			if codex.WorkDir() != tt.workDir {
				t.Errorf("WorkDir() = %q, want %q", codex.WorkDir(), tt.workDir)
			}

			if codex.Timeout() != tt.expectedTimeout {
				t.Errorf("Timeout() = %v, want %v", codex.Timeout(), tt.expectedTimeout)
			}
		})
	}
}

func TestCodex_Name(t *testing.T) {
	codex := NewCodex("/tmp", 5*time.Minute)

	expected := "CODEX"
	if got := codex.Name(); got != expected {
		t.Errorf("Name() = %q, want %q", got, expected)
	}
}

func TestCodex_ImplementsFighter(t *testing.T) {
	// This test verifies that Codex implements the Fighter interface
	var _ Fighter = (*Codex)(nil)
}

func TestCodex_buildReviewPrompt(t *testing.T) {
	codex := NewCodex("/tmp", 5*time.Minute)

	result := codex.buildReviewPrompt()

	// Verify the prompt contains expected elements
	expectedContents := []string{
		"LGTM: No issues found",
		"ISSUE:",
	}

	for _, expected := range expectedContents {
		if !strings.Contains(result, expected) {
			t.Errorf("buildReviewPrompt() should contain %q", expected)
		}
	}
}

func TestCodex_parseReviewOutput_LGTM(t *testing.T) {
	codex := NewCodex("/tmp", 5*time.Minute)

	tests := []struct {
		name   string
		output string
	}{
		{
			name:   "exact LGTM response",
			output: "LGTM: No issues found",
		},
		{
			name:   "LGTM lowercase",
			output: "lgtm: no issues found",
		},
		{
			name:   "contains LGTM",
			output: "After reviewing the code, LGTM - looks good to merge.",
		},
		{
			name:   "no issues found phrase",
			output: "I reviewed the diff and found no issues found in the code.",
		},
		{
			name:   "LGTM with extra text",
			output: "The code looks clean.\nLGTM: No issues found\nGood job!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := codex.parseReviewOutput(tt.output)

			if result.HasIssues {
				t.Errorf("parseReviewOutput(%q) HasIssues = true, want false", tt.output)
			}

			if len(result.Issues) != 0 {
				t.Errorf("parseReviewOutput(%q) Issues = %v, want empty", tt.output, result.Issues)
			}

			if result.RawOutput != tt.output {
				t.Errorf("parseReviewOutput() RawOutput = %q, want %q", result.RawOutput, tt.output)
			}
		})
	}
}

func TestCodex_parseReviewOutput_WithIssues(t *testing.T) {
	codex := NewCodex("/tmp", 5*time.Minute)

	tests := []struct {
		name           string
		output         string
		expectedIssues []string
	}{
		{
			name:   "single issue",
			output: "ISSUE: Missing error handling in auth.go:45",
			expectedIssues: []string{
				"Missing error handling in auth.go:45",
			},
		},
		{
			name: "multiple issues",
			output: `Here are the issues I found:
ISSUE: Missing error handling in auth.go:45
ISSUE: SQL injection vulnerability in users.go:23
ISSUE: Unused variable in main.go:12`,
			expectedIssues: []string{
				"Missing error handling in auth.go:45",
				"SQL injection vulnerability in users.go:23",
				"Unused variable in main.go:12",
			},
		},
		{
			name: "issues with mixed case prefix",
			output: `Issue: First problem
ISSUE: Second problem
issue: Third problem`,
			expectedIssues: []string{
				"First problem",
				"Second problem",
				"Third problem",
			},
		},
		{
			name: "issues with extra whitespace",
			output: `   ISSUE:   Whitespace before and after
ISSUE:Normal issue`,
			expectedIssues: []string{
				"Whitespace before and after",
				"Normal issue",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := codex.parseReviewOutput(tt.output)

			if !result.HasIssues {
				t.Errorf("parseReviewOutput(%q) HasIssues = false, want true", tt.output)
			}

			if len(result.Issues) != len(tt.expectedIssues) {
				t.Errorf("parseReviewOutput() got %d issues, want %d", len(result.Issues), len(tt.expectedIssues))
				t.Logf("Got issues: %v", result.Issues)
				t.Logf("Want issues: %v", tt.expectedIssues)
				return
			}

			for i, expected := range tt.expectedIssues {
				if result.Issues[i] != expected {
					t.Errorf("parseReviewOutput() Issues[%d] = %q, want %q", i, result.Issues[i], expected)
				}
			}
		})
	}
}

func TestCodex_parseReviewOutput_MixedOutput(t *testing.T) {
	codex := NewCodex("/tmp", 5*time.Minute)

	// Test output with various non-issue lines mixed in
	output := `I've reviewed the code and found the following problems:

Some general observations about the code structure...

ISSUE: The function lacks input validation

More commentary here...

ISSUE: Race condition in concurrent access

Final thoughts: Please fix these issues.`

	result := codex.parseReviewOutput(output)

	if !result.HasIssues {
		t.Error("parseReviewOutput() HasIssues = false, want true")
	}

	expectedIssues := []string{
		"The function lacks input validation",
		"Race condition in concurrent access",
	}

	if len(result.Issues) != len(expectedIssues) {
		t.Errorf("parseReviewOutput() got %d issues, want %d", len(result.Issues), len(expectedIssues))
		return
	}

	for i, expected := range expectedIssues {
		if result.Issues[i] != expected {
			t.Errorf("parseReviewOutput() Issues[%d] = %q, want %q", i, result.Issues[i], expected)
		}
	}
}

func TestCodex_parseReviewOutput_EmptyIssue(t *testing.T) {
	codex := NewCodex("/tmp", 5*time.Minute)

	// Test that empty ISSUE: lines are ignored
	output := `ISSUE: Valid issue
ISSUE:
ISSUE:
ISSUE: Another valid issue`

	result := codex.parseReviewOutput(output)

	if !result.HasIssues {
		t.Error("parseReviewOutput() HasIssues = false, want true")
	}

	expectedIssues := []string{
		"Valid issue",
		"Another valid issue",
	}

	if len(result.Issues) != len(expectedIssues) {
		t.Errorf("parseReviewOutput() got %d issues, want %d", len(result.Issues), len(expectedIssues))
		t.Logf("Got: %v", result.Issues)
	}
}

func TestCodex_parseReviewOutput_NoIssuesNoLGTM(t *testing.T) {
	codex := NewCodex("/tmp", 5*time.Minute)

	// If output has no LGTM and no ISSUE: lines, it should have no issues
	output := "The code looks fine to me. Nice work!"

	result := codex.parseReviewOutput(output)

	if result.HasIssues {
		t.Error("parseReviewOutput() HasIssues = true, want false (no ISSUE: lines found)")
	}

	if len(result.Issues) != 0 {
		t.Errorf("parseReviewOutput() Issues = %v, want empty", result.Issues)
	}
}

func TestCodex_parseReviewOutput_RawOutputPreserved(t *testing.T) {
	codex := NewCodex("/tmp", 5*time.Minute)

	output := "ISSUE: Something wrong\nMore text here"

	result := codex.parseReviewOutput(output)

	if result.RawOutput != output {
		t.Errorf("parseReviewOutput() RawOutput = %q, want %q", result.RawOutput, output)
	}
}
