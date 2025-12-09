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

// Execute runs Claude Code CLI with the provided prompt and optional image path.
// It uses the context for timeout/cancellation support.
// The command executed is: claude -p "<prompt>" --dangerously-skip-permissions
// If imagePath is provided, it is included in the prompt for Claude to analyze.
func (c *Claude) Execute(ctx context.Context, prompt string, imagePath string) (string, error) {
	// Check if claude is installed
	if _, err := exec.LookPath("claude"); err != nil {
		return "", fmt.Errorf("claude CLI not found in PATH: %w", err)
	}

	// Create context with timeout if not already set
	execCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// If image path is provided, include it in the prompt
	// Claude Code can read images when provided as file paths
	finalPrompt := prompt
	if imagePath != "" {
		finalPrompt = fmt.Sprintf("%s\n\n[Image attached: %s]", prompt, imagePath)
	}

	// Build and execute the command
	cmd := exec.CommandContext(execCtx, "claude", "-p", finalPrompt, "--dangerously-skip-permissions")
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

// Review executes Claude to review a git diff and returns the parsed review result.
// It sends the diff as a prompt asking for code review.
func (c *Claude) Review(ctx context.Context, gitDiff string) (*types.ReviewResult, error) {
	// Build the review prompt
	reviewPrompt := c.buildReviewPrompt(gitDiff)

	// Execute Claude with the review prompt (no image for reviews)
	output, err := c.Execute(ctx, reviewPrompt, "")
	if err != nil {
		return nil, err
	}

	// Parse the output and return the review result
	return c.parseReviewOutput(output), nil
}

// buildReviewPrompt constructs the review prompt for Claude.
func (c *Claude) buildReviewPrompt(gitDiff string) string {
	return fmt.Sprintf(`Review the following git diff for issues.
Find real issues: bugs, vulnerabilities, bad practices, missing error handling.
If NO issues respond "LGTM: No issues found".
If issues found, list each as "ISSUE: [description]".

Git diff:
%s`, gitDiff)
}

// parseReviewOutput uses an LLM to intelligently parse the review output.
// This allows handling any review format without rigid pattern matching.
func (c *Claude) parseReviewOutput(output string) *types.ReviewResult {
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
func (c *Claude) parseReviewOutputFallback(output string) *types.ReviewResult {
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

// WorkDir returns the working directory configured for this Claude instance.
func (c *Claude) WorkDir() string {
	return c.workDir
}

// Timeout returns the timeout configured for this Claude instance.
func (c *Claude) Timeout() time.Duration {
	return c.timeout
}
