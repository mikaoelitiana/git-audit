package ui

import (
	"fmt"
	"strings"

	"github.com/you/git-audit/internal/git"
	"github.com/you/git-audit/internal/theme"
)

func renderCoupling(t *theme.Theme, data []git.CouplingEntry, err error, loading bool, scroll, width, height int) string {
	var b strings.Builder
	b.WriteString(t.CmdBlock.Width(width - 4).Render(cmdLine(t, git.CoChangeCmd)))
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
		b.WriteString(t.InsightOk.Width(width - 4).Render(t.GreenB.Render("✓  ") + t.Green.Render("No strong coupling detected (no pair changed together ≥3 times).")))
		return b.String()
	}
	b.WriteString(t.InsightWarn.Width(width - 4).Render(
		t.AmberB.Render("▲ insight  ") +
			t.Amber.Render("These file pairs are always committed together — hidden coupling, consider extracting shared logic."),
	))
	b.WriteString("\n\n")

	halfW := (width - 22) / 2
	if halfW < 15 {
		halfW = 15
	}
	b.WriteString("  " + t.TableHeader.Render(rpad("#", 3)) + "  " +
		t.TableHeader.Render(pad("FILE A", halfW)) + "  " +
		t.TableHeader.Render(pad("FILE B", halfW)) + "  " +
		t.TableHeader.Render(rpad("TOGETHER", 8)) + "  " +
		t.TableHeader.Render(pad("COUPLING", barWidth)) + "\n")
	b.WriteString("  " + divider(t, width-4) + "\n")

	shown := 0
	for i := scroll; i < len(data) && shown < height-11; i++ {
		e := data[i]
		barStyle := t.Blue
		if e.Pct > 75 {
			barStyle = t.RedB
		} else if e.Pct > 40 {
			barStyle = t.Amber
		}
		b.WriteString("  " +
			t.Muted.Render(rpad(fmt.Sprintf("%d", i+1), 3)) + "  " +
			pad(truncatePath(e.FileA, halfW), halfW) + "  " +
			pad(truncatePath(e.FileB, halfW), halfW) + "  " +
			t.Amber.Render(rpad(fmt.Sprintf("%d", e.Together), 8)) + "  " +
			t.Bar(barWidth, e.Pct, barStyle) + "\n")
		shown++
	}
	if scroll+shown < len(data) {
		b.WriteString(t.Muted.Render(fmt.Sprintf("  … %d more", len(data)-scroll-shown)))
	}
	return b.String()
}
