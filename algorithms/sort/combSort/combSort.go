package combSort

// CombSort is a sorting algorithm that uses a gap sequence to sort the array.
// The gap sequence is a sequence of integers where the next number is the previous number
// multiplied by 1.3.
// The algorithm is a variation of the bubble sort algorithm.

func CombSort(arr []int) []int {
	if len(arr) < 2 {
		return arr
	}

	gap := len(arr)
	shrink := 1.3
	sorted := false

	for !sorted {
		// Shrink the gap
		gap = int(float64(gap) / shrink)
		// If the gap is 1, the array is sorted
		if gap <= 1 {
			gap = 1
			sorted = true
		}

		// Loop through the array and swap the elements if they are in the wrong order
		for i := 0; i < len(arr)-gap; i++ {
			if arr[i] > arr[i+gap] {
				arr[i], arr[i+gap] = arr[i+gap], arr[i]
				sorted = false
			}
		}
	}

	return arr
}
