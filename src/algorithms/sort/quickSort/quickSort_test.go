package quickSort

import "testing"

func TestQuickSort(t *testing.T) {
  arr := []int{2, 4, 3, 5, 1}
  arr = QuickSort(arr)
  if arr[0] != 1 || arr[1] != 2 || arr[2] != 3 || arr[3] != 4 || arr[4] != 5 {
    t.Error("QuickSort() failed. Got", arr, "Expected 1 2 3 4 5")
  }

  t.Log("QuickSort() passed.")
}
