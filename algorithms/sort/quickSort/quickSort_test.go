package quickSort

import (
	"math/rand"
	"testing"
)

func TestQuickSort(t *testing.T) {
	arr := []int{2, 4, 3, 5, 1}
	arr = QuickSort(arr)
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

// perm - return a slice of random numbers with range n
//
//	@param {int} n - range of the slice
//	@return {[]int} - slice of random numbers
func perm(n int) (out []int) {
	// rand.Perm returns a slice of n random numbers
	return append(out, rand.Perm(n)...)
}

// BenchmarkQuickSort - run the QuickSort function b.N times
//
//	@param {testing.B} b - testing.B
//	@return - void
func BenchmarkQuickSort(b *testing.B) {
	// Benchmark a 100
	benchmarkQuickSort(100, b)
}

// benchmarkQuickSort - run the QuickSort function b.N times
//
//	@param {int} x - size of the slice
//	@param {testing.B} b - testing.B
//	@return - void
func benchmarkQuickSort(x int, b *testing.B) {
	// run the QuickSort function b.N times
	for i := 0; i < b.N; i++ {
		// stop timer
		b.StopTimer()
		// get a slice of random numbers with range x
		values := perm(x)
		// start timer
		b.StartTimer()
		// run the QuickSort function
		_ = QuickSort(values)
	}
}
