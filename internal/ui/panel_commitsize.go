package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/you/git-audit/internal/git"
	"github.com/you/git-audit/internal/theme"
)

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
