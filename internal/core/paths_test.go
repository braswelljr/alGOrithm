package core

import (
	"strings"
	"testing"
)

// TestPathsAreConsistent checks every path: a unique ID, a known level, real
// prerequisites with no cycle, and steps that name real entries and run.
func TestPathsAreConsistent(t *testing.T) {
	paths := Paths()
	byID := map[string]Path{}
	for _, p := range paths {
		if p.ID == "" || p.Title == "" || p.Summary == "" || p.Audience == "" || len(p.Steps) == 0 {
			t.Errorf("%q: missing id, title, summary, audience or steps", p.ID)
		}
		if _, dup := byID[p.ID]; dup {
			t.Errorf("%s: duplicate path id", p.ID)
		}
		byID[p.ID] = p
		if p.Level != Beginner && p.Level != Intermediate && p.Level != Advanced {
			t.Errorf("%s: level %q", p.ID, p.Level)
		}
		if strings.Contains(p.Summary, "—") {
			t.Errorf("%s: em dash in the summary", p.ID)
		}
	}
	for _, p := range paths {
		for _, pre := range p.Prerequisites {
			if _, ok := byID[pre]; !ok {
				t.Errorf("%s: prerequisite %q is not a path", p.ID, pre)
			}
		}
		for i, s := range p.Steps {
			e, ok := Lookup(s.Entry)
			if !ok {
				t.Errorf("%s step %d: no entry %q", p.ID, i+1, s.Entry)
				continue
			}
			if s.Note == "" || strings.Contains(s.Note, "—") {
				t.Errorf("%s step %d: the note is empty or has an em dash", p.ID, i+1)
			}
			in := *e.Example
			if s.Input != nil {
				in = *s.Input
			}
			if _, err := Execute(e.ID, in, -1); err != nil {
				t.Errorf("%s step %d (%s): %v", p.ID, i+1, e.ID, err)
			}
		}
	}
	// No path may require itself, however indirectly.
	state := map[string]int{} // 0 unseen, 1 on the stack, 2 done
	var visit func(id string, trail []string)
	visit = func(id string, trail []string) {
		switch state[id] {
		case 1:
			t.Errorf("prerequisite cycle: %s -> %s", strings.Join(trail, " -> "), id)
			return
		case 2:
			return
		}
		state[id] = 1
		for _, pre := range byID[id].Prerequisites {
			visit(pre, append(trail, id))
		}
		state[id] = 2
	}
	for _, p := range paths {
		visit(p.ID, nil)
	}
}

func TestPathByID(t *testing.T) {
	if p, ok := PathByID("sorting-from-zero"); !ok || p.Title == "" {
		t.Fatal("sorting-from-zero not found")
	}
	if _, ok := PathByID("nope"); ok {
		t.Fatal("PathByID(nope) found something")
	}
}
