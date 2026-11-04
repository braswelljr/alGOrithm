package main

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/braswelljr/alGOrithm/internal/core"
)

// Learn is docs/learn.md: every learning path with its steps, what to notice
// at each, and the command that runs it.
func Learn() string {
	var b strings.Builder
	b.WriteString("# Learning paths\n\nOrdered walks through the catalog, generated from `core.Paths()` (`go run ./tools/gendocs`). Each step is an entry to run, what to notice while it runs, and the command that shows it; `algorithm learn ID` prints the same in the terminal, and the site's Learn page opens each step in the visualizer.\n\n")
	b.WriteString("| Path | Level | For | Steps | First |\n|------|-------|-----|-------|-------|\n")
	for _, p := range core.Paths() {
		fmt.Fprintf(&b, "| [%s](#%s) | %s | %s | %d | %s |\n", p.Title, anchor(p.Title), p.Level, p.Audience, len(p.Steps), links(p.Prerequisites))
	}
	for _, p := range core.Paths() {
		v := p.View()
		fmt.Fprintf(&b, "\n## %s\n\n%s\n\n- Level: %s, for %s.\n", p.Title, p.Summary, p.Level, p.Audience)
		if len(p.Prerequisites) > 0 {
			fmt.Fprintf(&b, "- Walk first: %s.\n", links(p.Prerequisites))
		}
		if p.Guide != "" {
			fmt.Fprintf(&b, "- Read: [%s](../%s).\n", p.Guide, p.Guide)
		}
		fmt.Fprintf(&b, "- In the terminal: `algorithm learn %s`.\n\n", p.ID)
		for i, st := range v.Steps {
			fmt.Fprintf(&b, "%d. **%s** (`%s`): %s\n\n   ```bash\n   %s\n   ```\n\n", i+1, st.Name, st.Entry, st.Note, strings.ReplaceAll(st.Command, "\n", "\n   "))
		}
	}
	return strings.TrimRight(b.String(), "\n") + "\n"
}

// links writes path IDs as links to their sections, or "none".
func links(ids []string) string {
	if len(ids) == 0 {
		return "none"
	}
	out := make([]string, len(ids))
	for i, id := range ids {
		title := id
		if p, ok := core.PathByID(id); ok {
			title = p.Title
		}
		out[i] = fmt.Sprintf("[%s](#%s)", title, anchor(title))
	}
	return strings.Join(out, ", ")
}

// anchor is the id a Markdown renderer gives a heading (GitHub's rule): lower
// case, letters, digits, spaces and hyphens kept, spaces turned to hyphens.
// Links point at the headings themselves, so the page needs no HTML anchors.
func anchor(heading string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(heading) {
		switch {
		case r == ' ':
			b.WriteByte('-')
		case r == '-' || r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
		}
	}
	return b.String()
}
