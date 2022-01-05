package heapSort

import "testing"

func TestHeapSort(t *testing.T) {
  arr := []int{2, 4, 3, 5, 1}
  arr = HeapSort(arr)
  if arr[0] != 1 || arr[1] != 2 || arr[2] != 3 || arr[3] != 4 || arr[4] != 5 {
    t.Error("HeapSort() failed. Got", arr, "Expected 1 2 3 4 5")
    return
  }

  t.Log("HeapSort() passed.")
}
