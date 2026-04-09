package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/you/git-audit/internal/git"
	"github.com/you/git-audit/internal/theme"
)

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
	for i := scroll; i < len(data) && shown < height-11; i++ {
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
