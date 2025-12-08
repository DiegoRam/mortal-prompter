package tui

import (
	"time"

	"github.com/diegoram/mortal-prompter/pkg/types"
)

// Observer is the interface for receiving orchestrator events
// This is defined in the tui package to avoid circular dependencies
// The orchestrator will implement a callback-based approach instead
type Observer interface {
	OnRoundStart(number int)
	OnFighterEnter(fighter string)
	OnFighterAction(action string)
	OnFighterFinish(fighter string, duration time.Duration)
	OnChangesDetected(fileCount int)
	OnIssuesFound(issues []string)
	OnNoIssues()
	OnSessionComplete(result *types.SessionResult, success bool)
	OnError(err error)
	OnConfirmationRequired(message string) bool
}

// ChannelObserver implements Observer by sending events to a channel
type ChannelObserver struct {
	eventChan    chan<- Event
	responseChan <-chan bool
}

// NewChannelObserver creates a new ChannelObserver
func NewChannelObserver(eventChan chan<- Event, responseChan <-chan bool) *ChannelObserver {
	return &ChannelObserver{
		eventChan:    eventChan,
		responseChan: responseChan,
	}
}

// OnRoundStart sends a round start event
func (o *ChannelObserver) OnRoundStart(number int) {
	o.eventChan <- Event{
		Type:    EventRoundStart,
		Payload: RoundStartPayload{Number: number},
	}
}

// OnFighterEnter sends a fighter enter event
func (o *ChannelObserver) OnFighterEnter(fighter string) {
	o.eventChan <- Event{
		Type:    EventFighterEnter,
		Payload: FighterEnterPayload{Fighter: fighter},
	}
}

// OnFighterAction sends a fighter action event
func (o *ChannelObserver) OnFighterAction(action string) {
	o.eventChan <- Event{
		Type:    EventFighterAction,
		Payload: FighterActionPayload{Action: action},
	}
}

// OnFighterFinish sends a fighter finish event
func (o *ChannelObserver) OnFighterFinish(fighter string, duration time.Duration) {
	o.eventChan <- Event{
		Type:    EventFighterFinish,
		Payload: FighterFinishPayload{Fighter: fighter, Duration: duration},
	}
}

// OnChangesDetected sends a changes detected event
func (o *ChannelObserver) OnChangesDetected(fileCount int) {
	o.eventChan <- Event{
		Type:    EventChangesDetected,
		Payload: ChangesDetectedPayload{FileCount: fileCount},
	}
}

// OnIssuesFound sends an issues found event
func (o *ChannelObserver) OnIssuesFound(issues []string) {
	o.eventChan <- Event{
		Type:    EventIssuesFound,
		Payload: IssuesFoundPayload{Issues: issues},
	}
}

// OnNoIssues sends a no issues event
func (o *ChannelObserver) OnNoIssues() {
	o.eventChan <- Event{
		Type:    EventNoIssues,
		Payload: nil,
	}
}

// OnSessionComplete sends a session complete event
func (o *ChannelObserver) OnSessionComplete(result *types.SessionResult, success bool) {
	o.eventChan <- Event{
		Type:    EventSessionComplete,
		Payload: SessionCompletePayload{Result: result, Success: success},
	}
}

// OnError sends an error event
func (o *ChannelObserver) OnError(err error) {
	o.eventChan <- Event{
		Type:    EventError,
		Payload: ErrorPayload{Error: err},
	}
}

// OnConfirmationRequired sends a confirmation required event and waits for response
func (o *ChannelObserver) OnConfirmationRequired(message string) bool {
	o.eventChan <- Event{
		Type:    EventConfirmationRequired,
		Payload: ConfirmationPayload{Message: message},
	}
	// Wait for response from TUI
	return <-o.responseChan
}
