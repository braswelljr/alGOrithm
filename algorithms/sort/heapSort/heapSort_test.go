package heapSort

import "testing"

func TestHeapSort(t *testing.T) {
	arr := []int{2, 4, 3, 5, 1, 6}
	arr = HeapSort(arr)
	if arr[0] != 1 || arr[1] != 2 || arr[2] != 3 || arr[3] != 4 || arr[4] != 5 {
		t.Error("HeapSort() failed. Got", arr, "Expected 1 2 3 4 5")
		return
	}

	t.Log("HeapSort() passed.")
}

func TestHeapSortLoop(t *testing.T) {
	slice := []int{2, 4, 3, 5, 4, 1, 6}
	slice = HeapSort(slice)
	for i := 0; i < len(slice)-1; i++ {
		if slice[i] > slice[i+1] {
			t.Error("HeapSort() failed. Got", slice[i], " < ", slice[i+1], ". Expected ", slice[i], " > ", slice[i+1])
		} else if slice[i] == slice[i+1] {
			t.Log("HeapSort() passed. Got", slice[i], " = ", slice[i+1])
		} else {
			t.Log("HeapSort() passed. Got", slice[i], " < ", slice[i+1])
		}
	}

	t.Log("Slice : ", slice)
}
