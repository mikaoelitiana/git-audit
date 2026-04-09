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
	if len(parts) == 0 {
		return cmd
	}
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
	if len([]rune(path)) <= max {
		return path
	}
	parts := strings.Split(path, "/")
	file := parts[len(parts)-1]
	if len(file) >= max {
		return theme.Truncate(file, max)
	}
	for i := len(parts) - 2; i >= 0; i-- {
		candidate := "…/" + strings.Join(parts[i:], "/")
		if len([]rune(candidate)) <= max {
			return candidate
		}
	}
	return theme.Truncate(file, max)
}

func pad(s string, n int) string {
	r := []rune(s)
	if len(r) >= n {
		return string(r[:n])
	}
	return s + strings.Repeat(" ", n-len(r))
}

func rpad(s string, n int) string {
	r := []rune(s)
	if len(r) >= n {
		return string(r[:n])
	}
	return strings.Repeat(" ", n-len(r)) + s
}

// ── PANEL 1: CHURN ────────────────────────────────────────────────────────────

func renderChurn(t *theme.Theme, data []git.ChurnEntry, err error, loading bool, scroll, width, height int) string {
	var b strings.Builder
	b.WriteString(t.CmdBlock.Width(width - 4).Render(cmdLine(t, git.ChurnCmd)))
	b.WriteString("\n")
	if loading {
		b.WriteString(t.Blue.Render("⟳ running git command…"))
		return b.String()
	}
	if err != nil {
		b.WriteString(t.RedB.Render("✗ error: ") + t.Dim.Render(err.Error()))
		return b.String()
	}
	if len(data) == 0 {
		b.WriteString(t.Dim.Render("  no results — is this a git repository with history?"))
		return b.String()
	}

	b.WriteString(t.InsightWarn.Width(width - 6).Render(
		t.AmberB.Render("▲ insight  ") +
			t.Amber.Render("High churn ≠ bad. But a file in both churn and bug lists is patch-on-patch territory — your highest risk."),
	))
	b.WriteString("\n\n")

	fileW := width - 46
	if fileW < 20 {
		fileW = 20
	}
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
		if e.Pct > 75 {
			countStyle = t.RedB
		} else if e.Pct < 30 {
			countStyle = t.Dim
		}

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
	if loading {
		b.WriteString(t.Blue.Render("⟳ running git command…"))
		return b.String()
	}
	if err != nil {
		b.WriteString(t.RedB.Render("✗ error: ") + t.Dim.Render(err.Error()))
		return b.String()
	}
	if len(data) == 0 {
		b.WriteString(t.Dim.Render("  no contributors found"))
		return b.String()
	}

	total, activeCount := 0, 0
	for _, c := range data {
		total += c.Commits
		if c.Active {
			activeCount++
		}
	}
	top := data[0]

	var insightStyle lipgloss.Style
	var insightText string
	if top.Pct >= 60 {
		insightStyle = t.InsightCrit
		activeStr := "still active"
		if !top.Active {
			activeStr = t.RedB.Render("⚠ NO LONGER ACTIVE")
		}
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
	if nameW < 20 {
		nameW = 20
	}
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
		if i == 0 {
			nameStyle = t.BlueB
		} else if !c.Active {
			nameStyle = t.Muted
		}
		barStyle := t.Blue
		if i == 0 {
			barStyle = t.AmberB
		} else if !c.Active {
			barStyle = t.Muted
		}
		statusStr := t.Green.Render("● active")
		if !c.Active {
			statusStr = t.Red.Render("○ gone  ")
		}

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
	if loading {
		b.WriteString(t.Blue.Render("⟳ running git command…"))
		return b.String()
	}
	if err != nil {
		b.WriteString(t.RedB.Render("✗ error: ") + t.Dim.Render(err.Error()))
		return b.String()
	}
	if len(data) == 0 {
		b.WriteString(t.Dim.Render("  no bug-related commits found"))
		return b.String()
	}

	overlap := 0
	for _, e := range data {
		if e.InChurn {
			overlap++
		}
	}

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
	if fileW < 20 {
		fileW = 20
	}
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
		if e.InChurn {
			fileStyle = t.RedB
			barStyle = t.Red
		} else if e.Pct > 60 {
			fileStyle = t.Amber
			barStyle = t.Amber
		}
		overlapStr := t.Muted.Render("  —      ")
		if e.InChurn {
			overlapStr = t.RedB.Render("  ✕ YES  ")
		}

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
	if loading {
		b.WriteString(t.Blue.Render("⟳ running git command…"))
		return b.String()
	}
	if err != nil {
		b.WriteString(t.RedB.Render("✗ error: ") + t.Dim.Render(err.Error()))
		return b.String()
	}
	if len(data) == 0 {
		b.WriteString(t.Dim.Render("  no commit history found"))
		return b.String()
	}

	total, maxC := 0, 0
	for _, e := range data {
		total += e.Count
		if e.Count > maxC {
			maxC = e.Count
		}
	}
	avg := float64(total) / float64(len(data))
	recent := data
	if len(data) > 6 {
		recent = data[len(data)-6:]
	}
	recentTotal := 0
	for _, e := range recent {
		recentTotal += e.Count
	}
	recentAvg := float64(recentTotal) / float64(len(recent))

	trendStr := t.Green.Render("→ steady")
	if recentAvg > avg*1.1 {
		trendStr = t.GreenB.Render("↑ accelerating")
	} else if recentAvg < avg*0.8 {
		trendStr = t.RedB.Render("↓ declining")
	}

	b.WriteString(t.InsightInfo.Width(width - 6).Render(
		t.Blue.Render(fmt.Sprintf("avg: %.0f/mo   recent avg: %.0f/mo   trend: ", avg, recentAvg)) + trendStr,
	))
	b.WriteString("\n\n")

	chartH := height - 12
	if chartH < 4 {
		chartH = 4
	}
	if chartH > 16 {
		chartH = 16
	}

	colW := 6
	maxCols := (width - 8) / colW
	visible := data
	if len(visible) > maxCols {
		visible = visible[len(visible)-maxCols:]
	}

	for row := 0; row < chartH; row++ {
		rowVal := float64(maxC) * float64(chartH-row) / float64(chartH)
		b.WriteString(t.Muted.Render(fmt.Sprintf("  %4d │", int(rowVal))))
		for _, e := range visible {
			filled := int(math.Ceil(float64(e.Count) / float64(maxC) * float64(chartH)))
			cellVal := chartH - row
			if cellVal <= filled {
				pct := float64(e.Count) / float64(maxC)
				var colStyle lipgloss.Style
				if pct > 0.75 {
					colStyle = t.GreenB
				} else if pct > 0.4 {
					colStyle = t.Amber
				} else {
					colStyle = t.Red
				}
				b.WriteString(colStyle.Render(" ████ "))
			} else {
				b.WriteString("      ")
			}
		}
		b.WriteString("\n")
	}
	b.WriteString(t.Muted.Render("       └"+strings.Repeat("──────", len(visible))) + "\n")
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

// ── PANEL 6: STALE FILES ──────────────────────────────────────────────────────

func renderStale(t *theme.Theme, data []git.StaleEntry, err error, loading bool, scroll, width, height int) string {
	var b strings.Builder
	b.WriteString(t.CmdBlock.Width(width - 4).Render(cmdLine(t, git.StaleLogCmd)))
	b.WriteString("\n")
	if loading {
		b.WriteString(t.Blue.Render("⟳ running git command…"))
		return b.String()
	}
	if err != nil {
		b.WriteString(t.RedB.Render("✗ error: ") + t.Dim.Render(err.Error()))
		return b.String()
	}
	if len(data) == 0 {
		b.WriteString(t.InsightOk.Width(width - 6).Render(t.GreenB.Render("✓  ") + t.Green.Render("No stale files — everything touched in the last year.")))
		return b.String()
	}
	b.WriteString(t.InsightWarn.Width(width - 6).Render(
		t.AmberB.Render("▲ insight  ") +
			t.Amber.Render(fmt.Sprintf("%d file(s) untouched for 1+ year — candidates for deletion or archival.", len(data))),
	))
	b.WriteString("\n\n")

	fileW := width - 38
	if fileW < 20 {
		fileW = 20
	}
	b.WriteString("  " + t.TableHeader.Render(rpad("#", 3)) + "  " +
		t.TableHeader.Render(pad("FILE", fileW)) + "  " +
		t.TableHeader.Render(rpad("LAST CHANGED", 12)) + "  " +
		t.TableHeader.Render(rpad("DAYS AGO", 8)) + "\n")
	b.WriteString("  " + divider(t, width-4) + "\n")

	shown := 0
	for i := scroll; i < len(data) && shown < height-8; i++ {
		e := data[i]
		age := t.Amber
		if e.DaysAgo > 730 {
			age = t.RedB
		} else if e.DaysAgo < 400 {
			age = t.Dim
		}
		daysStr := fmt.Sprintf("%d", e.DaysAgo)
		if e.DaysAgo == 9999 {
			daysStr = "never"
		}
		b.WriteString("  " +
			t.Muted.Render(rpad(fmt.Sprintf("%d", i+1), 3)) + "  " +
			pad(truncatePath(e.File, fileW), fileW) + "  " +
			t.Dim.Render(rpad(e.LastChanged, 12)) + "  " +
			age.Render(rpad(daysStr, 8)) + "\n")
		shown++
	}
	if scroll+shown < len(data) {
		b.WriteString(t.Muted.Render(fmt.Sprintf("  … %d more", len(data)-scroll-shown)))
	}
	return b.String()
}

// ── PANEL 7: LONG-LIVED BRANCHES ─────────────────────────────────────────────

func renderBranches(t *theme.Theme, data []git.BranchEntry, err error, loading bool, scroll, width, height int) string {
	var b strings.Builder
	b.WriteString(t.CmdBlock.Width(width - 4).Render(cmdLine(t, git.BranchListCmd)))
	b.WriteString("\n")
	if loading {
		b.WriteString(t.Blue.Render("⟳ running git command…"))
		return b.String()
	}
	if err != nil {
		b.WriteString(t.RedB.Render("✗ error: ") + t.Dim.Render(err.Error()))
		return b.String()
	}
	if len(data) == 0 {
		b.WriteString(t.Dim.Render("  no local branches found"))
		return b.String()
	}

	old := 0
	for _, e := range data {
		if e.DaysAgo > 90 {
			old++
		}
	}
	var insightStyle lipgloss.Style
	var insightText string
	if old > 0 {
		insightStyle = t.InsightWarn
		insightText = t.AmberB.Render(fmt.Sprintf("▲ %d branch(es) older than 90 days  ", old)) +
			t.Amber.Render("Long-lived branches increase merge risk.")
	} else {
		insightStyle = t.InsightOk
		insightText = t.GreenB.Render("✓  ") + t.Green.Render("All branches are recent — good branch hygiene.")
	}
	b.WriteString(insightStyle.Width(width - 6).Render(insightText))
	b.WriteString("\n\n")

	nameW := width - 44
	if nameW < 20 {
		nameW = 20
	}
	b.WriteString("  " + t.TableHeader.Render(rpad("#", 3)) + "  " +
		t.TableHeader.Render(pad("BRANCH", nameW)) + "  " +
		t.TableHeader.Render(rpad("LAST COMMIT", 11)) + "  " +
		t.TableHeader.Render(rpad("AGE (days)", 10)) + "  " +
		t.TableHeader.Render("AUTHOR") + "\n")
	b.WriteString("  " + divider(t, width-4) + "\n")

	shown := 0
	for i := scroll; i < len(data) && shown < height-8; i++ {
		e := data[i]
		ageStyle := t.Green
		if e.DaysAgo > 180 {
			ageStyle = t.RedB
		} else if e.DaysAgo > 90 {
			ageStyle = t.Amber
		}
		b.WriteString("  " +
			t.Muted.Render(rpad(fmt.Sprintf("%d", i+1), 3)) + "  " +
			t.Base.Render(pad(theme.Truncate(e.Name, nameW), nameW)) + "  " +
			t.Dim.Render(rpad(e.Date, 11)) + "  " +
			ageStyle.Render(rpad(fmt.Sprintf("%d", e.DaysAgo), 10)) + "  " +
			t.Muted.Render(theme.Truncate(e.Author, 20)) + "\n")
		shown++
	}
	if scroll+shown < len(data) {
		b.WriteString(t.Muted.Render(fmt.Sprintf("  … %d more", len(data)-scroll-shown)))
	}
	return b.String()
}

// ── PANEL 8: CO-CHANGE COUPLING ───────────────────────────────────────────────

func renderCoupling(t *theme.Theme, data []git.CouplingEntry, err error, loading bool, scroll, width, height int) string {
	var b strings.Builder
	b.WriteString(t.CmdBlock.Width(width - 4).Render(cmdLine(t, git.CoChangeCmd)))
	b.WriteString("\n")
	if loading {
		b.WriteString(t.Blue.Render("⟳ running git command…"))
		return b.String()
	}
	if err != nil {
		b.WriteString(t.RedB.Render("✗ error: ") + t.Dim.Render(err.Error()))
		return b.String()
	}
	if len(data) == 0 {
		b.WriteString(t.InsightOk.Width(width - 6).Render(t.GreenB.Render("✓  ") + t.Green.Render("No strong coupling detected (no pair changed together ≥3 times).")))
		return b.String()
	}
	b.WriteString(t.InsightWarn.Width(width - 6).Render(
		t.AmberB.Render("▲ insight  ") +
			t.Amber.Render("These file pairs are always committed together — hidden coupling, consider extracting shared logic."),
	))
	b.WriteString("\n\n")

	halfW := (width - 22) / 2
	if halfW < 15 {
		halfW = 15
	}
	b.WriteString("  " + t.TableHeader.Render(rpad("#", 3)) + "  " +
		t.TableHeader.Render(pad("FILE A", halfW)) + "  " +
		t.TableHeader.Render(pad("FILE B", halfW)) + "  " +
		t.TableHeader.Render(rpad("TOGETHER", 8)) + "  " +
		t.TableHeader.Render(pad("COUPLING", barWidth)) + "\n")
	b.WriteString("  " + divider(t, width-4) + "\n")

	shown := 0
	for i := scroll; i < len(data) && shown < height-8; i++ {
		e := data[i]
		barStyle := t.Blue
		if e.Pct > 75 {
			barStyle = t.RedB
		} else if e.Pct > 40 {
			barStyle = t.Amber
		}
		b.WriteString("  " +
			t.Muted.Render(rpad(fmt.Sprintf("%d", i+1), 3)) + "  " +
			pad(truncatePath(e.FileA, halfW), halfW) + "  " +
			pad(truncatePath(e.FileB, halfW), halfW) + "  " +
			t.Amber.Render(rpad(fmt.Sprintf("%d", e.Together), 8)) + "  " +
			t.Bar(barWidth, e.Pct, barStyle) + "\n")
		shown++
	}
	if scroll+shown < len(data) {
		b.WriteString(t.Muted.Render(fmt.Sprintf("  … %d more", len(data)-scroll-shown)))
	}
	return b.String()
}

// ── PANEL 9: FRESH FILES ──────────────────────────────────────────────────────

func renderFresh(t *theme.Theme, data []git.FreshEntry, err error, loading bool, scroll, width, height int) string {
	var b strings.Builder
	b.WriteString(t.CmdBlock.Width(width - 4).Render(cmdLine(t, git.FreshCmd)))
	b.WriteString("\n")
	if loading {
		b.WriteString(t.Blue.Render("⟳ running git command…"))
		return b.String()
	}
	if err != nil {
		b.WriteString(t.RedB.Render("✗ error: ") + t.Dim.Render(err.Error()))
		return b.String()
	}

	if len(data) == 0 {
		b.WriteString(t.Dim.Render("  no new files in the last 90 days"))
		return b.String()
	}

	b.WriteString(t.InsightInfo.Width(width - 6).Render(
		t.Blue.Render(fmt.Sprintf("ℹ  %d new file(s) added in the last 90 days — onboarding surface area.", len(data))),
	))
	b.WriteString("\n\n")

	fileW := width - 42
	if fileW < 20 {
		fileW = 20
	}
	authorW := 18
	b.WriteString("  " + t.TableHeader.Render(rpad("#", 3)) + "  " +
		t.TableHeader.Render(pad("FILE", fileW)) + "  " +
		t.TableHeader.Render(rpad("DATE", 10)) + "  " +
		t.TableHeader.Render(rpad("DAYS AGO", 8)) + "  " +
		t.TableHeader.Render(pad("AUTHOR", authorW)) + "\n")
	b.WriteString("  " + divider(t, width-4) + "\n")

	shown := 0
	for i := scroll; i < len(data) && shown < height-11; i++ {
		e := data[i]
		ageStyle := t.GreenB
		if e.DaysAgo > 60 {
			ageStyle = t.Dim
		} else if e.DaysAgo > 30 {
			ageStyle = t.Green
		}
		b.WriteString("  " +
			t.Muted.Render(rpad(fmt.Sprintf("%d", i+1), 3)) + "  " +
			pad(truncatePath(e.File, fileW), fileW) + "  " +
			t.Dim.Render(rpad(e.Date, 10)) + "  " +
			ageStyle.Render(rpad(fmt.Sprintf("%d", e.DaysAgo), 8)) + "  " +
			t.Muted.Render(theme.Truncate(e.Author, authorW)) + "\n")
		shown++
	}
	if scroll+shown < len(data) {
		b.WriteString(t.Muted.Render(fmt.Sprintf("  … %d more", len(data)-scroll-shown)))
	}
	return b.String()
}

// ── PANEL 10: OWNERSHIP DRIFT ─────────────────────────────────────────────────

func renderOwnership(t *theme.Theme, data []git.OwnershipEntry, err error, loading bool, scroll, width, height int) string {
	var b strings.Builder
	b.WriteString(t.CmdBlock.Width(width - 4).Render(cmdLine(t, git.OwnerNewCmd)))
	b.WriteString("\n")
	if loading {
		b.WriteString(t.Blue.Render("⟳ running git command…"))
		return b.String()
	}
	if err != nil {
		b.WriteString(t.RedB.Render("✗ error: ") + t.Dim.Render(err.Error()))
		return b.String()
	}
	if len(data) == 0 {
		b.WriteString(t.InsightOk.Width(width - 6).Render(t.GreenB.Render("✓  ") + t.Green.Render("No ownership changes detected.")))
		return b.String()
	}

	drifted := 0
	for _, e := range data {
		if e.Drifted {
			drifted++
		}
	}
	var insightStyle lipgloss.Style
	var insightText string
	if drifted > 0 {
		insightStyle = t.InsightWarn
		insightText = t.AmberB.Render(fmt.Sprintf("▲ %d file(s) changed primary owner  ", drifted)) +
			t.Amber.Render("New owner may lack context — good candidate for pair review.")
	} else {
		insightStyle = t.InsightOk
		insightText = t.GreenB.Render("✓  ") + t.Green.Render("No ownership drift — stable authorship.")
	}
	b.WriteString(insightStyle.Width(width - 6).Render(insightText))
	b.WriteString("\n\n")

	fileW := (width - 46) / 2
	if fileW < 15 {
		fileW = 15
	}
	b.WriteString("  " + t.TableHeader.Render(rpad("#", 3)) + "  " +
		t.TableHeader.Render(pad("FILE", fileW)) + "  " +
		t.TableHeader.Render(pad("OLD OWNER", fileW)) + "  " +
		t.TableHeader.Render(pad("NEW OWNER", fileW)) + "  " +
		t.TableHeader.Render("DRIFT?") + "\n")
	b.WriteString("  " + divider(t, width-4) + "\n")

	shown := 0
	for i := scroll; i < len(data) && shown < height-8; i++ {
		e := data[i]
		driftStr := t.Muted.Render("  —    ")
		if e.Drifted {
			driftStr = t.AmberB.Render("  ✕ YES")
		}
		oldStr := e.OldOwner
		if oldStr == "" {
			oldStr = "—"
		}
		b.WriteString("  " +
			t.Muted.Render(rpad(fmt.Sprintf("%d", i+1), 3)) + "  " +
			pad(truncatePath(e.File, fileW), fileW) + "  " +
			t.Dim.Render(pad(theme.Truncate(oldStr, fileW), fileW)) + "  " +
			t.Base.Render(pad(theme.Truncate(e.NewOwner, fileW), fileW)) + "  " +
			driftStr + "\n")
		shown++
	}
	if scroll+shown < len(data) {
		b.WriteString(t.Muted.Render(fmt.Sprintf("  … %d more", len(data)-scroll-shown)))
	}
	return b.String()
}

// ── PANEL 11: TEST RATIO ──────────────────────────────────────────────────────

func renderTestRatio(t *theme.Theme, data []git.TestRatioEntry, err error, loading bool, scroll, width, height int) string {
	var b strings.Builder
	b.WriteString(t.CmdBlock.Width(width - 4).Render(cmdLine(t, git.TestRatioCmd)))
	b.WriteString("\n")
	if loading {
		b.WriteString(t.Blue.Render("⟳ running git command…"))
		return b.String()
	}
	if err != nil {
		b.WriteString(t.RedB.Render("✗ error: ") + t.Dim.Render(err.Error()))
		return b.String()
	}
	if len(data) == 0 {
		b.WriteString(t.Dim.Render("  no commits found in the last year"))
		return b.String()
	}

	testTotal, srcTotal := 0, 0
	for _, e := range data {
		if e.IsTest {
			testTotal += e.Count
		} else {
			srcTotal += e.Count
		}
	}
	total := testTotal + srcTotal
	var testPct float64
	if total > 0 {
		testPct = float64(testTotal) / float64(total) * 100
	}

	var insightStyle lipgloss.Style
	var insightText string
	switch {
	case testPct < 10:
		insightStyle = t.InsightCrit
		insightText = t.RedB.Render(fmt.Sprintf("✕ only %.0f%% of changes are in test files  ", testPct)) +
			t.Red.Render("Tests are not keeping pace with source changes.")
	case testPct < 25:
		insightStyle = t.InsightWarn
		insightText = t.AmberB.Render(fmt.Sprintf("▲ %.0f%% test coverage proxy  ", testPct)) +
			t.Amber.Render("Test changes lag behind source — consider increasing test investment.")
	default:
		insightStyle = t.InsightOk
		insightText = t.GreenB.Render(fmt.Sprintf("✓  %.0f%% of changes include tests  ", testPct)) +
			t.Green.Render("Tests are keeping pace with source changes.")
	}
	b.WriteString(insightStyle.Width(width - 6).Render(insightText))
	b.WriteString("\n\n")

	fileW := width - 38
	if fileW < 20 {
		fileW = 20
	}
	b.WriteString("  " + t.TableHeader.Render(rpad("#", 3)) + "  " +
		t.TableHeader.Render(pad("FILE", fileW)) + "  " +
		t.TableHeader.Render(rpad("CHANGES", 7)) + "  " +
		t.TableHeader.Render(pad("FREQ", barWidth)) + "  " +
		t.TableHeader.Render("TYPE") + "\n")
	b.WriteString("  " + divider(t, width-4) + "\n")

	shown := 0
	for i := scroll; i < len(data) && shown < height-8; i++ {
		e := data[i]
		typeStr := t.Blue.Render("src ")
		barStyle := t.Blue
		if e.IsTest {
			typeStr = t.Cyan.Render("test")
			barStyle = t.Cyan
		}
		b.WriteString("  " +
			t.Muted.Render(rpad(fmt.Sprintf("%d", i+1), 3)) + "  " +
			pad(truncatePath(e.File, fileW), fileW) + "  " +
			t.Dim.Render(rpad(fmt.Sprintf("%d", e.Count), 7)) + "  " +
			t.Bar(barWidth, e.Pct, barStyle) + "  " +
			typeStr + "\n")
		shown++
	}
	if scroll+shown < len(data) {
		b.WriteString(t.Muted.Render(fmt.Sprintf("  … %d more", len(data)-scroll-shown)))
	}
	return b.String()
}

// ── PANEL 12: COMMIT SIZE DISTRIBUTION ───────────────────────────────────────

func renderCommitSizes(t *theme.Theme, data []git.CommitSizeBucket, err error, loading bool, width, height int) string {
	var b strings.Builder
	b.WriteString(t.CmdBlock.Width(width - 4).Render(cmdLine(t, git.CommitSizeCmd)))
	b.WriteString("\n")
	if loading {
		b.WriteString(t.Blue.Render("⟳ running git command…"))
		return b.String()
	}
	if err != nil {
		b.WriteString(t.RedB.Render("✗ error: ") + t.Dim.Render(err.Error()))
		return b.String()
	}

	total := 0
	for _, b2 := range data {
		total += b2.Count
	}
	if total == 0 {
		b.WriteString(t.Dim.Render("  no commits found in the last year"))
		return b.String()
	}

	hugeIdx := len(data) - 1
	hugePct := 0.0
	if len(data) > 0 {
		hugePct = data[hugeIdx].Pct
	}
	var insightStyle lipgloss.Style
	var insightText string
	if hugePct > 10 {
		insightStyle = t.InsightWarn
		insightText = t.AmberB.Render(fmt.Sprintf("▲ %.0f%% of commits are huge (1000+ lines)  ", hugePct)) +
			t.Amber.Render("Large commits are harder to review and riskier to revert.")
	} else {
		insightStyle = t.InsightOk
		insightText = t.GreenB.Render("✓  ") + t.Green.Render(fmt.Sprintf("%d commits analysed — commit size distribution looks healthy.", total))
	}
	b.WriteString(insightStyle.Width(width - 6).Render(insightText))
	b.WriteString("\n\n")

	barW := width - 32
	if barW < 20 {
		barW = 20
	}
	b.WriteString("  " + t.TableHeader.Render(pad("SIZE BUCKET", 22)) + "  " +
		t.TableHeader.Render(rpad("COUNT", 6)) + "  " +
		t.TableHeader.Render(rpad("PCT", 5)) + "  " +
		t.TableHeader.Render(pad("DISTRIBUTION", barW)) + "\n")
	b.WriteString("  " + divider(t, width-4) + "\n")

	for _, bucket := range data {
		barStyle := t.Green
		if bucket.Pct > 30 && strings.Contains(bucket.Label, "large") {
			barStyle = t.Amber
		}
		if strings.Contains(bucket.Label, "huge") && bucket.Pct > 5 {
			barStyle = t.RedB
		}
		b.WriteString("  " +
			pad(bucket.Label, 22) + "  " +
			t.Dim.Render(rpad(fmt.Sprintf("%d", bucket.Count), 6)) + "  " +
			t.Amber.Render(rpad(fmt.Sprintf("%.0f%%", bucket.Pct), 5)) + "  " +
			t.Bar(barW, bucket.Pct, barStyle) + "\n")
	}
	return b.String()
}

// ── PANEL 13: MERGE FREQUENCY ─────────────────────────────────────────────────

func renderMergeFreq(t *theme.Theme, data []git.MergeFreqEntry, err error, loading bool, width, height int) string {
	var b strings.Builder
	b.WriteString(t.CmdBlock.Width(width - 4).Render(cmdLine(t, git.MergeFreqCmd)))
	b.WriteString("\n")
	if loading {
		b.WriteString(t.Blue.Render("⟳ running git command…"))
		return b.String()
	}
	if err != nil {
		b.WriteString(t.RedB.Render("✗ error: ") + t.Dim.Render(err.Error()))
		return b.String()
	}
	if len(data) == 0 {
		b.WriteString(t.Dim.Render("  no merge commits found — repo may use rebase workflow"))
		return b.String()
	}

	total, maxC := 0, 0
	for _, e := range data {
		total += e.Count
		if e.Count > maxC {
			maxC = e.Count
		}
	}
	avg := float64(total) / float64(len(data))
	recent := data
	if len(data) > 3 {
		recent = data[len(data)-3:]
	}
	recentTotal := 0
	for _, e := range recent {
		recentTotal += e.Count
	}
	recentAvg := float64(recentTotal) / float64(len(recent))

	trendStr := t.Green.Render("→ steady")
	if recentAvg > avg*1.2 {
		trendStr = t.AmberB.Render("↑ increasing")
	} else if recentAvg < avg*0.7 {
		trendStr = t.Dim.Render("↓ declining")
	}

	b.WriteString(t.InsightInfo.Width(width - 6).Render(
		t.Blue.Render(fmt.Sprintf("total merges: %d   avg/mo: %.0f   recent avg: %.0f   trend: ", total, avg, recentAvg)) + trendStr,
	))
	b.WriteString("\n\n")

	chartH := height - 12
	if chartH < 4 {
		chartH = 4
	}
	if chartH > 16 {
		chartH = 16
	}
	colW := 6
	maxCols := (width - 8) / colW
	visible := data
	if len(visible) > maxCols {
		visible = visible[len(visible)-maxCols:]
	}

	for row := 0; row < chartH; row++ {
		rowVal := float64(maxC) * float64(chartH-row) / float64(chartH)
		b.WriteString(t.Muted.Render(fmt.Sprintf("  %4d │", int(rowVal))))
		for _, e := range visible {
			filled := int(math.Ceil(float64(e.Count) / float64(maxC) * float64(chartH)))
			if chartH-row <= filled {
				pct := float64(e.Count) / float64(maxC)
				var colStyle lipgloss.Style
				if pct > 0.75 {
					colStyle = t.RedB
				} else if pct > 0.4 {
					colStyle = t.Amber
				} else {
					colStyle = t.Blue
				}
				b.WriteString(colStyle.Render(" ████ "))
			} else {
				b.WriteString("      ")
			}
		}
		b.WriteString("\n")
	}
	b.WriteString(t.Muted.Render("       └"+strings.Repeat("──────", len(visible))) + "\n")
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
	if loading {
		b.WriteString(t.Blue.Render("⟳ running git command…"))
		return b.String()
	}
	if err != nil && len(data) == 0 {
		b.WriteString(t.RedB.Render("✗ error: ") + t.Dim.Render(err.Error()))
		return b.String()
	}

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

	if len(data) == 0 {
		b.WriteString(t.Dim.Render("  (no results)"))
		return b.String()
	}

	msgW := width - 30
	if msgW < 20 {
		msgW = 20
	}
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
		kStyle, ok := kindColor[e.Kind]
		if !ok {
			kStyle = t.Dim
		}
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
