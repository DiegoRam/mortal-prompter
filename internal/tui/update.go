package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/diegoram/mortal-prompter/internal/clipboard"
)

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textarea.SetWidth(msg.Width - 10)
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - 10
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tickMsg:
		// Periodic tick to ensure UI updates - continue ticking only during battle
		if m.view == ViewBattle {
			// Update blink animation counter
			m.blinkCounter++
			if m.blinkCounter >= blinkInterval {
				m.blinkOn = !m.blinkOn
				m.blinkCounter = 0
			}
			return m, tick()
		}
		return m, nil

	case eventMsg:
		return m.handleEvent(msg.event)

	case battleStartedMsg:
		m.battleStarted = true
		return m, tea.Batch(m.spinner.Tick, tick(), waitForEvent(m.eventChan))

	case battleFinishedMsg:
		m.view = ViewResults
		m.sessionResult = msg.result
		m.sessionSuccess = msg.success
		m.sessionError = msg.err
		return m, nil

	case clearImageMessageMsg:
		m.imageMessage = ""
		return m, nil
	}

	// Update sub-components based on current view
	switch m.view {
	case ViewPrompt:
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		cmds = append(cmds, cmd)

	case ViewBattle:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// handleKeyMsg handles keyboard input
func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.view {
	case ViewFighterSelect:
		return m.handleFighterSelectKeys(msg)
	case ViewPrompt:
		return m.handlePromptKeys(msg)
	case ViewBattle:
		return m.handleBattleKeys(msg)
	case ViewResults:
		return m.handleResultsKeys(msg)
	case ViewConfirmation:
		return m.handleConfirmationKeys(msg)
	}
	return m, nil
}

// handleFighterSelectKeys handles keys in the fighter selection view
func (m Model) handleFighterSelectKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit

	case tea.KeyEnter:
		// Move to prompt view
		m.view = ViewPrompt
		m.textarea.Focus()
		return m, nil

	case tea.KeyUp, tea.KeyDown, tea.KeyTab:
		// Switch between implementer and reviewer fields
		if m.fighterSelectField == FieldImplementer {
			m.fighterSelectField = FieldReviewer
		} else {
			m.fighterSelectField = FieldImplementer
		}
		return m, nil

	case tea.KeyLeft:
		// Move to previous fighter option
		m.moveFighterSelection(-1)
		return m, nil

	case tea.KeyRight:
		// Move to next fighter option
		m.moveFighterSelection(1)
		return m, nil
	}

	return m, nil
}

// handlePromptKeys handles keys in the prompt view
func (m Model) handlePromptKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Check for submit: Ctrl+Enter, Ctrl+S, or Escape then Enter
	isCtrlEnter := msg.Type == tea.KeyEnter && msg.Alt
	isCtrlS := msg.Type == tea.KeyCtrlS

	switch {
	// Only Ctrl+C quits in prompt view - allow typing "q" normally
	case msg.Type == tea.KeyCtrlC:
		return m, tea.Quit

	case isCtrlEnter || isCtrlS:
		m.prompt = m.textarea.Value()
		if m.prompt == "" {
			return m, nil
		}
		// Mark battle as started and quit - the main.go will create a new TUI
		// with the orchestrator for the actual battle
		m.battleStarted = true
		return m, tea.Quit

	// Handle Ctrl+V for paste image
	case key.Matches(msg, m.keys.PasteImage):
		return m.handleImagePaste()

	// Handle Ctrl+X for remove image
	case key.Matches(msg, m.keys.RemoveImage):
		return m.handleImageRemove()

	default:
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		return m, cmd
	}
}

// handleImagePaste reads an image from the clipboard and attaches it
func (m Model) handleImagePaste() (tea.Model, tea.Cmd) {
	// Clear any previous message
	m.imageMessage = ""

	// Try to read image from clipboard
	imgData, err := clipboard.ReadImage()
	if err != nil {
		if err == clipboard.ErrNoImageInClipboard {
			m.imageMessage = "No image found in clipboard"
		} else {
			m.imageMessage = "Failed to read clipboard: " + err.Error()
		}
		// Clear message after a delay
		return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
			return clearImageMessageMsg{}
		})
	}

	// Save image to temp file in output directory
	outputDir := m.config.OutputDir
	if outputDir == "" {
		outputDir = ".mortal-prompter"
	}

	filePath, err := clipboard.SaveToFile(imgData.Data, outputDir)
	if err != nil {
		m.imageMessage = "Failed to save image: " + err.Error()
		return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
			return clearImageMessageMsg{}
		})
	}

	// Store the attachment
	m.attachedImage = &ImageAttachment{
		Data:     imgData.Data,
		FilePath: filePath,
		Width:    imgData.Width,
		Height:   imgData.Height,
		AddedAt:  time.Now(),
	}

	m.imageMessage = "Image attached successfully!"
	return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return clearImageMessageMsg{}
	})
}

// handleImageRemove removes the attached image
func (m Model) handleImageRemove() (tea.Model, tea.Cmd) {
	if m.attachedImage == nil {
		m.imageMessage = "No image attached"
		return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
			return clearImageMessageMsg{}
		})
	}

	m.attachedImage = nil
	m.imageMessage = "Image removed"
	return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return clearImageMessageMsg{}
	})
}

// clearImageMessageMsg is sent to clear the temporary image message
type clearImageMessageMsg struct{}

// handleBattleKeys handles keys in the battle view
func (m Model) handleBattleKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit):
		// Send abort signal and quit
		return m, tea.Quit

	case key.Matches(msg, m.keys.Details):
		m.showDetails = !m.showDetails
		return m, nil

	case msg.Type == tea.KeyCtrlC:
		return m, tea.Quit
	}
	return m, nil
}

// handleResultsKeys handles keys in the results view
func (m Model) handleResultsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit), msg.Type == tea.KeyCtrlC, msg.Type == tea.KeyEnter:
		return m, tea.Quit

	case key.Matches(msg, m.keys.ViewDiff):
		m.showDetails = !m.showDetails
		return m, nil
	}
	return m, nil
}

// handleConfirmationKeys handles keys in the confirmation dialog
func (m Model) handleConfirmationKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Confirm):
		m.responseChan <- true
		m.view = ViewBattle
		return m, tea.Batch(m.spinner.Tick, tick(), waitForEvent(m.eventChan))

	case key.Matches(msg, m.keys.Deny), key.Matches(msg, m.keys.Cancel):
		m.responseChan <- false
		m.view = ViewBattle
		return m, tea.Batch(m.spinner.Tick, tick(), waitForEvent(m.eventChan))

	case msg.Type == tea.KeyCtrlC:
		m.responseChan <- false
		return m, tea.Quit
	}
	return m, nil
}

// handleEvent processes events from the orchestrator
func (m Model) handleEvent(event Event) (tea.Model, tea.Cmd) {
	switch event.Type {
	case EventRoundStart:
		if payload, ok := event.Payload.(RoundStartPayload); ok {
			m.currentRound = payload.Number
			m.rounds = append(m.rounds, RoundDisplay{
				Number:       payload.Number,
				Status:       "in_progress",
				CurrentPhase: "implementer",
			})
			m.implementerState = FighterIdle
			m.reviewerState = FighterIdle
		}

	case EventFighterEnter:
		if payload, ok := event.Payload.(FighterEnterPayload); ok {
			// Clear current action when a new fighter enters
			m.currentAction = ""
			// Check if it's the implementer or reviewer entering
			if payload.Fighter == m.implementerName {
				m.implementerState = FighterActive
				m.reviewerState = FighterIdle
				if len(m.rounds) > 0 {
					m.rounds[len(m.rounds)-1].CurrentPhase = "implementer"
				}
			} else {
				// Reviewer enters - implementer should already be finished via FighterFinish event
				m.reviewerState = FighterActive
				if len(m.rounds) > 0 {
					m.rounds[len(m.rounds)-1].CurrentPhase = "reviewer"
				}
			}
		}

	case EventFighterAction:
		if payload, ok := event.Payload.(FighterActionPayload); ok {
			m.currentAction = payload.Action
			// Update fighter states based on who is performing the action
			if payload.Fighter == m.implementerName || payload.Fighter == "Claude Code" {
				m.implementerState = FighterActive
				m.reviewerState = FighterIdle
			} else {
				m.reviewerState = FighterActive
				// Only mark implementer as finished if was active
				if m.implementerState == FighterActive {
					m.implementerState = FighterFinished
				}
			}
		}

	case EventFighterFinish:
		if payload, ok := event.Payload.(FighterFinishPayload); ok {
			if payload.Fighter == m.implementerName || payload.Fighter == "Claude Code" {
				m.implementerState = FighterFinished
				if len(m.rounds) > 0 {
					m.rounds[len(m.rounds)-1].ClaudeDone = true
				}
			} else {
				m.reviewerState = FighterFinished
				if len(m.rounds) > 0 {
					m.rounds[len(m.rounds)-1].CodexDone = true
				}
			}
		}

	case EventIssuesFound:
		if payload, ok := event.Payload.(IssuesFoundPayload); ok {
			if len(m.rounds) > 0 {
				m.rounds[len(m.rounds)-1].Issues = payload.Issues
				m.rounds[len(m.rounds)-1].Status = "completed"
			}
		}

	case EventNoIssues:
		if len(m.rounds) > 0 {
			m.rounds[len(m.rounds)-1].Status = "completed"
		}

	case EventSessionComplete:
		if payload, ok := event.Payload.(SessionCompletePayload); ok {
			m.view = ViewResults
			m.sessionResult = payload.Result
			m.sessionSuccess = payload.Success
		}
		return m, nil // Don't wait for more events

	case EventError:
		if payload, ok := event.Payload.(ErrorPayload); ok {
			m.sessionError = payload.Error
			m.view = ViewResults
			m.sessionSuccess = false
		}
		return m, nil

	case EventConfirmationRequired:
		if payload, ok := event.Payload.(ConfirmationPayload); ok {
			m.confirmMessage = payload.Message
			m.view = ViewConfirmation
		}
		return m, nil // Don't wait, we need user input
	}

	// Continue listening for events AND keep ticking for UI updates
	return m, tea.Batch(
		m.spinner.Tick,
		tick(),
		waitForEvent(m.eventChan),
	)
}

// GetPrompt returns the entered prompt (called after battle starts)
func (m Model) GetPrompt() string {
	return m.prompt
}

// IsBattleStarted returns true if the user submitted the prompt
func (m Model) IsBattleStarted() bool {
	return m.battleStarted
}
