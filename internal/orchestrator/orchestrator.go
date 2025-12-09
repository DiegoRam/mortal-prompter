// Package orchestrator implements the main battle loop between Claude Code and Codex.
// It manages the iterative development and code review cycle.
package orchestrator

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/diegoram/mortal-prompter/internal/config"
	"github.com/diegoram/mortal-prompter/internal/fighters"
	"github.com/diegoram/mortal-prompter/internal/git"
	"github.com/diegoram/mortal-prompter/internal/logger"
	"github.com/diegoram/mortal-prompter/pkg/types"
)

// Observer is the interface for receiving orchestrator events
type Observer interface {
	OnRoundStart(number int)
	OnFighterEnter(fighter string)
	OnFighterAction(fighter, action string)
	OnFighterFinish(fighter string, duration time.Duration)
	OnChangesDetected(fileCount int)
	OnIssuesFound(issues []string)
	OnNoIssues()
	OnSessionComplete(result *types.SessionResult, success bool)
	OnError(err error)
	OnConfirmationRequired(message string) bool
}

// Orchestrator manages the code review battle between Claude Code and Codex.
type Orchestrator struct {
	config *config.Config
	claude *fighters.Claude
	codex  *fighters.Codex
	git    *git.Git
	logger *logger.Logger

	// Observer for TUI updates (optional)
	observer Observer

	// Session state
	rounds       []types.Round
	currentRound int
	state        types.SessionState
	startTime    time.Time

	// Image path for multimodal prompts (only used in first round)
	imagePath string
}

// New creates a new Orchestrator instance with the provided configuration and logger.
func New(cfg *config.Config, log *logger.Logger) *Orchestrator {
	return &Orchestrator{
		config:       cfg,
		claude:       fighters.NewClaude(cfg.WorkDir, fighters.DefaultTimeout),
		codex:        fighters.NewCodex(cfg.WorkDir, fighters.DefaultTimeout),
		git:          git.New(cfg.WorkDir),
		logger:       log,
		rounds:       make([]types.Round, 0),
		currentRound: 0,
		state:        types.StateInitializing,
	}
}

// NewWithObserver creates a new Orchestrator instance with an observer for TUI updates.
func NewWithObserver(cfg *config.Config, log *logger.Logger, observer Observer) *Orchestrator {
	o := New(cfg, log)
	o.observer = observer
	return o
}

// SetPrompt allows setting the prompt after creation (for TUI mode)
func (o *Orchestrator) SetPrompt(prompt string) {
	o.config.Prompt = prompt
}

// SetImagePath sets the image path for multimodal prompts (only used in first round)
func (o *Orchestrator) SetImagePath(imagePath string) {
	o.imagePath = imagePath
}

// Run executes the main battle loop and returns the session result.
// The loop continues until:
// - Codex finds no issues (LGTM) -> Success
// - Max iterations reached and user declines to continue -> Aborted
// - An error occurs -> Failed
func (o *Orchestrator) Run(ctx context.Context) (*types.SessionResult, error) {
	o.startTime = time.Now()
	o.state = types.StateRunning

	// Verify this is a git repository
	if !o.git.IsGitRepo() {
		o.state = types.StateFailed
		err := fmt.Errorf("working directory is not a git repository: %s", o.config.WorkDir)
		o.notifyError(err)
		return nil, err
	}

	currentPrompt := o.config.Prompt
	var previousIssues []string

	for {
		select {
		case <-ctx.Done():
			o.state = types.StateAborted
			result := o.buildResult(false)
			o.notifySessionComplete(result, false)
			return result, ctx.Err()
		default:
		}

		o.currentRound++

		// Notify round start
		o.notifyRoundStart(o.currentRound)

		// Check if we've hit max iterations
		if o.currentRound > o.config.MaxIterations {
			o.state = types.StateWaitingConfirmation
			if !o.promptContinue() {
				o.state = types.StateAborted
				if o.logger != nil {
					o.logger.Info("Session aborted by user after max iterations")
				}
				result := o.buildResult(false)
				o.notifySessionComplete(result, false)
				return result, nil
			}
			o.state = types.StateRunning
		}

		// Execute round
		round, err := o.executeRound(ctx, o.currentRound, currentPrompt, previousIssues)
		if err != nil {
			o.state = types.StateFailed
			if o.logger != nil {
				o.logger.Error(err)
			}
			o.notifyError(err)
			return o.buildResult(false), err
		}

		o.rounds = append(o.rounds, *round)

		// Check if we're done (no issues found)
		if !round.HasIssues {
			o.state = types.StateCompleted
			if o.logger != nil {
				o.logger.NoIssues()
				o.logger.FinalVictory(len(o.rounds), time.Since(o.startTime))
			}
			o.notifyNoIssues()

			// Auto-commit if enabled
			if o.config.AutoCommit {
				if err := o.autoCommit(); err != nil {
					if o.logger != nil {
						o.logger.Error(fmt.Errorf("auto-commit failed: %w", err))
					}
				}
			}

			result := o.buildResult(true)
			o.notifySessionComplete(result, true)
			return result, nil
		}

		// Prepare for next round
		if o.logger != nil {
			o.logger.IssuesFound(round.Issues)
			o.logger.PreparingNextRound()
		}
		o.notifyIssuesFound(round.Issues)

		previousIssues = round.Issues
		currentPrompt = o.config.Prompt // Base prompt stays the same, issues are added by BuildPromptWithIssues

		// Interactive mode: ask before each round
		if o.config.Interactive && o.currentRound < o.config.MaxIterations {
			o.state = types.StateWaitingConfirmation
			if !o.promptNextRound() {
				o.state = types.StateAborted
				if o.logger != nil {
					o.logger.Info("Session aborted by user")
				}
				result := o.buildResult(false)
				o.notifySessionComplete(result, false)
				return result, nil
			}
			o.state = types.StateRunning
		}
	}
}

// executeRound runs a single round of the battle.
func (o *Orchestrator) executeRound(ctx context.Context, number int, basePrompt string, previousIssues []string) (*types.Round, error) {
	roundStart := time.Now()

	round := &types.Round{
		Number:    number,
		Timestamp: roundStart,
	}

	if o.logger != nil {
		o.logger.RoundStart(number)
	}

	// Build the prompt (includes issues if any)
	prompt := o.claude.BuildPromptWithIssues(basePrompt, previousIssues)
	round.ClaudePrompt = prompt

	// Execute Claude
	if o.logger != nil {
		o.logger.FighterEnter(o.claude.Name())
		o.logger.FighterAction("Claude Code implementing changes...")
		o.logger.CLIInput(o.claude.Name(), prompt)
	}
	o.notifyFighterEnter(o.claude.Name())
	o.notifyFighterAction("Claude Code", "Implementing changes...")

	// Only pass image path on first round (subsequent rounds focus on issues)
	imagePath := ""
	if number == 1 && o.imagePath != "" {
		imagePath = o.imagePath
		if o.logger != nil {
			o.logger.Info(fmt.Sprintf("Including image: %s", imagePath))
		}
	}

	claudeStart := time.Now()
	claudeOutput, err := o.claude.Execute(ctx, prompt, imagePath)
	claudeDuration := time.Since(claudeStart)

	if err != nil {
		if o.logger != nil {
			o.logger.CLIOutput(o.claude.Name(), claudeOutput)
		}
		return nil, fmt.Errorf("claude execution failed: %w", err)
	}

	round.ClaudeOutput = claudeOutput
	if o.logger != nil {
		o.logger.CLIOutput(o.claude.Name(), claudeOutput)
		o.logger.FighterFinish(o.claude.Name(), claudeDuration)
	}
	o.notifyFighterFinish(o.claude.Name(), claudeDuration)

	// Get git diff
	if o.logger != nil {
		o.logger.FighterAction("Capturing git diff...")
	}
	o.notifyFighterAction("Claude Code", "Capturing git diff...")

	// Stage all changes first to capture everything
	if err := o.git.StageAll(); err != nil {
		return nil, fmt.Errorf("failed to stage changes: %w", err)
	}

	diff, err := o.git.GetStagedDiff()
	if err != nil {
		return nil, fmt.Errorf("failed to get git diff: %w", err)
	}

	round.GitDiff = diff

	// Log the git diff
	if o.logger != nil {
		o.logger.GitDiff(diff)
	}

	// Check if there are any changes
	if strings.TrimSpace(diff) == "" {
		if o.logger != nil {
			o.logger.Info("No changes detected in this round")
		}
		// If no changes, we consider it as no issues (nothing to review)
		round.HasIssues = false
		round.Duration = time.Since(roundStart)
		return round, nil
	}

	// Log changes detected
	fileCount := countFilesInDiff(diff)
	if o.logger != nil {
		o.logger.ChangesDetected(fileCount)
	}
	o.notifyChangesDetected(fileCount)

	// Execute Codex review
	if o.logger != nil {
		o.logger.FighterEnter(o.codex.Name())
		o.logger.FighterAction("Codex reviewing changes...")
		o.logger.CLIInput(o.codex.Name(), "codex review --uncommitted")
	}
	o.notifyFighterEnter(o.codex.Name())
	o.notifyFighterAction("Codex", "Reviewing changes...")

	codexStart := time.Now()
	reviewResult, err := o.codex.Review(ctx, diff)
	codexDuration := time.Since(codexStart)

	if err != nil {
		return nil, fmt.Errorf("codex review failed: %w", err)
	}

	if o.logger != nil {
		o.logger.CLIOutput(o.codex.Name(), reviewResult.RawOutput)
		o.logger.FighterFinish(o.codex.Name(), codexDuration)
	}
	o.notifyFighterFinish(o.codex.Name(), codexDuration)

	round.CodexReview = reviewResult.RawOutput
	round.HasIssues = reviewResult.HasIssues
	round.Issues = reviewResult.Issues
	round.Duration = time.Since(roundStart)

	return round, nil
}

// buildResult constructs the final SessionResult.
func (o *Orchestrator) buildResult(success bool) *types.SessionResult {
	result := &types.SessionResult{
		Success:       success,
		TotalRounds:   len(o.rounds),
		TotalDuration: time.Since(o.startTime),
		Rounds:        o.rounds,
	}

	// Get final diff (all changes combined)
	if diff, err := o.git.GetStagedDiff(); err == nil {
		result.FinalDiff = diff
	}

	// Extract modified files from rounds
	filesMap := make(map[string]bool)
	for _, round := range o.rounds {
		for _, file := range extractFilesFromDiff(round.GitDiff) {
			filesMap[file] = true
		}
	}

	result.FilesModified = make([]string, 0, len(filesMap))
	for file := range filesMap {
		result.FilesModified = append(result.FilesModified, file)
	}

	return result
}

// autoCommit creates a git commit with the configured message.
func (o *Orchestrator) autoCommit() error {
	o.logger.Info("Auto-committing changes...")

	message := fmt.Sprintf("%s\n\nMortal Prompter session:\n- Rounds: %d\n- Duration: %s",
		o.config.CommitMessage,
		len(o.rounds),
		time.Since(o.startTime).Round(time.Second),
	)

	if err := o.git.Commit(message); err != nil {
		if err == git.ErrNoChanges {
			o.logger.Info("No changes to commit")
			return nil
		}
		return err
	}

	o.logger.Info("Changes committed successfully")
	return nil
}

// promptContinue asks the user if they want to continue after max iterations.
func (o *Orchestrator) promptContinue() bool {
	// If observer is set, use it for confirmation
	if o.observer != nil {
		msg := fmt.Sprintf("Maximum iterations (%d) reached. Continue?", o.config.MaxIterations)
		return o.observer.OnConfirmationRequired(msg)
	}

	// Otherwise use terminal prompt
	fmt.Printf("\n⚠️  Maximum iterations (%d) reached.\n", o.config.MaxIterations)
	fmt.Print("Continue for another round? [y/N]: ")

	return readYesNo()
}

// promptNextRound asks the user if they want to proceed with the next round (interactive mode).
func (o *Orchestrator) promptNextRound() bool {
	// If observer is set, use it for confirmation
	if o.observer != nil {
		return o.observer.OnConfirmationRequired("Proceed to next round?")
	}

	// Otherwise use terminal prompt
	fmt.Print("\nProceed to next round? [Y/n]: ")
	return readYesNoDefault(true)
}

// readYesNo reads a yes/no response from stdin (default: no).
func readYesNo() bool {
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

// readYesNoDefault reads a yes/no response with a configurable default.
func readYesNoDefault(defaultYes bool) bool {
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response == "" {
		return defaultYes
	}

	return response == "y" || response == "yes"
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

// extractFilesFromDiff extracts file paths from a git diff.
func extractFilesFromDiff(diff string) []string {
	files := make([]string, 0)
	lines := strings.Split(diff, "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "+++ b/") {
			file := strings.TrimPrefix(line, "+++ b/")
			if file != "/dev/null" {
				files = append(files, file)
			}
		}
	}

	return files
}

// GetState returns the current session state.
func (o *Orchestrator) GetState() types.SessionState {
	return o.state
}

// GetCurrentRound returns the current round number.
func (o *Orchestrator) GetCurrentRound() int {
	return o.currentRound
}

// GetRounds returns the history of all rounds.
func (o *Orchestrator) GetRounds() []types.Round {
	return o.rounds
}

// Observer notification methods

func (o *Orchestrator) notifyRoundStart(number int) {
	if o.observer != nil {
		o.observer.OnRoundStart(number)
	}
}

func (o *Orchestrator) notifyFighterEnter(fighter string) {
	if o.observer != nil {
		o.observer.OnFighterEnter(fighter)
	}
}

func (o *Orchestrator) notifyFighterAction(fighter, action string) {
	if o.observer != nil {
		o.observer.OnFighterAction(fighter, action)
	}
}

func (o *Orchestrator) notifyFighterFinish(fighter string, duration time.Duration) {
	if o.observer != nil {
		o.observer.OnFighterFinish(fighter, duration)
	}
}

func (o *Orchestrator) notifyChangesDetected(fileCount int) {
	if o.observer != nil {
		o.observer.OnChangesDetected(fileCount)
	}
}

func (o *Orchestrator) notifyIssuesFound(issues []string) {
	if o.observer != nil {
		o.observer.OnIssuesFound(issues)
	}
}

func (o *Orchestrator) notifyNoIssues() {
	if o.observer != nil {
		o.observer.OnNoIssues()
	}
}

func (o *Orchestrator) notifySessionComplete(result *types.SessionResult, success bool) {
	if o.observer != nil {
		o.observer.OnSessionComplete(result, success)
	}
}

func (o *Orchestrator) notifyError(err error) {
	if o.observer != nil {
		o.observer.OnError(err)
	}
}
