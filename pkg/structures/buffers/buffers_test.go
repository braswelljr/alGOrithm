package buffers_test

import (
	"math/rand"
	"reflect"
	"strings"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/structures/buffers"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

func word(r *rand.Rand) string {
	b := make([]byte, r.Intn(12))
	for i := range b {
		b[i] = byte('a' + r.Intn(26))
	}
	return string(b)
}

func TestGapBufferAgainstAString(t *testing.T) {
	r := rand.New(rand.NewSource(6))
	g := buffers.NewGapBuffer("hello")
	model, cursor := "hello", 5
	for step := 0; step < 4000; step++ {
		switch r.Intn(4) {
		case 0:
			cursor = min(max(cursor+r.Intn(21)-10, 0), len(model))
			g.Move(cursor)
		case 1:
			w := word(r)
			g.Insert(w)
			model = model[:cursor] + w + model[cursor:]
			cursor += len(w)
		case 2:
			n := min(r.Intn(5), cursor)
			if got := g.Backspace(n); got != n {
				t.Fatalf("Backspace deleted %d, want %d", got, n)
			}
			model = model[:cursor-n] + model[cursor:]
			cursor -= n
		default:
			n := min(r.Intn(5), len(model)-cursor)
			g.Delete(n)
			model = model[:cursor] + model[cursor+n:]
		}
		if g.String() != model || g.Cursor() != cursor || g.Len() != len(model) {
			t.Fatalf("step %d: %q at %d, want %q at %d", step, g.String(), g.Cursor(), model, cursor)
		}
	}
}

func TestRopeAgainstAString(t *testing.T) {
	r := rand.New(rand.NewSource(7))
	rope := buffers.NewRope("the quick brown fox jumps over the lazy dog")
	model := rope.String()
	for step := 0; step < 3000; step++ {
		i := r.Intn(len(model) + 1)
		if r.Intn(2) == 0 {
			w := word(r)
			rope.Insert(i, w)
			model = model[:i] + w + model[i:]
		} else {
			n := min(r.Intn(8), len(model)-i)
			rope.Delete(i, n)
			model = model[:i] + model[i+n:]
		}
		if step%500 == 0 {
			rope.Rebalance()
		}
		if rope.String() != model || rope.Len() != len(model) {
			t.Fatalf("step %d: %q, want %q", step, rope.String(), model)
		}
		if len(model) > 0 {
			j := r.Intn(len(model))
			if b, ok := rope.Index(j); !ok || b != model[j] {
				t.Fatalf("Index(%d) = %q", j, b)
			}
		}
	}
	a, b := rope.Split(len(model) / 2)
	if a.Concat(b).String() != model {
		t.Error("Split then Concat changed the text")
	}
	if rope.String() != model {
		t.Error("Split changed the rope")
	}
}

func TestGapBufferSnapshot(t *testing.T) {
	g := buffers.NewGapBuffer("a b")
	g.Move(1)
	s := g.Snapshot()
	if s.Layout != "row" || len(s.Nodes) != 8 || len(s.Edges) != 0 {
		t.Fatalf("%+v", s)
	}
	// "a", then the gap of 5 with the cursor on its first cell, then " b"
	labels, roles := []string{}, []string{}
	for _, n := range s.Nodes {
		labels = append(labels, n.Label)
		roles = append(roles, n.Role)
	}
	wantLabels := []string{"a", "_", "_", "_", "_", "_", "␠", "b"}
	wantRoles := []string{"", "cursor", "gap", "gap", "gap", "gap", "", ""}
	if !reflect.DeepEqual(labels, wantLabels) || !reflect.DeepEqual(roles, wantRoles) {
		t.Errorf("labels %q roles %q", labels, roles)
	}
	gaps := 0
	for _, l := range labels {
		if l == "_" {
			gaps++
		}
	}
	if gaps != g.Gap() {
		t.Errorf("%d gap cells, Gap() = %d", gaps, g.Gap())
	}

	full := buffers.NewGapBuffer("")
	full.Insert("12345678") // fills the 8-byte gap exactly
	if full.Gap() != 0 {
		t.Fatalf("gap %d", full.Gap())
	}
	for _, n := range full.Snapshot().Nodes {
		if n.Role != "" {
			t.Errorf("a full buffer has no cursor cell: %+v", n)
		}
	}
	if full.Backspace(-1) != 0 || full.Delete(10) != 0 || full.Backspace(100) != 8 || full.String() != "" {
		t.Errorf("clamping: %q", full.String())
	}
}

func TestRopeDepthAndSnapshot(t *testing.T) {
	cases := []struct {
		text  string
		depth int
	}{
		{"", 0},
		{"abc", 1},
		{"abcdefgh", 1},
		{"abcdefghijklmnop", 2},
		{strings.Repeat("x", 64), 4},
	}
	for _, c := range cases {
		if d := buffers.NewRope(c.text).Depth(); d != c.depth {
			t.Errorf("depth of a %d-byte rope %d, want %d", len(c.text), d, c.depth)
		}
	}

	r := buffers.NewRope("abcdefghijklmnop")
	s := r.Snapshot()
	want := trace.Snapshot{
		Layout: "tree",
		Nodes: []trace.Node{
			{ID: 0, Label: "weight 8", Role: "root", Parent: -1},
			{ID: 1, Label: `"abcdefgh"`, Parent: 0},
			{ID: 2, Label: `"ijklmnop"`, Parent: 0},
		},
		Edges: []trace.Edge{{From: 0, To: 1, Label: "left"}, {From: 0, To: 2, Label: "right"}},
	}
	if !reflect.DeepEqual(s, want) {
		t.Errorf("snapshot %+v\nwant %+v", s, want)
	}
	if e := buffers.NewRope("").Snapshot(); e.Nodes == nil || len(e.Nodes) != 0 {
		t.Errorf("an empty rope's snapshot %+v", e)
	}

	t.Run("unbalanced inserts deepen it and Rebalance restores the depth", func(t *testing.T) {
		r := buffers.NewRope("")
		for i := 0; i < 32; i++ {
			r.Insert(0, "ab")
		}
		deep := r.Depth()
		text := r.String()
		r.Rebalance()
		if r.String() != text || r.Depth() >= deep || r.Depth() != 4 {
			t.Errorf("depth %d before, %d after rebalancing 64 bytes (want 4)", deep, r.Depth())
		}
	})

	if _, ok := r.Index(-1); ok {
		t.Error("Index(-1) is out of range")
	}
	if _, ok := r.Index(r.Len()); ok {
		t.Error("Index(Len) is out of range")
	}
}
