package insertionSort

import (
  "math/rand"
  "testing"
)

func TestInsertionSort(t *testing.T) {
  arr := []int{2, 4, 3, 5, 1}
  arr = InsertionSort(arr)
  if arr[0] != 1 || arr[1] != 2 || arr[2] != 3 || arr[3] != 4 || arr[4] != 5 {
    t.Error("InsertionSort() failed. Got", arr, "Expected 1 2 3 4 5")
  }

  t.Log("InsertionSort() passed.")
}

func TestInsertionSortLoop(t *testing.T) {
  slice := []int{2, 4, 3, 5, 4, 1, 6}
  slice = InsertionSort(slice)
  //InsertionSort(slice)
  for i := 0; i < len(slice)-1; i++ {
    if slice[i] > slice[i+1] {
      t.Error("InsertionSort() failed. Got", slice[i], " < ", slice[i+1], ". Expected ", slice[i], " > ", slice[i+1])
    } else if slice[i] == slice[i+1] {
      t.Log("InsertionSort() passed. Got", slice[i], " = ", slice[i+1])
    } else {
      t.Log("InsertionSort() passed. Got", slice[i], " < ", slice[i+1])
    }
  }
  //SelectionSort(slice)

  t.Log("Slice : ", slice)
}

// returns a slice of random numbers with range n
func perm(n int) (out []int) {
  for _, v := range rand.Perm(n) {
    out = append(out, v)
  }
  return
}

func BenchmarkInsertionSort(b *testing.B) {
  // Benchmark a 100
  benchmarkInsertionSort(100, b)
}

func benchmarkInsertionSort(x int, b *testing.B) {
  for i := 0; i < b.N; i++ {
    b.StopTimer()
    values := perm(x)
    b.StartTimer()
    values = InsertionSort(values)
  }
}
