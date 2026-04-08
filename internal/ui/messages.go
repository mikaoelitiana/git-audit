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

type MsgStaleLoaded struct {
	Data []git.StaleEntry
	Err  error
}

type MsgBranchesLoaded struct {
	Data []git.BranchEntry
	Err  error
}

type MsgCouplingLoaded struct {
	Data []git.CouplingEntry
	Err  error
}

type MsgFreshLoaded struct {
	Data []git.FreshEntry
	Err  error
}

type MsgOwnershipLoaded struct {
	Data []git.OwnershipEntry
	Err  error
}

type MsgTestRatioLoaded struct {
	Data []git.TestRatioEntry
	Err  error
}

type MsgCommitSizesLoaded struct {
	Data []git.CommitSizeBucket
	Err  error
}

type MsgMergeFreqLoaded struct {
	Data []git.MergeFreqEntry
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

func loadStale(cwd string) tea.Cmd {
	return func() tea.Msg {
		data, err := git.StaleFiles(cwd)
		return MsgStaleLoaded{Data: data, Err: err}
	}
}

func loadBranches(cwd string) tea.Cmd {
	return func() tea.Msg {
		data, err := git.LongBranches(cwd)
		return MsgBranchesLoaded{Data: data, Err: err}
	}
}

func loadCoupling(cwd string) tea.Cmd {
	return func() tea.Msg {
		data, err := git.CoChange(cwd)
		return MsgCouplingLoaded{Data: data, Err: err}
	}
}

func loadFresh(cwd string) tea.Cmd {
	return func() tea.Msg {
		data, err := git.FreshFiles(cwd)
		return MsgFreshLoaded{Data: data, Err: err}
	}
}

func loadOwnership(cwd string) tea.Cmd {
	return func() tea.Msg {
		data, err := git.OwnershipDrift(cwd)
		return MsgOwnershipLoaded{Data: data, Err: err}
	}
}

func loadTestRatio(cwd string) tea.Cmd {
	return func() tea.Msg {
		data, err := git.TestRatio(cwd)
		return MsgTestRatioLoaded{Data: data, Err: err}
	}
}

func loadCommitSizes(cwd string) tea.Cmd {
	return func() tea.Msg {
		data, err := git.CommitSizes(cwd)
		return MsgCommitSizesLoaded{Data: data, Err: err}
	}
}

func loadMergeFreq(cwd string) tea.Cmd {
	return func() tea.Msg {
		data, err := git.MergeFrequency(cwd)
		return MsgMergeFreqLoaded{Data: data, Err: err}
	}
}
