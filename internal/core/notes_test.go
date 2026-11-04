package core

import (
	"strings"
	"testing"
)

// TestNotesNameRealEntries keeps every note attached to a real entry, written
// as sentences, and enough of them to cover the core of the catalog.
func TestNotesNameRealEntries(t *testing.T) {
	for id, n := range notes {
		if _, ok := Lookup(id); !ok {
			t.Errorf("notes[%q]: no such entry", id)
		}
		for name, text := range map[string]string{"TLDR": n.TLDR, "KeyIdea": n.KeyIdea, "WhenToUse": n.WhenToUse, "Pitfall": n.Pitfall} {
			if !strings.HasSuffix(text, ".") && !strings.HasSuffix(text, "?") {
				t.Errorf("notes[%q].%s should be a sentence ending in a period: %q", id, name, text)
			}
			if strings.ContainsRune(text, '—') {
				t.Errorf("notes[%q].%s has an em dash", id, name)
			}
		}
	}
	if len(notes) < 120 {
		t.Errorf("%d notes; want at least 120", len(notes))
	}
	levels := map[string]int{}
	for id := range notes {
		e, _ := Lookup(id)
		levels[e.Level]++
	}
	for _, l := range []string{Beginner, Intermediate, Advanced} {
		if levels[l] == 0 {
			t.Errorf("no notes for %s entries", l)
		}
	}
	if e, _ := Lookup("merge-sort"); e.Notes == nil || e.Notes.TLDR == "" {
		t.Error("Entries() does not copy Notes")
	}
}
