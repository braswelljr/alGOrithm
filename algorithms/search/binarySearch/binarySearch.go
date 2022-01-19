package binarySearch

// BinarySearch finds the position of the target in the array or list.
func BinarySearch(list []int, target int) int {
  low := 0
  high := len(list) - 1

  // loop until low is greater than high
  for low <= high {
    // calculate the midpoint
    mid := low + (high-low)/2

    // if the midpoint is the target, return it
    if list[mid] == target {
      return mid
      // if the midpoint is greater than the target, set high to mid - 1
    } else if list[mid] < target {
      low = mid + 1
    } else {
      high = mid - 1
    }
  }

  // if target is not found, return -1
  return -1
}
