// Package main is the entry point for the mortal-prompter CLI application.
// Mortal Prompter orchestrates a code review battle between Claude Code and OpenAI Codex.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fatih/color"
	"github.com/diegoram/mortal-prompter/internal/config"
	"github.com/diegoram/mortal-prompter/internal/logger"
	"github.com/diegoram/mortal-prompter/internal/orchestrator"
	"github.com/diegoram/mortal-prompter/internal/reporter"
	"github.com/diegoram/mortal-prompter/internal/tui"
	"github.com/diegoram/mortal-prompter/pkg/types"
	"github.com/spf13/cobra"
)

// Version information - set at build time via ldflags
var (
	Version   = "dev"
	BuildTime = "unknown"
)

// Color definitions for arcade-style output
var (
	titleColor   = color.New(color.FgHiYellow, color.Bold)
	successColor = color.New(color.FgHiGreen, color.Bold)
	infoColor    = color.New(color.FgHiCyan)
	errorColor   = color.New(color.FgHiRed, color.Bold)
)

func main() {
	if err := execute(); err != nil {
		errorColor.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// execute sets up and runs the root Cobra command.
func execute() error {
	cfg := config.New()

	rootCmd := &cobra.Command{
		Use:   "mortal-prompter",
		Short: "Orchestrate code review battles between Claude Code and Codex",
		Long: `Mortal Prompter - A CLI that orchestrates a development and code review loop
between Claude Code (implementer) and OpenAI Codex (reviewer).

The tool acts as a referee in a Mortal Kombat-style battle:
  - CLAUDE CODE (Fighter 1): Executes development/implementation tasks
  - CODEX (Fighter 2): Reviews the code and finds issues

The loop continues until Codex finds no more issues or the iteration limit is reached.

Example usage:
  mortal-prompter -p "implement JWT authentication" --auto-commit -v
  mortal-prompter --prompt "add unit tests for users module" -m 5 -i`,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if version flag was explicitly requested
			versionFlag, _ := cmd.Flags().GetBool("version")
			if versionFlag {
				printVersion()
				return nil
			}

			// Determine if we should use CLI mode:
			// - If --no-tui is set, use CLI mode
			// - If -p/--prompt is provided, use CLI mode (for backwards compatibility)
			useCLI := cfg.NoTUI || cfg.Prompt != ""

			if useCLI {
				return runCLI(cfg)
			}
			return runTUI(cfg)
		},
	}

	// Bind configuration flags
	cfg.BindFlags(rootCmd)

	// Add version flag
	rootCmd.Flags().Bool("version", false, "Display version information and exit")

	return rootCmd.Execute()
}

// printBanner displays the arcade-style startup banner.
func printBanner() {
	banner := `
╔══════════════════════════════════════════════════════════════════════════╗
║                                                                          ║
║   ███╗   ███╗ ██████╗ ██████╗ ████████╗ █████╗ ██╗                       ║
║   ████╗ ████║██╔═══██╗██╔══██╗╚══██╔══╝██╔══██╗██║                       ║
║   ██╔████╔██║██║   ██║██████╔╝   ██║   ███████║██║                       ║
║   ██║╚██╔╝██║██║   ██║██╔══██╗   ██║   ██╔══██║██║                       ║
║   ██║ ╚═╝ ██║╚██████╔╝██║  ██║   ██║   ██║  ██║███████╗                  ║
║   ╚═╝     ╚═╝ ╚═════╝ ╚═╝  ╚═╝   ╚═╝   ╚═╝  ╚═╝╚══════╝                  ║
║                                                                          ║
║   ██████╗ ██████╗  ██████╗ ███╗   ███╗██████╗ ████████╗███████╗██████╗   ║
║   ██╔══██╗██╔══██╗██╔═══██╗████╗ ████║██╔══██╗╚══██╔══╝██╔════╝██╔══██╗  ║
║   ██████╔╝██████╔╝██║   ██║██╔████╔██║██████╔╝   ██║   █████╗  ██████╔╝  ║
║   ██╔═══╝ ██╔══██╗██║   ██║██║╚██╔╝██║██╔═══╝    ██║   ██╔══╝  ██╔══██╗  ║
║   ██║     ██║  ██║╚██████╔╝██║ ╚═╝ ██║██║        ██║   ███████╗██║  ██║  ║
║   ╚═╝     ╚═╝  ╚═╝ ╚═════╝ ╚═╝     ╚═╝╚═╝        ╚═╝   ╚══════╝╚═╝  ╚═╝  ║
║                                                                          ║
╚══════════════════════════════════════════════════════════════════════════╝
`
	titleColor.Print(banner)
	fmt.Println()
	successColor.Println("                         FIGHT!")
	fmt.Println()
	infoColor.Printf("       Claude Code vs Codex - Code Review Battle Arena\n")
	infoColor.Printf("       Version: %s (built: %s)\n", Version, BuildTime)
	fmt.Println()
	fmt.Println("══════════════════════════════════════════════════════════════════════════")
	fmt.Println()
}

// printVersion displays version information.
func printVersion() {
	fmt.Printf("mortal-prompter version %s\n", Version)
	fmt.Printf("Built: %s\n", BuildTime)
}

// runTUI runs the TUI-based interface
func runTUI(cfg *config.Config) error {
	// Validate and prepare working directory
	if err := validateWorkDir(cfg); err != nil {
		return err
	}

	// Initialize logger (for file logging, even in TUI mode)
	log, err := logger.New(cfg.OutputDir, cfg.Verbose)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer log.Close()

	// Create TUI model
	model := tui.NewModel(cfg)

	// Setup context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	// Create and run the program
	p := tea.NewProgram(model, tea.WithAltScreen())

	// Run TUI - it will handle prompts and battle internally
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	m := finalModel.(tui.Model)

	// If battle wasn't started (user quit from prompt), exit gracefully
	if !m.IsBattleStarted() {
		return nil
	}

	// Get the prompt from TUI
	prompt := m.GetPrompt()
	if prompt == "" {
		return nil
	}

	// Now run the actual battle (TUI was just for input)
	// Set the prompt in config
	cfg.Prompt = prompt

	// Create a new TUI model for battle phase
	battleModel := tui.NewModel(cfg)
	battleModel.SetBattleStarted(prompt)

	// Create observer using the battle model's channels
	observer := tui.NewChannelObserver(battleModel.GetEventChannel(), battleModel.GetResponseChannel())

	// Enable silent mode on logger - TUI handles display
	log.SetSilentMode(true)

	// Create orchestrator with observer
	orch := orchestrator.NewWithObserver(cfg, log, observer)

	// Run orchestrator in goroutine
	var result *types.SessionResult
	var orchErr error
	done := make(chan struct{})

	go func() {
		result, orchErr = orch.Run(ctx)
		close(done)
	}()

	// Run battle TUI
	battleProgram := tea.NewProgram(battleModel, tea.WithAltScreen())
	_, tuiErr := battleProgram.Run()
	if tuiErr != nil {
		cancel()
		return fmt.Errorf("TUI error: %w", tuiErr)
	}

	// Wait for orchestrator to finish
	<-done

	// Generate report if we have results
	if result != nil {
		rep := reporter.New(cfg.OutputDir)
		reportPath, reportErr := rep.GenerateReport(result, prompt)
		if reportErr == nil {
			infoColor.Printf("Report: %s\n", reportPath)
		}

		// Print summary
		if result.Success {
			successColor.Printf("\nSession completed successfully in %d round(s)\n", result.TotalRounds)
		} else {
			infoColor.Printf("\nSession ended after %d round(s)\n", result.TotalRounds)
		}

		infoColor.Printf("Log file: %s\n", log.GetLogFilePath())
	}

	if orchErr != nil {
		return orchErr
	}

	return nil
}

// runCLI runs the original CLI-based interface
func runCLI(cfg *config.Config) error {
	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return err
	}

	// Print banner and start
	printBanner()

	// Initialize logger
	log, err := logger.New(cfg.OutputDir, cfg.Verbose)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer log.Close()

	// Setup context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Info("Received interrupt signal, shutting down...")
		cancel()
	}()

	// Log session start
	log.Info(fmt.Sprintf("Initial prompt: %s", cfg.Prompt))
	log.Info(fmt.Sprintf("Working directory: %s", cfg.WorkDir))
	log.Info(fmt.Sprintf("Max iterations: %d", cfg.MaxIterations))
	fmt.Println()

	// Initialize and run orchestrator
	orch := orchestrator.New(cfg, log)
	result, err := orch.Run(ctx)

	if err != nil {
		return err
	}

	// Generate battle report
	rep := reporter.New(cfg.OutputDir)
	reportPath, reportErr := rep.GenerateReport(result, cfg.Prompt)
	if reportErr != nil {
		log.Error(fmt.Errorf("failed to generate report: %w", reportErr))
	}

	// Print summary
	if result.Success {
		successColor.Printf("\nSession completed successfully in %d round(s)\n", result.TotalRounds)
	} else {
		infoColor.Printf("\nSession ended after %d round(s)\n", result.TotalRounds)
	}

	infoColor.Printf("Log file: %s\n", log.GetLogFilePath())
	if reportErr == nil {
		infoColor.Printf("Report: %s\n", reportPath)
	}

	return nil
}

// validateWorkDir validates and resolves the working directory
func validateWorkDir(cfg *config.Config) error {
	absWorkDir, err := filepath.Abs(cfg.WorkDir)
	if err != nil {
		return fmt.Errorf("invalid working directory: %w", err)
	}

	info, err := os.Stat(absWorkDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("working directory does not exist: %s", absWorkDir)
		}
		return fmt.Errorf("cannot access working directory: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("working directory path is not a directory: %s", absWorkDir)
	}

	cfg.WorkDir = absWorkDir

	// Resolve output directory relative to working directory
	if !filepath.IsAbs(cfg.OutputDir) {
		cfg.OutputDir = filepath.Join(cfg.WorkDir, cfg.OutputDir)
	}

	return cfg.EnsureOutputDir()
}
