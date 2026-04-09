package ui

import "testing"

// ── pad ───────────────────────────────────────────────────────────────────────

func TestPad(t *testing.T) {
	tests := []struct {
		s    string
		n    int
		want string
	}{
		{"hi", 5, "hi   "},       // right-pads with spaces
		{"hello", 5, "hello"},    // exact length — no change
		{"toolong", 4, "tool"},   // truncates to n
		{"", 3, "   "},           // empty string pads to n spaces
		{"a", 1, "a"},            // single char, exact
	}
	for _, tc := range tests {
		got := pad(tc.s, tc.n)
		if got != tc.want {
			t.Errorf("pad(%q, %d) = %q, want %q", tc.s, tc.n, got, tc.want)
		}
	}
}

// ── rpad ──────────────────────────────────────────────────────────────────────

func TestRpad(t *testing.T) {
	tests := []struct {
		s    string
		n    int
		want string
	}{
		{"hi", 5, "   hi"},       // left-pads with spaces
		{"hello", 5, "hello"},    // exact length — no change
		{"toolong", 4, "tool"},   // truncates to n
		{"", 3, "   "},           // empty string pads to n spaces
		{"a", 1, "a"},            // single char, exact
	}
	for _, tc := range tests {
		got := rpad(tc.s, tc.n)
		if got != tc.want {
			t.Errorf("rpad(%q, %d) = %q, want %q", tc.s, tc.n, got, tc.want)
		}
	}
}

// ── truncatePath ──────────────────────────────────────────────────────────────

func TestTruncatePath(t *testing.T) {
	tests := []struct {
		path string
		max  int
		want string
	}{
		// Short enough — returned as-is.
		{"src/main.go", 20, "src/main.go"},
		{"main.go", 7, "main.go"},

		// Path fits but with some room.
		{"a/b/c.go", 10, "a/b/c.go"},

		// Filename alone fits; show as much of the dir as possible.
		{"very/long/path/to/file.go", 15, "…/to/file.go"},

		// Filename (19 chars) longer than max=18 — falls back to theme.Truncate.
		{"a/b/c/d/e/f/g/h/verylongfilename.go", 18, "verylongfilename.…"},

		// Filename itself is longer than max — truncate with ellipsis via theme.Truncate.
		{"reallylongfilename.go", 10, "reallylon…"},
	}
	for _, tc := range tests {
		got := truncatePath(tc.path, tc.max)
		if got != tc.want {
			t.Errorf("truncatePath(%q, %d) = %q, want %q", tc.path, tc.max, got, tc.want)
		}
	}
}
