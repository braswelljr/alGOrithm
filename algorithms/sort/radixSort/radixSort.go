package radixSort

import (
  "strconv"
)

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

  // add the zero padding
  x := "1"
  // loop through the max and add the zero padding
  for i := 0; i < len(strconv.Itoa(max)); i++ {
    x += "0"
  }
  // convert the max to a string
  padding, _ := strconv.Atoi(x)

  // create a list of buckets
  buckets := make([][]int, padding)

  // loop through the list and sort it
  for _, element := range list {
    // get the bucket index
    bucketIndex := element % padding

    // append the element to the bucket
    buckets[bucketIndex] = append(buckets[bucketIndex], element)
  }
  // create a new list
  list = []int{}

  // loop through the buckets and sort them
  for _, bucket := range buckets {
    // sort the bucket
    sortedBucket := RadixSort(bucket)

    // append the sorted bucket to the list
    list = append(list, sortedBucket...)
  }

  // return the sorted list
  return list
}
