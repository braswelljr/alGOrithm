package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRealWorldNamesRealEntries fails when realworld.go names an entry that
// does not exist, leaves a field empty, or points at a missing example.
func TestRealWorldNamesRealEntries(t *testing.T) {
	ids := map[string]bool{}
	for _, e := range Entries() {
		ids[e.ID] = true
	}
	examples := 0
	for id, uses := range realWorld {
		if !ids[id] {
			t.Errorf("realworld.go names %q, which is not an entry", id)
		}
		for _, u := range uses {
			if u.Title == "" || u.Where == "" || !strings.HasSuffix(u.Where, ".") {
				t.Errorf("%s: a use needs a title and a full sentence: %+v", id, u)
			}
			if strings.Contains(u.Where, "—") {
				t.Errorf("%s: no em dashes", id)
			}
			if u.Example != "" {
				examples++
				if _, err := os.Stat(filepath.Join("..", "..", u.Example, "main.go")); err != nil {
					t.Errorf("%s: example %s has no main.go", id, u.Example)
				}
			}
		}
	}
	if len(realWorld) < 150 {
		t.Errorf("only %d entries have real-world uses; keep at least 150", len(realWorld))
	}
	t.Logf("%d entries with real-world uses, %d linking an example", len(realWorld), examples)
}
