// Package theme provides Gruvbox Dark and Gruvbox Light palettes,
// terminal background detection via OSC 11, and a Theme struct that
// carries all lipgloss styles so callers never reference raw colours.
package theme

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

// ── VARIANT ───────────────────────────────────────────────────────────────────

// Variant is either Dark or Light.
type Variant int

const (
	Dark  Variant = iota
	Light         // Gruvbox Light
)

func (v Variant) String() string {
	if v == Light {
		return "light"
	}
	return "dark"
}

// ── PALETTE ───────────────────────────────────────────────────────────────────

type palette struct {
	Bg, Bg1, Bg2, Bg3             string
	Border, BorderHi              string
	Fg, FgDim, FgMuted            string
	Green, GreenDim               string
	Amber, AmberDim               string
	Red, RedDim                   string
	Blue, BlueDim                 string
	Cyan, Purple, Orange          string
	BadgeRedBg, BadgeAmberBg      string
	BadgeAmberDkBg, BadgeGreenBg  string
}

var darkPalette = palette{
	Bg: "#1d2021", Bg1: "#282828", Bg2: "#32302f", Bg3: "#3c3836",
	Border: "#504945", BorderHi: "#665c54",
	Fg: "#ebdbb2", FgDim: "#a89984", FgMuted: "#7c6f64",
	Green: "#b8bb26", GreenDim: "#98971a",
	Amber: "#fabd2f", AmberDim: "#d79921",
	Red: "#fb4934", RedDim: "#cc241d",
	Blue: "#83a598", BlueDim: "#458588",
	Cyan: "#8ec07c", Purple: "#d3869b", Orange: "#fe8019",
	BadgeRedBg: "#cc241d", BadgeAmberBg: "#5a3a1a",
	BadgeAmberDkBg: "#3d3000", BadgeGreenBg: "#2a3a1a",
}

var lightPalette = palette{
	Bg: "#fbf1c7", Bg1: "#f2e5bc", Bg2: "#ebdbb2", Bg3: "#d5c4a1",
	Border: "#bdae93", BorderHi: "#a89984",
	Fg: "#3c3836", FgDim: "#504945", FgMuted: "#7c6f64",
	Green: "#79740e", GreenDim: "#b8bb26",
	Amber: "#b57614", AmberDim: "#d79921",
	Red: "#9d0006", RedDim: "#cc241d",
	Blue: "#076678", BlueDim: "#458588",
	Cyan: "#427b58", Purple: "#8f3f71", Orange: "#af3a03",
	BadgeRedBg: "#f9d8d6", BadgeAmberBg: "#fef3d0",
	BadgeAmberDkBg: "#f8eec4", BadgeGreenBg: "#e6efc6",
}

// ── THEME STRUCT ──────────────────────────────────────────────────────────────

// Theme carries every lipgloss style used by the app.
type Theme struct {
	Variant Variant
	P       palette

	Base, Dim, Muted, Bold lipgloss.Style

	Green, GreenB         lipgloss.Style
	Amber, AmberB         lipgloss.Style
	Red, RedB             lipgloss.Style
	Blue, BlueB           lipgloss.Style
	Cyan, Purple, Orange  lipgloss.Style

	Border, ActiveBorder  lipgloss.Style

	AppTitle, TitleBar    lipgloss.Style
	TabActive, TabInactive lipgloss.Style
	StatusBar, StatusKey, StatusMode lipgloss.Style
	SidebarItem, SidebarActive       lipgloss.Style
	TableHeader                      lipgloss.Style

	CmdBlock, CmdPrompt, CmdKeyword  lipgloss.Style
	CmdFlag, CmdString, CmdPipe, CmdUtil lipgloss.Style

	RiskCritical, RiskHigh, RiskMedium, RiskLow lipgloss.Style
	InsightWarn, InsightCrit, InsightOk, InsightInfo lipgloss.Style
}

// New returns a fully built Theme for the given variant.
func New(v Variant) *Theme {
	t := &Theme{Variant: v}
	t.apply()
	return t
}

// Toggle flips between Dark and Light and rebuilds every style.
func (t *Theme) Toggle() {
	if t.Variant == Dark {
		t.Variant = Light
	} else {
		t.Variant = Dark
	}
	t.apply()
}

func (t *Theme) apply() {
	p := darkPalette
	if t.Variant == Light {
		p = lightPalette
	}
	t.P = p
	c := func(h string) lipgloss.Color { return lipgloss.Color(h) }

	t.Base  = lipgloss.NewStyle().Foreground(c(p.Fg))
	t.Dim   = lipgloss.NewStyle().Foreground(c(p.FgDim))
	t.Muted = lipgloss.NewStyle().Foreground(c(p.FgMuted))
	t.Bold  = lipgloss.NewStyle().Foreground(c(p.Fg)).Bold(true)

	t.Green  = lipgloss.NewStyle().Foreground(c(p.Green))
	t.GreenB = lipgloss.NewStyle().Foreground(c(p.Green)).Bold(true)
	t.Amber  = lipgloss.NewStyle().Foreground(c(p.Amber))
	t.AmberB = lipgloss.NewStyle().Foreground(c(p.Amber)).Bold(true)
	t.Red    = lipgloss.NewStyle().Foreground(c(p.Red))
	t.RedB   = lipgloss.NewStyle().Foreground(c(p.Red)).Bold(true)
	t.Blue   = lipgloss.NewStyle().Foreground(c(p.Blue))
	t.BlueB  = lipgloss.NewStyle().Foreground(c(p.Blue)).Bold(true)
	t.Cyan   = lipgloss.NewStyle().Foreground(c(p.Cyan))
	t.Purple = lipgloss.NewStyle().Foreground(c(p.Purple))
	t.Orange = lipgloss.NewStyle().Foreground(c(p.Orange))

	t.Border = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(c(p.Border))
	t.ActiveBorder = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(c(p.Blue))

	t.AppTitle = lipgloss.NewStyle().Background(c(p.Blue)).Foreground(c(p.Bg)).Bold(true).PaddingLeft(1).PaddingRight(1)
	t.TitleBar = lipgloss.NewStyle().Background(c(p.Bg3)).Foreground(c(p.Fg)).Bold(true).PaddingLeft(1).PaddingRight(1)

	t.TabActive = lipgloss.NewStyle().
		Background(c(p.Bg)).Foreground(c(p.Amber)).Bold(true).
		PaddingLeft(1).PaddingRight(1).Underline(true)
	t.TabInactive = lipgloss.NewStyle().
		Background(c(p.Bg)).Foreground(c(p.FgDim)).PaddingLeft(1).PaddingRight(1)

	t.StatusBar  = lipgloss.NewStyle().Background(c(p.Bg3)).Foreground(c(p.FgDim)).PaddingLeft(1)
	t.StatusKey  = lipgloss.NewStyle().Background(c(p.Bg3)).Foreground(c(p.Blue)).Bold(true)
	t.StatusMode = lipgloss.NewStyle().Background(c(p.Blue)).Foreground(c(p.Bg)).Bold(true).PaddingLeft(1).PaddingRight(1)

	t.SidebarItem   = lipgloss.NewStyle().PaddingLeft(1)
	t.SidebarActive = lipgloss.NewStyle().Background(c(p.Bg3)).Foreground(c(p.Amber)).Bold(true).PaddingLeft(1)

	t.TableHeader = lipgloss.NewStyle().Foreground(c(p.FgMuted)).Bold(true)

	t.CmdBlock = lipgloss.NewStyle().
		Background(c(p.Bg2)).Foreground(c(p.Amber)).
		PaddingLeft(1).PaddingRight(1).
		Border(lipgloss.RoundedBorder()).BorderForeground(c(p.Border))
	t.CmdPrompt  = lipgloss.NewStyle().Foreground(c(p.FgMuted))
	t.CmdKeyword = lipgloss.NewStyle().Foreground(c(p.Blue)).Bold(true)
	t.CmdFlag    = lipgloss.NewStyle().Foreground(c(p.Cyan))
	t.CmdString  = lipgloss.NewStyle().Foreground(c(p.Amber))
	t.CmdPipe    = lipgloss.NewStyle().Foreground(c(p.FgMuted))
	t.CmdUtil    = lipgloss.NewStyle().Foreground(c(p.Green))

	t.RiskCritical = lipgloss.NewStyle().Background(c(p.BadgeRedBg)).Foreground(c(p.Red)).Bold(true).PaddingLeft(1).PaddingRight(1)
	t.RiskHigh     = lipgloss.NewStyle().Background(c(p.BadgeAmberBg)).Foreground(c(p.Amber)).Bold(true).PaddingLeft(1).PaddingRight(1)
	t.RiskMedium   = lipgloss.NewStyle().Background(c(p.BadgeAmberDkBg)).Foreground(c(p.AmberDim)).PaddingLeft(1).PaddingRight(1)
	t.RiskLow      = lipgloss.NewStyle().Background(c(p.BadgeGreenBg)).Foreground(c(p.GreenDim)).PaddingLeft(1).PaddingRight(1)

	t.InsightWarn = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(c(p.AmberDim)).Foreground(c(p.Amber)).PaddingLeft(1).PaddingRight(1)
	t.InsightCrit = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(c(p.RedDim)).Foreground(c(p.Red)).PaddingLeft(1).PaddingRight(1)
	t.InsightOk   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(c(p.GreenDim)).Foreground(c(p.Green)).PaddingLeft(1).PaddingRight(1)
	t.InsightInfo = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(c(p.BlueDim)).Foreground(c(p.Blue)).PaddingLeft(1).PaddingRight(1)
}

// ── HELPERS ───────────────────────────────────────────────────────────────────

func (t *Theme) RiskStyle(pct float64) lipgloss.Style {
	switch {
	case pct > 75:
		return t.RiskCritical
	case pct > 50:
		return t.RiskHigh
	case pct > 25:
		return t.RiskMedium
	default:
		return t.RiskLow
	}
}

func RiskLabel(pct float64) string {
	switch {
	case pct > 75: return "critical"
	case pct > 50: return "high"
	case pct > 25: return "medium"
	default:        return "low"
	}
}

func (t *Theme) Bar(width int, pct float64, fillStyle lipgloss.Style) string {
	if width <= 0 { return "" }
	filled := int(float64(width) * pct / 100)
	if filled > width { filled = width }
	empty := width - filled
	s := ""
	if filled > 0 { s += fillStyle.Render(strings.Repeat("█", filled)) }
	if empty  > 0 { s += t.Muted.Render(strings.Repeat("░", empty))  }
	return s
}

func Truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n { return s }
	return string(r[:n-1]) + "…"
}

// ── TERMINAL BACKGROUND DETECTION ────────────────────────────────────────────

// DetectVariant queries the terminal background via OSC 11, falling back
// to env-var heuristics.  Call this once before starting the Bubble Tea
// program (the raw-mode switch would eat the response otherwise).
//
// Supported terminals: kitty, Alacritty, WezTerm, iTerm2, GNOME Terminal,
// Konsole, foot, and most xterm-compatible emulators.
//
// Override at any time with:  GIT_AUDIT_THEME=light  or  =dark
func DetectVariant() Variant {
	// Env override wins
	if v := os.Getenv("GIT_AUDIT_THEME"); v != "" {
		if strings.ToLower(v) == "light" { return Light }
		return Dark
	}

	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return detectFromEnv()
	}
	defer tty.Close()

	// Switch to raw mode so the terminal's OSC 11 response is not echoed
	// back to the screen as visible garbage characters.
	oldState, err := term.MakeRaw(int(tty.Fd()))
	if err != nil {
		return detectFromEnv()
	}
	defer term.Restore(int(tty.Fd()), oldState) //nolint:errcheck

	// OSC 11 — query background colour
	if _, err := fmt.Fprint(tty, "\x1b]11;?\x07"); err != nil {
		return detectFromEnv()
	}

	tty.SetReadDeadline(time.Now().Add(150 * time.Millisecond)) //nolint:errcheck
	buf := make([]byte, 64)
	n, err := tty.Read(buf)
	if err != nil || n == 0 {
		return detectFromEnv()
	}

	resp := string(buf[:n])
	// Response: ESC ] 11 ; rgb:RRRR/GGGG/BBBB BEL
	idx := strings.Index(resp, "rgb:")
	if idx < 0 {
		return detectFromEnv()
	}
	rgb := resp[idx+4:]
	parts := strings.SplitN(rgb, "/", 2)
	if len(parts) < 1 || len(parts[0]) < 2 {
		return detectFromEnv()
	}
	var redByte int
	fmt.Sscanf(parts[0][:2], "%x", &redByte)
	if redByte > 0x7f {
		return Light
	}
	return Dark
}

// detectFromEnv uses COLORFGBG as a secondary signal.
func detectFromEnv() Variant {
	if cfg := os.Getenv("COLORFGBG"); cfg != "" {
		parts := strings.Split(cfg, ";")
		var bg int
		fmt.Sscanf(parts[len(parts)-1], "%d", &bg)
		if bg >= 8 { return Light }
	}
	return Dark
}
