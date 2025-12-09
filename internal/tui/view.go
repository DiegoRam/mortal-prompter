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

	// Fighter sprite styles
	leftFighterColor := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF")).Bold(true)
	rightFighterColor := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6600")).Bold(true)
	impactStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")).Bold(true)

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

	// Build fighter names line with proper alignment
	// Layout: [space][name][padding][VS][padding][name][space]
	vsText := "VS"
	leftPad := (W - len(implementerName) - len(vsText) - len(reviewerName)) / 3
	rightPad := W - len(implementerName) - leftPad - len(vsText) - leftPad - len(reviewerName)

	line1 := implementerNameStyled + strings.Repeat(" ", leftPad) +
		vsText +
		strings.Repeat(" ", leftPad) + reviewerNameStyled +
		strings.Repeat(" ", rightPad)
	sb.WriteString("║" + line1 + "║\n")

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

	// Status line with same alignment as names
	statusLeftPad := (W - len(implementerStatusText) - 2 - len(reviewerStatusText)) / 3
	statusRightPad := W - len(implementerStatusText) - statusLeftPad - 2 - statusLeftPad - len(reviewerStatusText)

	line3 := implementerStatusStyled + strings.Repeat(" ", statusLeftPad) +
		"  " +
		strings.Repeat(" ", statusLeftPad) + reviewerStatusStyled +
		strings.Repeat(" ", statusRightPad)
	sb.WriteString("║" + line3 + "║\n")

	// === ANIMATED FIGHTER SPRITES ===
	// Get current animation frames based on state
	leftFrames := m.getLeftFighterFrames()
	rightFrames := m.getRightFighterFrames()

	// Select frame based on animation counter
	leftFrameIdx := (m.animFrame / 2) % len(leftFrames)
	rightFrameIdx := (m.animFrame / 2) % len(rightFrames)

	leftSprite := leftFrames[leftFrameIdx]
	rightSprite := rightFrames[rightFrameIdx]

	// Get impact effect if both are fighting
	showImpact := m.leftAttacking && m.rightAttacking

	// Render fighter arena (5 lines for fighters)
	const fighterWidth = 12
	const gapWidth = W - 2*fighterWidth // gap between fighters

	for i := range leftSprite {
		leftLine := leftSprite[i]
		rightLine := rightSprite[i]

		// Pad sprites to fixed width
		for len(leftLine) < fighterWidth {
			leftLine += " "
		}
		for len(rightLine) < fighterWidth {
			rightLine = " " + rightLine
		}

		// Build the arena line
		var arenaLine string

		// Add impact effects in the middle during combat
		if showImpact && i == 1 && m.blinkOn {
			// Both fighters attacking - show clash!
			impact := impactStyle.Render("*CLASH*")
			impactLen := 7
			impactPad := (gapWidth - impactLen) / 2
			gap := strings.Repeat(" ", impactPad) + impact + strings.Repeat(" ", gapWidth-impactPad-impactLen)
			arenaLine = leftFighterColor.Render(leftLine) + gap + rightFighterColor.Render(rightLine)
		} else if m.leftAttacking && i == 1 && m.blinkOn {
			// Left attacking effect - show punch!
			effect := impactStyle.Render(">>>*")
			arenaLine = leftFighterColor.Render(leftLine) + effect + strings.Repeat(" ", gapWidth-4) + rightFighterColor.Render(rightLine)
		} else if m.rightAttacking && i == 1 && m.blinkOn {
			// Right attacking effect - show punch!
			effect := impactStyle.Render("*<<<")
			arenaLine = leftFighterColor.Render(leftLine) + strings.Repeat(" ", gapWidth-4) + effect + rightFighterColor.Render(rightLine)
		} else {
			gap := strings.Repeat(" ", gapWidth)
			arenaLine = leftFighterColor.Render(leftLine) + gap + rightFighterColor.Render(rightLine)
		}

		// Calculate actual display width (without ANSI codes)
		displayWidth := fighterWidth + gapWidth + fighterWidth
		sb.WriteString(padLine(arenaLine, displayWidth))
	}

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

	// Box width = 60 characters inside (between ║ borders)
	const boxW = 60

	if m.sessionSuccess {
		// Victory banner - each line is exactly 60 chars inside
		victory := `
╔════════════════════════════════════════════════════════════╗
║                                                            ║
║    ██╗   ██╗██╗ ██████╗████████╗ ██████╗ ██████╗ ██╗   ██╗ ║
║    ██║   ██║██║██╔════╝╚══██╔══╝██╔═══██╗██╔══██╗╚██╗ ██╔╝ ║
║    ██║   ██║██║██║        ██║   ██║   ██║██████╔╝ ╚████╔╝  ║
║    ╚██╗ ██╔╝██║██║        ██║   ██║   ██║██╔══██╗  ╚██╔╝   ║
║     ╚████╔╝ ██║╚██████╗   ██║   ╚██████╔╝██║  ██║   ██║    ║
║      ╚═══╝  ╚═╝ ╚═════╝   ╚═╝    ╚═════╝ ╚═╝  ╚═╝   ╚═╝    ║
║                                                            ║`
		sb.WriteString(VictoryStyle.Render(victory))
		sb.WriteString("\n")

		if m.sessionResult != nil && m.sessionResult.TotalRounds == 1 {
			sb.WriteString(VictoryStyle.Render("║                    FLAWLESS VICTORY!                      ║"))
		} else {
			sb.WriteString(VictoryStyle.Render("║                        YOU WIN!                           ║"))
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
		sb.WriteString(DefeatStyle.Render("║                    SESSION ABORTED                         ║"))
		sb.WriteString("\n")
	}

	sb.WriteString("╠════════════════════════════════════════════════════════════╣\n")

	// Stats - use fixed width formatting
	if m.sessionResult != nil {
		rounds := fmt.Sprintf("Rounds: %d", m.sessionResult.TotalRounds)
		duration := fmt.Sprintf("Duration: %s", m.sessionResult.TotalDuration.Round(time.Second))
		files := fmt.Sprintf("Files: %d", len(m.sessionResult.FilesModified))
		statsContent := fmt.Sprintf("  %-12s  │  %-16s  │  %-10s  ", rounds, duration, files)
		// Pad to exactly boxW
		for len(statsContent) < boxW {
			statsContent += " "
		}
		sb.WriteString("║" + statsContent + "║\n")
	}

	// Error message if any
	if m.sessionError != nil {
		sb.WriteString("╠════════════════════════════════════════════════════════════╣\n")
		errText := "  Error: " + truncateString(m.sessionError.Error(), boxW-12)
		for len(errText) < boxW {
			errText += " "
		}
		sb.WriteString(ErrorStyle.Render("║" + errText + "║"))
		sb.WriteString("\n")
	}

	// Report path
	if m.reportPath != "" {
		sb.WriteString("╠════════════════════════════════════════════════════════════╣\n")
		reportText := "  Report: " + truncateString(m.reportPath, boxW-12)
		for len(reportText) < boxW {
			reportText += " "
		}
		sb.WriteString(InfoStyle.Render("║" + reportText + "║"))
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

// Fighter sprite frames - idle stance for left fighter (facing right)
var leftIdleFrames = [][]string{
	{
		"   .O.   ",
		"  --|--  ",
		"   /|    ",
		"  | |    ",
		"  d b    ",
	},
	{
		"   .O.   ",
		"  --|--  ",
		"   /|    ",
		"  / \\    ",
		" d   b   ",
	},
}

// Fighter sprite frames - attack stance for left fighter
var leftAttackFrames = [][]string{
	{
		"   .O.   ",
		"  --|==> ",
		"   /|    ",
		"  | |    ",
		"  d b    ",
	},
	{
		"   .O_   ",
		"  --|===>",
		"   /|    ",
		"  | |    ",
		"  d b    ",
	},
	{
		"   .O)   ",
		"  --X===>",
		"   /|    ",
		"  | |    ",
		"  d b    ",
	},
}

// Fighter sprite frames - idle stance for right fighter (facing left)
var rightIdleFrames = [][]string{
	{
		"   .O.   ",
		"  --|--  ",
		"    |\\   ",
		"    | |  ",
		"    d b  ",
	},
	{
		"   .O.   ",
		"  --|--  ",
		"    |\\   ",
		"    / \\  ",
		"   d   b ",
	},
}

// Fighter sprite frames - attack stance for right fighter
var rightAttackFrames = [][]string{
	{
		"   .O.   ",
		" <==|--  ",
		"    |\\   ",
		"    | |  ",
		"    d b  ",
	},
	{
		"   _O.   ",
		"<===|--  ",
		"    |\\   ",
		"    | |  ",
		"    d b  ",
	},
	{
		"   (O.   ",
		"<===X--  ",
		"    |\\   ",
		"    | |  ",
		"    d b  ",
	},
}

// getLeftFighterFrames returns the appropriate frames for the left fighter based on state
func (m Model) getLeftFighterFrames() [][]string {
	if m.leftAttacking {
		return leftAttackFrames
	}
	return leftIdleFrames
}

// getRightFighterFrames returns the appropriate frames for the right fighter based on state
func (m Model) getRightFighterFrames() [][]string {
	if m.rightAttacking {
		return rightAttackFrames
	}
	return rightIdleFrames
}
