// Package components contains TUI components for mortal-prompter.
package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	healthBarFull  = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
	healthBarEmpty = lipgloss.NewStyle().Foreground(lipgloss.Color("#333333"))
)

// HealthBar renders an arcade-style health bar
// current and max are values to display, width is the total character width
func HealthBar(current, max, width int) string {
	if max <= 0 {
		max = 1
	}
	if current < 0 {
		current = 0
	}
	if current > max {
		current = max
	}

	// Calculate filled portion
	innerWidth := width - 2 // Account for brackets
	if innerWidth < 1 {
		innerWidth = 1
	}

	filled := (current * innerWidth) / max
	empty := innerWidth - filled

	// Build the bar
	var sb strings.Builder
	sb.WriteString("[")
	sb.WriteString(healthBarFull.Render(strings.Repeat("█", filled)))
	sb.WriteString(healthBarEmpty.Render(strings.Repeat("░", empty)))
	sb.WriteString("]")

	return sb.String()
}

// ProgressBar renders a progress bar with percentage
func ProgressBar(percent float64, width int) string {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	innerWidth := width - 2
	if innerWidth < 1 {
		innerWidth = 1
	}

	filled := int(percent * float64(innerWidth) / 100)
	empty := innerWidth - filled

	var sb strings.Builder
	sb.WriteString("[")
	sb.WriteString(healthBarFull.Render(strings.Repeat("█", filled)))
	if filled < innerWidth {
		sb.WriteString(healthBarFull.Render("▓"))
		empty--
	}
	if empty > 0 {
		sb.WriteString(healthBarEmpty.Render(strings.Repeat("░", empty)))
	}
	sb.WriteString("]")

	return sb.String()
}

// FighterHealthBar renders a health bar specific for fighters with label
func FighterHealthBar(name string, state string, width int) string {
	var stateStyle lipgloss.Style
	var barColor lipgloss.Color

	switch state {
	case "active", "fighting":
		stateStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FFFF"))
		barColor = lipgloss.Color("#00FFFF")
	case "waiting":
		stateStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
		barColor = lipgloss.Color("#333333")
	case "done", "finished":
		stateStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
		barColor = lipgloss.Color("#00FF00")
	default:
		stateStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
		barColor = lipgloss.Color("#FFFFFF")
	}

	nameStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))

	// Build the bar with custom color
	barFull := lipgloss.NewStyle().Foreground(barColor)
	innerWidth := width - 2
	if innerWidth < 1 {
		innerWidth = 10
	}

	bar := "[" + barFull.Render(strings.Repeat("█", innerWidth)) + "]"

	return nameStyle.Render(name) + "\n" + bar + "\n" + stateStyle.Render(strings.ToUpper(state))
}
