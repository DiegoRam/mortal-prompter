// Package logger provides arcade-style terminal output and file logging
// for the mortal-prompter CLI application.
package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
)

// Logger handles both terminal output with colors/emojis and file logging
// for the mortal-prompter application.
type Logger struct {
	verbose    bool
	silentMode bool // When true, don't print to terminal (for TUI mode)
	logFile    *os.File
	outputDir  string
	spinner    *spinner.Spinner
	mu         sync.Mutex

	// Writers for terminal output (allows injection for testing)
	stdout io.Writer
	stderr io.Writer
}

// New creates a new Logger instance.
// It creates the output directory if it doesn't exist and initializes
// a log file with the current timestamp.
func New(outputDir string, verbose bool) (*Logger, error) {
	// Use default output directory if not specified
	if outputDir == "" {
		outputDir = ".mortal-prompter"
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create log file with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logFileName := fmt.Sprintf("session-%s.log", timestamp)
	logFilePath := filepath.Join(outputDir, logFileName)

	logFile, err := os.Create(logFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}

	return &Logger{
		verbose:   verbose,
		logFile:   logFile,
		outputDir: outputDir,
		spinner:   nil,
		stdout:    os.Stdout,
		stderr:    os.Stderr,
	}, nil
}

// SetOutputWriters allows setting custom writers for testing purposes.
func (l *Logger) SetOutputWriters(stdout, stderr io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.stdout = stdout
	l.stderr = stderr
}

// SetSilentMode enables or disables silent mode.
// In silent mode, the logger only writes to the log file, not to terminal.
// This is useful when running in TUI mode.
func (l *Logger) SetSilentMode(silent bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.silentMode = silent
}

// Close closes the log file handle.
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.StopSpinnerInternal()

	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}

// writeToFile writes a message to the log file without colors or emojis.
func (l *Logger) writeToFile(format string, args ...interface{}) {
	if l.logFile == nil {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	msg := fmt.Sprintf(format, args...)
	// Strip emojis and special characters for clean logs
	cleanMsg := stripEmojis(msg)
	fmt.Fprintf(l.logFile, "[%s] %s\n", timestamp, cleanMsg)
}

// stripEmojis removes common emojis from a string for clean log output.
func stripEmojis(s string) string {
	replacer := strings.NewReplacer(
		"\U0001F3AE", "", // Game controller
		"\U0001F94A", "", // Boxing glove
		"\u23F3", "",     // Hourglass
		"\u2705", "",     // Check mark
		"\u26A0\uFE0F", "", // Warning sign
		"\u26A0", "",     // Warning sign (without variation selector)
		"\U0001F4DD", "", // Memo
		"\U0001F50D", "", // Magnifying glass
		"\U0001F504", "", // Arrows
		"\U0001F3C6", "", // Trophy
		"\u274C", "",     // Cross mark
		"\U0001F535", "", // Blue circle
		"\U0001F7E2", "", // Green circle
	)
	return replacer.Replace(s)
}

// printToTerminal prints a message to stdout (unless in silent mode).
func (l *Logger) printToTerminal(msg string) {
	if l.silentMode {
		return
	}
	fmt.Fprintln(l.stdout, msg)
}

// printToTerminalErr prints a message to stderr (unless in silent mode).
func (l *Logger) printToTerminalErr(msg string) {
	if l.silentMode {
		return
	}
	fmt.Fprintln(l.stderr, msg)
}

// RoundStart displays a round banner with the round number.
func (l *Logger) RoundStart(number int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.StopSpinnerInternal()

	banner := fmt.Sprintf(`
%s
%s MORTAL PROMPTER - ROUND %d
%s`,
		strings.Repeat("\u2550", 60),
		"\U0001F3AE", // Game controller emoji
		number,
		strings.Repeat("\u2550", 60),
	)

	yellow := color.New(color.FgYellow, color.Bold)
	l.printToTerminal(yellow.Sprint(banner))
	l.writeToFile("ROUND %d START", number)
}

// FighterEnter displays a message when a fighter enters the arena.
func (l *Logger) FighterEnter(name string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.StopSpinnerInternal()

	msg := fmt.Sprintf("\U0001F94A %s enters the arena...", name)
	cyan := color.New(color.FgCyan, color.Bold)
	l.printToTerminal(cyan.Sprint(msg))
	l.writeToFile("%s enters the arena", name)
}

// FighterAction displays a message during fighter execution and starts a spinner.
func (l *Logger) FighterAction(action string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.StopSpinnerInternal()

	l.writeToFile("%s", action)

	// Don't start spinner in silent mode
	if l.silentMode {
		return
	}

	// Create and start spinner with hourglass prefix
	l.spinner = spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	l.spinner.Writer = l.stdout
	l.spinner.Prefix = "\u23F3 "
	l.spinner.Suffix = " " + action
	l.spinner.Start()
}

// FighterFinish displays a message when a fighter completes its task.
func (l *Logger) FighterFinish(name string, duration time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.StopSpinnerInternal()

	msg := fmt.Sprintf("\u2705 %s finishes! (took %s)", name, formatDuration(duration))
	green := color.New(color.FgGreen, color.Bold)
	l.printToTerminal(green.Sprint(msg))
	l.writeToFile("%s finished (took %s)", name, formatDuration(duration))
}

// IssuesFound displays the issues found by the reviewer.
func (l *Logger) IssuesFound(issues []string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.StopSpinnerInternal()

	msg := fmt.Sprintf("\u26A0\uFE0F  CODEX found %d issue(s)!", len(issues))
	yellow := color.New(color.FgYellow, color.Bold)
	l.printToTerminal(yellow.Sprint(msg))
	l.printToTerminal("")

	l.writeToFile("Issues found: %d", len(issues))

	for i, issue := range issues {
		issueMsg := fmt.Sprintf("   ISSUE %d: %s", i+1, issue)
		l.printToTerminal(color.YellowString(issueMsg))
		l.writeToFile("  ISSUE %d: %s", i+1, issue)
	}

	l.printToTerminal("")
}

// NoIssues displays a success message when no issues are found.
func (l *Logger) NoIssues() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.StopSpinnerInternal()

	msg := "\u2705 LGTM - No issues found!"
	green := color.New(color.FgGreen, color.Bold)
	l.printToTerminal(green.Sprint(msg))
	l.writeToFile("LGTM - No issues found")
}

// FinalVictory displays the final victory banner.
func (l *Logger) FinalVictory(totalRounds int, totalDuration time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.StopSpinnerInternal()

	banner := fmt.Sprintf(`
%s
%s FLAWLESS VICTORY!
%s

   Total Rounds:   %d
   Total Duration: %s

%s FINISH HIM! %s
%s`,
		strings.Repeat("\u2550", 60),
		"\U0001F3C6", // Trophy emoji
		strings.Repeat("\u2550", 60),
		totalRounds,
		formatDuration(totalDuration),
		"\U0001F94A", // Boxing glove
		"\U0001F94A",
		strings.Repeat("\u2550", 60),
	)

	green := color.New(color.FgGreen, color.Bold)
	l.printToTerminal(green.Sprint(banner))
	l.writeToFile("FLAWLESS VICTORY! Total rounds: %d, Duration: %s", totalRounds, formatDuration(totalDuration))
}

// Error displays an error message in red.
func (l *Logger) Error(err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.StopSpinnerInternal()

	msg := fmt.Sprintf("\u274C ERROR: %s", err.Error())
	red := color.New(color.FgRed, color.Bold)
	l.printToTerminalErr(red.Sprint(msg))
	l.writeToFile("ERROR: %s", err.Error())
}

// Info displays an info message in cyan.
func (l *Logger) Info(msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.StopSpinnerInternal()

	infoMsg := fmt.Sprintf("\U0001F535 %s", msg)
	cyan := color.New(color.FgCyan)
	l.printToTerminal(cyan.Sprint(infoMsg))
	l.writeToFile("INFO: %s", msg)
}

// Debug displays a debug message only when verbose mode is enabled.
func (l *Logger) Debug(msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.verbose {
		return
	}

	l.StopSpinnerInternal()

	debugMsg := fmt.Sprintf("[DEBUG] %s", msg)
	gray := color.New(color.FgHiBlack)
	l.printToTerminal(gray.Sprint(debugMsg))
	l.writeToFile("DEBUG: %s", msg)
}

// StartSpinner starts a spinner animation with the given message.
func (l *Logger) StartSpinner(message string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.StopSpinnerInternal()

	// Don't start spinner in silent mode
	if l.silentMode {
		return
	}

	l.spinner = spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	l.spinner.Writer = l.stdout
	l.spinner.Suffix = " " + message
	l.spinner.Start()
}

// StopSpinner stops the current spinner animation.
func (l *Logger) StopSpinner() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.StopSpinnerInternal()
}

// StopSpinnerInternal stops the spinner without acquiring lock (for internal use).
func (l *Logger) StopSpinnerInternal() {
	if l.spinner != nil && l.spinner.Active() {
		l.spinner.Stop()
		l.spinner = nil
	}
}

// formatDuration formats a duration in a human-readable way.
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	if seconds == 0 {
		return fmt.Sprintf("%dm", minutes)
	}
	return fmt.Sprintf("%dm %ds", minutes, seconds)
}

// GetLogFilePath returns the path to the current log file.
func (l *Logger) GetLogFilePath() string {
	if l.logFile == nil {
		return ""
	}
	return l.logFile.Name()
}

// GetOutputDir returns the output directory path.
func (l *Logger) GetOutputDir() string {
	return l.outputDir
}

// IsVerbose returns whether verbose mode is enabled.
func (l *Logger) IsVerbose() bool {
	return l.verbose
}

// ChangesDetected logs the number of files modified.
func (l *Logger) ChangesDetected(fileCount int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.StopSpinnerInternal()

	msg := fmt.Sprintf("\U0001F4DD Changes detected: %d file(s) modified", fileCount)
	cyan := color.New(color.FgCyan)
	l.printToTerminal(cyan.Sprint(msg))
	l.writeToFile("Changes detected: %d file(s) modified", fileCount)
}

// PreparingNextRound displays a message before the next round.
func (l *Logger) PreparingNextRound() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.StopSpinnerInternal()

	msg := "\U0001F504 Preparing next round..."
	yellow := color.New(color.FgYellow)
	l.printToTerminal(yellow.Sprint(msg))
	l.writeToFile("Preparing next round")
}
