package list

import (
	"github.com/braswelljr/alGOrithm/src/utils"
	"testing"
)

// TestList tests the list data structure.
func TestList(t *testing.T) {
	x, _ := utils.TypeToInterface([][]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}})
	l := New()
	l.Push(1)
	l.Push(x)
	l.Pop()

	if l.Size() != 2 {
		t.Errorf("Expected length of 2, got %d", l.Size())
	}
}
