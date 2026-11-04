package searchtrees_test

import (
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/structures/searchtrees"
)

// TestLenAndContains checks the BST's and the AVL tree's size and membership:
// a duplicate is not counted twice, and an absent value is not found.
func TestLenAndContains(t *testing.T) {
	type set interface {
		Insert(int) bool
		Len() int
		Contains(int) bool
	}
	for name, s := range map[string]set{"bst": searchtrees.NewBST[int](), "avl": searchtrees.NewAVL[int]()} {
		for _, v := range []int{5, 3, 8, 3, 1} {
			s.Insert(v)
		}
		if s.Len() != 4 {
			t.Errorf("%s: Len() = %d, want 4 (3 was inserted twice)", name, s.Len())
		}
		for _, v := range []int{1, 3, 5, 8} {
			if !s.Contains(v) {
				t.Errorf("%s: Contains(%d) = false", name, v)
			}
		}
		if s.Contains(4) {
			t.Errorf("%s: Contains(4) = true for a value never inserted", name)
		}
	}
}

// TestDegreeAndLevels checks the B-tree keeps the degree it was made with and
// the skip list's levels grow with it but stay within one level a value.
func TestDegreeAndLevels(t *testing.T) {
	if d := searchtrees.NewBTree[int](3).Degree(); d != 3 {
		t.Errorf("Degree() = %d, want 3", d)
	}
	s := searchtrees.NewSkipList[int](1)
	for v := 0; v < 64; v++ {
		s.Insert(v)
	}
	if l := s.Levels(); l < 1 || l > 64 {
		t.Errorf("64 values in %d levels", l)
	}
}
