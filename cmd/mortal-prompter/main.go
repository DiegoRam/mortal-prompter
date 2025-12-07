// Package main is the entry point for the mortal-prompter CLI application.
// Mortal Prompter orchestrates a code review battle between Claude Code and OpenAI Codex.
package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/minimalart/mortal-prompter/internal/config"
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

			// Validate configuration
			if err := cfg.Validate(); err != nil {
				return err
			}

			// Print banner and start
			printBanner()

			// TODO: Phase 2+ - Initialize and run orchestrator
			infoColor.Printf("Initial prompt: %s\n", cfg.Prompt)
			infoColor.Printf("Working directory: %s\n", cfg.WorkDir)
			infoColor.Printf("Max iterations: %d\n", cfg.MaxIterations)

			fmt.Println()
			color.New(color.FgHiYellow).Println("Orchestrator implementation coming in Phase 2...")

			return nil
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
║   ███╗   ███╗ ██████╗ ██████╗ ████████╗ █████╗ ██╗                        ║
║   ████╗ ████║██╔═══██╗██╔══██╗╚══██╔══╝██╔══██╗██║                        ║
║   ██╔████╔██║██║   ██║██████╔╝   ██║   ███████║██║                        ║
║   ██║╚██╔╝██║██║   ██║██╔══██╗   ██║   ██╔══██║██║                        ║
║   ██║ ╚═╝ ██║╚██████╔╝██║  ██║   ██║   ██║  ██║███████╗                   ║
║   ╚═╝     ╚═╝ ╚═════╝ ╚═╝  ╚═╝   ╚═╝   ╚═╝  ╚═╝╚══════╝                   ║
║                                                                          ║
║   ██████╗ ██████╗  ██████╗ ███╗   ███╗██████╗ ████████╗███████╗██████╗    ║
║   ██╔══██╗██╔══██╗██╔═══██╗████╗ ████║██╔══██╗╚══██╔══╝██╔════╝██╔══██╗   ║
║   ██████╔╝██████╔╝██║   ██║██╔████╔██║██████╔╝   ██║   █████╗  ██████╔╝   ║
║   ██╔═══╝ ██╔══██╗██║   ██║██║╚██╔╝██║██╔═══╝    ██║   ██╔══╝  ██╔══██╗   ║
║   ██║     ██║  ██║╚██████╔╝██║ ╚═╝ ██║██║        ██║   ███████╗██║  ██║   ║
║   ╚═╝     ╚═╝  ╚═╝ ╚═════╝ ╚═╝     ╚═╝╚═╝        ╚═╝   ╚══════╝╚═╝  ╚═╝   ║
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
