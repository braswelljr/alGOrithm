package mergeSort

// MergeSort sorts the given slice of integers using the merge sort algorithm.
// The algorithm is stable.
// The algorithm runs in O(n log n) time.
// The algorithm does not require additional memory.
func MergeSort(slice []int) []int {
  // Base case.
  if len(slice) <= 1 {
    return slice
  }

  // Recursive case. First, divide the list into equal-sized sub lists
  // consisting of the first half and second half of the list.
  // This assumes lists start at index 0.
  var mid int
  var left []int
  var right []int

  // If the list is odd, the middle element is the median.
  if len(slice)%2 != 0 {
    mid = len(slice) / 2
    left = slice[:mid+1]
    right = slice[mid+1:]
  } else {
    mid = len(slice) / 2
    left = slice[:mid]
    right = slice[mid:]
  }

  //Recursively sort the sub lists.
  left = MergeSort(left)
  right = MergeSort(right)

  // Merge the two sorted sub lists.
  // Merge the two sorted lists into a single sorted list.
  var result []int
  for len(left) > 0 && len(right) > 0 {
    if left[0] < right[0] {
      result = append(result, left[0])
      left = left[1:]
    } else {
      result = append(result, right[0])
      right = right[1:]
    }
  }

  // Append and return the remaining elements of the list.
  return append(append(result, left...), right...)
}
