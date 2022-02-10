package insertionSort

// InsertionSort : Insertion sorts an array by taking each element in turn and inserting it into the correct position in the array.
func InsertionSort(array []int) []int {
	// check for length of array and empty array
	if len(array) < 2 {
		return array
	}

	// iterate over array
	for i := 1; i < len(array); i++ {
		// iterate over array from the beginning to the current index
		for j := 0; j < i; j++ {
			// check if the current element is smaller than the element at the current index
			if array[i] < array[j] {
				// swap the elements
				array[i], array[j] = array[j], array[i]
			}
		}
	}

	// return the sorted array
	return array
}
