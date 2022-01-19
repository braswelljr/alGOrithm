package jumpSearch

import "math"

// JumpSearch is a search algorithm that works by comparing the value of the element to be searched with the value of the next element in the array.
func JumpSearch(list []int, x int) int {
  n := len(list)
  // step size
  step := int(math.Sqrt(float64(n)))
  prev := 0
  // find the block where the element is present (if it is present)
  for list[min(step, n)-1] < x {
    // if the element is not present in the list
    prev = step  // update previous block
    step += step // update step size
    // if the element is not present in the list
    if prev >= n {
      return -1
    }
  }
  // find the element in the block
  for i := prev; i < min(step, n); i++ {
    // if the element is found
    if list[i] == x {
      return i
    }
  }

  // if the element is not found
  return -1
}

// min returns the minimum of two integers
func min(a, b int) int {
  if a < b {
    return a
  }
  return b
}
