package ui

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/you/git-audit/internal/git"
	"github.com/you/git-audit/internal/theme"
)

// ── PANEL IDS ─────────────────────────────────────────────────────────────────

const (
	PanelChurn = iota
	PanelBusFactor
	PanelBugs
	PanelVelocity
	PanelFirefight
	PanelStale
	PanelBranches
	PanelCoupling
	PanelFresh
	PanelOwnership
	PanelTestRatio
	PanelCommitSize
	PanelMergeFreq
	numPanels
)

var panelTitles = [numPanels]string{
	"Churn Hotspots", "Bus Factor", "Bug Clusters", "Velocity", "Firefighting",
	"Stale Files", "Long Branches", "Co-change", "Fresh Files",
	"Ownership Drift", "Test Ratio", "Commit Sizes", "Merge Freq",
}
var panelIcons = [numPanels]string{"⬆", "◉", "⬡", "~", "!", "⧖", "⌥", "⊕", "✦", "⤵", "⌗", "∑", "⊗"}

// ── MODEL ─────────────────────────────────────────────────────────────────────

type Model struct {
	cwd    string
	width  int
	height int

	activePanel int
	scroll      [numPanels]int

	churnData    []git.ChurnEntry
	churnErr     error
	churnLoading bool
	churnFiles   map[string]bool

	busData    []git.Contributor
	busErr     error
	busLoading bool

	bugData    []git.BugEntry
	bugErr     error
	bugLoading bool

	velData    []git.VelocityEntry
	velErr     error
	velLoading bool

	fireData    []git.HotfixEntry
	fireErr     error
	fireLoading bool

	staleData    []git.StaleEntry
	staleErr     error
	staleLoading bool

	branchData    []git.BranchEntry
	branchErr     error
	branchLoading bool

	couplingData    []git.CouplingEntry
	couplingErr     error
	couplingLoading bool

	freshData    []git.FreshEntry
	freshErr     error
	freshLoading bool

	ownerData    []git.OwnershipEntry
	ownerErr     error
	ownerLoading bool

	testData    []git.TestRatioEntry
	testErr     error
	testLoading bool

	sizeData    []git.CommitSizeBucket
	sizeErr     error
	sizeLoading bool

	mergeData    []git.MergeFreqEntry
	mergeErr     error
	mergeLoading bool

	branch       string
	totalCommits int

	statusMsg   string
	statusTimer time.Time
	spinner     spinner.Model

	// Theme — pointer so Toggle mutates in place
	theme *theme.Theme
}

func New(cwd string, v theme.Variant) Model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot

	t := theme.New(v)
	sp.Style = t.Blue

	m := Model{
		cwd:             cwd,
		churnLoading:    true,
		busLoading:      true,
		bugLoading:      true,
		velLoading:      true,
		fireLoading:     true,
		staleLoading:    true,
		branchLoading:   true,
		couplingLoading: true,
		freshLoading:    true,
		ownerLoading:    true,
		testLoading:     true,
		sizeLoading:     true,
		mergeLoading:    true,
		churnFiles:      make(map[string]bool),
		branch:       git.CurrentBranch(cwd),
		totalCommits: git.TotalCommits(cwd),
		spinner:      sp,
		theme:        t,
	}
	return m
}

// ── INIT ─────────────────────────────────────────────────────────────────────

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		loadChurn(m.cwd),
		loadBusFactor(m.cwd),
		loadVelocity(m.cwd),
		loadFirefight(m.cwd),
		loadStale(m.cwd),
		loadBranches(m.cwd),
		loadCoupling(m.cwd),
		loadFresh(m.cwd),
		loadOwnership(m.cwd),
		loadTestRatio(m.cwd),
		loadCommitSizes(m.cwd),
		loadMergeFreq(m.cwd),
	)
}

// ── UPDATE ────────────────────────────────────────────────────────────────────

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case MsgChurnLoaded:
		m.churnLoading = false
		m.churnData, m.churnErr = msg.Data, msg.Err
		m.churnFiles = make(map[string]bool)
		for _, e := range m.churnData { m.churnFiles[e.File] = true }
		return m, loadBugs(m.cwd, m.churnFiles)

	case MsgBusFactorLoaded:
		m.busLoading = false
		m.busData, m.busErr = msg.Data, msg.Err
		return m, nil

	case MsgBugLoaded:
		m.bugLoading = false
		m.bugData, m.bugErr = msg.Data, msg.Err
		return m, nil

	case MsgVelocityLoaded:
		m.velLoading = false
		m.velData, m.velErr = msg.Data, msg.Err
		return m, nil

	case MsgFirefightLoaded:
		m.fireLoading = false
		m.fireData, m.fireErr = msg.Data, msg.Err
		return m, nil

	case MsgStaleLoaded:
		m.staleLoading = false
		m.staleData, m.staleErr = msg.Data, msg.Err
		return m, nil

	case MsgBranchesLoaded:
		m.branchLoading = false
		m.branchData, m.branchErr = msg.Data, msg.Err
		return m, nil

	case MsgCouplingLoaded:
		m.couplingLoading = false
		m.couplingData, m.couplingErr = msg.Data, msg.Err
		return m, nil

	case MsgFreshLoaded:
		m.freshLoading = false
		m.freshData, m.freshErr = msg.Data, msg.Err
		return m, nil

	case MsgOwnershipLoaded:
		m.ownerLoading = false
		m.ownerData, m.ownerErr = msg.Data, msg.Err
		return m, nil

	case MsgTestRatioLoaded:
		m.testLoading = false
		m.testData, m.testErr = msg.Data, msg.Err
		return m, nil

	case MsgCommitSizesLoaded:
		m.sizeLoading = false
		m.sizeData, m.sizeErr = msg.Data, msg.Err
		return m, nil

	case MsgMergeFreqLoaded:
		m.mergeLoading = false
		m.mergeData, m.mergeErr = msg.Data, msg.Err
		return m, nil

	case MsgStatusClear:
		m.statusMsg = ""
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		m.activePanel = int(msg.String()[0] - '1')
	case "0":
		if numPanels > 9 {
			m.activePanel = 9
		}

	case "tab", "l", "right":
		m.activePanel = (m.activePanel + 1) % numPanels

	case "shift+tab", "h", "left":
		m.activePanel = (m.activePanel - 1 + numPanels) % numPanels

	case "j", "down":
		m.scroll[m.activePanel]++

	case "k", "up":
		if m.scroll[m.activePanel] > 0 { m.scroll[m.activePanel]-- }

	case "g":
		m.scroll[m.activePanel] = 0

	case "G":
		m.scroll[m.activePanel] = 9999

	case "T", "t":
		m.theme.Toggle()
		m.spinner.Style = m.theme.Blue
		label := "dark"
		if m.theme.Variant == theme.Light { label = "light" }
		m.setStatus(fmt.Sprintf("theme: %s", label))

	case "r":
		m.scroll[m.activePanel] = 0
		m.setStatus("re-running command…")
		switch m.activePanel {
		case PanelChurn:       m.churnLoading = true;    return m, loadChurn(m.cwd)
		case PanelBusFactor:   m.busLoading = true;      return m, loadBusFactor(m.cwd)
		case PanelBugs:        m.bugLoading = true;      return m, loadBugs(m.cwd, m.churnFiles)
		case PanelVelocity:    m.velLoading = true;      return m, loadVelocity(m.cwd)
		case PanelFirefight:   m.fireLoading = true;     return m, loadFirefight(m.cwd)
		case PanelStale:       m.staleLoading = true;    return m, loadStale(m.cwd)
		case PanelBranches:    m.branchLoading = true;   return m, loadBranches(m.cwd)
		case PanelCoupling:    m.couplingLoading = true; return m, loadCoupling(m.cwd)
		case PanelFresh:       m.freshLoading = true;    return m, loadFresh(m.cwd)
		case PanelOwnership:   m.ownerLoading = true;    return m, loadOwnership(m.cwd)
		case PanelTestRatio:   m.testLoading = true;     return m, loadTestRatio(m.cwd)
		case PanelCommitSize:  m.sizeLoading = true;     return m, loadCommitSizes(m.cwd)
		case PanelMergeFreq:   m.mergeLoading = true;    return m, loadMergeFreq(m.cwd)
		}

	case "y":
		_, cmd := panelCmd(m.activePanel)
		if err := copyToClipboard(cmd); err == nil {
			m.setStatus("copied to clipboard ✓")
		} else {
			m.setStatus("$ " + cmd)
		}
	}
	return m, statusClearCmd(3 * time.Second)
}

func (m *Model) setStatus(s string) {
	m.statusMsg = s
	m.statusTimer = time.Now()
}

func statusClearCmd(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(_ time.Time) tea.Msg { return MsgStatusClear{} })
}

func panelCmd(panel int) (string, string) {
	cmds := [numPanels][2]string{
		{"Churn Hotspots", git.ChurnCmd},
		{"Bus Factor", git.BusFactorCmd},
		{"Bug Clusters", git.BugCmd},
		{"Velocity", git.VelocityCmd},
		{"Firefighting", git.FirefightCmd},
		{"Stale Files", git.StaleLogCmd},
		{"Long Branches", git.BranchListCmd},
		{"Co-change Coupling", git.CoChangeCmd},
		{"Fresh Files", git.FreshCmd},
		{"Ownership Drift", git.OwnerNewCmd},
		{"Test Ratio", git.TestRatioCmd},
		{"Commit Sizes", git.CommitSizeCmd},
		{"Merge Frequency", git.MergeFreqCmd},
	}
	return cmds[panel][0], cmds[panel][1]
}

func copyToClipboard(s string) error {
	for _, tool := range []string{"xclip -selection clipboard", "xsel --clipboard --input", "pbcopy", "wl-copy"} {
		parts := strings.Fields(tool)
		c := exec.Command(parts[0], parts[1:]...)
		c.Stdin = strings.NewReader(s)
		if err := c.Run(); err == nil { return nil }
	}
	return fmt.Errorf("no clipboard tool found")
}

// ── VIEW ──────────────────────────────────────────────────────────────────────

func (m Model) View() string {
	if m.width == 0 { return "loading…" }
	if m.width < 60 || m.height < 18 { return "Terminal too small — resize to at least 60×18\n" }

	var b strings.Builder
	b.WriteString(m.titleBar())
	b.WriteString("\n")
	b.WriteString(m.body())
	b.WriteString(m.statusBar())
	return b.String()
}

// ── TITLE BAR ────────────────────────────────────────────────────────────────

func (m Model) titleBar() string {
	t := m.theme
	appTitle := t.AppTitle.Render(" git-audit ")

	// Theme indicator badge
	themeIcon := "🌙"
	if m.theme.Variant == theme.Light { themeIcon = "☀" }
	themeBadge := t.Dim.Render(fmt.Sprintf(" %s ", themeIcon))

	var tabs strings.Builder
	for i := 0; i < numPanels; i++ {
		key := fmt.Sprintf("%d", i+1)
		if i == 9 {
			key = "0"
		} else if i > 9 {
			key = fmt.Sprintf("+%d", i-9)
		}
		label := fmt.Sprintf(" %s:%s ", key, panelIcons[i])
		if i == m.activePanel {
			tabs.WriteString(t.TabActive.Render(label))
		} else {
			tabs.WriteString(t.TabInactive.Render(label))
		}
		if i < numPanels-1 { tabs.WriteString(t.Muted.Render("│")) }
	}

	right := t.TitleBar.Render(
		t.Muted.Render(git.RepoName(m.cwd)) + "  " +
			t.Dim.Render(m.branch) + "  " +
			t.Muted.Render(time.Now().Format("15:04:05")),
	)

	tabStr := tabs.String()
	appW   := lipgloss.Width(appTitle)
	tabW   := lipgloss.Width(tabStr)
	rightW := lipgloss.Width(right)
	badgeW := lipgloss.Width(themeBadge)
	fill   := m.width - appW - tabW - rightW - badgeW - 1
	if fill < 0 { fill = 0 }
	gap := lipgloss.NewStyle().Background(lipgloss.Color(t.P.Bg3)).Render(strings.Repeat(" ", fill))

	return appTitle + tabStr + gap + themeBadge + right
}

// ── BODY ─────────────────────────────────────────────────────────────────────

func (m Model) body() string {
	t := m.theme
	sidebarW := 20
	contentW := m.width - sidebarW - 1
	bodyH    := m.height - 2

	sidebar  := m.sidebar(sidebarW, bodyH)
	content  := m.panelContent(contentW, bodyH)
	divCol   := t.Muted.Render(strings.Repeat("│\n", bodyH))

	return lipgloss.JoinHorizontal(lipgloss.Top, sidebar, divCol, content)
}

func (m Model) sidebar(w, h int) string {
	t := m.theme
	var b strings.Builder

	loadingOf := func(p int) bool {
		return [numPanels]bool{
			m.churnLoading, m.busLoading, m.bugLoading, m.velLoading, m.fireLoading,
			m.staleLoading, m.branchLoading, m.couplingLoading, m.freshLoading,
			m.ownerLoading, m.testLoading, m.sizeLoading, m.mergeLoading,
		}[p]
	}
	errOf := func(p int) bool {
		return ([numPanels]error{
			m.churnErr, m.busErr, m.bugErr, m.velErr, m.fireErr,
			m.staleErr, m.branchErr, m.couplingErr, m.freshErr,
			m.ownerErr, m.testErr, m.sizeErr, m.mergeErr,
		})[p] != nil
	}

	b.WriteString("\n")
	for i := 0; i < numPanels; i++ {
		var dot string
		switch {
		case loadingOf(i): dot = t.Amber.Render("⟳")
		case errOf(i):     dot = t.Red.Render("✗")
		default:           dot = t.Green.Render("●")
		}

		num   := t.BlueB.Render(fmt.Sprintf("%d", i+1))
		label := fmt.Sprintf(" %s %s %s  %s", num, panelIcons[i], theme.Truncate(panelTitles[i], w-8), dot)

		var style lipgloss.Style
		if i == m.activePanel {
			style = lipgloss.NewStyle().
				Background(lipgloss.Color(t.P.Bg3)).
				Foreground(lipgloss.Color(t.P.Amber)).
				Bold(true).Width(w)
		} else {
			style = lipgloss.NewStyle().
				Foreground(lipgloss.Color(t.P.FgDim)).Width(w)
		}
		b.WriteString(style.Render(label) + "\n\n")
	}

	remaining := h - numPanels*2 - 2
	if remaining > 7 {
		b.WriteString(strings.Repeat("\n", remaining-7))
		hints := [][2]string{{"j/k", "scroll"}, {"Tab", "panel"}, {"r", "reload"}, {"y", "copy cmd"}, {"T", "toggle theme"}, {"q", "quit"}}
		for _, kv := range hints {
			b.WriteString(t.BlueB.Render(fmt.Sprintf(" %-4s", kv[0])) + t.Muted.Render(" "+kv[1]) + "\n")
		}
	}
	return b.String()
}

func (m Model) panelContent(w, h int) string {
	t := m.theme
	scroll := m.scroll[m.activePanel]
	var content string

	switch m.activePanel {
	case PanelChurn:      content = renderChurn(t, m.churnData, m.churnErr, m.churnLoading, scroll, w, h)
	case PanelBusFactor:  content = renderBusFactor(t, m.busData, m.busErr, m.busLoading, scroll, w, h)
	case PanelBugs:       content = renderBugs(t, m.bugData, m.bugErr, m.bugLoading, scroll, w, h)
	case PanelVelocity:   content = renderVelocity(t, m.velData, m.velErr, m.velLoading, w, h)
	case PanelFirefight:  content = renderFirefighting(t, m.fireData, m.fireErr, m.fireLoading, scroll, w, h)
	case PanelStale:      content = renderStale(t, m.staleData, m.staleErr, m.staleLoading, scroll, w, h)
	case PanelBranches:   content = renderBranches(t, m.branchData, m.branchErr, m.branchLoading, scroll, w, h)
	case PanelCoupling:   content = renderCoupling(t, m.couplingData, m.couplingErr, m.couplingLoading, scroll, w, h)
	case PanelFresh:      content = renderFresh(t, m.freshData, m.freshErr, m.freshLoading, scroll, w, h)
	case PanelOwnership:  content = renderOwnership(t, m.ownerData, m.ownerErr, m.ownerLoading, scroll, w, h)
	case PanelTestRatio:  content = renderTestRatio(t, m.testData, m.testErr, m.testLoading, scroll, w, h)
	case PanelCommitSize: content = renderCommitSizes(t, m.sizeData, m.sizeErr, m.sizeLoading, w, h)
	case PanelMergeFreq:  content = renderMergeFreq(t, m.mergeData, m.mergeErr, m.mergeLoading, w, h)
	}

	title := t.BlueB.Render("  ── ") + t.AmberB.Render(panelTitles[m.activePanel]) + t.Muted.Render(" ──")
	return title + "\n" + content
}

// ── STATUS BAR ───────────────────────────────────────────────────────────────

func (m Model) statusBar() string {
	t := m.theme
	mode := t.StatusMode.Render(" NORMAL ")

	var msg string
	if m.statusMsg != "" {
		msg = t.Blue.Render("  " + m.statusMsg)
	} else {
		msg = t.Dim.Render("  " + m.cwd)
	}

	right := t.StatusKey.Render(fmt.Sprintf(" git-audit v1.0 [%s] ", m.theme.Variant))
	rightW := lipgloss.Width(right)
	modeW  := lipgloss.Width(mode)
	msgW   := m.width - modeW - rightW - 1
	if msgW < 0 { msgW = 0 }

	msgStyled := lipgloss.NewStyle().
		Background(lipgloss.Color(t.P.Bg3)).
		Width(msgW).
		Render(msg)

	return mode + msgStyled + right
}
