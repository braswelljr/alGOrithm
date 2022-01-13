package selectionSort

// SelectionSort : Selection Sort is a sorting algorithm that works by repeatedly finding the smallest element and placing it at the beginning of the array.
func SelectionSort(array []int) []int {
  // check if array is empty
  if len(array) < 2 {
    return array
  }

  // loop through array
  for i := 0; i < len(array); i++ {
    // find smallest element in array
    min := i
    for j := i + 1; j < len(array); j++ {
      if array[j] < array[min] {
        min = j
      }
    }
    // swap the smallest element with first element
    array[i], array[min] = array[min], array[i]
  }

  return array
}
