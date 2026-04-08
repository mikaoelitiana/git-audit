package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/you/git-audit/internal/git"
	"github.com/you/git-audit/internal/theme"
	"github.com/you/git-audit/internal/ui"
)

func main() {
	cwd := "."
	if len(os.Args) > 1 {
		cwd = os.Args[1]
	}

	abs, err := filepath.Abs(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if !git.IsGitRepo(abs) {
		fmt.Fprintf(os.Stderr, "error: not a git repository: %s\n", abs)
		fmt.Fprintln(os.Stderr, "usage: git-audit [path/to/repo]")
		os.Exit(1)
	}

	// Detect terminal background BEFORE entering alt-screen raw mode.
	// Once Bubble Tea takes over the terminal the OSC 11 response would
	// be swallowed. DetectVariant respects GIT_AUDIT_THEME env override.
	variant := theme.DetectVariant()

	p := tea.NewProgram(
		ui.New(abs, variant),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
