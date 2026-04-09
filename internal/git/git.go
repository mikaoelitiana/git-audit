package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
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

const BusFactorCmd = `git shortlog -sn --no-merges HEAD`
const BusFactorRecentCmd = `git shortlog -sn --no-merges --since="6 months ago" HEAD`

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

// ── COMMAND 6: STALE FILES ────────────────────────────────────────────────────

const StaleLogCmd = `git log --format='>>>%ad' --date=format:'%Y-%m-%d' --name-only --since="5 years ago"`
const StaleListCmd = `git ls-files`

type StaleEntry struct {
	File        string
	LastChanged string
	DaysAgo     int
}

func StaleFiles(cwd string) ([]StaleEntry, error) {
	logLines, _ := runIn(cwd, StaleLogCmd)
	lastSeen := make(map[string]string)
	var curDate string
	for _, l := range logLines {
		if strings.HasPrefix(l, ">>>") {
			curDate = l[3:]
			continue
		}
		if l == "" || curDate == "" {
			continue
		}
		if _, exists := lastSeen[l]; !exists {
			lastSeen[l] = curDate
		}
	}
	allFiles, err := runIn(cwd, StaleListCmd)
	if err != nil {
		return nil, err
	}
	cutoff := time.Now().AddDate(-1, 0, 0)
	var entries []StaleEntry
	for _, f := range allFiles {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		date, ok := lastSeen[f]
		if !ok {
			entries = append(entries, StaleEntry{File: f, LastChanged: "never", DaysAgo: 9999})
			continue
		}
		t, err := time.Parse("2006-01-02", date)
		if err != nil {
			continue
		}
		if t.Before(cutoff) {
			entries = append(entries, StaleEntry{File: f, LastChanged: date, DaysAgo: int(time.Since(t).Hours() / 24)})
		}
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].DaysAgo > entries[j].DaysAgo })
	if len(entries) > 30 {
		entries = entries[:30]
	}
	return entries, nil
}

// ── COMMAND 7: LONG-LIVED BRANCHES ───────────────────────────────────────────

const BranchListCmd = `git for-each-ref --sort=-committerdate refs/heads/ --format='%(refname:short)|%(committerdate:short)|%(authorname)'`

type BranchEntry struct {
	Name    string
	Date    string
	DaysAgo int
	Author  string
}

func LongBranches(cwd string) ([]BranchEntry, error) {
	lines, err := runIn(cwd, BranchListCmd)
	if err != nil {
		return nil, err
	}
	var entries []BranchEntry
	for _, l := range lines {
		parts := strings.SplitN(l, "|", 3)
		if len(parts) < 2 {
			continue
		}
		name := strings.TrimSpace(parts[0])
		dateStr := strings.TrimSpace(parts[1])
		author := ""
		if len(parts) == 3 {
			author = strings.TrimSpace(parts[2])
		}
		daysAgo := 0
		if t, err := time.Parse("2006-01-02", dateStr); err == nil {
			daysAgo = int(time.Since(t).Hours() / 24)
		}
		entries = append(entries, BranchEntry{Name: name, Date: dateStr, DaysAgo: daysAgo, Author: author})
	}
	return entries, nil
}

// ── COMMAND 8: CO-CHANGE COUPLING ─────────────────────────────────────────────

const CoChangeCmd = `git log --name-only --format='>>>%H' --since="1 year ago" --no-merges`

type CouplingEntry struct {
	FileA    string
	FileB    string
	Together int
	Pct      float64
}

func CoChange(cwd string) ([]CouplingEntry, error) {
	lines, err := runIn(cwd, CoChangeCmd)
	if err != nil {
		return nil, err
	}
	var commitFiles []string
	var allCommits [][]string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if strings.HasPrefix(l, ">>>") {
			if len(commitFiles) >= 2 && len(commitFiles) <= 20 {
				allCommits = append(allCommits, commitFiles)
			}
			commitFiles = nil
			continue
		}
		if l != "" {
			commitFiles = append(commitFiles, l)
		}
	}
	if len(commitFiles) >= 2 && len(commitFiles) <= 20 {
		allCommits = append(allCommits, commitFiles)
	}
	pairCount := make(map[[2]string]int)
	for _, files := range allCommits {
		sort.Strings(files)
		for i := 0; i < len(files); i++ {
			for j := i + 1; j < len(files); j++ {
				pairCount[[2]string{files[i], files[j]}]++
			}
		}
	}
	var entries []CouplingEntry
	for pair, count := range pairCount {
		if count >= 3 {
			entries = append(entries, CouplingEntry{FileA: pair[0], FileB: pair[1], Together: count})
		}
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Together > entries[j].Together })
	if len(entries) > 20 {
		entries = entries[:20]
	}
	if len(entries) > 0 {
		max := entries[0].Together
		for i := range entries {
			entries[i].Pct = float64(entries[i].Together) / float64(max) * 100
		}
	}
	return entries, nil
}

// ── COMMAND 9: FRESH FILES ────────────────────────────────────────────────────

const FreshCmd = `git log --diff-filter=A --since="90 days ago" --name-only --format='>>>%ad|%an' --date=short`

type FreshEntry struct {
	File    string
	Date    string
	Author  string
	DaysAgo int
}

func FreshFiles(cwd string) ([]FreshEntry, error) {
	lines, err := runIn(cwd, FreshCmd)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]bool)
	var entries []FreshEntry
	var curDate, curAuthor string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if strings.HasPrefix(l, ">>>") {
			meta := l[3:]
			parts := strings.SplitN(meta, "|", 2)
			curDate = parts[0]
			if len(parts) == 2 {
				curAuthor = parts[1]
			}
			continue
		}
		if l == "" || curDate == "" || seen[l] {
			continue
		}
		seen[l] = true
		daysAgo := 0
		if t, err := time.Parse("2006-01-02", curDate); err == nil {
			daysAgo = int(time.Since(t).Hours() / 24)
		}
		entries = append(entries, FreshEntry{File: l, Date: curDate, Author: curAuthor, DaysAgo: daysAgo})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].DaysAgo < entries[j].DaysAgo })
	return entries, nil
}

// ── COMMAND 10: OWNERSHIP DRIFT ──────────────────────────────────────────────

const OwnerOldCmd = `git log --format='>>>%an' --name-only --since="1 year ago" --until="6 months ago"`
const OwnerNewCmd = `git log --format='>>>%an' --name-only --since="6 months ago"`

type OwnershipEntry struct {
	File     string
	OldOwner string
	NewOwner string
	Drifted  bool
}

func parseOwnerMap(lines []string) map[string]map[string]int {
	m := make(map[string]map[string]int)
	var curAuthor string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if strings.HasPrefix(l, ">>>") {
			curAuthor = l[3:]
			continue
		}
		if l == "" || curAuthor == "" {
			continue
		}
		if m[l] == nil {
			m[l] = make(map[string]int)
		}
		m[l][curAuthor]++
	}
	return m
}

func topOwner(counts map[string]int) string {
	best, bestN := "", 0
	for name, n := range counts {
		if n > bestN {
			best, bestN = name, n
		}
	}
	return best
}

func OwnershipDrift(cwd string) ([]OwnershipEntry, error) {
	oldLines, _ := runIn(cwd, OwnerOldCmd)
	newLines, _ := runIn(cwd, OwnerNewCmd)
	oldMap := parseOwnerMap(oldLines)
	newMap := parseOwnerMap(newLines)
	var entries []OwnershipEntry
	for file, newCounts := range newMap {
		newOwner := topOwner(newCounts)
		oldOwner := ""
		if oldCounts, ok := oldMap[file]; ok {
			oldOwner = topOwner(oldCounts)
		}
		drifted := oldOwner != "" && oldOwner != newOwner
		if drifted || oldOwner == "" {
			entries = append(entries, OwnershipEntry{File: file, OldOwner: oldOwner, NewOwner: newOwner, Drifted: drifted})
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Drifted != entries[j].Drifted {
			return entries[i].Drifted
		}
		return entries[i].File < entries[j].File
	})
	if len(entries) > 30 {
		entries = entries[:30]
	}
	return entries, nil
}

// ── COMMAND 11: TEST RATIO ────────────────────────────────────────────────────

const TestRatioCmd = `git log --format='' --name-only --since="1 year ago" --no-merges`

type TestRatioEntry struct {
	File   string
	Count  int
	IsTest bool
	Pct    float64
}

var testPatterns = []string{"_test.go", "test_", ".spec.", ".test.", "_spec.", "Test.java", "Spec.java"}
var testDirs = []string{"/test/", "/tests/", "/spec/", "/__tests__/", "/testing/"}

func isTestFile(path string) bool {
	lower := strings.ToLower(path)
	for _, p := range testPatterns {
		if strings.Contains(lower, strings.ToLower(p)) {
			return true
		}
	}
	for _, d := range testDirs {
		if strings.Contains("/"+lower+"/", d) {
			return true
		}
	}
	return false
}

func TestRatio(cwd string) ([]TestRatioEntry, error) {
	lines, err := runIn(cwd, TestRatioCmd)
	if err != nil {
		return nil, err
	}
	counts := make(map[string]int)
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" {
			counts[l]++
		}
	}
	var entries []TestRatioEntry
	for file, count := range counts {
		entries = append(entries, TestRatioEntry{File: file, Count: count, IsTest: isTestFile(file)})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Count > entries[j].Count })
	if len(entries) > 30 {
		entries = entries[:30]
	}
	if len(entries) > 0 {
		max := entries[0].Count
		for i := range entries {
			entries[i].Pct = float64(entries[i].Count) / float64(max) * 100
		}
	}
	return entries, nil
}

// ── COMMAND 12: COMMIT SIZE DISTRIBUTION ─────────────────────────────────────

const CommitSizeCmd = `git log --shortstat --format='%H' --since="1 year ago" --no-merges`

type CommitSizeBucket struct {
	Label string
	Count int
	Pct   float64
}

func CommitSizes(cwd string) ([]CommitSizeBucket, error) {
	lines, err := runIn(cwd, CommitSizeCmd)
	if err != nil {
		return nil, err
	}
	buckets := []CommitSizeBucket{
		{Label: "tiny   (1–10)"},
		{Label: "small  (11–50)"},
		{Label: "medium (51–200)"},
		{Label: "large  (201–1000)"},
		{Label: "huge   (1000+)"},
	}
	limits := []int{10, 50, 200, 1000, 1 << 30}
	numRe := regexp.MustCompile(`\d+`)
	for _, l := range lines {
		if !strings.Contains(l, "insertion") && !strings.Contains(l, "deletion") {
			continue
		}
		nums := numRe.FindAllString(l, -1)
		total := 0
		for i, n := range nums {
			if i == 0 {
				continue // first number is "N files changed"
			}
			v, _ := strconv.Atoi(n)
			total += v
		}
		if total == 0 {
			continue
		}
		for i, lim := range limits {
			if total <= lim {
				buckets[i].Count++
				break
			}
		}
	}
	totalCommits := 0
	for _, b := range buckets {
		totalCommits += b.Count
	}
	if totalCommits > 0 {
		for i := range buckets {
			buckets[i].Pct = float64(buckets[i].Count) / float64(totalCommits) * 100
		}
	}
	return buckets, nil
}

// ── COMMAND 13: MERGE FREQUENCY ──────────────────────────────────────────────

const MergeFreqCmd = `git log --merges --format='%ad' --date=format:'%Y-%m'`

type MergeFreqEntry struct {
	Month string
	Count int
}

func MergeFrequency(cwd string) ([]MergeFreqEntry, error) {
	lines, err := runIn(cwd, MergeFreqCmd)
	if err != nil {
		return nil, err
	}
	monthRe := regexp.MustCompile(`^\d{4}-\d{2}$`)
	counts := make(map[string]int)
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if monthRe.MatchString(l) {
			counts[l]++
		}
	}
	var entries []MergeFreqEntry
	for month, count := range counts {
		entries = append(entries, MergeFreqEntry{Month: month, Count: count})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Month < entries[j].Month })
	return entries, nil
}

// ── COMMAND 5: FIREFIGHTING ───────────────────────────────────────────────────

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
