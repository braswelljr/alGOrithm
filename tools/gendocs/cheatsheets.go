package main

import (
	"fmt"
	"strings"

	"github.com/braswelljr/alGOrithm/internal/core"
)

// Cheatsheets is docs/cheatsheets.md: every category on one table, each
// entry's time, space and key idea (its notes, or its summary when it has
// none), the printable counterpart of the site's /cheatsheets page.
func Cheatsheets() string {
	var b strings.Builder
	b.WriteString("# Cheat sheets\n\nEvery category on one table: what each entry costs and the idea that makes it work. Generated from the registry (`go run ./tools/gendocs`); the site's /cheatsheets page is the same, printable. `algorithm info ID` gives the TL;DR, when to use it and the pitfall.\n")
	topic := ""
	for _, c := range core.Categories {
		entries := core.InCategory(c.ID)
		if len(entries) == 0 {
			continue
		}
		if c.Topic != topic {
			topic = c.Topic
			fmt.Fprintf(&b, "\n## %s\n", topic)
		}
		fmt.Fprintf(&b, "\n### %s\n\n%s\n\n| Entry | Time | Space | Key idea |\n|-------|------|-------|----------|\n", c.Name, c.Summary)
		for _, e := range entries {
			idea := e.Summary
			if e.Notes != nil {
				idea = e.Notes.KeyIdea
			}
			fmt.Fprintf(&b, "| %s | %s | %s | %s |\n", cell(e.Name), cell(e.Time), cell(e.Space), cell(idea))
		}
	}
	return b.String()
}
