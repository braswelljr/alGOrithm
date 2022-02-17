package radixSort

// RadixSort is a non-comparative integer sorting algorithm.
func RadixSort(list []int) []int {
  if len(list) < 2 {
    return list
  }

  // Find the maximum number in the list
  max := list[0]

  // loop through the list and find the max
  for _, element := range list {
    // if the element is greater than the max, set the max to the element
    if element > max {
      // update the max
      max = element
    }
  }

  // create a bucket for each digit
  buckets := make([][]int, 10)
  // loop through the list and sort it
  for _, element := range list {
    // get the digit of the element
    digit := element / max
    // append the element to the bucket
    buckets[digit] = append(buckets[digit], element)
  }

  //  loop through the buckets and sort them
  for _, bucket := range buckets {
    // sort the bucket
    bucket = RadixSort(bucket)
  }

  // create a new list to hold the sorted list
  sortedList := make([]int, 0)
  // loop through the buckets and append the elements to the sorted list
  for _, bucket := range buckets {
    sortedList = append(sortedList, bucket...)
  }

  // return the sorted list
  return sortedList
}
