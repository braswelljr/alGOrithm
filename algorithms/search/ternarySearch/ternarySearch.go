package ternarySearch

// TernarySearch is a search algorithm that works by dividing the search space into three parts:
// the left, middle, and right parts. The middle part is searched first, and if the value is not
// found, the left and right parts are searched.
// The idea is to divide the search space into three parts, and search the middle part first.
// If the value is not found, the left and right parts are searched.
// The time complexity of ternary search is O(log3(n)).
// The space complexity of ternary search is O(1).
func TernarySearch(list []int, target int) int {
	left, right := 0, len(list)-1
	for left <= right {
		mid := left + (right-left)/3
		if list[mid] == target {
			return mid
		} else if list[mid] < target {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}
	return -1
}
