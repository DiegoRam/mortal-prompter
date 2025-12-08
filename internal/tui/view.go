package tui

import (
	"fmt"
	"strings"
	"time"
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
	// All lines are exactly 62 characters wide (60 content + 2 borders)
	const W = 60
	var lines []string

	border := strings.Repeat("=", W)

	lines = append(lines, "+"+border+"+")
	lines = append(lines, "|"+centerText("M O R T A L   P R O M P T E R", W)+"|")
	lines = append(lines, "|"+centerText(fmt.Sprintf("ROUND %d", m.currentRound), W)+"|")
	lines = append(lines, "+"+border+"+")

	// Fighter names and status
	claudeStatus := "WAITING"
	codexStatus := "WAITING"

	switch m.claudeState {
	case FighterActive:
		claudeStatus = "FIGHTING"
	case FighterFinished:
		claudeStatus = "DONE"
	}

	switch m.codexState {
	case FighterActive:
		codexStatus = "FIGHTING"
	case FighterFinished:
		codexStatus = "DONE"
	}

	// Fighter line: "CLAUDE CODE          VS          CODEX"
	fighterLine := fmt.Sprintf("%-20s VS %20s", "CLAUDE CODE", "CODEX")
	lines = append(lines, "|"+centerText(fighterLine, W)+"|")

	// Health bars
	barLine := fmt.Sprintf("%-20s    %20s", "[##########]", "[##########]")
	lines = append(lines, "|"+centerText(barLine, W)+"|")

	// Status line
	statusLine := fmt.Sprintf("%-20s    %20s", claudeStatus, codexStatus)
	lines = append(lines, "|"+centerText(statusLine, W)+"|")

	lines = append(lines, "+"+border+"+")

	// Round history
	for _, round := range m.rounds {
		var icon string
		switch round.Status {
		case "completed":
			if len(round.Issues) > 0 {
				icon = "!"
			} else {
				icon = "+"
			}
		case "in_progress":
			icon = "*"
		default:
			icon = "x"
		}

		content := fmt.Sprintf("%s Round %d: %s", icon, round.Number, round.Status)
		if len(round.Issues) > 0 {
			content += fmt.Sprintf(" (%d issues)", len(round.Issues))
		}
		if round.Duration > 0 {
			content += fmt.Sprintf(" [%s]", round.Duration.Round(time.Second))
		}

		lines = append(lines, "|"+padRight(" "+content, W)+"|")
	}

	// Current action
	if m.currentAction != "" {
		action := "  > " + m.currentAction
		if len(action) > W-1 {
			action = action[:W-4] + "..."
		}
		lines = append(lines, "|"+padRight(action, W)+"|")
	}

	lines = append(lines, "+"+border+"+")
	lines = append(lines, "|"+padRight(" d: details | q: abort | ?: help", W)+"|")
	lines = append(lines, "+"+border+"+")

	return strings.Join(lines, "\n") + "\n"
}

// centerText centers text within a given width
func centerText(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	padding := (width - len(s)) / 2
	return strings.Repeat(" ", padding) + s + strings.Repeat(" ", width-padding-len(s))
}

// padRight pads a string to the right to reach the specified width
func padRight(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	return s + strings.Repeat(" ", width-len(s))
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
