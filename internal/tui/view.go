package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/diegoram/mortal-prompter/internal/fighters"
)

// View renders the current view
func (m Model) View() string {
	switch m.view {
	case ViewFighterSelect:
		return m.viewFighterSelect()
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

// viewFighterSelect renders the fighter selection view
func (m Model) viewFighterSelect() string {
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

	// Working directory display
	if m.config != nil && m.config.WorkDir != "" {
		workDir := m.config.WorkDir
		// Truncate if too long (keep end of path which is more meaningful)
		maxLen := 60
		if len(workDir) > maxLen {
			workDir = "..." + workDir[len(workDir)-maxLen+3:]
		}
		sb.WriteString(HelpStyle.Render("  📁 Working Directory: " + workDir))
		sb.WriteString("\n\n")
	}

	sb.WriteString(SuccessStyle.Render("                         CHOOSE YOUR FIGHTERS!"))
	sb.WriteString("\n\n")
	sb.WriteString("═══════════════════════════════════════════════════════════════════════════")
	sb.WriteString("\n\n")

	// Fighter selection
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Bold(true)
	unselectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")).Bold(true)
	activeFieldStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF")).Bold(true)

	// Implementer selection
	implementerLabel := "  IMPLEMENTER: "
	if m.fighterSelectField == FieldImplementer {
		implementerLabel = activeFieldStyle.Render("▶ IMPLEMENTER: ")
	} else {
		implementerLabel = labelStyle.Render("  IMPLEMENTER: ")
	}
	sb.WriteString(implementerLabel)
	sb.WriteString(m.renderFighterOptions(m.implementerType, selectedStyle, unselectedStyle))
	sb.WriteString("\n\n")

	// Reviewer selection
	reviewerLabel := "  REVIEWER:    "
	if m.fighterSelectField == FieldReviewer {
		reviewerLabel = activeFieldStyle.Render("▶ REVIEWER:    ")
	} else {
		reviewerLabel = labelStyle.Render("  REVIEWER:    ")
	}
	sb.WriteString(reviewerLabel)
	sb.WriteString(m.renderFighterOptions(m.reviewerType, selectedStyle, unselectedStyle))
	sb.WriteString("\n\n")

	sb.WriteString("═══════════════════════════════════════════════════════════════════════════")
	sb.WriteString("\n\n")

	// Help
	sb.WriteString(HelpStyle.Render("  ←/→: select fighter  •  ↑/↓: switch field  •  enter: continue  •  ctrl+c: quit"))
	sb.WriteString("\n")

	return sb.String()
}

// renderFighterOptions renders the fighter options for selection
func (m Model) renderFighterOptions(selected fighters.FighterType, selectedStyle, unselectedStyle lipgloss.Style) string {
	var parts []string
	for _, ft := range m.availableFighters {
		name := strings.ToUpper(string(ft))
		if ft == selected {
			parts = append(parts, selectedStyle.Render("["+name+"]"))
		} else {
			parts = append(parts, unselectedStyle.Render(" "+name+" "))
		}
	}
	return strings.Join(parts, "  ")
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

	// Image attachment indicator
	if m.attachedImage != nil {
		sb.WriteString("  ")
		sb.WriteString(ImageAttachedStyle.Render("[IMAGE ATTACHED]"))
		sb.WriteString(" ")
		imageInfo := fmt.Sprintf("%dx%d PNG", m.attachedImage.Width, m.attachedImage.Height)
		sb.WriteString(ImageInfoStyle.Render(imageInfo))
		sb.WriteString("\n")
		sb.WriteString(HelpStyle.Render("  ctrl+x: remove image"))
		sb.WriteString("\n\n")
	}

	// Show temporary image message if any
	if m.imageMessage != "" {
		sb.WriteString("  ")
		sb.WriteString(InfoStyle.Render(m.imageMessage))
		sb.WriteString("\n\n")
	}

	// Prompt label
	sb.WriteString(TitleStyle.Render("  Enter your prompt for Claude Code:"))
	sb.WriteString("\n\n")

	// Textarea
	sb.WriteString("  ")
	sb.WriteString(m.textarea.View())
	sb.WriteString("\n\n")

	// Help - include paste image hint
	helpText := "  ctrl+s: submit  •  ctrl+v: paste image  •  ctrl+c: quit"
	sb.WriteString(HelpStyle.Render(helpText))
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
	implementerStatusText := "WAITING"
	reviewerStatusText := "WAITING"

	switch m.implementerState {
	case FighterActive:
		implementerStatusText = "FIGHTING"
	case FighterFinished:
		implementerStatusText = "DONE"
	}

	switch m.reviewerState {
	case FighterActive:
		reviewerStatusText = "FIGHTING"
	case FighterFinished:
		reviewerStatusText = "DONE"
	}

	// Build fighter display with proper spacing
	// Layout: ║ IMPLEMENTER         VS         REVIEWER              ║
	// We have W=60 chars inside the box
	// Left section: 25 chars, VS: 4 chars (with spaces), Right section: 25 chars, padding: 6
	const colWidth = 25

	// Get fighter names (use defaults if not set)
	implementerName := m.implementerName
	if implementerName == "" {
		implementerName = strings.ToUpper(string(m.implementerType))
	}
	reviewerName := m.reviewerName
	if reviewerName == "" {
		reviewerName = strings.ToUpper(string(m.reviewerType))
	}

	// Fighter names line
	implementerNameStyled := fighterStyle.Render(implementerName)
	reviewerNameStyled := fighterStyle.Render(reviewerName)

	line1 := " " + implementerNameStyled + strings.Repeat(" ", colWidth-1-len(implementerName)) +
		"VS" +
		strings.Repeat(" ", colWidth-len(reviewerName)) + reviewerNameStyled +
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

	// Status line - use blinking effect for "FIGHTING" state
	var implementerStatusStyled, reviewerStatusStyled string
	switch m.implementerState {
	case FighterActive:
		// Use blinking style for active fighter
		if m.blinkOn {
			implementerStatusStyled = FightingBlinkOnStyle.Render(implementerStatusText)
		} else {
			implementerStatusStyled = FightingBlinkOffStyle.Render(implementerStatusText)
		}
	case FighterFinished:
		implementerStatusStyled = infoStyle.Render(implementerStatusText)
	default:
		implementerStatusStyled = waitingStyle.Render(implementerStatusText)
	}

	switch m.reviewerState {
	case FighterActive:
		// Use blinking style for active fighter
		if m.blinkOn {
			reviewerStatusStyled = FightingBlinkOnStyle.Render(reviewerStatusText)
		} else {
			reviewerStatusStyled = FightingBlinkOffStyle.Render(reviewerStatusText)
		}
	case FighterFinished:
		reviewerStatusStyled = infoStyle.Render(reviewerStatusText)
	default:
		reviewerStatusStyled = waitingStyle.Render(reviewerStatusText)
	}

	line3 := " " + implementerStatusStyled + strings.Repeat(" ", colWidth-1-len(implementerStatusText)) +
		"  " +
		strings.Repeat(" ", colWidth-len(reviewerStatusText)) + reviewerStatusStyled +
		strings.Repeat(" ", W-2*colWidth-2-1)
	sb.WriteString("║" + line3 + "║\n")

	sb.WriteString(midBorder + "\n")

	// Prompt section
	if m.prompt != "" {
		// Show image indicator if attached
		if m.attachedImage != nil {
			imgLabel := " [+IMG] "
			imgLabelWidth := len(imgLabel)
			styledImgLabel := ImageAttachedStyle.Render(imgLabel)
			sb.WriteString(padLine(styledImgLabel, imgLabelWidth))
		}

		promptLabel := " PROMPT: "
		promptLabelWidth := len(promptLabel)
		styledLabel := warningStyle.Render(promptLabel)
		sb.WriteString(padLine(styledLabel, promptLabelWidth))

		// Show the prompt, truncating if necessary and wrapping to multiple lines
		maxPromptWidth := W - 4 // Leave space for "  " prefix and padding
		promptText := m.prompt
		// Replace newlines with spaces for display
		promptText = strings.ReplaceAll(promptText, "\n", " ")
		promptText = strings.ReplaceAll(promptText, "\r", "")

		// Truncate if too long (show first 2 lines worth)
		maxTotalLen := maxPromptWidth * 2
		if len(promptText) > maxTotalLen {
			promptText = promptText[:maxTotalLen-3] + "..."
		}

		// Split into lines
		for len(promptText) > 0 {
			lineLen := min(len(promptText), maxPromptWidth)
			lineText := "  " + promptText[:lineLen]
			lineWidth := len(lineText)
			sb.WriteString(padLine(lineText, lineWidth))
			promptText = promptText[lineLen:]
		}

		sb.WriteString(midBorder + "\n")
	}

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

	// Log file path
	if m.logFilePath != "" {
		logLabel := " LOG: "
		// Truncate path if too long
		maxPathLen := W - len(logLabel) - 2
		logPath := m.logFilePath
		if len(logPath) > maxPathLen {
			logPath = "..." + logPath[len(logPath)-maxPathLen+3:]
		}
		logLine := logLabel + logPath
		logWidth := len(logLine)
		styledLog := " " + waitingStyle.Render("LOG:") + " " + waitingStyle.Render(logPath)
		sb.WriteString(padLine(styledLog, logWidth))
		sb.WriteString(midBorder + "\n")
	}

	// Working directory
	if m.config != nil && m.config.WorkDir != "" {
		dirLabel := " DIR: "
		// Truncate path if too long
		maxPathLen := W - len(dirLabel) - 2
		dirPath := m.config.WorkDir
		if len(dirPath) > maxPathLen {
			dirPath = "..." + dirPath[len(dirPath)-maxPathLen+3:]
		}
		dirLine := dirLabel + dirPath
		dirWidth := len(dirLine)
		styledDir := " " + waitingStyle.Render("DIR:") + " " + waitingStyle.Render(dirPath)
		sb.WriteString(padLine(styledDir, dirWidth))
		sb.WriteString(midBorder + "\n")
	}

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
