package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// ── TYPES ────────────────────────────────────────────────────────────────────

type ChurnEntry struct {
	File  string
	Count int
	Pct   float64 // relative to max, 0–100
}

type Contributor struct {
	Name    string
	Commits int
	Pct     float64 // of total
	Active  bool    // committed in last 6 months
}

type BugEntry struct {
	File     string
	Count    int
	Pct      float64
	InChurn  bool
}

type VelocityEntry struct {
	Month string // "2024-03"
	Count int
}

type HotfixEntry struct {
	Hash    string
	Message string
	Kind    string // revert | hotfix | rollback | emergency
}

// ── RUNNER ───────────────────────────────────────────────────────────────────

func run(cwd, cmd string) ([]string, error) {
	out, err := exec.Command("sh", "-c", cmd).Output()
	_ = cwd // exec inherits env; caller sets Dir via the command string or we cd
	if err != nil {
		// grep returns exit 1 when no matches — that's OK
		if len(out) > 0 {
			return splitLines(string(out)), nil
		}
		return nil, fmt.Errorf("%w", err)
	}
	return splitLines(string(out)), nil
}

func runIn(cwd, cmd string) ([]string, error) {
	c := exec.Command("sh", "-c", cmd)
	c.Dir = cwd
	out, err := c.Output()
	if err != nil {
		if len(out) > 0 {
			return splitLines(string(out)), nil
		}
		return nil, err
	}
	return splitLines(string(out)), nil
}

func splitLines(s string) []string {
	var lines []string
	for _, l := range strings.Split(s, "\n") {
		if l != "" {
			lines = append(lines, l)
		}
	}
	return lines
}

// RepoName returns the base name of the repo directory.
func RepoName(path string) string {
	return filepath.Base(filepath.Clean(path))
}

// IsGitRepo returns true if cwd is inside a git repository.
func IsGitRepo(cwd string) bool {
	c := exec.Command("git", "rev-parse", "--git-dir")
	c.Dir = cwd
	return c.Run() == nil
}

// CurrentBranch returns the current branch name.
func CurrentBranch(cwd string) string {
	lines, err := runIn(cwd, "git rev-parse --abbrev-ref HEAD")
	if err != nil || len(lines) == 0 {
		return "unknown"
	}
	return lines[0]
}

// TotalCommits returns total commit count.
func TotalCommits(cwd string) int {
	lines, err := runIn(cwd, "git rev-list --count HEAD")
	if err != nil || len(lines) == 0 {
		return 0
	}
	n, _ := strconv.Atoi(strings.TrimSpace(lines[0]))
	return n
}

// ── COMMAND 1: CHURN ─────────────────────────────────────────────────────────

const ChurnCmd = `git log --format=format: --name-only --since="1 year ago" | sort | uniq -c | sort -nr | head -20`

var countFileRe = regexp.MustCompile(`^\s*(\d+)\s+(.+)$`)

func Churn(cwd string) ([]ChurnEntry, error) {
	lines, err := runIn(cwd, ChurnCmd)
	if err != nil {
		return nil, err
	}
	var entries []ChurnEntry
	for _, l := range lines {
		m := countFileRe.FindStringSubmatch(l)
		if m == nil {
			continue
		}
		n, _ := strconv.Atoi(m[1])
		entries = append(entries, ChurnEntry{File: strings.TrimSpace(m[2]), Count: n})
	}
	if len(entries) > 0 {
		max := entries[0].Count
		for i := range entries {
			entries[i].Pct = float64(entries[i].Count) / float64(max) * 100
		}
	}
	return entries, nil
}

// ── COMMAND 2: BUS FACTOR ────────────────────────────────────────────────────

const BusFactorCmd = `git shortlog -sn --no-merges`
const BusFactorRecentCmd = `git shortlog -sn --no-merges --since="6 months ago"`

func BusFactor(cwd string) ([]Contributor, error) {
	lines, err := runIn(cwd, BusFactorCmd)
	if err != nil {
		return nil, err
	}

	recentLines, _ := runIn(cwd, BusFactorRecentCmd)
	recentSet := make(map[string]bool)
	for _, l := range recentLines {
		m := countFileRe.FindStringSubmatch(l)
		if m != nil {
			recentSet[strings.TrimSpace(m[2])] = true
		}
	}

	var contribs []Contributor
	total := 0
	for _, l := range lines {
		m := countFileRe.FindStringSubmatch(l)
		if m == nil {
			continue
		}
		n, _ := strconv.Atoi(m[1])
		name := strings.TrimSpace(m[2])
		total += n
		contribs = append(contribs, Contributor{
			Name:    name,
			Commits: n,
			Active:  recentSet[name],
		})
	}
	if total > 0 {
		for i := range contribs {
			contribs[i].Pct = float64(contribs[i].Commits) / float64(total) * 100
		}
	}
	return contribs, nil
}

// ── COMMAND 3: BUG CLUSTERS ──────────────────────────────────────────────────

const BugCmd = `git log -i -E --grep="fix|bug|broken" --name-only --format='' | sort | uniq -c | sort -nr | head -20`

func BugClusters(cwd string, churnFiles map[string]bool) ([]BugEntry, error) {
	lines, err := runIn(cwd, BugCmd)
	if err != nil {
		return nil, err
	}
	var entries []BugEntry
	for _, l := range lines {
		m := countFileRe.FindStringSubmatch(l)
		if m == nil {
			continue
		}
		n, _ := strconv.Atoi(m[1])
		file := strings.TrimSpace(m[2])
		entries = append(entries, BugEntry{
			File:    file,
			Count:   n,
			InChurn: churnFiles[file],
		})
	}
	if len(entries) > 0 {
		max := entries[0].Count
		for i := range entries {
			entries[i].Pct = float64(entries[i].Count) / float64(max) * 100
		}
	}
	return entries, nil
}

// ── COMMAND 4: VELOCITY ──────────────────────────────────────────────────────

const VelocityCmd = `git log --format='%ad' --date=format:'%Y-%m' | sort | uniq -c`

func Velocity(cwd string) ([]VelocityEntry, error) {
	lines, err := runIn(cwd, VelocityCmd)
	if err != nil {
		return nil, err
	}
	monthRe := regexp.MustCompile(`^\s*(\d+)\s+(\d{4}-\d{2})$`)
	var entries []VelocityEntry
	for _, l := range lines {
		m := monthRe.FindStringSubmatch(l)
		if m == nil {
			continue
		}
		n, _ := strconv.Atoi(m[1])
		entries = append(entries, VelocityEntry{Month: m[2], Count: n})
	}
	return entries, nil
}

// ── COMMAND 5: FIREFIGHTING ──────────────────────────────────────────────────

const FirefightCmd = `git log --oneline --since="1 year ago" | grep -iE 'revert|hotfix|emergency|rollback'`

func Firefighting(cwd string) ([]HotfixEntry, error) {
	lines, err := runIn(cwd, FirefightCmd)
	// grep returns exit 1 on no match — treat as empty, not error
	if err != nil && len(lines) == 0 {
		return nil, nil
	}
	var entries []HotfixEntry
	for _, l := range lines {
		parts := strings.SplitN(strings.TrimSpace(l), " ", 2)
		hash := parts[0]
		msg := ""
		if len(parts) > 1 {
			msg = parts[1]
		}
		lower := strings.ToLower(msg)
		kind := "emergency"
		switch {
		case strings.Contains(lower, "revert"):
			kind = "revert"
		case strings.Contains(lower, "hotfix"):
			kind = "hotfix"
		case strings.Contains(lower, "rollback"):
			kind = "rollback"
		}
		entries = append(entries, HotfixEntry{Hash: hash, Message: msg, Kind: kind})
	}
	return entries, nil
}
