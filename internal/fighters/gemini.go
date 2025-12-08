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

// Gemini represents the Gemini CLI fighter.
// It can act as both implementer and reviewer via the gemini CLI tool.
type Gemini struct {
	workDir string
	timeout time.Duration
}

// Ensure Gemini implements the Fighter interface.
var _ Fighter = (*Gemini)(nil)

// NewGemini creates a new Gemini fighter instance.
// workDir specifies the working directory for command execution.
// timeout specifies the maximum duration for command execution.
func NewGemini(workDir string, timeout time.Duration) *Gemini {
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	return &Gemini{
		workDir: workDir,
		timeout: timeout,
	}
}

// Name returns the display name of the Gemini fighter.
func (g *Gemini) Name() string {
	return "GEMINI"
}

// Execute runs Gemini CLI with the provided prompt and returns the output.
// It uses the context for timeout/cancellation support.
// The command executed is: gemini -p "<prompt>"
func (g *Gemini) Execute(ctx context.Context, prompt string) (string, error) {
	// Check if gemini is installed
	if _, err := exec.LookPath("gemini"); err != nil {
		return "", fmt.Errorf("gemini CLI not found in PATH: %w", err)
	}

	// Create context with timeout if not already set
	execCtx, cancel := context.WithTimeout(ctx, g.timeout)
	defer cancel()

	// Build and execute the command
	cmd := exec.CommandContext(execCtx, "gemini", "-p", prompt)
	cmd.Dir = g.workDir

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
			return combinedOutput, fmt.Errorf("gemini execution timed out after %v", g.timeout)
		}
		if execCtx.Err() == context.Canceled {
			return combinedOutput, fmt.Errorf("gemini execution was cancelled")
		}
		return combinedOutput, fmt.Errorf("gemini execution failed: %w", err)
	}

	return combinedOutput, nil
}

// Review executes Gemini to review a git diff and returns the parsed review result.
// It sends the diff as a prompt asking for code review.
func (g *Gemini) Review(ctx context.Context, gitDiff string) (*types.ReviewResult, error) {
	// Check if gemini is installed
	if _, err := exec.LookPath("gemini"); err != nil {
		return nil, fmt.Errorf("gemini CLI not found in PATH: %w", err)
	}

	// Create context with timeout if not already set
	execCtx, cancel := context.WithTimeout(ctx, g.timeout)
	defer cancel()

	// Build the review prompt
	reviewPrompt := g.buildReviewPrompt(gitDiff)

	// Build and execute the command
	cmd := exec.CommandContext(execCtx, "gemini", "-p", reviewPrompt)
	cmd.Dir = g.workDir

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
			return nil, fmt.Errorf("gemini execution timed out after %v", g.timeout)
		}
		if execCtx.Err() == context.Canceled {
			return nil, fmt.Errorf("gemini execution was cancelled")
		}
		return nil, fmt.Errorf("gemini execution failed: %w", err)
	}

	// Parse the output and return the review result
	return g.parseReviewOutput(combinedOutput), nil
}

// buildReviewPrompt constructs the review prompt for Gemini.
func (g *Gemini) buildReviewPrompt(gitDiff string) string {
	return fmt.Sprintf(`Review the following git diff for issues.
Find real issues: bugs, vulnerabilities, bad practices, missing error handling.
If NO issues respond "LGTM: No issues found".
If issues found, list each as "ISSUE: [description]".

Git diff:
%s`, gitDiff)
}

// parseReviewOutput uses an LLM to intelligently parse the review output.
// This allows handling any review format without rigid pattern matching.
func (g *Gemini) parseReviewOutput(output string) *types.ReviewResult {
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

	// Execute the parsing prompt
	parseOutput, err := g.Execute(ctx, parsePrompt)
	if err != nil {
		// Fallback to simple heuristic if LLM call fails
		return g.parseReviewOutputFallback(output)
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
func (g *Gemini) parseReviewOutputFallback(output string) *types.ReviewResult {
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

// BuildPromptWithIssues constructs a prompt for Gemini that includes
// previous issues found during code review.
// If there are no previous issues, it returns the basePrompt as-is.
// If there are issues, it builds a structured prompt requesting corrections.
func (g *Gemini) BuildPromptWithIssues(basePrompt string, previousIssues []string) string {
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

// WorkDir returns the working directory configured for this Gemini instance.
func (g *Gemini) WorkDir() string {
	return g.workDir
}

// Timeout returns the timeout configured for this Gemini instance.
func (g *Gemini) Timeout() time.Duration {
	return g.timeout
}
