package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dgr8akki/nano-gameplays/internal/app"
)

const defaultModel = "claude-sonnet-4-6"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "identify" {
		if code := runIdentify(os.Args[2:]); code != 0 {
			os.Exit(code)
		}
		return
	}
	if code := runTUI(os.Args[1:]); code != 0 {
		os.Exit(code)
	}
}

func runTUI(args []string) int {
	fs := flag.NewFlagSet("nano-gameplays", flag.ExitOnError)
	startDir := fs.String("start-dir", "", "directory the file picker opens in (default: cwd)")
	modelID := fs.String("model", envOrDefault("NANO_GAMEPLAYS_MODEL", defaultModel), "Claude model ID")
	noTUI := fs.Bool("no-tui", false, "refuse to start the TUI; behave as identify")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *noTUI {
		fmt.Fprintln(os.Stderr, "--no-tui requires the identify subcommand with --video")
		return 2
	}
	dir := *startDir
	if dir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot resolve cwd: %v\n", err)
			return 2
		}
		dir = cwd
	}

	m := buildModel(dir, *modelID)
	prog := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := prog.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "tui error: %v\n", err)
		return 1
	}
	if mm, ok := finalModel.(app.Model); ok && mm.FinalPayload != nil {
		os.Stdout.Write(mm.FinalPayload.JSON)
		os.Stdout.WriteString("\n")
		return 0
	}
	return 1
}

func runIdentify(args []string) int {
	fs := flag.NewFlagSet("identify", flag.ExitOnError)
	video := fs.String("video", "", "path to a gameplay video (required)")
	modelID := fs.String("model", envOrDefault("NANO_GAMEPLAYS_MODEL", defaultModel), "Claude model ID")
	asJSON := fs.Bool("json", false, "emit HandoffPayload JSON to stdout")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *video == "" {
		fmt.Fprintln(os.Stderr, "identify: --video PATH is required")
		return 2
	}
	return runIdentifyOnce(*video, *modelID, *asJSON)
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// runIdentifyOnce and buildModel are wired in identify.go and wire.go after
// US1/US2 components are implemented.
