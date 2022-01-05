package quickSort

func QuickSort(array []int) []int {
  // array is empty or less than 1 element
  if len(array) <= 1 {
    return array
  }

  // choose pivot
  pivot := array[0]
  // create left and right arrays
  left := make([]int, 0)
  right := make([]int, 0)

  // partition array
  for _, value := range array {
    if value < pivot {
      left = append(left, value)
    } else {
      right = append(right, value)
    }
  }

  // recursively sort left and right arrays
  left = QuickSort(left)
  right = QuickSort(right)

  // combine and return left and right arrays
  return append(append(left, pivot), right...)
}
