package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// View renders the current view
func (m Model) View() string {
	switch m.view {
	case ViewPrompt:
		return m.viewPrompt()
	case ViewBattle:
		return m.viewBattle()
	case ViewResults:
		return m.viewResults()
	case ViewConfirmation:
		return m.viewConfirmation()
	default:
		return "Unknown view"
	}
}

// viewPrompt renders the prompt input view
func (m Model) viewPrompt() string {
	var sb strings.Builder

	// Banner
	banner := `
╔════════════════════════════════════════════════════════════════════════╗
║                                                                        ║
║  ███╗   ███╗ ██████╗ ██████╗ ████████╗ █████╗ ██╗                      ║
║  ████╗ ████║██╔═══██╗██╔══██╗╚══██╔══╝██╔══██╗██║                      ║
║  ██╔████╔██║██║   ██║██████╔╝   ██║   ███████║██║                      ║
║  ██║╚██╔╝██║██║   ██║██╔══██╗   ██║   ██╔══██║██║                      ║
║  ██║ ╚═╝ ██║╚██████╔╝██║  ██║   ██║   ██║  ██║███████╗                 ║
║  ╚═╝     ╚═╝ ╚═════╝ ╚═╝  ╚═╝   ╚═╝   ╚═╝  ╚═╝╚══════╝                 ║
║                                                                        ║
║  ██████╗ ██████╗  ██████╗ ███╗   ███╗██████╗ ████████╗███████╗██████╗  ║
║  ██╔══██╗██╔══██╗██╔═══██╗████╗ ████║██╔══██╗╚══██╔══╝██╔════╝██╔══██╗ ║
║  ██████╔╝██████╔╝██║   ██║██╔████╔██║██████╔╝   ██║   █████╗  ██████╔╝ ║
║  ██╔═══╝ ██╔══██╗██║   ██║██║╚██╔╝██║██╔═══╝    ██║   ██╔══╝  ██╔══██╗ ║
║  ██║     ██║  ██║╚██████╔╝██║ ╚═╝ ██║██║        ██║   ███████╗██║  ██║ ║
║  ╚═╝     ╚═╝  ╚═╝ ╚═════╝ ╚═╝     ╚═╝╚═╝        ╚═╝   ╚══════╝╚═╝  ╚═╝ ║
║                                                                        ║
╚════════════════════════════════════════════════════════════════════════╝`

	sb.WriteString(TitleStyle.Render(banner))
	sb.WriteString("\n\n")
	sb.WriteString(SuccessStyle.Render("                           CHOOSE YOUR TASK!"))
	sb.WriteString("\n\n")
	sb.WriteString(InfoStyle.Render("         Claude Code vs Codex - Code Review Battle Arena"))
	sb.WriteString("\n\n")
	sb.WriteString("═══════════════════════════════════════════════════════════════════════════")
	sb.WriteString("\n\n")

	// Prompt label
	sb.WriteString(TitleStyle.Render("  Enter your prompt for Claude Code:"))
	sb.WriteString("\n\n")

	// Textarea
	sb.WriteString("  ")
	sb.WriteString(m.textarea.View())
	sb.WriteString("\n\n")

	// Help
	sb.WriteString(HelpStyle.Render("  ctrl+s: submit  •  ctrl+c: quit"))
	sb.WriteString("\n")

	return sb.String()
}

// viewBattle renders the battle view
func (m Model) viewBattle() string {
	// Styles
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")).Bold(true)
	fighterStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF")).Bold(true)
	activeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Bold(true)
	waitingStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF"))
	warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00"))

	var sb strings.Builder

	// Box width (display characters, not bytes)
	const W = 60

	// Helper to pad a line to width W (between ║ borders)
	padLine := func(content string, contentWidth int) string {
		pad := max(0, W-contentWidth)
		return "║" + content + strings.Repeat(" ", pad) + "║\n"
	}

	// Header
	topBorder := "╔" + strings.Repeat("═", W) + "╗"
	sb.WriteString(titleStyle.Render(topBorder) + "\n")

	title := "M O R T A L   P R O M P T E R"
	titlePad := (W - len(title)) / 2
	titleLine := strings.Repeat(" ", titlePad) + title + strings.Repeat(" ", W-titlePad-len(title))
	sb.WriteString(titleStyle.Render("║"+titleLine+"║") + "\n")

	roundText := fmt.Sprintf("ROUND %d", m.currentRound)
	roundPad := (W - len(roundText)) / 2
	roundLine := strings.Repeat(" ", roundPad) + roundText + strings.Repeat(" ", W-roundPad-len(roundText))
	sb.WriteString(titleStyle.Render("║"+roundLine+"║") + "\n")

	midBorder := "╠" + strings.Repeat("═", W) + "╣"
	sb.WriteString(titleStyle.Render(midBorder) + "\n")

	// Fighter status text (plain, for width calculation)
	claudeStatusText := "WAITING"
	codexStatusText := "WAITING"

	switch m.claudeState {
	case FighterActive:
		claudeStatusText = "FIGHTING"
	case FighterFinished:
		claudeStatusText = "DONE"
	}

	switch m.codexState {
	case FighterActive:
		codexStatusText = "FIGHTING"
	case FighterFinished:
		codexStatusText = "DONE"
	}

	// Build fighter display with proper spacing
	// Layout: ║ CLAUDE CODE         VS         CODEX              ║
	// We have W=60 chars inside the box
	// Left section: 25 chars, VS: 4 chars (with spaces), Right section: 25 chars, padding: 6
	const colWidth = 25

	// Fighter names line
	claudeNamePlain := "CLAUDE CODE"
	codexNamePlain := "CODEX"
	claudeNameStyled := fighterStyle.Render(claudeNamePlain)
	codexNameStyled := fighterStyle.Render(codexNamePlain)

	line1 := " " + claudeNameStyled + strings.Repeat(" ", colWidth-1-len(claudeNamePlain)) +
		"VS" +
		strings.Repeat(" ", colWidth-len(codexNamePlain)) + codexNameStyled +
		strings.Repeat(" ", W-2*colWidth-2-1)
	sb.WriteString("║" + line1 + "║\n")

	// Health bars line
	barPlain := "[##########]"
	barStyled := activeStyle.Render(barPlain)
	line2 := " " + barStyled + strings.Repeat(" ", colWidth-1-len(barPlain)) +
		"  " +
		strings.Repeat(" ", colWidth-len(barPlain)) + barStyled +
		strings.Repeat(" ", W-2*colWidth-2-1)
	sb.WriteString("║" + line2 + "║\n")

	// Status line
	var claudeStatusStyled, codexStatusStyled string
	switch m.claudeState {
	case FighterActive:
		claudeStatusStyled = activeStyle.Render(claudeStatusText)
	case FighterFinished:
		claudeStatusStyled = infoStyle.Render(claudeStatusText)
	default:
		claudeStatusStyled = waitingStyle.Render(claudeStatusText)
	}

	switch m.codexState {
	case FighterActive:
		codexStatusStyled = activeStyle.Render(codexStatusText)
	case FighterFinished:
		codexStatusStyled = infoStyle.Render(codexStatusText)
	default:
		codexStatusStyled = waitingStyle.Render(codexStatusText)
	}

	line3 := " " + claudeStatusStyled + strings.Repeat(" ", colWidth-1-len(claudeStatusText)) +
		"  " +
		strings.Repeat(" ", colWidth-len(codexStatusText)) + codexStatusStyled +
		strings.Repeat(" ", W-2*colWidth-2-1)
	sb.WriteString("║" + line3 + "║\n")

	sb.WriteString(midBorder + "\n")

	// Round history
	for _, round := range m.rounds {
		var icon, status string
		var style lipgloss.Style

		switch round.Status {
		case "completed":
			if len(round.Issues) > 0 {
				icon = "!"
				style = warningStyle
			} else {
				icon = "+"
				style = activeStyle
			}
		case "in_progress":
			icon = "*"
			style = infoStyle
		default:
			icon = "x"
			style = warningStyle
		}

		status = round.Status
		if len(round.Issues) > 0 {
			status += fmt.Sprintf(" (%d issues)", len(round.Issues))
		}
		if round.Duration > 0 {
			status += fmt.Sprintf(" [%s]", round.Duration.Round(time.Second))
		}

		content := fmt.Sprintf(" %s Round %d: %s", icon, round.Number, status)
		contentWidth := len(content)
		styledContent := " " + style.Render(fmt.Sprintf("%s Round %d: %s", icon, round.Number, status))
		sb.WriteString(padLine(styledContent, contentWidth))
	}

	// Current action
	if m.currentAction != "" {
		actionText := m.currentAction
		if len(actionText) > W-8 {
			actionText = actionText[:W-11] + "..."
		}
		content := "  > " + actionText
		contentWidth := len(content)
		styledContent := "  " + infoStyle.Render("> "+actionText)
		sb.WriteString(padLine(styledContent, contentWidth))
	}

	sb.WriteString(midBorder + "\n")

	// Help line
	helpText := " d: details | q: abort | ?: help"
	helpWidth := len(helpText)
	sb.WriteString(padLine(helpText, helpWidth))

	bottomBorder := "╚" + strings.Repeat("═", W) + "╝"
	sb.WriteString(bottomBorder + "\n")

	return sb.String()
}


// viewResults renders the results view
func (m Model) viewResults() string {
	var sb strings.Builder

	if m.sessionSuccess {
		// Victory banner
		victory := `
╔════════════════════════════════════════════════════════════╗
║                                                            ║
║     ██╗   ██╗██╗ ██████╗████████╗ ██████╗ ██████╗ ██╗   ██╗║
║     ██║   ██║██║██╔════╝╚══██╔══╝██╔═══██╗██╔══██╗╚██╗ ██╔╝║
║     ██║   ██║██║██║        ██║   ██║   ██║██████╔╝ ╚████╔╝ ║
║     ╚██╗ ██╔╝██║██║        ██║   ██║   ██║██╔══██╗  ╚██╔╝  ║
║      ╚████╔╝ ██║╚██████╗   ██║   ╚██████╔╝██║  ██║   ██║   ║
║       ╚═══╝  ╚═╝ ╚═════╝   ╚═╝    ╚═════╝ ╚═╝  ╚═╝   ╚═╝   ║
║                                                            ║`
		sb.WriteString(VictoryStyle.Render(victory))
		sb.WriteString("\n")

		if m.sessionResult != nil && m.sessionResult.TotalRounds == 1 {
			sb.WriteString(VictoryStyle.Render("║                   FLAWLESS VICTORY!                       ║"))
		} else {
			sb.WriteString(VictoryStyle.Render("║                      YOU WIN!                              ║"))
		}
		sb.WriteString("\n")
	} else {
		// Defeat/Aborted banner
		defeat := `
╔════════════════════════════════════════════════════════════╗
║                                                            ║
║               ███████╗███╗   ██╗██████╗                    ║
║               ██╔════╝████╗  ██║██╔══██╗                   ║
║               █████╗  ██╔██╗ ██║██║  ██║                   ║
║               ██╔══╝  ██║╚██╗██║██║  ██║                   ║
║               ███████╗██║ ╚████║██████╔╝                   ║
║               ╚══════╝╚═╝  ╚═══╝╚═════╝                    ║
║                                                            ║`
		sb.WriteString(DefeatStyle.Render(defeat))
		sb.WriteString("\n")
		sb.WriteString(DefeatStyle.Render("║                   SESSION ABORTED                          ║"))
		sb.WriteString("\n")
	}

	sb.WriteString("╠════════════════════════════════════════════════════════════╣\n")

	// Stats
	if m.sessionResult != nil {
		statsLine := fmt.Sprintf("║  Rounds: %-3d │  Duration: %-10s │  Files: %-3d        ║",
			m.sessionResult.TotalRounds,
			m.sessionResult.TotalDuration.Round(time.Second),
			len(m.sessionResult.FilesModified))
		sb.WriteString(statsLine)
		sb.WriteString("\n")
	}

	// Error message if any
	if m.sessionError != nil {
		sb.WriteString("╠════════════════════════════════════════════════════════════╣\n")
		errLine := fmt.Sprintf("║  Error: %-50s ║", truncateString(m.sessionError.Error(), 50))
		sb.WriteString(ErrorStyle.Render(errLine))
		sb.WriteString("\n")
	}

	// Report path
	if m.reportPath != "" {
		sb.WriteString("╠════════════════════════════════════════════════════════════╣\n")
		reportLine := fmt.Sprintf("║  Report: %-49s ║", truncateString(m.reportPath, 49))
		sb.WriteString(InfoStyle.Render(reportLine))
		sb.WriteString("\n")
	}

	sb.WriteString("╠════════════════════════════════════════════════════════════╣\n")
	sb.WriteString(HelpStyle.Render("║  v: view diff   │   enter/q: exit                          ║"))
	sb.WriteString("\n")
	sb.WriteString("╚════════════════════════════════════════════════════════════╝\n")

	return sb.String()
}

// viewConfirmation renders the confirmation dialog
func (m Model) viewConfirmation() string {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString("╔════════════════════════════════════════════════════════════╗\n")
	sb.WriteString("║                      CONFIRMATION                          ║\n")
	sb.WriteString("╠════════════════════════════════════════════════════════════╣\n")

	// Message
	msgLine := fmt.Sprintf("║  %-56s  ║", m.confirmMessage)
	sb.WriteString(WarningStyle.Render(msgLine))
	sb.WriteString("\n")

	sb.WriteString("╠════════════════════════════════════════════════════════════╣\n")
	sb.WriteString("║           [Y] Continue          [N] Abort                  ║\n")
	sb.WriteString("╚════════════════════════════════════════════════════════════╝\n")

	return sb.String()
}

// Helper functions

// truncateString truncates a string to a maximum length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
