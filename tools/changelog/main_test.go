package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const history = "# Changelog\n\n## [Unreleased]\n\n## [1.8.3] - 2026-10-25\n\n### Fixed\n\n- **A corrected result.** Old behavior is explained.\n"

func invoke(t *testing.T, text string, args ...string) (int, string, string) {
	t.Helper()
	p := filepath.Join(t.TempDir(), "CHANGELOG.md")
	if err := os.WriteFile(p, []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
	args = append(args, "-file", p)
	var out, errOut bytes.Buffer
	code := run(args, &out, &errOut)
	return code, out.String(), errOut.String()
}

func TestCheck(t *testing.T) {
	for _, text := range []string{history, "\ufeff" + strings.ReplaceAll(history, "\n", "\r\n"), strings.Replace(history, "1.8.3", "1.9.0-rc.1+build.5", 1)} {
		code, out, errOut := invoke(t, text, "check")
		if code != 0 || !strings.Contains(out, "valid") || errOut != "" {
			t.Fatalf("code=%d out=%q err=%q", code, out, errOut)
		}
	}
}

func TestMalformedChangelogs(t *testing.T) {
	tests := []struct{ name, text, diagnostic string }{
		{"missing unreleased", strings.Replace(history, "## [Unreleased]", "", 1), "Unreleased"},
		{"duplicate unreleased", history + "\n## [Unreleased]\n", "duplicate"},
		{"duplicate version", history + "\n## [1.8.3] - 2026-10-26\n\n### Fixed\n- Again\n", "duplicate"},
		{"invalid version", strings.Replace(history, "1.8.3", "1.08.3", 1), "heading"},
		{"invalid prerelease", strings.Replace(history, "1.8.3", "1.9.0-rc.01", 1), "heading"},
		{"undated release", strings.Replace(history, " - 2026-10-25", "", 1), "heading"},
		{"pending release", strings.Replace(history, "2026-10-25", "Unreleased", 1), "heading"},
		{"invalid date", strings.Replace(history, "2026-10-25", "2026-02-30", 1), "date"},
		{"empty release", "## [Unreleased]\n\n## [1.8.3] - 2026-10-25\n", "empty"},
		{"empty category", history + "\n### Security\n<!-- TODO -->\n", "empty"},
		{"unknown category", strings.Replace(history, "### Fixed", "### Planned", 1), "category"},
		{"category order", history + "\n### Added\n- Added\n", "order"},
		{"category duplicate", history + "\n### Fixed\n- Again\n", "duplicate"},
		{"conflict", history + "\n<<<<<<< HEAD\n", "conflict"},
		{"conflict end", history + "\n>>>>>>> other\n", "conflict"},
		{"fenced placeholder", "## [Unreleased]\n## [1.8.3] - 2026-10-25\n```\n### Fixed\n- not an entry\n```\n", "empty"},
		{"unclosed fence", history + "\n```go\nexample\n", "unclosed"},
		{"unclosed comment", history + "\n<!-- comment\n", "unclosed"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			code, out, errOut := invoke(t, tc.text, "check")
			if code != 1 || out != "" || !strings.Contains(errOut, tc.diagnostic) {
				t.Fatalf("code=%d out=%q err=%q; want %q", code, out, errOut, tc.diagnostic)
			}
		})
	}
}

func TestExtractExactFinalizedSection(t *testing.T) {
	body := "### Fixed\n\n- **First fix.**\n\n```markdown\n## [99.0.0] - 2026-10-25\n<<<<<<< example\n```\n"
	text := "# Changelog\n## [Unreleased]\n### Added\n- Still pending\n\n## [1.8.3] - 2026-10-25\n\n" + body + "\n## [1.8.2] - 2023-01-13\n### Fixed\n- Older fix\n\n[1.8.3]: https://example.com/v1.8.3\n"
	for _, version := range []string{"1.8.3", "v1.8.3"} {
		code, out, errOut := invoke(t, text, "extract", "-version", version)
		if code != 0 || out != body || errOut != "" {
			t.Fatalf("code=%d out=%q err=%q", code, out, errOut)
		}
	}
	code, out, errOut := invoke(t, text, "check", "-version", "v1.8.3")
	if code != 0 || out == "" || errOut != "" {
		t.Fatalf("code=%d out=%q err=%q", code, out, errOut)
	}
	code, out, errOut = invoke(t, strings.ReplaceAll(history, "1.8.3", "1.8.30"), "extract", "-version", "1.8.3")
	if code != 1 || out != "" || !strings.Contains(errOut, "not found") {
		t.Fatalf("code=%d out=%q err=%q", code, out, errOut)
	}
}

func TestExtractOmitsComparisonDefinitions(t *testing.T) {
	code, out, errOut := invoke(t, history+"\n[1.8.3]: https://example.com/tag\n", "extract", "-version", "1.8.3")
	if code != 0 || strings.Contains(out, "example.com") || errOut != "" {
		t.Fatalf("code=%d out=%q err=%q", code, out, errOut)
	}
}

func TestCommentsCannotInventReleaseSections(t *testing.T) {
	text := strings.Replace(history, "### Fixed", "<!--\n## [99.0.0] - 2026-10-25\n### Other\n-->\n### Fixed", 1)
	code, _, errOut := invoke(t, text, "check")
	if code != 0 || errOut != "" {
		t.Fatalf("code=%d err=%q", code, errOut)
	}
}

func TestTildeFencesAndCommentsDoNotPopulateRelease(t *testing.T) {
	text := "## [Unreleased]\n## [1.8.3] - 2026-10-25\n<!-- - a hidden entry -->\n~~~markdown\n### Fixed\n- another example\n~~~\n"
	code, out, errOut := invoke(t, text, "extract", "-version", "1.8.3")
	if code != 1 || out != "" || !strings.Contains(errOut, "empty") {
		t.Fatalf("code=%d out=%q err=%q", code, out, errOut)
	}
}

func TestCommandErrors(t *testing.T) {
	for _, args := range [][]string{nil, {"unknown"}, {"extract"}, {"extract", "-version", "Unreleased"}, {"check", "-version", "1.08.3"}, {"check", "-version", "v"}, {"check", "extra"}, {"check", "-unknown"}} {
		code, out, errOut := invoke(t, history, args...)
		if code != 2 || out != "" || errOut == "" {
			t.Fatalf("%v: code=%d out=%q err=%q", args, code, out, errOut)
		}
	}
	var out, errOut bytes.Buffer
	if code := run([]string{"check", "-file", filepath.Join(t.TempDir(), "missing.md")}, &out, &errOut); code != 1 || out.Len() != 0 || errOut.Len() == 0 {
		t.Fatalf("code=%d out=%q err=%q", code, out.String(), errOut.String())
	}
}

type failedWriter struct{}

func (failedWriter) Write([]byte) (int, error) { return 0, errors.New("output unavailable") }

func TestFailedOutputIsNotSuccess(t *testing.T) {
	p := filepath.Join(t.TempDir(), "CHANGELOG.md")
	if err := os.WriteFile(p, []byte(history), 0600); err != nil {
		t.Fatal(err)
	}
	var errOut bytes.Buffer
	code := run([]string{"extract", "-version", "1.8.3", "-file", p}, failedWriter{}, &errOut)
	if code != 1 || !strings.Contains(errOut.String(), "output unavailable") {
		t.Fatalf("code=%d err=%q", code, errOut.String())
	}
}
