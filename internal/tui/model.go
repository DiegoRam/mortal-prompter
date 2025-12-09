package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/diegoram/mortal-prompter/internal/config"
	"github.com/diegoram/mortal-prompter/internal/fighters"
	"github.com/diegoram/mortal-prompter/pkg/types"
)

// ViewState represents the current view of the TUI
type ViewState int

const (
	ViewFighterSelect ViewState = iota
	ViewPrompt
	ViewBattle
	ViewResults
	ViewConfirmation
)

// FighterSelectField represents which field is being edited in fighter selection
type FighterSelectField int

const (
	FieldImplementer FighterSelectField = iota
	FieldReviewer
)

// FighterState represents the state of a fighter during battle
type FighterState int

const (
	FighterIdle FighterState = iota
	FighterActive
	FighterFinished
)

// RoundDisplay holds display data for a single round
type RoundDisplay struct {
	Number       int
	Status       string // "in_progress", "completed", "failed"
	Issues       []string
	Duration     time.Duration
	ClaudeDone   bool
	CodexDone    bool
	CurrentPhase string // "claude", "codex", "diff"
}

// Model is the main bubbletea model for the TUI
type Model struct {
	// View state
	view ViewState

	// Configuration
	config *config.Config

	// Components
	textarea textarea.Model
	spinner  spinner.Model
	viewport viewport.Model
	help     help.Model
	keys     KeyMap

	// Dimensions
	width  int
	height int

	// Fighter selection
	implementerType     fighters.FighterType
	reviewerType        fighters.FighterType
	fighterSelectField  FighterSelectField
	availableFighters   []fighters.FighterType

	// Session data
	prompt             string
	rounds             []RoundDisplay
	currentRound       int
	implementerState   FighterState
	reviewerState      FighterState
	implementerName    string
	reviewerName       string
	currentAction      string
	sessionResult      *types.SessionResult
	sessionSuccess     bool
	sessionError       error

	// Async communication
	eventChan    chan Event
	responseChan chan bool

	// Confirmation dialog state
	confirmMessage string
	confirmDefault bool

	// Battle started flag
	battleStarted bool

	// Detail view toggle
	showDetails bool

	// Start time for duration display
	startTime time.Time

	// Report path
	reportPath string

	// Log file path
	logFilePath string

	// Blink animation state for "FIGHTING" text
	blinkOn      bool
	blinkCounter int
}

// blinkInterval defines how many ticks before toggling blink state
// At 100ms tick rate, 5 ticks = 500ms per blink state
const blinkInterval = 5

// NewModel creates a new TUI model
func NewModel(cfg *config.Config) Model {
	// Initialize textarea for prompt input
	ta := textarea.New()
	ta.Placeholder = "Enter your prompt for the implementer..."
	ta.CharLimit = 10000
	ta.SetWidth(70)
	ta.SetHeight(8)

	// Initialize spinner
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = FighterActiveStyle

	// Initialize viewport for scrolling
	vp := viewport.New(80, 20)

	// Initialize help
	h := help.New()

	return Model{
		view:              ViewFighterSelect,
		config:            cfg,
		textarea:          ta,
		spinner:           sp,
		viewport:          vp,
		help:              h,
		keys:              DefaultKeyMap(),
		rounds:            make([]RoundDisplay, 0),
		eventChan:         make(chan Event, 100),
		responseChan:      make(chan bool, 1),
		width:             80,
		height:            24,
		implementerType:   cfg.Implementer,
		reviewerType:      cfg.Reviewer,
		availableFighters: fighters.AllFighterTypes(),
		fighterSelectField: FieldImplementer,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		textarea.Blink,
		m.spinner.Tick,
		tick(), // Always tick for UI updates
	}

	// If battle is already started (e.g., created via SetBattleStarted),
	// start listening for orchestrator events immediately
	if m.battleStarted {
		cmds = append(cmds, waitForEvent(m.eventChan))
	}

	return tea.Batch(cmds...)
}

// GetEventChannel returns the event channel for the orchestrator observer
func (m *Model) GetEventChannel() chan Event {
	return m.eventChan
}

// GetResponseChannel returns the response channel for confirmations
func (m *Model) GetResponseChannel() chan bool {
	return m.responseChan
}

// SetBattleStarted sets the model to battle mode with the given prompt
func (m *Model) SetBattleStarted(prompt string) {
	m.prompt = prompt
	m.battleStarted = true
	m.view = ViewBattle
	// Explicitly reset fighter states to WAITING
	m.implementerState = FighterIdle
	m.reviewerState = FighterIdle
	m.currentRound = 0
	m.currentAction = ""
	m.rounds = make([]RoundDisplay, 0)
	m.startTime = time.Now()
}

// GetImplementerType returns the selected implementer type
func (m Model) GetImplementerType() fighters.FighterType {
	return m.implementerType
}

// GetReviewerType returns the selected reviewer type
func (m Model) GetReviewerType() fighters.FighterType {
	return m.reviewerType
}

// SetFighterNames sets the display names for the fighters
func (m *Model) SetFighterNames(implementer, reviewer string) {
	m.implementerName = implementer
	m.reviewerName = reviewer
}

// SetReportPath sets the report path for display in results
func (m *Model) SetReportPath(path string) {
	m.reportPath = path
}

// SetLogFilePath sets the log file path for display in battle view
func (m *Model) SetLogFilePath(path string) {
	m.logFilePath = path
}

// moveFighterSelection moves the fighter selection by delta (-1 or +1)
func (m *Model) moveFighterSelection(delta int) {
	if len(m.availableFighters) == 0 {
		return
	}

	var currentIdx int
	var currentType fighters.FighterType

	if m.fighterSelectField == FieldImplementer {
		currentType = m.implementerType
	} else {
		currentType = m.reviewerType
	}

	// Find current index
	for i, ft := range m.availableFighters {
		if ft == currentType {
			currentIdx = i
			break
		}
	}

	// Calculate new index with wrapping
	newIdx := (currentIdx + delta + len(m.availableFighters)) % len(m.availableFighters)

	// Update the appropriate field
	if m.fighterSelectField == FieldImplementer {
		m.implementerType = m.availableFighters[newIdx]
	} else {
		m.reviewerType = m.availableFighters[newIdx]
	}
}

// eventMsg wraps an event from the orchestrator
type eventMsg struct {
	event Event
}

// battleStartedMsg signals that the battle has started
type battleStartedMsg struct{}

// battleFinishedMsg signals that the battle has finished
type battleFinishedMsg struct {
	result  *types.SessionResult
	success bool
	err     error
}

// tickMsg is used for periodic UI updates
type tickMsg time.Time

// tick returns a command that sends a tickMsg after a short delay
func tick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// waitForEvent creates a command that waits for an event from the orchestrator
func waitForEvent(eventChan <-chan Event) tea.Cmd {
	return func() tea.Msg {
		event, ok := <-eventChan
		if !ok {
			return nil
		}
		return eventMsg{event}
	}
}
