package bubbleSort

import (
  "testing"
)

var (
  slice = []int{5, 4, 3, 2, 1}
  z     = []int{4, 3, 3, 5, 1044}
)

func TestBubbleSort(t *testing.T) {
  slice = BubbleSort(slice)
  if slice[0] != 1 || slice[1] != 2 || slice[2] != 3 || slice[3] != 4 || slice[4] != 5 {
    t.Error("BubbleSort() failed. Got", slice, ". Expected 1 2 3 4 5")
    return
  }

  t.Logf("BubbleSort() passed. Got %v", slice)
}

func TestBubbleSortLoop(t *testing.T) {
  slice = BubbleSort(slice)
  for i := 0; i < len(slice)-1; i++ {
    if slice[i] > slice[i+1] {
      t.Error("BubbleSort() failed. Got", slice[i], " < ", slice[i+1], ". Expected ", slice[i], " > ", slice[i+1])
      return
    }
    if slice[i] == slice[i+1] {
      t.Log("BubbleSort() passed. Got", slice[i], " = ", slice[i+1])
      return
    }
    t.Log("BubbleSort() passed. Got", slice[i], " < ", slice[i+1])
  }
}
