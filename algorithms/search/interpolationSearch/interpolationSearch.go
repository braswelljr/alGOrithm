package interpolationSearch

// InterpolationSearch is an algorithm for searching for a key in an array that has been ordered by numerical values assigned to the keys (key values).
func InterpolationSearch(list []int, target int) int {
	low, high := 0, len(list)-1
	var mid int

	// loop while low is greater than high
	for low <= high {
		mid = low + (high-low)*(target-list[low])/(list[high]-list[low])
		if list[mid] == target {
			return mid
		} else if list[mid] < target {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}

	return -1
}
