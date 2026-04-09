package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/you/git-audit/internal/git"
	"github.com/you/git-audit/internal/theme"
)

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
	b.WriteString(insightStyle.Width(width - 4).Render(insightText))
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
	for i := scroll; i < len(data) && shown < height-11; i++ {
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
