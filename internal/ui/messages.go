package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/you/git-audit/internal/git"
)

// ── MESSAGES ─────────────────────────────────────────────────────────────────

type MsgChurnLoaded struct {
	Data []git.ChurnEntry
	Err  error
}

type MsgBusFactorLoaded struct {
	Data []git.Contributor
	Err  error
}

type MsgBugLoaded struct {
	Data []git.BugEntry
	Err  error
}

type MsgVelocityLoaded struct {
	Data []git.VelocityEntry
	Err  error
}

type MsgFirefightLoaded struct {
	Data []git.HotfixEntry
	Err  error
}

type MsgStatusClear struct{}

// ── COMMANDS ─────────────────────────────────────────────────────────────────

func loadChurn(cwd string) tea.Cmd {
	return func() tea.Msg {
		data, err := git.Churn(cwd)
		return MsgChurnLoaded{Data: data, Err: err}
	}
}

func loadBusFactor(cwd string) tea.Cmd {
	return func() tea.Msg {
		data, err := git.BusFactor(cwd)
		return MsgBusFactorLoaded{Data: data, Err: err}
	}
}

func loadBugs(cwd string, churnFiles map[string]bool) tea.Cmd {
	return func() tea.Msg {
		data, err := git.BugClusters(cwd, churnFiles)
		return MsgBugLoaded{Data: data, Err: err}
	}
}

func loadVelocity(cwd string) tea.Cmd {
	return func() tea.Msg {
		data, err := git.Velocity(cwd)
		return MsgVelocityLoaded{Data: data, Err: err}
	}
}

func loadFirefight(cwd string) tea.Cmd {
	return func() tea.Msg {
		data, err := git.Firefighting(cwd)
		return MsgFirefightLoaded{Data: data, Err: err}
	}
}
