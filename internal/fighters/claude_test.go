package fighters

import (
	"strings"
	"testing"
	"time"
)

func TestNewClaude(t *testing.T) {
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
			claude := NewClaude(tt.workDir, tt.timeout)

			if claude.WorkDir() != tt.workDir {
				t.Errorf("WorkDir() = %q, want %q", claude.WorkDir(), tt.workDir)
			}

			if claude.Timeout() != tt.expectedTimeout {
				t.Errorf("Timeout() = %v, want %v", claude.Timeout(), tt.expectedTimeout)
			}
		})
	}
}

func TestClaude_Name(t *testing.T) {
	claude := NewClaude("/tmp", 5*time.Minute)

	expected := "CLAUDE CODE"
	if got := claude.Name(); got != expected {
		t.Errorf("Name() = %q, want %q", got, expected)
	}
}

func TestClaude_ImplementsFighter(t *testing.T) {
	// This test verifies that Claude implements the Fighter interface
	var _ Fighter = (*Claude)(nil)
}

func TestClaude_BuildPromptWithIssues_NoIssues(t *testing.T) {
	claude := NewClaude("/tmp", 5*time.Minute)

	basePrompt := "Implement JWT authentication"
	result := claude.BuildPromptWithIssues(basePrompt, nil)

	if result != basePrompt {
		t.Errorf("BuildPromptWithIssues() with nil issues = %q, want %q", result, basePrompt)
	}

	result = claude.BuildPromptWithIssues(basePrompt, []string{})

	if result != basePrompt {
		t.Errorf("BuildPromptWithIssues() with empty issues = %q, want %q", result, basePrompt)
	}
}

func TestClaude_BuildPromptWithIssues_WithIssues(t *testing.T) {
	claude := NewClaude("/tmp", 5*time.Minute)

	basePrompt := "Implement JWT authentication"
	issues := []string{
		"Missing error handling in auth.go:45",
		"SQL injection vulnerability in users.go:23",
		"Unused variable in main.go:12",
	}

	result := claude.BuildPromptWithIssues(basePrompt, issues)

	// Verify the result contains the expected structure
	if !strings.Contains(result, "CONTEXTO:") {
		t.Error("BuildPromptWithIssues() should contain 'CONTEXTO:'")
	}

	if !strings.Contains(result, "ISSUES ENCONTRADOS EN LA REVISION ANTERIOR:") {
		t.Error("BuildPromptWithIssues() should contain issues header")
	}

	if !strings.Contains(result, "TAREA: Corrige los issues mencionados arriba.") {
		t.Error("BuildPromptWithIssues() should contain task instruction")
	}

	// Verify all issues are included
	for _, issue := range issues {
		if !strings.Contains(result, "- "+issue) {
			t.Errorf("BuildPromptWithIssues() should contain issue: %q", issue)
		}
	}

	// Verify the base prompt is NOT included when there are issues
	if strings.Contains(result, basePrompt) {
		t.Error("BuildPromptWithIssues() should NOT contain base prompt when issues exist")
	}
}

func TestClaude_BuildPromptWithIssues_SingleIssue(t *testing.T) {
	claude := NewClaude("/tmp", 5*time.Minute)

	issues := []string{"Critical bug in handler"}
	result := claude.BuildPromptWithIssues("base prompt", issues)

	if !strings.Contains(result, "- Critical bug in handler") {
		t.Error("BuildPromptWithIssues() should contain the single issue")
	}

	// Count occurrences of "- " to ensure only one issue is listed
	count := strings.Count(result, "\n- ")
	if count != 1 {
		t.Errorf("BuildPromptWithIssues() should have exactly 1 issue, got %d", count)
	}
}

func TestClaude_BuildPromptWithIssues_StructureIntegrity(t *testing.T) {
	claude := NewClaude("/tmp", 5*time.Minute)

	issues := []string{"Issue 1", "Issue 2"}
	result := claude.BuildPromptWithIssues("base", issues)

	// Verify the prompt has proper structure (order of sections)
	contextoIdx := strings.Index(result, "CONTEXTO:")
	issuesIdx := strings.Index(result, "ISSUES ENCONTRADOS")
	tareaIdx := strings.Index(result, "TAREA:")

	if contextoIdx == -1 || issuesIdx == -1 || tareaIdx == -1 {
		t.Fatal("BuildPromptWithIssues() missing required sections")
	}

	if contextoIdx > issuesIdx {
		t.Error("CONTEXTO should come before ISSUES")
	}

	if issuesIdx > tareaIdx {
		t.Error("ISSUES should come before TAREA")
	}
}
