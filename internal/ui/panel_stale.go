package ui

import (
	"fmt"
	"strings"

	"github.com/you/git-audit/internal/git"
	"github.com/you/git-audit/internal/theme"
)

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
	for i := scroll; i < len(data) && shown < height-11; i++ {
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
