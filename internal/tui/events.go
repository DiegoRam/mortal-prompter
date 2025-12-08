package tui

import (
	"time"

	"github.com/diegoram/mortal-prompter/pkg/types"
)

// EventType identifies the type of event from the orchestrator
type EventType int

const (
	EventRoundStart EventType = iota
	EventFighterEnter
	EventFighterAction
	EventFighterFinish
	EventChangesDetected
	EventIssuesFound
	EventNoIssues
	EventSessionComplete
	EventError
	EventConfirmationRequired
)

// Event represents an event from the orchestrator
type Event struct {
	Type    EventType
	Payload interface{}
}

// RoundStartPayload contains data for round start events
type RoundStartPayload struct {
	Number int
}

// FighterEnterPayload contains data for fighter enter events
type FighterEnterPayload struct {
	Fighter string
}

// FighterActionPayload contains data for fighter action events
type FighterActionPayload struct {
	Action string
}

// FighterFinishPayload contains data for fighter finish events
type FighterFinishPayload struct {
	Fighter  string
	Duration time.Duration
}

// ChangesDetectedPayload contains data for changes detected events
type ChangesDetectedPayload struct {
	FileCount int
}

// IssuesFoundPayload contains data for issues found events
type IssuesFoundPayload struct {
	Issues []string
}

// SessionCompletePayload contains data for session complete events
type SessionCompletePayload struct {
	Result  *types.SessionResult
	Success bool
}

// ErrorPayload contains data for error events
type ErrorPayload struct {
	Error error
}

// ConfirmationPayload contains data for confirmation required events
type ConfirmationPayload struct {
	Message string
}
