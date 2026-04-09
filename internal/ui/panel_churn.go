package ui

import (
	"fmt"
	"strings"

	"github.com/you/git-audit/internal/git"
	"github.com/you/git-audit/internal/theme"
)

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
