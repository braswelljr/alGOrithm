package bubbleSort

// BubbleSort (SORT ALGORITHM) is a simple sorting algorithm that repeatedly steps through the list.
// Time: O(n^2)
// Space: O(1)
func BubbleSort(slice []int) []int {
	// check if slice is empty
	if len(slice) <= 1 {
		return slice
	}
	// Loop through the slice
	for i := 0; i < len(slice); i++ {
		// Loop with the length - i - 1
		for j := 0; j < len(slice)-i-1; j++ {
			// If the current element is greater than the next element
			if slice[j] > slice[j+1] {
				// swap the elements
				slice[j], slice[j+1] = slice[j+1], slice[j]
			}
		}
	}
	// return the sorted slice
	return slice
}
