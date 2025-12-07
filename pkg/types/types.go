// Package types defines shared types used across the mortal-prompter application.
package types

import "time"

// Round represents a single iteration in the code review battle between Claude and Codex.
// Each round consists of Claude implementing changes and Codex reviewing them.
type Round struct {
	// Number is the sequential round number (1-indexed)
	Number int

	// ClaudePrompt is the prompt sent to Claude Code for this round
	ClaudePrompt string

	// ClaudeOutput is the raw output captured from Claude Code execution
	ClaudeOutput string

	// GitDiff contains the git diff of changes made by Claude in this round
	GitDiff string

	// CodexReview contains the raw review output from Codex
	CodexReview string

	// HasIssues indicates whether Codex found any issues in this round
	HasIssues bool

	// Issues is the list of specific issues found by Codex
	Issues []string

	// Duration is how long this round took to complete
	Duration time.Duration

	// Timestamp is when this round started
	Timestamp time.Time
}

// ReviewResult represents the parsed output from Codex's code review.
type ReviewResult struct {
	// HasIssues indicates whether any issues were found during review
	HasIssues bool

	// Issues is the list of specific issues identified
	Issues []string

	// RawOutput is the complete raw output from Codex
	RawOutput string
}

// SessionResult represents the final outcome of a mortal-prompter session.
type SessionResult struct {
	// Success indicates whether the session completed successfully (no issues remaining)
	Success bool

	// TotalRounds is the number of rounds executed during the session
	TotalRounds int

	// TotalDuration is the total time the session took
	TotalDuration time.Duration

	// Rounds contains the history of all rounds in the session
	Rounds []Round

	// FinalDiff contains the cumulative git diff of all changes
	FinalDiff string

	// FilesModified is a list of all files that were modified during the session
	FilesModified []string
}

// FighterType represents the type of LLM fighter.
type FighterType string

const (
	// FighterClaude represents Claude Code (the implementer)
	FighterClaude FighterType = "CLAUDE CODE"

	// FighterCodex represents OpenAI Codex (the reviewer)
	FighterCodex FighterType = "CODEX"
)

// SessionState represents the current state of a mortal-prompter session.
type SessionState string

const (
	// StateInitializing indicates the session is starting up
	StateInitializing SessionState = "initializing"

	// StateRunning indicates the session is actively running rounds
	StateRunning SessionState = "running"

	// StateWaitingConfirmation indicates the session is waiting for user input
	StateWaitingConfirmation SessionState = "waiting_confirmation"

	// StateCompleted indicates the session finished successfully
	StateCompleted SessionState = "completed"

	// StateAborted indicates the session was aborted
	StateAborted SessionState = "aborted"

	// StateFailed indicates the session failed due to an error
	StateFailed SessionState = "failed"
)
