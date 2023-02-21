package selectionSort

import (
	"math/rand"
	"testing"
)

func TestSelectionSort(t *testing.T) {
	arr := []int{2, 4, 3, 5, 1}
	arr = SelectionSort(arr)
	if arr[0] != 1 || arr[1] != 2 || arr[2] != 3 || arr[3] != 4 || arr[4] != 5 {
		t.Error("SelectionSort() failed. Got", arr, "Expected 1 2 3 4 5")
	}

	t.Log("SelectionSort() passed.")
}

func TestSelectionSortLoop(t *testing.T) {
	slice := []int{2, 4, 3, 5, 4, 1, 6}
	slice = SelectionSort(slice)
	for i := 0; i < len(slice)-1; i++ {
		if slice[i] > slice[i+1] {
			t.Error("SelectionSort() failed. Got", slice[i], " < ", slice[i+1], ". Expected ", slice[i], " > ", slice[i+1])
		} else if slice[i] == slice[i+1] {
			t.Log("SelectionSort() passed. Got", slice[i], " = ", slice[i+1])
		} else {
			t.Log("SelectionSort() passed. Got", slice[i], " < ", slice[i+1])
		}
	}

	t.Log("Slice : ", slice)
}

// perm - return a slice of random numbers with range n
//
//	@param {int} n - range of the slice
//	@return {[]int} - slice of random numbers
func perm(n int) (out []int) {
	// rand.Perm returns a slice of n random numbers
	return append(out, rand.Perm(n)...)
}

// BenchmarkRadixSort - run the RadixSort function b.N times
//
//	@param {testing.B} b - testing.B
//	@return - void
func BenchmarkSelectionSort(b *testing.B) {
	// Benchmark a 100
	benchmarkSelectionSort(100, b)
}

// benchmarkSelectionSort - run the SelectionSort function b.N times
//
//	@param {int} x - size of the slice
//	@param {testing.B} b - testing.B
//	@return - void
func benchmarkSelectionSort(x int, b *testing.B) {
	// run the SelectionSort function b.N times
	for i := 0; i < b.N; i++ {
		// stop timer
		b.StopTimer()
		// get a slice of random numbers
		values := perm(x)
		// start timer
		b.StartTimer()
		// run the SelectionSort function
		_ = SelectionSort(values)
	}
}
