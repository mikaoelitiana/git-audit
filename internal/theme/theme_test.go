package theme

import (
	"fmt"
	"strings"
	"testing"
)

// ── Variant.String ────────────────────────────────────────────────────────────

func TestVariantString(t *testing.T) {
	if got := Dark.String(); got != "dark" {
		t.Errorf("Dark.String() = %q, want %q", got, "dark")
	}
	if got := Light.String(); got != "light" {
		t.Errorf("Light.String() = %q, want %q", got, "light")
	}
}

// ── RiskLabel ─────────────────────────────────────────────────────────────────

func TestRiskLabel(t *testing.T) {
	tests := []struct {
		pct  float64
		want string
	}{
		{100, "critical"},
		{76, "critical"},
		{75.1, "critical"},
		// boundary: 75 is NOT > 75
		{75, "high"},
		{51, "high"},
		{50.1, "high"},
		// boundary: 50 is NOT > 50
		{50, "medium"},
		{26, "medium"},
		{25.1, "medium"},
		// boundary: 25 is NOT > 25
		{25, "low"},
		{10, "low"},
		{0, "low"},
	}
	for _, tc := range tests {
		got := RiskLabel(tc.pct)
		if got != tc.want {
			t.Errorf("RiskLabel(%.1f) = %q, want %q", tc.pct, got, tc.want)
		}
	}
}

// ── Truncate ──────────────────────────────────────────────────────────────────

func TestTruncate(t *testing.T) {
	tests := []struct {
		s    string
		n    int
		want string
	}{
		{"hello", 10, "hello"},        // shorter than limit — no change
		{"hello", 5, "hello"},         // exactly at limit — no change
		{"hello world", 5, "hell…"},   // truncated: 4 chars + ellipsis = 5
		{"hello world", 6, "hello…"},  // truncated: 5 chars + ellipsis = 6
		{"", 5, ""},                   // empty string
		{"ab", 1, "…"},               // only room for ellipsis
		{"abc", 2, "a…"},             // one char + ellipsis
	}
	for _, tc := range tests {
		got := Truncate(tc.s, tc.n)
		if got != tc.want {
			t.Errorf("Truncate(%q, %d) = %q, want %q", tc.s, tc.n, got, tc.want)
		}
	}
}

// ── Theme.Toggle ──────────────────────────────────────────────────────────────

func TestThemeToggle(t *testing.T) {
	th := New(Dark)
	if th.Variant != Dark {
		t.Fatal("expected Dark variant after New(Dark)")
	}

	th.Toggle()
	if th.Variant != Light {
		t.Errorf("after Toggle: expected Light, got %v", th.Variant)
	}

	// Palette should have changed — spot-check one colour.
	if th.P.Bg == darkPalette.Bg {
		t.Error("after Toggle to Light, palette background should differ from dark palette")
	}

	th.Toggle()
	if th.Variant != Dark {
		t.Errorf("after second Toggle: expected Dark, got %v", th.Variant)
	}
	if th.P.Bg != darkPalette.Bg {
		t.Error("after Toggle back to Dark, palette background should match dark palette")
	}
}

// ── Theme.RiskStyle ───────────────────────────────────────────────────────────

func TestRiskStyle(t *testing.T) {
	th := New(Dark)

	// lipgloss.Style contains func fields and cannot be compared with ==.
	// GetForeground returns a lipgloss.TerminalColor (interface). The concrete
	// type is lipgloss.Color (a string alias), so fmt.Sprint gives the hex
	// colour string — stable and TTY-independent.
	fgOf := func(pct float64) string {
		return fmt.Sprint(th.RiskStyle(pct).GetForeground())
	}

	// Within the same tier the foreground colour must be identical.
	if fgOf(100) != fgOf(76) {
		t.Error("100% and 76% should map to the same (critical) tier")
	}
	if fgOf(75) != fgOf(51) {
		t.Error("75% and 51% should map to the same (high) tier")
	}
	if fgOf(50) != fgOf(26) {
		t.Error("50% and 26% should map to the same (medium) tier")
	}
	if fgOf(25) != fgOf(0) {
		t.Error("25% and 0% should map to the same (low) tier")
	}

	// At each boundary the tier — and thus the foreground colour — must change.
	if fgOf(76) == fgOf(75) {
		t.Error("76% (critical) and 75% (high) should map to different tiers")
	}
	if fgOf(51) == fgOf(50) {
		t.Error("51% (high) and 50% (medium) should map to different tiers")
	}
	if fgOf(26) == fgOf(25) {
		t.Error("26% (medium) and 25% (low) should map to different tiers")
	}
}

// ── Theme.Bar ─────────────────────────────────────────────────────────────────

func TestBar(t *testing.T) {
	th := New(Dark)

	t.Run("zero width returns empty string", func(t *testing.T) {
		got := th.Bar(0, 50, th.Green)
		if got != "" {
			t.Errorf("Bar(0, 50, ...) = %q, want empty string", got)
		}
	})

	t.Run("100 pct fills bar with filled blocks only", func(t *testing.T) {
		got := th.Bar(10, 100, th.Green)
		if !strings.Contains(got, "█") {
			t.Error("100% bar should contain filled blocks")
		}
		if strings.Contains(got, "░") {
			t.Error("100% bar should not contain empty blocks")
		}
	})

	t.Run("0 pct fills bar with empty blocks only", func(t *testing.T) {
		got := th.Bar(10, 0, th.Green)
		if strings.Contains(got, "█") {
			t.Error("0% bar should not contain filled blocks")
		}
		if !strings.Contains(got, "░") {
			t.Error("0% bar should contain empty blocks")
		}
	})

	t.Run("50 pct contains both block types", func(t *testing.T) {
		got := th.Bar(10, 50, th.Green)
		if !strings.Contains(got, "█") {
			t.Error("50% bar should contain filled blocks")
		}
		if !strings.Contains(got, "░") {
			t.Error("50% bar should contain empty blocks")
		}
	})

	t.Run("pct over 100 is capped", func(t *testing.T) {
		got := th.Bar(10, 200, th.Green)
		if strings.Contains(got, "░") {
			t.Error("200% bar should be capped at full — no empty blocks")
		}
	})
}

// ── DetectVariant env override ────────────────────────────────────────────────

func TestDetectVariantEnvOverride(t *testing.T) {
	t.Setenv("GIT_AUDIT_THEME", "light")
	if got := DetectVariant(); got != Light {
		t.Errorf("GIT_AUDIT_THEME=light: expected Light, got %v", got)
	}

	t.Setenv("GIT_AUDIT_THEME", "dark")
	if got := DetectVariant(); got != Dark {
		t.Errorf("GIT_AUDIT_THEME=dark: expected Dark, got %v", got)
	}

	// Unknown value defaults to Dark.
	t.Setenv("GIT_AUDIT_THEME", "solarized")
	if got := DetectVariant(); got != Dark {
		t.Errorf("GIT_AUDIT_THEME=solarized: expected Dark fallback, got %v", got)
	}
}

// ── detectFromEnv ─────────────────────────────────────────────────────────────

func TestDetectFromEnv(t *testing.T) {
	t.Run("light background via COLORFGBG", func(t *testing.T) {
		t.Setenv("COLORFGBG", "0;15") // bg=15 → light
		if got := detectFromEnv(); got != Light {
			t.Errorf("COLORFGBG=0;15: expected Light, got %v", got)
		}
	})

	t.Run("dark background via COLORFGBG", func(t *testing.T) {
		t.Setenv("COLORFGBG", "15;0") // bg=0 → dark
		if got := detectFromEnv(); got != Dark {
			t.Errorf("COLORFGBG=15;0: expected Dark, got %v", got)
		}
	})

	t.Run("missing COLORFGBG defaults to dark", func(t *testing.T) {
		t.Setenv("COLORFGBG", "")
		if got := detectFromEnv(); got != Dark {
			t.Errorf("empty COLORFGBG: expected Dark, got %v", got)
		}
	})
}
