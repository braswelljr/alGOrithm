package mergeSort

import "testing"

var (
	slice = []int{5, 4, 3, 2, 1}
)

func TestMergeSort(t *testing.T) {
	slice = MergeSort(slice)

	t.Log(slice)
	if slice[0] != 1 || slice[1] != 2 || slice[2] != 3 || slice[3] != 4 || slice[4] != 5 {
		t.Error("MergeSort() failed. Got", slice, ". Expected 1 2 3 4 5")
		return
	}

	t.Logf("BubbleSort() passed. Got %v", slice)

}

func TestMergeSortLoop(t *testing.T) {
	// sort slice
	slice = MergeSort(slice)
	// Check if the slice is sorted
	for i := 0; i < len(slice)-1; i++ {
		if slice[i] > slice[i+1] {
			t.Error("The slice is not sorted. Expected ", slice[i], " < ", slice[i+1])
		}

		if slice[i] == slice[i+1] {
			t.Log("MergeSort() passed. Got", slice[i], " = ", slice[i+1])
		}

		t.Log("MergeSort() passed. Got", slice[i], " < ", slice[i+1])
	}
}
