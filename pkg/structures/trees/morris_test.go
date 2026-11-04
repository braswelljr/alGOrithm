package trees_test

import (
	"slices"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/structures/trees"
)

// TestMorrisAndString checks Morris's threaded inorder against the recursive
// one, that it leaves the tree as it found it, and that String is the
// serialized form Parse reads back.
func TestMorrisAndString(t *testing.T) {
	tree, err := trees.BuildFromPreIn([]int{4, 2, 1, 3, 6, 5, 7}, []int{1, 2, 3, 4, 5, 6, 7})
	if err != nil {
		t.Fatal(err)
	}
	before := tree.String()
	if got := trees.Morris(tree); !slices.Equal(got, trees.Inorder(tree)) {
		t.Errorf("Morris = %v, Inorder = %v", got, trees.Inorder(tree))
	}
	if tree.String() != before {
		t.Error("Morris left the tree threaded")
	}
	back, err := trees.Parse(before)
	if err != nil || back.String() != before {
		t.Errorf("String() %q does not read back (%v)", before, err)
	}
	// An empty tree is a Tree with no root.
	if got := trees.Morris(&trees.Tree{}); len(got) != 0 {
		t.Errorf("Morris of an empty tree = %v", got)
	}
}
