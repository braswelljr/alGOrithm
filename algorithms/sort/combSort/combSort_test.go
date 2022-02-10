package combSort

import "testing"

var (
	x = []int{5, 3, 2, 4, 1}
)

// Sorts an array of integers using the comb sort algorithm.
func TestCombSort(t *testing.T) {
	x = CombSort(x)
	if x[0] != 1 || x[1] != 2 || x[2] != 3 || x[3] != 4 || x[4] != 5 {
		t.Error("CombSort() failed. Got", x, "Expected 1 2 3 4 5")
		return
	}
	t.Logf("CombSort() passed. Got %v", x)
}

// Test the comb sort algorithm against the built-in sort function.
func TestCombSortLoop(t *testing.T) {
	// sort x
	x = CombSort(x)
	// Check if the x is sorted
	for i := 0; i < len(x)-1; i++ {
		if x[i] > x[i+1] {
			t.Error("The x is not sorted. Expected ", x[i], " > ", x[i+1])
		}

		if x[i] == x[i+1] {
			t.Log("MergeSort() passed. Got", x[i], " = ", x[i+1])
		}

		t.Log("MergeSort() passed. Got", x[i], " < ", x[i+1])
	}
}
