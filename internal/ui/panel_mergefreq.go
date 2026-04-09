package ui

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/you/git-audit/internal/git"
	"github.com/you/git-audit/internal/theme"
)

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

	b.WriteString(t.InsightInfo.Width(width - 4).Render(
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
