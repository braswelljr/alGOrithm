package quickSort

import (
  "math/rand"
  "testing"
)

func TestQuickSort(t *testing.T) {
  arr := []int{2, 4, 3, 5, 1}
  //arr = QuickSort(arr)
  if arr[0] != 1 || arr[1] != 2 || arr[2] != 3 || arr[3] != 4 || arr[4] != 5 {
    t.Error("QuickSort() failed. Got", arr, "Expected 1 2 3 4 5")
  }

  t.Log("QuickSort() passed.")
}

func TestQuickSortLoop(t *testing.T) {
  slice := []int{2, 4, 3, 5, 4, 1, 6}
  slice = QuickSort(slice)
  //QuickSort(slice)
  for i := 0; i < len(slice)-1; i++ {
    if slice[i] > slice[i+1] {
      t.Error("QuickSort() failed. Got", slice[i], " < ", slice[i+1], ". Expected ", slice[i], " > ", slice[i+1])
    } else if slice[i] == slice[i+1] {
      t.Log("QuickSort() passed. Got", slice[i], " = ", slice[i+1])
    } else {
      t.Log("QuickSort() passed. Got", slice[i], " < ", slice[i+1])
    }
  }

  t.Log("Slice : ", slice)
}

// returns a slice of random numbers with range n
func perm(n int) (out []int) {
  for _, v := range rand.Perm(n) {
    out = append(out, v)
  }
  return
}

func BenchmarkQuickSort(b *testing.B) {
  // Benchmark a 100
  benchmarkQuickSort(100, b)
}

func benchmarkQuickSort(x int, b *testing.B) {
  for i := 0; i < b.N; i++ {
    b.StopTimer()
    values := perm(x)
    b.StartTimer()
    values = QuickSort(values)
  }
}
