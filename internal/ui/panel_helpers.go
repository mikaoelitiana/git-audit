package ui

import (
	"strings"

	"github.com/you/git-audit/internal/theme"
)

const barWidth = 14

// cmdLine formats a git command string with syntax-highlighted tokens.
func cmdLine(t *theme.Theme, cmd string) string {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return cmd
	}
	var out strings.Builder
	out.WriteString(t.CmdPrompt.Render("$ "))
	for i, p := range parts {
		switch {
		case p == "git":
			out.WriteString(t.CmdKeyword.Render(p))
		case i == 1:
			out.WriteString(" " + t.Blue.Render(p))
		case strings.HasPrefix(p, "--") || strings.HasPrefix(p, "-"):
			out.WriteString(" " + t.CmdFlag.Render(p))
		case strings.HasPrefix(p, `"`) || strings.HasPrefix(p, `'`):
			out.WriteString(" " + t.CmdString.Render(p))
		case p == "|":
			out.WriteString(" " + t.CmdPipe.Render(p))
		case p == "sort" || p == "uniq" || p == "head" || p == "grep":
			out.WriteString(" " + t.CmdUtil.Render(p))
		default:
			out.WriteString(" " + t.Dim.Render(p))
		}
	}
	return out.String()
}

// divider renders a full-width horizontal rule.
func divider(t *theme.Theme, width int) string {
	return t.Muted.Render(strings.Repeat("─", width))
}

// truncatePath shortens a file path to fit within max runes,
// preserving the filename and as much of the directory as possible.
func truncatePath(path string, max int) string {
	if len([]rune(path)) <= max {
		return path
	}
	parts := strings.Split(path, "/")
	file := parts[len(parts)-1]
	if len(file) >= max {
		return theme.Truncate(file, max)
	}
	for i := len(parts) - 2; i >= 0; i-- {
		candidate := "…/" + strings.Join(parts[i:], "/")
		if len([]rune(candidate)) <= max {
			return candidate
		}
	}
	return theme.Truncate(file, max)
}

// pad right-pads s to exactly n runes (truncates if longer).
func pad(s string, n int) string {
	r := []rune(s)
	if len(r) >= n {
		return string(r[:n])
	}
	return s + strings.Repeat(" ", n-len(r))
}

// rpad left-pads s to exactly n runes (truncates if longer).
func rpad(s string, n int) string {
	r := []rune(s)
	if len(r) >= n {
		return string(r[:n])
	}
	return strings.Repeat(" ", n-len(r)) + s
}
