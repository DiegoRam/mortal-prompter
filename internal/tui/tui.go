package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/minimalart/mortal-prompter/internal/config"
)

// Run starts the TUI and returns the entered prompt when the user submits
// Returns the prompt string and any error that occurred
func Run(cfg *config.Config) (*Model, error) {
	m := NewModel(cfg)

	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	model := finalModel.(Model)
	return &model, nil
}

// RunWithProgram starts the TUI and returns the program for external control
func RunWithProgram(cfg *config.Config) (*tea.Program, *Model) {
	m := NewModel(cfg)
	p := tea.NewProgram(m, tea.WithAltScreen())
	return p, &m
}
