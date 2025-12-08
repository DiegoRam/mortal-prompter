// Package tui implements the bubbletea-based terminal UI for mortal-prompter.
package tui

import "github.com/charmbracelet/lipgloss"

// Arcade color palette
var (
	ColorYellow  = lipgloss.Color("#FFFF00")
	ColorCyan    = lipgloss.Color("#00FFFF")
	ColorGreen   = lipgloss.Color("#00FF00")
	ColorRed     = lipgloss.Color("#FF0000")
	ColorMagenta = lipgloss.Color("#FF00FF")
	ColorWhite   = lipgloss.Color("#FFFFFF")
	ColorGray    = lipgloss.Color("#666666")
	ColorDarkBg  = lipgloss.Color("#1a1a2e")
)

// Styles for the TUI
var (
	// Main container style
	ContainerStyle = lipgloss.NewStyle().
			Padding(1, 2)

	// Banner and title styles
	BannerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorYellow).
			BorderStyle(lipgloss.DoubleBorder()).
			BorderForeground(ColorYellow).
			Padding(0, 2).
			Align(lipgloss.Center)

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorYellow)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorCyan)

	// Round header style
	RoundHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorYellow).
				Background(lipgloss.Color("#333")).
				Padding(0, 2)

	// Fighter styles
	FighterActiveStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorCyan)

	FighterWaitingStyle = lipgloss.NewStyle().
				Foreground(ColorGray)

	FighterNameStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorWhite)

	// Status styles
	SuccessStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorGreen)

	ErrorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorRed)

	WarningStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorYellow)

	InfoStyle = lipgloss.NewStyle().
			Foreground(ColorCyan)

	// Box styles
	BoxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ColorCyan).
			Padding(1, 2)

	DoubleBoxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.DoubleBorder()).
			BorderForeground(ColorYellow).
			Padding(0, 1)

	// Help style
	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorGray)

	// Prompt view specific
	TextareaStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ColorCyan).
			Padding(0, 1)

	// Health bar styles
	HealthBarFullStyle = lipgloss.NewStyle().
				Foreground(ColorGreen)

	HealthBarEmptyStyle = lipgloss.NewStyle().
				Foreground(ColorGray)

	// Round list item styles
	RoundCompleteStyle = lipgloss.NewStyle().
				Foreground(ColorGreen)

	RoundInProgressStyle = lipgloss.NewStyle().
				Foreground(ColorYellow)

	RoundFailedStyle = lipgloss.NewStyle().
				Foreground(ColorRed)

	// Victory/Defeat styles
	VictoryStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorGreen).
			Align(lipgloss.Center)

	DefeatStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorRed).
			Align(lipgloss.Center)
)
