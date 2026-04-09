package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/you/git-audit/internal/git"
	"github.com/you/git-audit/internal/theme"
)

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

	activeCount := 0
	for _, c := range data {
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
