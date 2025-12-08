// Package reporter generates markdown battle reports for mortal-prompter sessions.
package reporter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/diegoram/mortal-prompter/pkg/types"
)

// Reporter generates markdown reports for battle sessions.
type Reporter struct {
	outputDir string
}

// New creates a new Reporter instance.
func New(outputDir string) *Reporter {
	return &Reporter{
		outputDir: outputDir,
	}
}

// GenerateReport creates a markdown battle report and writes it to a file.
// Returns the path to the generated report file.
func (r *Reporter) GenerateReport(result *types.SessionResult, initialPrompt string) (string, error) {
	// Ensure output directory exists
	if err := os.MkdirAll(r.outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate filename with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("report-%s.md", timestamp)
	filepath := filepath.Join(r.outputDir, filename)

	// Generate report content
	content := r.generateContent(result, initialPrompt)

	// Write to file
	if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write report file: %w", err)
	}

	return filepath, nil
}

// generateContent creates the markdown content for the report.
func (r *Reporter) generateContent(result *types.SessionResult, initialPrompt string) string {
	var sb strings.Builder

	// Header
	sb.WriteString("# Mortal Prompter - Battle Report\n\n")

	// Summary section
	r.writeSummary(&sb, result, initialPrompt)

	// Round history
	r.writeRoundHistory(&sb, result.Rounds)

	// Final changes
	r.writeFinalChanges(&sb, result)

	// Files modified
	r.writeFilesModified(&sb, result.FilesModified)

	return sb.String()
}

// writeSummary writes the summary section of the report.
func (r *Reporter) writeSummary(sb *strings.Builder, result *types.SessionResult, initialPrompt string) {
	sb.WriteString("## Summary\n\n")
	sb.WriteString(fmt.Sprintf("- **Initial Prompt:** %s\n", initialPrompt))
	sb.WriteString(fmt.Sprintf("- **Total Rounds:** %d\n", result.TotalRounds))
	sb.WriteString(fmt.Sprintf("- **Total Duration:** %s\n", formatDuration(result.TotalDuration)))

	if result.Success {
		sb.WriteString("- **Result:** SUCCESS - FLAWLESS VICTORY\n")
	} else {
		sb.WriteString("- **Result:** ABORTED\n")
	}
	sb.WriteString("\n")
}

// writeRoundHistory writes the detailed history of each round.
func (r *Reporter) writeRoundHistory(sb *strings.Builder, rounds []types.Round) {
	if len(rounds) == 0 {
		return
	}

	sb.WriteString("## Round History\n\n")

	for _, round := range rounds {
		sb.WriteString(fmt.Sprintf("### Round %d\n\n", round.Number))

		// Task description
		if round.Number == 1 {
			sb.WriteString(fmt.Sprintf("**Claude Code Task:** %s\n\n", truncatePrompt(round.ClaudePrompt, 200)))
		} else {
			sb.WriteString("**Claude Code Task:** Fix issues from previous review\n\n")
		}

		// Duration
		sb.WriteString(fmt.Sprintf("**Duration:** %s\n\n", formatDuration(round.Duration)))

		// Files changed
		filesChanged := countFilesInDiff(round.GitDiff)
		sb.WriteString(fmt.Sprintf("**Files Changed:** %d\n\n", filesChanged))

		// Review result
		if !round.HasIssues {
			sb.WriteString("**Codex Review:** LGTM - No issues found\n\n")
		} else {
			sb.WriteString(fmt.Sprintf("**Codex Review:** %d issue(s) found\n\n", len(round.Issues)))
			for i, issue := range round.Issues {
				sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, issue))
			}
			sb.WriteString("\n")
		}

		sb.WriteString("---\n\n")
	}
}

// writeFinalChanges writes the final git diff section.
func (r *Reporter) writeFinalChanges(sb *strings.Builder, result *types.SessionResult) {
	sb.WriteString("## Final Changes\n\n")

	if result.FinalDiff == "" {
		sb.WriteString("*No changes recorded*\n\n")
		return
	}

	// Truncate very long diffs
	diff := result.FinalDiff
	if len(diff) > 10000 {
		diff = diff[:10000] + "\n\n... (truncated, see git diff for full changes)"
	}

	sb.WriteString("```diff\n")
	sb.WriteString(diff)
	if !strings.HasSuffix(diff, "\n") {
		sb.WriteString("\n")
	}
	sb.WriteString("```\n\n")
}

// writeFilesModified writes the list of modified files.
func (r *Reporter) writeFilesModified(sb *strings.Builder, files []string) {
	sb.WriteString("## Files Modified\n\n")

	if len(files) == 0 {
		sb.WriteString("*No files modified*\n\n")
		return
	}

	for _, file := range files {
		sb.WriteString(fmt.Sprintf("- `%s`\n", file))
	}
	sb.WriteString("\n")
}

// formatDuration formats a duration in a human-readable way.
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	if seconds == 0 {
		return fmt.Sprintf("%dm", minutes)
	}
	return fmt.Sprintf("%dm %ds", minutes, seconds)
}

// truncatePrompt truncates a prompt to maxLen characters, adding ellipsis if needed.
func truncatePrompt(prompt string, maxLen int) string {
	// Remove newlines for cleaner display
	prompt = strings.ReplaceAll(prompt, "\n", " ")
	prompt = strings.Join(strings.Fields(prompt), " ") // Normalize whitespace

	if len(prompt) <= maxLen {
		return prompt
	}
	return prompt[:maxLen-3] + "..."
}

// countFilesInDiff counts the number of files in a git diff.
func countFilesInDiff(diff string) int {
	count := 0
	lines := strings.Split(diff, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git") {
			count++
		}
	}
	return count
}
