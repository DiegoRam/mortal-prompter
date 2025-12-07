package fighters

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// DefaultTimeout is the default timeout for Claude CLI execution.
const DefaultTimeout = 5 * time.Minute

// Claude represents the Claude Code fighter (the implementer).
// It wraps the claude CLI tool for executing development tasks.
type Claude struct {
	workDir string
	timeout time.Duration
}

// Ensure Claude implements the Fighter interface.
var _ Fighter = (*Claude)(nil)

// NewClaude creates a new Claude fighter instance.
// workDir specifies the working directory for command execution.
// timeout specifies the maximum duration for command execution.
func NewClaude(workDir string, timeout time.Duration) *Claude {
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	return &Claude{
		workDir: workDir,
		timeout: timeout,
	}
}

// Name returns the display name of the Claude fighter.
func (c *Claude) Name() string {
	return "CLAUDE CODE"
}

// Execute runs Claude Code CLI with the provided prompt and returns the output.
// It uses the context for timeout/cancellation support.
// The command executed is: claude -p "<prompt>" --dangerously-skip-permissions
func (c *Claude) Execute(ctx context.Context, prompt string) (string, error) {
	// Check if claude is installed
	if _, err := exec.LookPath("claude"); err != nil {
		return "", fmt.Errorf("claude CLI not found in PATH: %w", err)
	}

	// Create context with timeout if not already set
	execCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Build and execute the command
	cmd := exec.CommandContext(execCtx, "claude", "-p", prompt, "--dangerously-skip-permissions")
	cmd.Dir = c.workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Combine stdout and stderr for complete output
	combinedOutput := stdout.String()
	if stderr.Len() > 0 {
		if combinedOutput != "" {
			combinedOutput += "\n"
		}
		combinedOutput += stderr.String()
	}

	if err != nil {
		// Check if context was cancelled or timed out
		if execCtx.Err() == context.DeadlineExceeded {
			return combinedOutput, fmt.Errorf("claude execution timed out after %v", c.timeout)
		}
		if execCtx.Err() == context.Canceled {
			return combinedOutput, fmt.Errorf("claude execution was cancelled")
		}
		return combinedOutput, fmt.Errorf("claude execution failed: %w", err)
	}

	return combinedOutput, nil
}

// BuildPromptWithIssues constructs a prompt for Claude that includes
// previous issues found during code review.
// If there are no previous issues, it returns the basePrompt as-is.
// If there are issues, it builds a structured prompt requesting corrections.
func (c *Claude) BuildPromptWithIssues(basePrompt string, previousIssues []string) string {
	if len(previousIssues) == 0 {
		return basePrompt
	}

	var sb strings.Builder
	sb.WriteString("CONTEXTO: Estas en una sesion de code review iterativo.\n\n")
	sb.WriteString("ISSUES ENCONTRADOS EN LA REVISION ANTERIOR:\n")

	for _, issue := range previousIssues {
		sb.WriteString("- ")
		sb.WriteString(issue)
		sb.WriteString("\n")
	}

	sb.WriteString("\nTAREA: Corrige los issues mencionados arriba.\n")
	sb.WriteString("No expliques los cambios, solo implementa las correcciones.\n")

	return sb.String()
}

// WorkDir returns the working directory configured for this Claude instance.
func (c *Claude) WorkDir() string {
	return c.workDir
}

// Timeout returns the timeout configured for this Claude instance.
func (c *Claude) Timeout() time.Duration {
	return c.timeout
}
