package quickSort

func QuickSort(array []int) []int {
	if len(array) < 2 {
		return array
	}

	// Pick the pivot
	pivot := array[0]

	// Partition the array
	// into values less than the pivot,
	// and values greater than the pivot
	less, greater := make([]int, 0), make([]int, 0)
	for _, value := range array[1:] {
		if value <= pivot {
			less = append(less, value)
		} else {
			greater = append(greater, value)
		}
	}

	// recursively sort the array
	less = QuickSort(less)
	greater = QuickSort(greater)

	// append and return results
	return append(append(less, pivot), greater...)
}
