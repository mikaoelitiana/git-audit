package ui

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/you/git-audit/internal/git"
	"github.com/you/git-audit/internal/theme"
)

const barWidth = 14

// ── HELPERS ───────────────────────────────────────────────────────────────────

func cmdLine(t *theme.Theme, cmd string) string {
	parts := strings.Fields(cmd)
	if len(parts) == 0 { return cmd }
	var out strings.Builder
	out.WriteString(t.CmdPrompt.Render("$ "))
	for i, p := range parts {
		switch {
		case p == "git":
			out.WriteString(t.CmdKeyword.Render(p))
		case i == 1:
			out.WriteString(" " + t.Blue.Render(p))
		case strings.HasPrefix(p, "--") || strings.HasPrefix(p, "-"):
			out.WriteString(" " + t.CmdFlag.Render(p))
		case strings.HasPrefix(p, `"`) || strings.HasPrefix(p, `'`):
			out.WriteString(" " + t.CmdString.Render(p))
		case p == "|":
			out.WriteString(" " + t.CmdPipe.Render(p))
		case p == "sort" || p == "uniq" || p == "head" || p == "grep":
			out.WriteString(" " + t.CmdUtil.Render(p))
		default:
			out.WriteString(" " + t.Dim.Render(p))
		}
	}
	return out.String()
}

func divider(t *theme.Theme, width int) string {
	return t.Muted.Render(strings.Repeat("─", width))
}

func truncatePath(path string, max int) string {
	if len([]rune(path)) <= max { return path }
	parts := strings.Split(path, "/")
	file := parts[len(parts)-1]
	if len(file) >= max { return theme.Truncate(file, max) }
	for i := len(parts) - 2; i >= 0; i-- {
		candidate := "…/" + strings.Join(parts[i:], "/")
		if len([]rune(candidate)) <= max { return candidate }
	}
	return theme.Truncate(file, max)
}

func pad(s string, n int) string {
	r := []rune(s)
	if len(r) >= n { return string(r[:n]) }
	return s + strings.Repeat(" ", n-len(r))
}

func rpad(s string, n int) string {
	r := []rune(s)
	if len(r) >= n { return string(r[:n]) }
	return strings.Repeat(" ", n-len(r)) + s
}

// ── PANEL 1: CHURN ────────────────────────────────────────────────────────────

func renderChurn(t *theme.Theme, data []git.ChurnEntry, err error, loading bool, scroll, width, height int) string {
	var b strings.Builder
	b.WriteString(t.CmdBlock.Width(width - 4).Render(cmdLine(t, git.ChurnCmd)))
	b.WriteString("\n")
	if loading { b.WriteString(t.Blue.Render("⟳ running git command…")); return b.String() }
	if err != nil { b.WriteString(t.RedB.Render("✗ error: ") + t.Dim.Render(err.Error())); return b.String() }
	if len(data) == 0 { b.WriteString(t.Dim.Render("  no results — is this a git repository with history?")); return b.String() }

	b.WriteString(t.InsightWarn.Width(width-6).Render(
		t.AmberB.Render("▲ insight  ") +
			t.Amber.Render("High churn ≠ bad. But a file in both churn and bug lists is patch-on-patch territory — your highest risk."),
	))
	b.WriteString("\n\n")

	fileW := width - 46
	if fileW < 20 { fileW = 20 }
	b.WriteString("  " + t.TableHeader.Render(rpad("#", 3)) + "  " +
		t.TableHeader.Render(pad("FILE", fileW)) + "  " +
		t.TableHeader.Render(rpad("CHANGES", 8)) + "  " +
		t.TableHeader.Render(pad("CHURN", barWidth)) + "  " +
		t.TableHeader.Render("RISK") + "\n")
	b.WriteString("  " + divider(t, width-4) + "\n")

	shown := 0
	for i := scroll; i < len(data) && shown < height-8; i++ {
		e := data[i]
		riskStyle := t.RiskStyle(e.Pct)
		countStyle := t.Amber
		if e.Pct > 75 { countStyle = t.RedB } else if e.Pct < 30 { countStyle = t.Dim }

		b.WriteString("  " +
			t.Muted.Render(rpad(fmt.Sprintf("%d", i+1), 3)) + "  " +
			pad(truncatePath(e.File, fileW), fileW) + "  " +
			countStyle.Render(rpad(fmt.Sprintf("%d", e.Count), 8)) + "  " +
			t.Bar(barWidth, e.Pct, countStyle) + "  " +
			riskStyle.Render(theme.RiskLabel(e.Pct)) + "\n")
		shown++
	}
	if scroll+shown < len(data) {
		b.WriteString(t.Muted.Render(fmt.Sprintf("  … %d more (j/k to scroll)", len(data)-scroll-shown)))
	}
	return b.String()
}

// ── PANEL 2: BUS FACTOR ──────────────────────────────────────────────────────

func renderBusFactor(t *theme.Theme, data []git.Contributor, err error, loading bool, scroll, width, height int) string {
	var b strings.Builder
	b.WriteString(t.CmdBlock.Width(width - 4).Render(cmdLine(t, git.BusFactorCmd)))
	b.WriteString("\n")
	if loading { b.WriteString(t.Blue.Render("⟳ running git command…")); return b.String() }
	if err != nil { b.WriteString(t.RedB.Render("✗ error: ") + t.Dim.Render(err.Error())); return b.String() }
	if len(data) == 0 { b.WriteString(t.Dim.Render("  no contributors found")); return b.String() }

	total, activeCount := 0, 0
	for _, c := range data { total += c.Commits; if c.Active { activeCount++ } }
	top := data[0]

	var insightStyle lipgloss.Style
	var insightText string
	if top.Pct >= 60 {
		insightStyle = t.InsightCrit
		activeStr := "still active"
		if !top.Active { activeStr = t.RedB.Render("⚠ NO LONGER ACTIVE") }
		insightText = t.RedB.Render("⬡ bus factor: 1  ") +
			t.Red.Render(fmt.Sprintf("%s owns %.0f%% of commits — %s", top.Name, top.Pct, activeStr))
	} else {
		insightStyle = t.InsightOk
		insightText = t.GreenB.Render("✓  ") +
			t.Green.Render(fmt.Sprintf("Bus factor OK. %d active contributors in last 6 months.", activeCount))
	}
	b.WriteString(insightStyle.Width(width - 6).Render(insightText))
	b.WriteString("\n\n")

	nameW := width - 52
	if nameW < 20 { nameW = 20 }
	actW := barWidth + 2
	b.WriteString("  " + t.TableHeader.Render(rpad("#", 3)) + "  " +
		t.TableHeader.Render(pad("CONTRIBUTOR", nameW)) + "  " +
		t.TableHeader.Render(rpad("COMMITS", 8)) + "  " +
		t.TableHeader.Render(rpad("PCT", 6)) + "  " +
		t.TableHeader.Render(pad("SHARE", actW)) + "  " +
		t.TableHeader.Render("STATUS") + "\n")
	b.WriteString("  " + divider(t, width-4) + "\n")

	shown := 0
	for i := scroll; i < len(data) && shown < height-8; i++ {
		c := data[i]
		nameStyle := t.Base
		if i == 0 { nameStyle = t.BlueB } else if !c.Active { nameStyle = t.Muted }
		barStyle := t.Blue
		if i == 0 { barStyle = t.AmberB } else if !c.Active { barStyle = t.Muted }
		statusStr := t.Green.Render("● active")
		if !c.Active { statusStr = t.Red.Render("○ gone  ") }

		b.WriteString("  " +
			t.Muted.Render(rpad(fmt.Sprintf("%d", i+1), 3)) + "  " +
			nameStyle.Render(pad(theme.Truncate(c.Name, nameW), nameW)) + "  " +
			t.Dim.Render(rpad(fmt.Sprintf("%d", c.Commits), 8)) + "  " +
			t.Amber.Render(rpad(fmt.Sprintf("%.1f%%", c.Pct), 6)) + "  " +
			t.Bar(actW, c.Pct, barStyle) + "  " +
			statusStr + "\n")
		shown++
	}
	if scroll+shown < len(data) {
		b.WriteString(t.Muted.Render(fmt.Sprintf("  … %d more", len(data)-scroll-shown)))
	}
	return b.String()
}

// ── PANEL 3: BUG CLUSTERS ────────────────────────────────────────────────────

func renderBugs(t *theme.Theme, data []git.BugEntry, err error, loading bool, scroll, width, height int) string {
	var b strings.Builder
	b.WriteString(t.CmdBlock.Width(width - 4).Render(cmdLine(t, git.BugCmd)))
	b.WriteString("\n")
	if loading { b.WriteString(t.Blue.Render("⟳ running git command…")); return b.String() }
	if err != nil { b.WriteString(t.RedB.Render("✗ error: ") + t.Dim.Render(err.Error())); return b.String() }
	if len(data) == 0 { b.WriteString(t.Dim.Render("  no bug-related commits found")); return b.String() }

	overlap := 0
	for _, e := range data { if e.InChurn { overlap++ } }

	var insightStyle lipgloss.Style
	var insightText string
	if overlap > 0 {
		insightStyle = t.InsightCrit
		insightText = t.RedB.Render(fmt.Sprintf("✕ %d file(s) in BOTH churn and bug lists  ", overlap)) +
			t.Red.Render("Patch-on-patch — highest priority for refactor.")
	} else {
		insightStyle = t.InsightOk
		insightText = t.GreenB.Render("✓  ") + t.Green.Render("No overlap with churn hotspots.")
	}
	b.WriteString(insightStyle.Width(width - 6).Render(insightText))
	b.WriteString("\n\n")

	fileW := width - 46
	if fileW < 20 { fileW = 20 }
	b.WriteString("  " + t.TableHeader.Render(rpad("#", 3)) + "  " +
		t.TableHeader.Render(pad("FILE", fileW)) + "  " +
		t.TableHeader.Render(rpad("BUG COMMITS", 11)) + "  " +
		t.TableHeader.Render(pad("FREQ", barWidth)) + "  " +
		t.TableHeader.Render("IN CHURN?") + "\n")
	b.WriteString("  " + divider(t, width-4) + "\n")

	shown := 0
	for i := scroll; i < len(data) && shown < height-8; i++ {
		e := data[i]
		fileStyle, barStyle := t.Base, t.Blue
		if e.InChurn { fileStyle = t.RedB; barStyle = t.Red } else if e.Pct > 60 { fileStyle = t.Amber; barStyle = t.Amber }
		overlapStr := t.Muted.Render("  —      ")
		if e.InChurn { overlapStr = t.RedB.Render("  ✕ YES  ") }

		b.WriteString("  " +
			t.Muted.Render(rpad(fmt.Sprintf("%d", i+1), 3)) + "  " +
			fileStyle.Render(pad(truncatePath(e.File, fileW), fileW)) + "  " +
			t.Dim.Render(rpad(fmt.Sprintf("%d", e.Count), 11)) + "  " +
			t.Bar(barWidth, e.Pct, barStyle) + "  " +
			overlapStr + "\n")
		shown++
	}
	if scroll+shown < len(data) {
		b.WriteString(t.Muted.Render(fmt.Sprintf("  … %d more", len(data)-scroll-shown)))
	}
	return b.String()
}

// ── PANEL 4: VELOCITY ────────────────────────────────────────────────────────

func renderVelocity(t *theme.Theme, data []git.VelocityEntry, err error, loading bool, width, height int) string {
	var b strings.Builder
	b.WriteString(t.CmdBlock.Width(width - 4).Render(cmdLine(t, git.VelocityCmd)))
	b.WriteString("\n")
	if loading { b.WriteString(t.Blue.Render("⟳ running git command…")); return b.String() }
	if err != nil { b.WriteString(t.RedB.Render("✗ error: ") + t.Dim.Render(err.Error())); return b.String() }
	if len(data) == 0 { b.WriteString(t.Dim.Render("  no commit history found")); return b.String() }

	total, maxC := 0, 0
	for _, e := range data { total += e.Count; if e.Count > maxC { maxC = e.Count } }
	avg := float64(total) / float64(len(data))
	recent := data
	if len(data) > 6 { recent = data[len(data)-6:] }
	recentTotal := 0
	for _, e := range recent { recentTotal += e.Count }
	recentAvg := float64(recentTotal) / float64(len(recent))

	trendStr := t.Green.Render("→ steady")
	if recentAvg > avg*1.1 { trendStr = t.GreenB.Render("↑ accelerating") } else if recentAvg < avg*0.8 { trendStr = t.RedB.Render("↓ declining") }

	b.WriteString(t.InsightInfo.Width(width-6).Render(
		t.Blue.Render(fmt.Sprintf("avg: %.0f/mo   recent avg: %.0f/mo   trend: ", avg, recentAvg)) + trendStr,
	))
	b.WriteString("\n\n")

	chartH := height - 12
	if chartH < 4 { chartH = 4 }
	if chartH > 16 { chartH = 16 }

	colW := 6
	maxCols := (width - 8) / colW
	visible := data
	if len(visible) > maxCols { visible = visible[len(visible)-maxCols:] }

	for row := 0; row < chartH; row++ {
		rowVal := float64(maxC) * float64(chartH-row) / float64(chartH)
		b.WriteString(t.Muted.Render(fmt.Sprintf("  %4d │", int(rowVal))))
		for _, e := range visible {
			filled := int(math.Ceil(float64(e.Count) / float64(maxC) * float64(chartH)))
			cellVal := chartH - row
			if cellVal <= filled {
				pct := float64(e.Count) / float64(maxC)
				var colStyle lipgloss.Style
				if pct > 0.75 { colStyle = t.GreenB } else if pct > 0.4 { colStyle = t.Amber } else { colStyle = t.Red }
				b.WriteString(colStyle.Render(" ████ "))
			} else {
				b.WriteString("      ")
			}
		}
		b.WriteString("\n")
	}
	b.WriteString(t.Muted.Render("       └" + strings.Repeat("──────", len(visible))) + "\n")
	b.WriteString("        ")
	for _, e := range visible {
		lbl := strings.Replace(e.Month[2:], "-", ".", 1)
		b.WriteString(t.Muted.Render(fmt.Sprintf("%-6s", lbl)))
	}
	b.WriteString("\n        ")
	for _, e := range visible {
		b.WriteString(t.Dim.Render(fmt.Sprintf("%-6d", e.Count)))
	}
	b.WriteString("\n")
	return b.String()
}

// ── PANEL 5: FIREFIGHTING ─────────────────────────────────────────────────────

func renderFirefighting(t *theme.Theme, data []git.HotfixEntry, err error, loading bool, scroll, width, height int) string {
	var b strings.Builder
	b.WriteString(t.CmdBlock.Width(width - 4).Render(cmdLine(t, git.FirefightCmd)))
	b.WriteString("\n")
	if loading { b.WriteString(t.Blue.Render("⟳ running git command…")); return b.String() }
	if err != nil && len(data) == 0 { b.WriteString(t.RedB.Render("✗ error: ") + t.Dim.Render(err.Error())); return b.String() }

	n := len(data)
	var insightStyle lipgloss.Style
	var insightText string
	switch {
	case n == 0:
		insightStyle = t.InsightOk
		insightText = t.GreenB.Render("✓  ") + t.Green.Render("Zero crisis commits — stable, or no descriptive commit messages.")
	case n <= 6:
		insightStyle = t.InsightOk
		insightText = t.GreenB.Render("✓  ") + t.Green.Render(fmt.Sprintf("%d crisis event(s) in the last year — within normal range.", n))
	default:
		insightStyle = t.InsightCrit
		insightText = t.RedB.Render(fmt.Sprintf("⚠  %d crisis events  ", n)) + t.Red.Render("Team may be in sustained firefighting mode — check test coverage and staging.")
	}
	b.WriteString(insightStyle.Width(width - 6).Render(insightText))
	b.WriteString("\n\n")

	if len(data) == 0 { b.WriteString(t.Dim.Render("  (no results)")); return b.String() }

	msgW := width - 30
	if msgW < 20 { msgW = 20 }
	b.WriteString("  " + t.TableHeader.Render(pad("HASH", 10)) + "  " +
		t.TableHeader.Render(pad("TYPE", 12)) + "  " +
		t.TableHeader.Render("MESSAGE") + "\n")
	b.WriteString("  " + divider(t, width-4) + "\n")

	kindColor := map[string]lipgloss.Style{
		"revert": t.RedB, "hotfix": t.AmberB,
		"rollback": t.Purple, "emergency": t.RedB,
	}
	shown := 0
	for i := scroll; i < len(data) && shown < height-8; i++ {
		e := data[i]
		kStyle := kindColor[e.Kind]
		if (kStyle == lipgloss.Style{}) { kStyle = t.Dim }
		b.WriteString("  " +
			t.Blue.Render(pad(e.Hash, 10)) + "  " +
			kStyle.Render(pad("["+e.Kind+"]", 12)) + "  " +
			t.Base.Render(theme.Truncate(e.Message, msgW)) + "\n")
		shown++
	}
	if scroll+shown < len(data) {
		b.WriteString(t.Muted.Render(fmt.Sprintf("  … %d more", len(data)-scroll-shown)))
	}
	return b.String()
}
