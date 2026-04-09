package git

import (
	"reflect"
	"testing"
)

// ── splitLines ────────────────────────────────────────────────────────────────

func TestSplitLines(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"", nil},
		{"hello", []string{"hello"}},
		{"a\nb\nc", []string{"a", "b", "c"}},
		{"a\n\nb", []string{"a", "b"}},   // empty lines are filtered
		{"a\n", []string{"a"}},            // trailing newline stripped
		{"\n\n", nil},                     // all-empty input
		{"  spaces  ", []string{"  spaces  "}}, // interior spaces preserved
	}
	for _, tc := range tests {
		got := splitLines(tc.input)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("splitLines(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

// ── RepoName ──────────────────────────────────────────────────────────────────

func TestRepoName(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/home/user/my-project", "my-project"},
		{"/home/user/my-project/", "my-project"}, // trailing slash stripped
		{"/some/deeply/nested/repo", "repo"},
		{".", "."},
	}
	for _, tc := range tests {
		got := RepoName(tc.path)
		if got != tc.want {
			t.Errorf("RepoName(%q) = %q, want %q", tc.path, got, tc.want)
		}
	}
}

// ── isTestFile ────────────────────────────────────────────────────────────────

func TestIsTestFile(t *testing.T) {
	tests := []struct {
		path   string
		isTest bool
	}{
		// _test.go suffix (Go)
		{"pkg/foo_test.go", true},
		{"internal/bar_test.go", true},

		// test_ prefix pattern
		{"test_something.py", true},
		{"src/test_helper.rb", true},

		// .spec. pattern
		{"src/app.spec.js", true},
		{"components/Button.spec.tsx", true},

		// .test. pattern
		{"src/app.test.ts", true},
		{"utils/format.test.js", true},

		// _spec. pattern
		{"models/user_spec.rb", true},

		// Test.java / Spec.java suffixes (case-insensitive)
		{"FooTest.java", true},
		{"BarSpec.java", true},
		{"src/com/example/ServiceTest.java", true},

		// /test/ directory
		{"test/foo.go", true},
		{"project/test/helpers.go", true},

		// /tests/ directory
		{"tests/integration.go", true},

		// /spec/ directory
		{"spec/models/user.rb", true},

		// /__tests__/ directory
		{"src/__tests__/App.js", true},

		// /testing/ directory
		{"internal/testing/mock.go", true},

		// NOT test files
		{"src/main.go", false},
		{"pkg/handler.go", false},
		{"internal/ui/model.go", false},
		{"cmd/server/main.go", false},

		// Tricky non-test names
		{"testing.go", false},         // no /testing/ directory component
		{"protest.go", false},         // "test" substring but not a test pattern
		{"attestation.go", false},     // same
		{"specfile.go", false},        // "_spec." not matched by "specfile"
		{"contest/main.go", false},    // directory "contest" ≠ "/test/"
	}
	for _, tc := range tests {
		got := isTestFile(tc.path)
		if got != tc.isTest {
			t.Errorf("isTestFile(%q) = %v, want %v", tc.path, got, tc.isTest)
		}
	}
}

// ── parseOwnerMap ─────────────────────────────────────────────────────────────

func TestParseOwnerMap(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		got := parseOwnerMap(nil)
		if len(got) != 0 {
			t.Errorf("expected empty map, got %v", got)
		}
	})

	t.Run("single author single file", func(t *testing.T) {
		lines := []string{">>>Alice", "src/foo.go"}
		got := parseOwnerMap(lines)
		if got["src/foo.go"]["Alice"] != 1 {
			t.Errorf("expected Alice:1 for src/foo.go, got %v", got["src/foo.go"])
		}
	})

	t.Run("same file multiple commits by same author", func(t *testing.T) {
		lines := []string{
			">>>Alice", "src/foo.go",
			">>>Alice", "src/foo.go",
			">>>Alice", "src/foo.go",
		}
		got := parseOwnerMap(lines)
		if got["src/foo.go"]["Alice"] != 3 {
			t.Errorf("expected Alice:3, got %v", got["src/foo.go"]["Alice"])
		}
	})

	t.Run("multiple authors same file", func(t *testing.T) {
		lines := []string{
			">>>Alice", "src/foo.go",
			">>>Bob", "src/foo.go",
			">>>Alice", "src/foo.go",
		}
		got := parseOwnerMap(lines)
		if got["src/foo.go"]["Alice"] != 2 {
			t.Errorf("expected Alice:2, got %d", got["src/foo.go"]["Alice"])
		}
		if got["src/foo.go"]["Bob"] != 1 {
			t.Errorf("expected Bob:1, got %d", got["src/foo.go"]["Bob"])
		}
	})

	t.Run("multiple files different authors", func(t *testing.T) {
		lines := []string{
			">>>Alice", "src/a.go", "src/b.go",
			">>>Bob", "src/c.go",
		}
		got := parseOwnerMap(lines)
		if got["src/a.go"]["Alice"] != 1 {
			t.Errorf("expected src/a.go Alice:1")
		}
		if got["src/b.go"]["Alice"] != 1 {
			t.Errorf("expected src/b.go Alice:1")
		}
		if got["src/c.go"]["Bob"] != 1 {
			t.Errorf("expected src/c.go Bob:1")
		}
	})

	t.Run("empty lines and no author prefix are skipped", func(t *testing.T) {
		lines := []string{"", "src/orphan.go", ">>>Alice", "", "src/foo.go"}
		got := parseOwnerMap(lines)
		// "src/orphan.go" appears before any author marker — must be ignored
		if _, ok := got["src/orphan.go"]; ok {
			t.Error("orphan file (before any author) should not appear in map")
		}
		if got["src/foo.go"]["Alice"] != 1 {
			t.Errorf("expected src/foo.go Alice:1")
		}
	})
}

// ── topOwner ──────────────────────────────────────────────────────────────────

func TestTopOwner(t *testing.T) {
	t.Run("empty map returns empty string", func(t *testing.T) {
		got := topOwner(map[string]int{})
		if got != "" {
			t.Errorf("expected empty string, got %q", got)
		}
	})

	t.Run("single contributor", func(t *testing.T) {
		got := topOwner(map[string]int{"Alice": 5})
		if got != "Alice" {
			t.Errorf("expected Alice, got %q", got)
		}
	})

	t.Run("clear winner", func(t *testing.T) {
		got := topOwner(map[string]int{"Alice": 10, "Bob": 3, "Carol": 1})
		if got != "Alice" {
			t.Errorf("expected Alice, got %q", got)
		}
	})
}
