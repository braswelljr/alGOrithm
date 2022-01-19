package jumpSearch

import "testing"

func TestJumpSearch(t *testing.T) {
	l := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	if JumpSearch(l, 10) != 9 {
		t.Error("Expected 9, got ", JumpSearch(l, 10))
		return
	}
	t.Log("JumpSearch() Passed! with ", JumpSearch(l, 10), " as result")
}
