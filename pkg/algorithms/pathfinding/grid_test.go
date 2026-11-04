package pathfinding

import (
	"strings"
	"testing"
)

func TestParseAndString(t *testing.T) {
	g, err := Parse("S.#\n.5.\n#.E\n")
	if err != nil {
		t.Fatal(err)
	}
	if g.Rows != 3 || g.Cols != 3 || g.Start != 0 || g.End != 8 || g.Cost[2] != 0 || g.Cost[4] != 5 {
		t.Fatalf("Parse = %+v", g)
	}
	if got := g.String(); got != "S.#\n.5.\n#.E\n" {
		t.Errorf("String = %q", got)
	}
	if got := g.Neighbors(4); len(got) != 4 {
		t.Errorf("Neighbors(4) = %v", got)
	}
	if got := g.Neighbors(0); len(got) != 2 {
		t.Errorf("Neighbors(0) = %v (the wall at 2 is not a neighbor of 1, and 0 has 1 and 3)", got)
	}
	if g.Distance(0, 8) != 4 {
		t.Errorf("Distance = %d", g.Distance(0, 8))
	}
}

func TestParseDefaultsAndErrors(t *testing.T) {
	g, err := Parse("..\n..")
	if err != nil || g.Start != 0 || g.End != 3 {
		t.Errorf("no S/E: %+v, %v", g, err)
	}
	for _, bad := range []string{"", "##\n##", "S.x"} {
		if _, err := Parse(bad); err == nil {
			t.Errorf("Parse(%q) succeeded", bad)
		}
	}
	if _, err := Parse(strings.Repeat(strings.Repeat(".", 200)+"\n", 200)); err == nil {
		t.Error("oversized grid accepted")
	}
}

func TestGenerate(t *testing.T) {
	for _, k := range Kinds {
		g, err := Generate(k, 21, 31, 7)
		if err != nil {
			t.Fatalf("Generate(%s): %v", k, err)
		}
		h, _ := Generate(k, 21, 31, 7)
		if g.String() != h.String() {
			t.Errorf("Generate(%s) not reproducible", k)
		}
		if g.Cost[g.Start] == 0 || g.Cost[g.End] == 0 {
			t.Errorf("%s: start or end is a wall", k)
		}
	}
	m, _ := Generate(Maze, 20, 20, 1) // rounded down to 19x19
	if m.Rows != 19 || m.Cols != 19 {
		t.Errorf("maze size %dx%d", m.Rows, m.Cols)
	}
	// A perfect maze is connected: every room reachable from the start.
	seen := map[int]bool{m.Start: true}
	queue := []int{m.Start}
	for len(queue) > 0 {
		c := queue[0]
		queue = queue[1:]
		for _, n := range m.Neighbors(c) {
			if !seen[n] {
				seen[n] = true
				queue = append(queue, n)
			}
		}
	}
	if !seen[m.End] {
		t.Error("maze end unreachable")
	}
	if _, err := Generate("spiral", 5, 5, 1); err == nil {
		t.Error("unknown kind accepted")
	}
}

func TestPathTo(t *testing.T) {
	prev := []int{-1, 0, 1, -1}
	if got := PathTo(prev, 0, 2); len(got) != 3 || got[0] != 0 || got[2] != 2 {
		t.Errorf("PathTo = %v", got)
	}
	if got := PathTo(prev, 0, 3); got != nil {
		t.Errorf("unreached = %v", got)
	}
	if got := PathTo(prev, 0, 0); len(got) != 1 {
		t.Errorf("start = %v", got)
	}
}
