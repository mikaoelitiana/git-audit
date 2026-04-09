package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/you/git-audit/internal/git"
	"github.com/you/git-audit/internal/theme"
)

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
	for i := scroll; i < len(data) && shown < height-11; i++ {
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
