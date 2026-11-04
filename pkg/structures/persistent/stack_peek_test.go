package persistent_test

import (
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/structures/persistent"
)

// TestPeekAndSnapshot checks that Peek sees the top without changing any
// version, an empty stack has none, and a snapshot draws one cell a value.
func TestPeekAndSnapshot(t *testing.T) {
	var empty persistent.Stack[int]
	if _, ok := empty.Peek(); ok {
		t.Error("an empty stack has a top")
	}
	one := empty.Push(1)
	two := one.Push(2)
	if v, ok := two.Peek(); !ok || v != 2 {
		t.Errorf("Peek() = %d, %v; want 2, true", v, ok)
	}
	if v, _ := one.Peek(); v != 1 {
		t.Errorf("the older version's top is %d, want 1", v)
	}
	if n := len(two.Snapshot().Nodes); n != 2 {
		t.Errorf("a two-value stack draws %d nodes", n)
	}
}
