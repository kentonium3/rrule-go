// Command changelog checks the root changelog and extracts finalized release notes.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"
)

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || (args[0] != "check" && args[0] != "extract") {
		fmt.Fprintln(stderr, "usage: go run ./tools/changelog {check|extract} [-file CHANGELOG.md] [-version vX.Y.Z]")
		return 2
	}
	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flags.SetOutput(stderr)
	path := flags.String("file", "CHANGELOG.md", "changelog path")
	version := flags.String("version", "", "finalized release version (optional v prefix)")
	if err := flags.Parse(args[1:]); err != nil {
		return 2
	}
	want := strings.TrimPrefix(*version, "v")
	if flags.NArg() != 0 || (args[0] == "extract" && want == "") || (*version != "" && !validVersion(want)) {
		fmt.Fprintln(stderr, "provide flags only; extract requires -version with a valid SemVer release")
		return 2
	}
	data, err := os.ReadFile(*path)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	sections, err := parse(string(data))
	if err != nil {
		fmt.Fprintf(stderr, "%s: %v\n", *path, err)
		return 1
	}
	var notes string
	if want != "" {
		found := false
		for _, s := range sections {
			if s.version == want {
				notes = strings.TrimSpace(strings.Join(s.body, "\n"))
				found = true
				break
			}
		}
		if !found {
			fmt.Fprintf(stderr, "%s: finalized release %s not found\n", *path, want)
			return 1
		}
	}
	if args[0] == "extract" {
		_, err = fmt.Fprintln(stdout, notes)
	} else {
		_, err = fmt.Fprintf(stdout, "%s: valid\n", *path)
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

var (
	versionRE   = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?(?:\+[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?$`)
	releaseRE   = regexp.MustCompile(`^## \[([^]]+)\] - ([0-9]{4}-[0-9]{2}-[0-9]{2})$`)
	referenceRE = regexp.MustCompile(`^\[[^]]+\]:\s+\S`)
)

func validVersion(v string) bool {
	m := versionRE.FindStringSubmatch(v)
	if m == nil {
		return false
	}
	for _, id := range strings.Split(m[4], ".") {
		if len(id) < 2 || id[0] != '0' {
			continue
		}
		if strings.Trim(id, "0123456789") == "" {
			return false
		}
	}
	return true
}

type section struct {
	version string // Empty for Unreleased; never eligible for extraction.
	body    []string
	entries int
}

// parse recognizes our deliberately small changelog format, not arbitrary Markdown.
// Fences and HTML comments cannot invent release boundaries or populated entries.
func parse(text string) ([]section, error) {
	categories := []string{"Breaking", "Upgrade Notes", "Added", "Changed", "Fixed", "Deprecated", "Removed", "Security"}
	var sections []section
	seen := map[string]bool{}
	category, rank, categoryEntries := "", -1, 0
	var fence byte
	fenceLength := 0
	inComment := false
	finish := func() error {
		if category != "" && categoryEntries == 0 {
			return fmt.Errorf("empty category %q", category)
		}
		if len(sections) > 0 {
			s := sections[len(sections)-1]
			if s.version != "" && s.entries == 0 {
				return fmt.Errorf("empty release %s", s.version)
			}
		}
		return nil
	}
	text = strings.TrimPrefix(strings.ReplaceAll(text, "\r\n", "\n"), "\ufeff")
	for number, raw := range strings.Split(text, "\n") {
		line := strings.TrimRight(raw, " \t")
		appendBody := func() {
			if len(sections) > 0 {
				sections[len(sections)-1].body = append(sections[len(sections)-1].body, raw)
			}
		}
		if fence != 0 {
			appendBody()
			marker, n, tail := fenceMarker(line)
			if marker == fence && n >= fenceLength && strings.TrimSpace(tail) == "" {
				fence = 0
			}
			continue
		}
		visible := stripComments(line, &inComment)
		if marker, n, _ := fenceMarker(visible); n >= 3 {
			fence, fenceLength = marker, n
			appendBody()
			continue
		}
		trim := strings.TrimSpace(visible)
		if strings.HasPrefix(trim, "<<<<<<<") || strings.HasPrefix(trim, "|||||||") || trim == "=======" || strings.HasPrefix(trim, ">>>>>>>") {
			return nil, fmt.Errorf("line %d: unresolved conflict marker", number+1)
		}
		if strings.HasPrefix(visible, "## ") {
			label, version := "Unreleased", ""
			if visible != "## [Unreleased]" {
				m := releaseRE.FindStringSubmatch(visible)
				if m == nil || !validVersion(m[1]) {
					return nil, fmt.Errorf("line %d: invalid release heading; use ## [X.Y.Z] - YYYY-MM-DD", number+1)
				}
				if _, err := time.Parse("2006-01-02", m[2]); err != nil {
					return nil, fmt.Errorf("line %d: invalid release date", number+1)
				}
				label, version = m[1], m[1]
			}
			if seen[label] {
				return nil, fmt.Errorf("line %d: duplicate section %s", number+1, label)
			}
			if len(sections) == 0 && label != "Unreleased" {
				return nil, fmt.Errorf("line %d: first section must be ## [Unreleased]", number+1)
			}
			if err := finish(); err != nil {
				return nil, err
			}
			seen[label] = true
			sections = append(sections, section{version: version})
			category, rank, categoryEntries = "", -1, 0
			continue
		}
		if strings.HasPrefix(visible, "### ") {
			if len(sections) == 0 {
				return nil, fmt.Errorf("line %d: category before a release section", number+1)
			}
			name, nextRank := strings.TrimPrefix(visible, "### "), -1
			for i, c := range categories {
				if name == c {
					nextRank = i
					break
				}
			}
			if nextRank < 0 {
				return nil, fmt.Errorf("line %d: unknown category %q", number+1, name)
			}
			if nextRank == rank {
				return nil, fmt.Errorf("line %d: duplicate category %q", number+1, name)
			}
			if nextRank < rank {
				return nil, fmt.Errorf("line %d: category %q is out of order or duplicated", number+1, name)
			}
			if category != "" && categoryEntries == 0 {
				return nil, fmt.Errorf("line %d: empty category %q", number+1, category)
			}
			category, rank, categoryEntries = name, nextRank, 0
		} else if strings.HasPrefix(visible, "##") {
			return nil, fmt.Errorf("line %d: invalid heading", number+1)
		} else if strings.HasPrefix(visible, "- ") && strings.TrimSpace(visible[2:]) != "" {
			if len(sections) == 0 || category == "" {
				return nil, fmt.Errorf("line %d: entry needs a category", number+1)
			}
			sections[len(sections)-1].entries++
			categoryEntries++
		}
		// Comparison definitions belong to the document, not the final release's notes.
		if !referenceRE.MatchString(visible) {
			appendBody()
		}
	}
	if fence != 0 || inComment {
		return nil, fmt.Errorf("unclosed code fence or HTML comment")
	}
	if len(sections) == 0 {
		return nil, fmt.Errorf("missing ## [Unreleased] section")
	}
	if err := finish(); err != nil {
		return nil, err
	}
	return sections, nil
}

func fenceMarker(line string) (byte, int, string) {
	s := strings.TrimLeft(line, " ")
	if len(line)-len(s) > 3 || len(s) == 0 || (s[0] != '`' && s[0] != '~') {
		return 0, 0, ""
	}
	i := 0
	for i < len(s) && s[i] == s[0] {
		i++
	}
	if i < 3 {
		return 0, 0, ""
	}
	return s[0], i, s[i:]
}

func stripComments(line string, open *bool) string {
	var visible strings.Builder
	for line != "" {
		if *open {
			i := strings.Index(line, "-->")
			if i < 0 {
				break
			}
			line, *open = line[i+3:], false
		} else {
			i := strings.Index(line, "<!--")
			if i < 0 {
				visible.WriteString(line)
				break
			}
			visible.WriteString(line[:i])
			line, *open = line[i+4:], true
		}
	}
	return visible.String()
}
