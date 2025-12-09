package fighters

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/diegoram/mortal-prompter/pkg/types"
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
// It uses `codex review --uncommitted` command.
func (c *Codex) Review(ctx context.Context, gitDiff string) (*types.ReviewResult, error) {
	// Check if codex is installed
	if _, err := exec.LookPath("codex"); err != nil {
		return nil, fmt.Errorf("codex CLI not found in PATH: %w", err)
	}

	// Create context with timeout if not already set
	execCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Build and execute the command using `codex review --uncommitted`
	cmd := exec.CommandContext(execCtx, "codex", "review", "--uncommitted")
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

// buildReviewPrompt constructs the review instructions for Codex (kept for testing).
func (c *Codex) buildReviewPrompt() string {
	return `Find real issues: bugs, vulnerabilities, bad practices, missing error handling. If NO issues respond "LGTM: No issues found". If issues found, list each as "ISSUE: [description]".`
}

// parseReviewOutput uses an LLM to intelligently parse the review output.
// This allows handling any review format without rigid pattern matching.
func (c *Codex) parseReviewOutput(output string) *types.ReviewResult {
	result := &types.ReviewResult{
		RawOutput: output,
		Issues:    []string{},
	}

	// Use the LLM to interpret the review output
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	parsePrompt := fmt.Sprintf(`Analyze this code review output and extract any issues found.

INSTRUCTIONS:
1. If the review indicates the code is good (LGTM, no issues, looks good, etc.), respond with exactly: NO_ISSUES
2. If there are issues, list each one on a separate line starting with "ISSUE: "
3. Be concise - just the issue description, no explanations

REVIEW OUTPUT:
%s

YOUR RESPONSE:`, output)

	// Execute the parsing prompt (no image for parsing)
	parseOutput, err := c.Execute(ctx, parsePrompt, "")
	if err != nil {
		// Fallback to simple heuristic if LLM call fails
		return c.parseReviewOutputFallback(output)
	}

	// Parse the LLM's structured response
	parseOutput = strings.TrimSpace(parseOutput)

	// Check for NO_ISSUES response
	if strings.Contains(strings.ToUpper(parseOutput), "NO_ISSUES") {
		result.HasIssues = false
		return result
	}

	// Extract issues from ISSUE: lines
	lines := strings.Split(parseOutput, "\n")
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
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

// parseReviewOutputFallback is a simple heuristic fallback when LLM parsing fails.
func (c *Codex) parseReviewOutputFallback(output string) *types.ReviewResult {
	result := &types.ReviewResult{
		RawOutput: output,
		Issues:    []string{},
	}

	outputLower := strings.ToLower(output)

	// Check for common "no issues" indicators
	noIssueIndicators := []string{"lgtm", "no issues", "looks good", "no problems", "code is clean"}
	for _, indicator := range noIssueIndicators {
		if strings.Contains(outputLower, indicator) {
			// But check if there are also issue indicators
			issueIndicators := []string{"[p1]", "[p2]", "[p3]", "issue:", "bug:", "error:", "problem:"}
			hasIssueIndicator := false
			for _, issueInd := range issueIndicators {
				if strings.Contains(outputLower, issueInd) {
					hasIssueIndicator = true
					break
				}
			}
			if !hasIssueIndicator {
				result.HasIssues = false
				return result
			}
		}
	}

	// If we get here, assume there might be issues - extract what we can
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		trimmedLower := strings.ToLower(trimmedLine)

		// Look for various issue patterns
		for _, prefix := range []string{"issue:", "bug:", "error:", "problem:", "[p1]", "[p2]", "[p3]", "[p4]", "- ["} {
			if strings.HasPrefix(trimmedLower, prefix) || strings.Contains(trimmedLine, "[P1]") || strings.Contains(trimmedLine, "[P2]") {
				result.Issues = append(result.Issues, trimmedLine)
				break
			}
		}
	}

	result.HasIssues = len(result.Issues) > 0
	return result
}

// Execute runs Codex CLI with the provided prompt and optional image path.
// It uses the context for timeout/cancellation support.
// If imagePath is provided, it is passed via the --image flag.
func (c *Codex) Execute(ctx context.Context, prompt string, imagePath string) (string, error) {
	// Check if codex is installed
	if _, err := exec.LookPath("codex"); err != nil {
		return "", fmt.Errorf("codex CLI not found in PATH: %w", err)
	}

	// Create context with timeout if not already set
	execCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Build command args
	args := []string{"-p", prompt}
	if imagePath != "" {
		// Codex uses --image flag for image input
		args = append(args, "--image", imagePath)
	}

	// Build and execute the command
	cmd := exec.CommandContext(execCtx, "codex", args...)
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
			return combinedOutput, fmt.Errorf("codex execution timed out after %v", c.timeout)
		}
		if execCtx.Err() == context.Canceled {
			return combinedOutput, fmt.Errorf("codex execution was cancelled")
		}
		return combinedOutput, fmt.Errorf("codex execution failed: %w", err)
	}

	return combinedOutput, nil
}

// BuildPromptWithIssues constructs a prompt for Codex that includes
// previous issues found during code review.
func (c *Codex) BuildPromptWithIssues(basePrompt string, previousIssues []string) string {
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

// WorkDir returns the working directory configured for this Codex instance.
func (c *Codex) WorkDir() string {
	return c.workDir
}

// Timeout returns the timeout configured for this Codex instance.
func (c *Codex) Timeout() time.Duration {
	return c.timeout
}
