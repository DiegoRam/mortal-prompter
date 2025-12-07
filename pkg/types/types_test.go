package types

import (
	"testing"
	"time"
)

func TestRoundStruct(t *testing.T) {
	now := time.Now()
	round := Round{
		Number:       1,
		ClaudePrompt: "implement feature X",
		ClaudeOutput: "implementation output",
		GitDiff:      "+ added line\n- removed line",
		CodexReview:  "ISSUE: missing error handling",
		HasIssues:    true,
		Issues:       []string{"missing error handling"},
		Duration:     45 * time.Second,
		Timestamp:    now,
	}

	if round.Number != 1 {
		t.Errorf("expected Number to be 1, got %d", round.Number)
	}
	if round.ClaudePrompt != "implement feature X" {
		t.Errorf("unexpected ClaudePrompt: %s", round.ClaudePrompt)
	}
	if !round.HasIssues {
		t.Error("expected HasIssues to be true")
	}
	if len(round.Issues) != 1 {
		t.Errorf("expected 1 issue, got %d", len(round.Issues))
	}
	if round.Duration != 45*time.Second {
		t.Errorf("expected duration of 45s, got %v", round.Duration)
	}
}

func TestReviewResultStruct(t *testing.T) {
	result := ReviewResult{
		HasIssues: false,
		Issues:    nil,
		RawOutput: "LGTM: No issues found",
	}

	if result.HasIssues {
		t.Error("expected HasIssues to be false")
	}
	if result.Issues != nil {
		t.Error("expected Issues to be nil")
	}
	if result.RawOutput != "LGTM: No issues found" {
		t.Errorf("unexpected RawOutput: %s", result.RawOutput)
	}
}

func TestSessionResultStruct(t *testing.T) {
	session := SessionResult{
		Success:       true,
		TotalRounds:   3,
		TotalDuration: 5 * time.Minute,
		Rounds:        make([]Round, 3),
		FinalDiff:     "+ all changes",
		FilesModified: []string{"main.go", "config.go"},
	}

	if !session.Success {
		t.Error("expected Success to be true")
	}
	if session.TotalRounds != 3 {
		t.Errorf("expected TotalRounds to be 3, got %d", session.TotalRounds)
	}
	if len(session.FilesModified) != 2 {
		t.Errorf("expected 2 modified files, got %d", len(session.FilesModified))
	}
}

func TestFighterTypeConstants(t *testing.T) {
	if FighterClaude != "CLAUDE CODE" {
		t.Errorf("unexpected FighterClaude value: %s", FighterClaude)
	}
	if FighterCodex != "CODEX" {
		t.Errorf("unexpected FighterCodex value: %s", FighterCodex)
	}
}

func TestSessionStateConstants(t *testing.T) {
	states := []SessionState{
		StateInitializing,
		StateRunning,
		StateWaitingConfirmation,
		StateCompleted,
		StateAborted,
		StateFailed,
	}

	expectedValues := []string{
		"initializing",
		"running",
		"waiting_confirmation",
		"completed",
		"aborted",
		"failed",
	}

	for i, state := range states {
		if string(state) != expectedValues[i] {
			t.Errorf("expected state %d to be %q, got %q", i, expectedValues[i], state)
		}
	}
}
