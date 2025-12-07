package fighters

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/minimalart/mortal-prompter/pkg/types"
)

// Codex represents the Codex fighter (the reviewer).
// It wraps the codex CLI tool for performing code reviews.
type Codex struct {
	workDir string
	timeout time.Duration
}

// Ensure Codex implements the Fighter interface.
var _ Fighter = (*Codex)(nil)

// NewCodex creates a new Codex fighter instance.
// workDir specifies the working directory for command execution.
// timeout specifies the maximum duration for command execution.
func NewCodex(workDir string, timeout time.Duration) *Codex {
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	return &Codex{
		workDir: workDir,
		timeout: timeout,
	}
}

// Name returns the display name of the Codex fighter.
func (c *Codex) Name() string {
	return "CODEX"
}

// Review executes Codex to review a git diff and returns the parsed review result.
// It builds a review prompt with the diff, executes the codex CLI, and parses the output.
func (c *Codex) Review(ctx context.Context, gitDiff string) (*types.ReviewResult, error) {
	// Check if codex is installed
	if _, err := exec.LookPath("codex"); err != nil {
		return nil, fmt.Errorf("codex CLI not found in PATH: %w", err)
	}

	// Build the review prompt
	prompt := c.buildReviewPrompt(gitDiff)

	// Create context with timeout if not already set
	execCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Build and execute the command
	cmd := exec.CommandContext(execCtx, "codex", "-p", prompt)
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
			return nil, fmt.Errorf("codex execution timed out after %v", c.timeout)
		}
		if execCtx.Err() == context.Canceled {
			return nil, fmt.Errorf("codex execution was cancelled")
		}
		return nil, fmt.Errorf("codex execution failed: %w", err)
	}

	// Parse the output and return the review result
	return c.parseReviewOutput(combinedOutput), nil
}

// buildReviewPrompt constructs the review prompt for Codex with the git diff.
func (c *Codex) buildReviewPrompt(gitDiff string) string {
	return fmt.Sprintf(`Actua como un code reviewer senior extremadamente exigente.

Revisa el siguiente diff de git y encuentra TODOS los problemas:
- Bugs o errores logicos
- Vulnerabilidades de seguridad
- Malas practicas
- Codigo duplicado
- Falta de manejo de errores
- Problemas de performance
- Violaciones de convenciones de Go

GIT DIFF:
%sdifffence
%s
%sdifffence

INSTRUCCIONES DE RESPUESTA:
- Si NO encuentras issues, responde EXACTAMENTE: "LGTM: No issues found"
- Si encuentras issues, lista cada uno en formato:
  ISSUE: [descripcion del problema]

Se conciso y especifico. No incluyas sugerencias opcionales, solo problemas reales.`, "```", gitDiff, "```")
}

// parseReviewOutput parses the raw output from Codex to extract review results.
// It detects LGTM responses and extracts individual issues from ISSUE: prefixed lines.
func (c *Codex) parseReviewOutput(output string) *types.ReviewResult {
	result := &types.ReviewResult{
		RawOutput: output,
		Issues:    []string{},
	}

	// Normalize the output for case-insensitive matching
	outputLower := strings.ToLower(output)

	// Check for LGTM or "no issues found" indicators
	if strings.Contains(outputLower, "lgtm") || strings.Contains(outputLower, "no issues found") {
		result.HasIssues = false
		return result
	}

	// Parse lines looking for ISSUE: prefix
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Check for ISSUE: prefix (case-insensitive)
		if len(trimmedLine) >= 6 && strings.EqualFold(trimmedLine[:6], "ISSUE:") {
			issue := strings.TrimSpace(trimmedLine[6:])
			if issue != "" {
				result.Issues = append(result.Issues, issue)
			}
		}
	}

	result.HasIssues = len(result.Issues) > 0
	return result
}

// WorkDir returns the working directory configured for this Codex instance.
func (c *Codex) WorkDir() string {
	return c.workDir
}

// Timeout returns the timeout configured for this Codex instance.
func (c *Codex) Timeout() time.Duration {
	return c.timeout
}
