package ranges_test

import (
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/structures/ranges"
)

// TestDisjointSparseLen checks that the table counts the values it was built on.
func TestDisjointSparseLen(t *testing.T) {
	sum := func(a, b int) int { return a + b }
	if n := ranges.NewDisjointSparseTable([]int{4, 1, 7, 2, 9}, sum).Len(); n != 5 {
		t.Errorf("Len() = %d, want 5", n)
	}
	if n := ranges.NewDisjointSparseTable(nil, sum).Len(); n != 0 {
		t.Errorf("an empty table has Len() %d", n)
	}
}
