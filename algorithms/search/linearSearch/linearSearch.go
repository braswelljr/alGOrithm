package linearSearch

// LinearSearch algorithm
// Time complexity: O(n)
// Space complexity: O(1)
// Returns the index of the element if found, -1 otherwise
func LinearSearch(list []int, element int) int {
	// Loop list
	for i, value := range list {
		// If the element is found, return its index
		if value == element {
			return i
		}
	}
	// Element not found
	return -1
}
