package buffers_test

import (
	"math/rand"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/structures/buffers"
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
