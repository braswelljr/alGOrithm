package interpolationSearch

import "testing"

// TestInterpolationSearch
func TestInterpolationSearch(t *testing.T) {
	var arr = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	var result = InterpolationSearch(arr, 5)
	if result != 4 {
		t.Error("InterpolationSearch test failed on [1, 2, 3, 4, 5, 6, 7, 8, 9, 10], expected ", 4, ", got ", result)
	}
}
