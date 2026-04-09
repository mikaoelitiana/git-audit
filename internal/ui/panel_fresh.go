package ui

import (
	"fmt"
	"strings"

	"github.com/you/git-audit/internal/git"
	"github.com/you/git-audit/internal/theme"
)

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

	b.WriteString(t.InsightInfo.Width(width - 4).Render(
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
