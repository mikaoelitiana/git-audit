package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/you/git-audit/internal/git"
	"github.com/you/git-audit/internal/theme"
)

func renderTestRatio(t *theme.Theme, data []git.TestRatioEntry, err error, loading bool, scroll, width, height int) string {
	var b strings.Builder
	b.WriteString(t.CmdBlock.Width(width - 4).Render(cmdLine(t, git.TestRatioCmd)))
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
		b.WriteString(t.Dim.Render("  no commits found in the last year"))
		return b.String()
	}

	testTotal, srcTotal := 0, 0
	for _, e := range data {
		if e.IsTest {
			testTotal += e.Count
		} else {
			srcTotal += e.Count
		}
	}
	total := testTotal + srcTotal
	var testPct float64
	if total > 0 {
		testPct = float64(testTotal) / float64(total) * 100
	}

	var insightStyle lipgloss.Style
	var insightText string
	switch {
	case testPct < 10:
		insightStyle = t.InsightCrit
		insightText = t.RedB.Render(fmt.Sprintf("✕ only %.0f%% of changes are in test files  ", testPct)) +
			t.Red.Render("Tests are not keeping pace with source changes.")
	case testPct < 25:
		insightStyle = t.InsightWarn
		insightText = t.AmberB.Render(fmt.Sprintf("▲ %.0f%% test coverage proxy  ", testPct)) +
			t.Amber.Render("Test changes lag behind source — consider increasing test investment.")
	default:
		insightStyle = t.InsightOk
		insightText = t.GreenB.Render(fmt.Sprintf("✓  %.0f%% of changes include tests  ", testPct)) +
			t.Green.Render("Tests are keeping pace with source changes.")
	}
	b.WriteString(insightStyle.Width(width - 6).Render(insightText))
	b.WriteString("\n\n")

	fileW := width - 38
	if fileW < 20 {
		fileW = 20
	}
	b.WriteString("  " + t.TableHeader.Render(rpad("#", 3)) + "  " +
		t.TableHeader.Render(pad("FILE", fileW)) + "  " +
		t.TableHeader.Render(rpad("CHANGES", 7)) + "  " +
		t.TableHeader.Render(pad("FREQ", barWidth)) + "  " +
		t.TableHeader.Render("TYPE") + "\n")
	b.WriteString("  " + divider(t, width-4) + "\n")

	shown := 0
	for i := scroll; i < len(data) && shown < height-8; i++ {
		e := data[i]
		typeStr := t.Blue.Render("src ")
		barStyle := t.Blue
		if e.IsTest {
			typeStr = t.Cyan.Render("test")
			barStyle = t.Cyan
		}
		b.WriteString("  " +
			t.Muted.Render(rpad(fmt.Sprintf("%d", i+1), 3)) + "  " +
			pad(truncatePath(e.File, fileW), fileW) + "  " +
			t.Dim.Render(rpad(fmt.Sprintf("%d", e.Count), 7)) + "  " +
			t.Bar(barWidth, e.Pct, barStyle) + "  " +
			typeStr + "\n")
		shown++
	}
	if scroll+shown < len(data) {
		b.WriteString(t.Muted.Render(fmt.Sprintf("  … %d more", len(data)-scroll-shown)))
	}
	return b.String()
}
