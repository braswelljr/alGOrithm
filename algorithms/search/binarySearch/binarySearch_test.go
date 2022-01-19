package binarySearch

import "testing"

func TestBinarySearch(t *testing.T) {
	t.Run("test binary search", func(t *testing.T) {
		nums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		target := 5
		got := BinarySearch(nums, target)
		want := 4
		if got != want {
			t.Errorf("got %d want %d given %v", got, want, nums)
		}
		t.Log("BinarySearch test passed with target:", target, "got:", got, "want:", want)
	})
}
