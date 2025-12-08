// Package fighters provides wrappers for the LLM CLI tools used in mortal-prompter battles.
// It defines the Claude, Codex, and Gemini fighters.
package fighters

import (
	"context"

	"github.com/diegoram/mortal-prompter/pkg/types"
)

// FighterType represents the type of fighter
type FighterType string

const (
	FighterTypeClaude FighterType = "claude"
	FighterTypeCodex  FighterType = "codex"
	FighterTypeGemini FighterType = "gemini"
)

// AllFighterTypes returns all available fighter types
func AllFighterTypes() []FighterType {
	return []FighterType{FighterTypeClaude, FighterTypeCodex, FighterTypeGemini}
}

// Fighter is the interface implemented by all LLM fighters.
type Fighter interface {
	// Name returns the display name of the fighter.
	Name() string
}

// Implementer is the interface for fighters that can implement code changes.
type Implementer interface {
	Fighter
	// Execute runs the fighter with the provided prompt and returns the output.
	Execute(ctx context.Context, prompt string) (string, error)
	// BuildPromptWithIssues constructs a prompt that includes previous issues.
	BuildPromptWithIssues(basePrompt string, previousIssues []string) string
}

// Reviewer is the interface for fighters that can review code.
type Reviewer interface {
	Fighter
	// Review executes a code review on the git diff and returns the result.
	Review(ctx context.Context, gitDiff string) (*types.ReviewResult, error)
}
