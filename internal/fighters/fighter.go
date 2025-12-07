// Package fighters provides wrappers for the LLM CLI tools used in mortal-prompter battles.
// It defines the Claude (implementer) and Codex (reviewer) fighters.
package fighters

// Fighter is the interface implemented by all LLM fighters (Claude and Codex).
type Fighter interface {
	// Name returns the display name of the fighter.
	Name() string
}
