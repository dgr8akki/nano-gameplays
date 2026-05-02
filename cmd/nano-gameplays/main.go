package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dgr8akki/nano-gameplays/internal/app"
)

const claudeInstallURL = "https://claude.com/claude-code"

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

// preflightClaudeBinary verifies the `claude` CLI is available on PATH.
// Returns 0 on success or 2 with a single-line stderr message naming the
// missing tool and pointing at the recovery action.
func preflightClaudeBinary() int {
	if _, err := exec.LookPath("claude"); err != nil {
		fmt.Fprintf(os.Stderr,
			"claude CLI not found in PATH; install Claude Code from %s\n",
			claudeInstallURL)
		return 2
	}
	return 0
}

func runTUI(args []string) int {
	fs := flag.NewFlagSet("nano-gameplays", flag.ExitOnError)
	startDir := fs.String("start-dir", "", "directory the file picker opens in (default: cwd)")
	modelID := fs.String("model", os.Getenv("NANO_GAMEPLAYS_MODEL"), "Claude model ID; empty = defer to claude CLI default")
	noTUI := fs.Bool("no-tui", false, "refuse to start the TUI; behave as identify")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *noTUI {
		fmt.Fprintln(os.Stderr, "--no-tui requires the identify subcommand with --video")
		return 2
	}
	if code := preflightClaudeBinary(); code != 0 {
		return code
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
	modelID := fs.String("model", os.Getenv("NANO_GAMEPLAYS_MODEL"), "Claude model ID; empty = defer to claude CLI default")
	asJSON := fs.Bool("json", false, "emit HandoffPayload JSON to stdout")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *video == "" {
		fmt.Fprintln(os.Stderr, "identify: --video PATH is required")
		return 2
	}
	if code := preflightClaudeBinary(); code != 0 {
		return code
	}
	return runIdentifyOnce(*video, *modelID, *asJSON)
}
