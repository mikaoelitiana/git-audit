package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/you/git-audit/internal/git"
	"github.com/you/git-audit/internal/theme"
)

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
		b.WriteString(t.InsightOk.Width(width - 4).Render(t.GreenB.Render("✓  ") + t.Green.Render("No ownership changes detected.")))
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
	b.WriteString(insightStyle.Width(width - 4).Render(insightText))
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
	for i := scroll; i < len(data) && shown < height-11; i++ {
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
